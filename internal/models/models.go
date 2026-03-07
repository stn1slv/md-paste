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

// Alignment represents the horizontal alignment of a table cell.
type Alignment int

const (
	// AlignNone indicates no specific alignment (default).
	AlignNone Alignment = iota
	// AlignLeft indicates left alignment.
	AlignLeft
	// AlignCenter indicates center alignment.
	AlignCenter
	// AlignRight indicates right alignment.
	AlignRight
)

// Table represents a structured grid of data extracted from the clipboard.
type Table struct {
	Rows      []Row
	HasHeader bool
}

// Row represents a single horizontal line of cells in a table.
type Row struct {
	Cells    []Cell
	IsHeader bool
}

// Cell represents an individual data point within a table.
type Cell struct {
	Content   string
	Alignment Alignment
	RowSpan   int
	ColSpan   int
}
