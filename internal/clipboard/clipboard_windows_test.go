//go:build windows

package clipboard

import (
	"fmt"
	"testing"
	"unicode/utf16"
	"unsafe"
)

func TestParseWindowsHTMLFormat(t *testing.T) {
	const fragment = "<p>hello <b>world</b></p>"
	const wrapperPrefix = "<html><body><!--StartFragment-->"
	const wrapperSuffix = "<!--EndFragment--></body></html>"

	body := wrapperPrefix + fragment + wrapperSuffix
	headerTmpl := "Version:0.9\nStartHTML:%010d\nEndHTML:%010d\nStartFragment:%010d\nEndFragment:%010d\n"

	// Compute offsets against a stub of the same length as the final header.
	stub := fmt.Sprintf(headerTmpl, 0, 0, 0, 0)
	startHTML := len(stub)
	endHTML := startHTML + len(body)
	startFragment := startHTML + len(wrapperPrefix)
	endFragment := startFragment + len(fragment)

	header := fmt.Sprintf(headerTmpl, startHTML, endHTML, startFragment, endFragment)
	payload := []byte(header + body)

	t.Run("prefers StartFragment/EndFragment", func(t *testing.T) {
		got := parseWindowsHTMLFormat(payload)
		if got != fragment {
			t.Fatalf("fragment mismatch:\n got: %q\nwant: %q", got, fragment)
		}
	})

	t.Run("falls back to StartHTML/EndHTML when fragment markers are missing", func(t *testing.T) {
		htmlOnlyTmpl := "Version:0.9\nStartHTML:%010d\nEndHTML:%010d\n"
		stub := fmt.Sprintf(htmlOnlyTmpl, 0, 0)
		htmlBody := "<html><body><p>hi</p></body></html>"
		startHTML := len(stub)
		endHTML := startHTML + len(htmlBody)
		header := fmt.Sprintf(htmlOnlyTmpl, startHTML, endHTML)
		got := parseWindowsHTMLFormat([]byte(header + htmlBody))
		if got != htmlBody {
			t.Fatalf("html-only mismatch:\n got: %q\nwant: %q", got, htmlBody)
		}
	})

	t.Run("falls back to first '<' when offsets are unparseable", func(t *testing.T) {
		raw := []byte("Version:0.9\nStartHTML:bogus\nEndHTML:bogus\n<p>direct</p>")
		if got := parseWindowsHTMLFormat(raw); got != "<p>direct</p>" {
			t.Fatalf("fallback mismatch:\n got: %q", got)
		}
	})

	t.Run("ignores out-of-range offsets", func(t *testing.T) {
		raw := []byte("Version:0.9\nStartHTML:9999\nEndHTML:99999\n<p>ok</p>")
		if got := parseWindowsHTMLFormat(raw); got != "<p>ok</p>" {
			t.Fatalf("range fallback mismatch:\n got: %q", got)
		}
	})
}

const gmemZeroInit = 0x0040

// allocGlobalBytes allocates zero-initialized global memory holding data.
// GMEM_ZEROINIT keeps any size rounding deterministic: slack bytes past the
// payload are guaranteed zero rather than garbage.
func allocGlobalBytes(t *testing.T, data []byte) uintptr {
	t.Helper()
	h, _, err := procGlobalAlloc.Call(gmemMoveable|gmemZeroInit, uintptr(len(data)))
	if h == 0 {
		t.Fatalf("GlobalAlloc: %v", err)
	}
	t.Cleanup(func() {
		procGlobalFree.Call(h) //nolint:errcheck
	})

	p, lockErr := lockGlobal(h)
	if lockErr != nil {
		t.Fatalf("GlobalLock: %v", lockErr)
	}
	copy(unsafe.Slice((*byte)(p), len(data)), data)
	unlockGlobal(h)
	return h
}

func TestReadGlobalMemBytes(t *testing.T) {
	t.Run("stops at null terminator", func(t *testing.T) {
		h := allocGlobalBytes(t, []byte("hello\x00garbage"))
		if got := string(readGlobalMemBytes(h)); got != "hello" {
			t.Fatalf("got %q, want %q", got, "hello")
		}
	})

	t.Run("bounded by allocation size when no terminator exists", func(t *testing.T) {
		// No null anywhere in the payload: the scan must stop at the
		// allocation boundary instead of running past it.
		data := []byte("abcdefgh")
		h := allocGlobalBytes(t, data)
		if got := string(readGlobalMemBytes(h)); got != string(data) {
			t.Fatalf("got %q, want %q", got, data)
		}
	})
}

func TestReadGlobalMemUTF16(t *testing.T) {
	encode := func(s string, terminated bool) []byte {
		u := utf16.Encode([]rune(s))
		if terminated {
			u = append(u, 0)
		}
		b := make([]byte, len(u)*2)
		for i, v := range u {
			b[i*2] = byte(v)
			b[i*2+1] = byte(v >> 8)
		}
		return b
	}

	t.Run("stops at null terminator", func(t *testing.T) {
		h := allocGlobalBytes(t, encode("héllo", true))
		if got := readGlobalMemUTF16(h); got != "héllo" {
			t.Fatalf("got %q, want %q", got, "héllo")
		}
	})

	t.Run("bounded by allocation size when no terminator exists", func(t *testing.T) {
		h := allocGlobalBytes(t, encode("wörld", false))
		if got := readGlobalMemUTF16(h); got != "wörld" {
			t.Fatalf("got %q, want %q", got, "wörld")
		}
	})
}
