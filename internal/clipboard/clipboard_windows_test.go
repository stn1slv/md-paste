//go:build windows

package clipboard

import (
	"fmt"
	"testing"
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
