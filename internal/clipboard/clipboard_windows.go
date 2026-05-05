//go:build windows

// Package clipboard provides native Windows clipboard integration.
package clipboard

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/stn1slv/md-paste/internal/models"
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002

	// openClipboard retries for up to ~500ms in 10ms increments. This covers
	// realistic hold times from clipboard managers, RDP redirection, and AV
	// hooks without making CLI failures feel sluggish.
	openClipboardAttempts = 50
	openClipboardBackoff  = 10 * time.Millisecond
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procOpenClipboard            = user32.NewProc("OpenClipboard")
	procCloseClipboard           = user32.NewProc("CloseClipboard")
	procEmptyClipboard           = user32.NewProc("EmptyClipboard")
	procGetClipboardData         = user32.NewProc("GetClipboardData")
	procSetClipboardData         = user32.NewProc("SetClipboardData")
	procRegisterClipboardFormatW = user32.NewProc("RegisterClipboardFormatW")
	procGlobalAlloc              = kernel32.NewProc("GlobalAlloc")
	procGlobalFree               = kernel32.NewProc("GlobalFree")
	procGlobalLock               = kernel32.NewProc("GlobalLock")
	procGlobalUnlock             = kernel32.NewProc("GlobalUnlock")

	cfHTMLFormat uint32
)

func init() {
	name, err := syscall.UTF16PtrFromString("HTML Format")
	if err == nil {
		r, _, _ := procRegisterClipboardFormatW.Call(uintptr(unsafe.Pointer(name)))
		cfHTMLFormat = uint32(r)
	}
}

// openClipboard opens the system clipboard, retrying briefly while another
// process owns it. Clipboard managers, browsers, and antivirus tools regularly
// hold the clipboard for short windows, so a single attempt is unreliable.
func openClipboard() error {
	var lastErr error
	for i := 0; i < openClipboardAttempts; i++ {
		r, _, err := procOpenClipboard.Call(0)
		if r != 0 {
			return nil
		}
		lastErr = err
		time.Sleep(openClipboardBackoff)
	}
	return fmt.Errorf("OpenClipboard: %w", lastErr)
}

func closeClipboard() {
	procCloseClipboard.Call() //nolint:errcheck
}

// lockGlobal calls GlobalLock and returns a pointer to the locked memory.
// On failure GlobalLock returns NULL and the syscall errno is surfaced via
// the second return so callers can wrap it for diagnostics.
//
// We reinterpret the uintptr result via a *unsafe.Pointer cast rather than
// the direct unsafe.Pointer(uintptr) form, which sidesteps go vet's
// unsafeptr false-positive for OS-level (non-GC) memory addresses.
func lockGlobal(h uintptr) (unsafe.Pointer, error) {
	r, _, err := syscall.Syscall(procGlobalLock.Addr(), 1, h, 0, 0)
	if r == 0 {
		return nil, err
	}
	return *(*unsafe.Pointer)(unsafe.Pointer(&r)), nil
}

func unlockGlobal(h uintptr) {
	syscall.Syscall(procGlobalUnlock.Addr(), 1, h, 0, 0) //nolint:errcheck
}

// readGlobalMemBytes reads null-terminated bytes from a global memory handle
// returned by GetClipboardData. Used for CF_HTML (UTF-8 encoded).
func readGlobalMemBytes(h uintptr) []byte {
	p, err := lockGlobal(h)
	if err != nil {
		return nil
	}
	defer unlockGlobal(h)

	var n int
	for *(*byte)(unsafe.Add(p, n)) != 0 {
		n++
	}

	buf := make([]byte, n)
	for i := range buf {
		buf[i] = *(*byte)(unsafe.Add(p, i))
	}
	return buf
}

// readGlobalMemUTF16 reads null-terminated UTF-16 LE from a global memory
// handle returned by GetClipboardData. Used for CF_UNICODETEXT.
func readGlobalMemUTF16(h uintptr) string {
	p, err := lockGlobal(h)
	if err != nil {
		return ""
	}
	defer unlockGlobal(h)

	var n int
	for *(*uint16)(unsafe.Add(p, n*2)) != 0 {
		n++
	}

	u16s := make([]uint16, n)
	for i := range u16s {
		u16s[i] = *(*uint16)(unsafe.Add(p, i*2))
	}
	return string(utf16.Decode(u16s))
}

