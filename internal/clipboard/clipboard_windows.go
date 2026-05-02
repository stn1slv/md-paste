//go:build windows

// Package clipboard provides native Windows clipboard integration.
package clipboard

import (
	"errors"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/stn1slv/md-paste/internal/models"
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
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

	cfHTMLFormat   uint32
	errWriteFailed = errors.New("clipboard write failed")
)

func init() {
	name, err := syscall.UTF16PtrFromString("HTML Format")
	if err == nil {
		r, _, _ := procRegisterClipboardFormatW.Call(uintptr(unsafe.Pointer(name)))
		cfHTMLFormat = uint32(r)
	}
}

func openClipboard() error {
	r, _, err := procOpenClipboard.Call(0)
	if r == 0 {
		return err
	}
	return nil
}

func closeClipboard() {
	procCloseClipboard.Call() //nolint:errcheck
}

// readGlobalMemBytes reads null-terminated bytes from a global memory handle
// returned by GetClipboardData. Used for CF_HTML (UTF-8 encoded).
func readGlobalMemBytes(h uintptr) []byte {
	ptr, _, _ := procGlobalLock.Call(h)
	if ptr == 0 {
		return nil
	}
	defer procGlobalUnlock.Call(h) //nolint:errcheck

	p := unsafe.Pointer(ptr)
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
	ptr, _, _ := procGlobalLock.Call(h)
	if ptr == 0 {
		return ""
	}
	defer procGlobalUnlock.Call(h) //nolint:errcheck

	p := unsafe.Pointer(ptr)
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

// parseWindowsHTMLFormat extracts the HTML document from Windows CF_HTML data.
// The format prepends a text header with byte-offset markers before the HTML.
func parseWindowsHTMLFormat(data []byte) string {
	s := string(data)
	var startHTML, endHTML int

	for _, line := range strings.SplitN(s, "\n", 20) {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "StartHTML:"); ok {
			if n, err := strconv.Atoi(strings.TrimSpace(after)); err == nil {
				startHTML = n
			}
		} else if after, ok := strings.CutPrefix(line, "EndHTML:"); ok {
			if n, err := strconv.Atoi(strings.TrimSpace(after)); err == nil {
				endHTML = n
			}
		}
	}

	if startHTML > 0 && endHTML > startHTML && endHTML <= len(data) {
		return string(data[startHTML:endHTML])
	}

	// Fallback: strip the text header by finding the first '<'
	if idx := strings.IndexByte(s, '<'); idx >= 0 {
		return s[idx:]
	}
	return s
}

// Read retrieves content from the Windows system clipboard.
func Read() (models.ClipboardContent, error) {
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
	encoded := utf16.Encode([]rune(text))
	byteSize := (len(encoded) + 1) * 2 // +1 for UTF-16 null terminator

	h, _, _ := procGlobalAlloc.Call(gmemMoveable, uintptr(byteSize))
	if h == 0 {
		return errWriteFailed
	}

	ptr, _, _ := procGlobalLock.Call(h)
	if ptr == 0 {
		procGlobalFree.Call(h) //nolint:errcheck
		return errWriteFailed
	}

	p := unsafe.Pointer(ptr)
	for i, v := range encoded {
		*(*uint16)(unsafe.Add(p, i*2)) = v
	}
	*(*uint16)(unsafe.Add(p, len(encoded)*2)) = 0
	procGlobalUnlock.Call(h) //nolint:errcheck

	if err := openClipboard(); err != nil {
		procGlobalFree.Call(h) //nolint:errcheck
		return err
	}
	defer closeClipboard()

	if r, _, _ := procEmptyClipboard.Call(); r == 0 {
		procGlobalFree.Call(h) //nolint:errcheck
		return errWriteFailed
	}

	// After a successful SetClipboardData the OS owns the handle; do not free it.
	if r, _, _ := procSetClipboardData.Call(cfUnicodeText, h); r == 0 {
		procGlobalFree.Call(h) //nolint:errcheck
		return errWriteFailed
	}

	return nil
}

// Clear empties the system clipboard.
func Clear() error {
	if err := openClipboard(); err != nil {
		return err
	}
	defer closeClipboard()

	if r, _, err := procEmptyClipboard.Call(); r == 0 {
		return err
	}
	return nil
}
