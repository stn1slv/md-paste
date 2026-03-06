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

void write_clipboard(const char *text) {
	@autoreleasepool {
		NSPasteboard *pb = [NSPasteboard generalPasteboard];
		[pb clearContents];

		NSString *str = [NSString stringWithUTF8String:text];
		[pb setString:str forType:NSPasteboardTypeString];
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
	"unsafe"

	"github.com/stn1slv/md-paste/internal/models"
)

const readFailure = 0

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
func WriteMarkdown(text string) error {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	C.write_clipboard(cText)
	return nil
}

// Clear empties the system clipboard.
func Clear() error {
	C.clear_clipboard()
	return nil
}
