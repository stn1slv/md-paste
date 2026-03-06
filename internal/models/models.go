// Package models defines the core data structures for md-paste.
package models

// ContentType represents the type of content on the clipboard.
type ContentType string

const (
	// ContentTypeHTML indicates the clipboard content is HTML.
	ContentTypeHTML ContentType = "HTML"
	// ContentTypePlainText indicates the clipboard content is plain text.
	ContentTypePlainText ContentType = "PlainText"
	// ContentTypeNone indicates the clipboard has no usable text.
	ContentTypeNone ContentType = "None"
)

// ClipboardContent represents the data retrieved from the macOS pasteboard.
type ClipboardContent struct {
	RawHTML     string
	PlainText   string
	ContentType ContentType
}

// MarkdownDocument represents the structured text result of the conversion process.
type MarkdownDocument struct {
	Content    string
	SourceType ContentType
}