// parseWindowsHTMLFormat extracts the HTML payload from Windows CF_HTML data.
// The format prepends a text header with byte-offset markers before the HTML.
// We prefer StartFragment/EndFragment (the user's actual selection) and fall
// back to StartHTML/EndHTML (the surrounding synthetic document) when the
// fragment markers are absent.
func parseWindowsHTMLFormat(data []byte) string {
	s := string(data)
	offsets := map[string]int{}

	for _, line := range strings.SplitN(s, "\n", 20) {
		line = strings.TrimSpace(line)
		for _, key := range []string{"StartFragment:", "EndFragment:", "StartHTML:", "EndHTML:"} {
			if after, ok := strings.CutPrefix(line, key); ok {
				if n, err := strconv.Atoi(strings.TrimSpace(after)); err == nil {
					offsets[strings.TrimSuffix(key, ":")] = n
				}
				break
			}
		}
	}

	slice := func(start, end int) (string, bool) {
		if start > 0 && end > start && end <= len(data) {
			return string(data[start:end]), true
		}
		return "", false
	}

	if out, ok := slice(offsets["StartFragment"], offsets["EndFragment"]); ok {
		return out
	}
	if out, ok := slice(offsets["StartHTML"], offsets["EndHTML"]); ok {
		return out
	}

	// Fallback: strip the text header by finding the first '<'
	if idx := strings.IndexByte(s, '<'); idx >= 0 {
		return s[idx:]
	}
	return s
}

// Read retrieves content from the Windows system clipboard.
func Read() (models.ClipboardContent, error) {
	// OpenClipboard associates the clipboard with the calling OS thread; every
	// subsequent clipboard syscall must run on that same thread or it fails
	// with ERROR_CLIPBOARD_NOT_OPEN. Pin the goroutine to prevent migration.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := openClipboard(); err != nil {
		return models.ClipboardContent{ContentType: models.ContentTypeNone}, err
	}
	defer closeClipboard()

	content := models.ClipboardContent{ContentType: models.ContentTypeNone}

	// Try the registered "HTML Format" first (CF_HTML, UTF-8 encoded).
	if cfHTMLFormat != 0 {
		if h, _, _ := procGetClipboardData.Call(uintptr(cfHTMLFormat)); h != 0 {
			if raw := readGlobalMemBytes(h); len(raw) > 0 {
				content.RawHTML = parseWindowsHTMLFormat(raw)
				content.ContentType = models.ContentTypeHTML
			}
		}
	}

	// Always attempt to read plain text (CF_UNICODETEXT).
	if h, _, _ := procGetClipboardData.Call(cfUnicodeText); h != 0 {
		if plain := readGlobalMemUTF16(h); plain != "" {
			content.PlainText = plain
			if content.ContentType != models.ContentTypeHTML {
				content.ContentType = models.ContentTypePlainText
			}
		}
	}

	return content, nil
}

// WriteMarkdown writes the converted Markdown string to the clipboard as
// CF_UNICODETEXT (UTF-16 LE with null terminator).
func WriteMarkdown(text string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	encoded := utf16.Encode([]rune(text))
	byteSize := (len(encoded) + 1) * 2 // +1 for UTF-16 null terminator

	h, _, allocErr := procGlobalAlloc.Call(gmemMoveable, uintptr(byteSize))
	if h == 0 {
		return fmt.Errorf("GlobalAlloc: %w", allocErr)
	}

	p, lockErr := lockGlobal(h)
	if lockErr != nil {
		procGlobalFree.Call(h) //nolint:errcheck
		return fmt.Errorf("GlobalLock: %w", lockErr)
	}

	for i, v := range encoded {
		*(*uint16)(unsafe.Add(p, i*2)) = v
	}
	*(*uint16)(unsafe.Add(p, len(encoded)*2)) = 0
	unlockGlobal(h)

	if err := openClipboard(); err != nil {
		procGlobalFree.Call(h) //nolint:errcheck
		return err
	}
	defer closeClipboard()

	if r, _, err := procEmptyClipboard.Call(); r == 0 {
		procGlobalFree.Call(h) //nolint:errcheck
		return fmt.Errorf("EmptyClipboard: %w", err)
	}

	// After a successful SetClipboardData the OS owns the handle; do not free it.
	if r, _, err := procSetClipboardData.Call(cfUnicodeText, h); r == 0 {
		procGlobalFree.Call(h) //nolint:errcheck
		return fmt.Errorf("SetClipboardData: %w", err)
	}

	return nil
}

// Clear empties the system clipboard.
func Clear() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := openClipboard(); err != nil {
		return err
	}
	defer closeClipboard()

	if r, _, err := procEmptyClipboard.Call(); r == 0 {
		return fmt.Errorf("EmptyClipboard: %w", err)
	}
	return nil
}
