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

// Every entry point catches NSException. An Objective-C exception is not a Go
// panic: it cannot be recovered on the Go side and would terminate the process,
// which for the resident menu bar app means the icon simply disappears.

// Returns 1 when at least one string was read, 0 when the pasteboard held
// nothing usable, and -1 when the pasteboard could not be read.
// *html and *plain are always initialized; either may stay NULL.
int read_clipboard(char **html, char **plain) {
	*html = NULL;
	*plain = NULL;

	@autoreleasepool {
		@try {
			NSPasteboard *pb = [NSPasteboard generalPasteboard];
			if (pb == nil) {
				return -1;
			}

			// UTF8String returns NULL if the string cannot be encoded; strdup(NULL)
			// is undefined behaviour, so guard every conversion.
			NSString *htmlStr = [pb stringForType:NSPasteboardTypeHTML];
			if (htmlStr != nil) {
				const char *utf8 = [htmlStr UTF8String];
				if (utf8 != NULL) {
					*html = strdup(utf8);
				}
			}

			NSString *plainStr = [pb stringForType:NSPasteboardTypeString];
			if (plainStr != nil) {
				const char *utf8 = [plainStr UTF8String];
				if (utf8 != NULL) {
					*plain = strdup(utf8);
				}
			}

			return (*html != NULL || *plain != NULL) ? 1 : 0;
		} @catch (NSException *e) {
			// Reading plain text happens after the HTML strdup, so an exception
			// can arrive with one buffer already allocated. The Go side discards
			// both pointers on a failure status, so free them here.
			if (*html != NULL) {
				free(*html);
				*html = NULL;
			}
			if (*plain != NULL) {
				free(*plain);
				*plain = NULL;
			}
			return -1;
		}
	}
}

// Returns 1 on success, 0 if the bytes could not be decoded as UTF-8, and -1 if
// the pasteboard rejected the write.
int write_clipboard(const char *text) {
	@autoreleasepool {
		@try {
			NSPasteboard *pb = [NSPasteboard generalPasteboard];
			if (pb == nil) {
				return -1;
			}

			NSString *str = [NSString stringWithUTF8String:text];
			if (str == nil) {
				// Defense in depth: the Go side already replaces invalid UTF-8,
				// but never pass nil to setString: which would raise.
				return 0;
			}

			[pb clearContents];
			return [pb setString:str forType:NSPasteboardTypeString] ? 1 : -1;
		} @catch (NSException *e) {
			return -1;
		}
	}
}

// Returns 1 on success and -1 if the pasteboard could not be cleared.
int clear_clipboard(void) {
	@autoreleasepool {
		@try {
			NSPasteboard *pb = [NSPasteboard generalPasteboard];
			if (pb == nil) {
				return -1;
			}
			[pb clearContents];
			return 1;
		} @catch (NSException *e) {
			return -1;
		}
	}
}
*/
import "C"

import (
	"errors"
	"strings"
	"sync"
	"unicode/utf8"
	"unsafe"

	"github.com/stn1slv/md-paste/internal/models"
)

// Status codes returned by the C entry points.
const (
	statusFailure = -1
	statusEmpty   = 0
)

// utf8Replacement substitutes for any invalid UTF-8 byte sequence before
// the string crosses the CGO boundary.
const utf8Replacement = "�"

// pasteboardMu serializes access to the shared NSPasteboard. AppKit's pasteboard
// is not thread-safe, and the menu bar app's global shortcut runs conversions on
// a background goroutine, so Read and WriteMarkdown can otherwise overlap. We
// serialize rather than hop to the main queue: pasteboard access is not
// documented as main-thread-only, and blocking the UI thread on clipboard I/O
// would risk stalling (or deadlocking) the event loop.
var pasteboardMu sync.Mutex

var (
	errReadFailed  = errors.New("clipboard read failed")
	errWriteFailed = errors.New("clipboard write failed")
	errClearFailed = errors.New("clipboard clear failed")
)

// Read retrieves content from the macOS system clipboard.
func Read() (models.ClipboardContent, error) {
	pasteboardMu.Lock()
	defer pasteboardMu.Unlock()

	var cHTML, cPlain *C.char

	//nolint:gocritic // CGO generated code might trigger dupSubExpr
	status := C.read_clipboard(&cHTML, &cPlain)
	if status == statusFailure {
		// Distinguish a broken pasteboard from an empty one: reporting a failure
		// as "nothing to convert" hides real errors from the user.
		return models.ClipboardContent{ContentType: models.ContentTypeNone}, errReadFailed
	}
	if status == statusEmpty {
		return models.ClipboardContent{ContentType: models.ContentTypeNone}, nil
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

	pasteboardMu.Lock()
	defer pasteboardMu.Unlock()

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	if C.write_clipboard(cText) != 1 {
		return errWriteFailed
	}
	return nil
}

// Clear empties the system clipboard.
func Clear() error {
	pasteboardMu.Lock()
	defer pasteboardMu.Unlock()

	if C.clear_clipboard() != 1 {
		return errClearFailed
	}
	return nil
}
