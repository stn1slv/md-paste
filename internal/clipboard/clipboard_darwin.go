//go:build darwin

// Package clipboard provides native macOS clipboard integration.
package clipboard

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit

#include <string.h>
#include <stdlib.h>
#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>

// Returns 1 if successful, 0 if no string found.
int read_clipboard(char **html, char **plain) {
	@autoreleasepool {
		NSPasteboard *pb = [NSPasteboard generalPasteboard];

		// Try HTML first
		NSString *htmlStr = [pb stringForType:NSPasteboardTypeHTML];
		if (htmlStr) {
			*html = strdup([htmlStr UTF8String]);
			*plain = NULL;

			// Also grab plain text just in case
			NSString *plainStr = [pb stringForType:NSPasteboardTypeString];
			if (plainStr) {
				*plain = strdup([plainStr UTF8String]);
			}
			return 1;
		}

		// Fallback to plain text
		NSString *plainStr = [pb stringForType:NSPasteboardTypeString];
		if (plainStr) {
			*plain = strdup([plainStr UTF8String]);
			*html = NULL;
			return 1;
		}

		return 0; // Empty or non-text
	}
}

// Returns 1 on success, 0 if the bytes could not be decoded as UTF-8.
int write_clipboard(const char *text) {
	@autoreleasepool {
		NSPasteboard *pb = [NSPasteboard generalPasteboard];

		NSString *str = [NSString stringWithUTF8String:text];
		if (str == nil) {
			// Defense in depth: Go-side already replaces invalid UTF-8,
			// but never pass nil to setString: which would raise.
			return 0;
		}

		[pb clearContents];
		[pb setString:str forType:NSPasteboardTypeString];
		return 1;
	}
}

void clear_clipboard() {
	@autoreleasepool {
		NSPasteboard *pb = [NSPasteboard generalPasteboard];
		[pb clearContents];
	}
}
*/
import "C"

import (
	"errors"
	"strings"
	"unicode/utf8"
	"unsafe"

	"github.com/stn1slv/md-paste/internal/models"
)

const (
	readFailure  = 0
	writeFailure = 0
)

// utf8Replacement substitutes for any invalid UTF-8 byte sequence before
// the string crosses the CGO boundary.
const utf8Replacement = "�"

var errWriteFailed = errors.New("clipboard write failed")

// Read retrieves content from the macOS system clipboard.
func Read() (models.ClipboardContent, error) {
	var cHTML, cPlain *C.char

	//nolint:gocritic // CGO generated code might trigger dupSubExpr
	success := C.read_clipboard(&cHTML, &cPlain)
	if success == readFailure {
		return models.ClipboardContent{
			ContentType: models.ContentTypeNone,
		}, nil
	}

	content := models.ClipboardContent{
		ContentType: models.ContentTypePlainText,
	}

	if cPlain != nil {
		content.PlainText = C.GoString(cPlain)
		C.free(unsafe.Pointer(cPlain))
	}

	if cHTML != nil {
		content.RawHTML = C.GoString(cHTML)
		if content.RawHTML != "" {
			content.ContentType = models.ContentTypeHTML
		}
		C.free(unsafe.Pointer(cHTML))
	}

	// Double check empty case even if success was true
	if content.RawHTML == "" && content.PlainText == "" {
		content.ContentType = models.ContentTypeNone
	}

	return content, nil
}

// WriteMarkdown writes the converted Markdown string back to the clipboard.
// Invalid UTF-8 byte sequences are replaced with U+FFFD before crossing the
// CGO boundary; otherwise NSString stringWithUTF8String: would return nil
// and setString: would raise an Objective-C exception.
func WriteMarkdown(text string) error {
	if !utf8.ValidString(text) {
		text = strings.ToValidUTF8(text, utf8Replacement)
	}

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	if C.write_clipboard(cText) == writeFailure {
		return errWriteFailed
	}
	return nil
}

// Clear empties the system clipboard.
func Clear() error {
	C.clear_clipboard()
	return nil
}
