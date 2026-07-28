// Package service implements the platform-neutral clipboard-to-Markdown
// pipeline shared by the CLI and the menu bar presentation layers.
package service

import (
	"fmt"

	"github.com/stn1slv/md-paste/internal/converter"
	"github.com/stn1slv/md-paste/internal/models"
)

// ReadFunc reads the current clipboard content.
type ReadFunc func() (models.ClipboardContent, error)

// WriteFunc writes Markdown text to a destination (clipboard, stdout, ...).
type WriteFunc func(string) error

// Outcome describes the result of a Convert call.
type Outcome int

const (
	// OutcomeEmpty means the clipboard held nothing convertible; no write happened.
	OutcomeEmpty Outcome = iota
	// OutcomeConverted means the content was converted and written.
	OutcomeConverted
)

// Hooks carries optional behavior injected by a caller.
type Hooks struct {
	// OnRawContent, when set, is called with the raw clipboard content before
	// conversion. Used by the CLI --save-raw flag.
	OnRawContent func(models.ClipboardContent) error
	// Sink, when set, overrides write as the destination for the Markdown.
	// Used by the CLI --stdout flag.
	Sink WriteFunc
}

// Convert runs the shared pipeline: read -> (optional raw hook) -> convert -> write.
// When the clipboard has no usable content it returns OutcomeEmpty and performs no
// write, preserving the CLI's silence-on-empty behavior.
func Convert(read ReadFunc, write WriteFunc, h Hooks) (Outcome, models.MarkdownDocument, error) {
	content, err := read()
	if err != nil {
		return OutcomeEmpty, models.MarkdownDocument{}, fmt.Errorf("failed to read clipboard: %w", err)
	}

	if content.ContentType == models.ContentTypeNone {
		return OutcomeEmpty, models.MarkdownDocument{}, nil
	}

	if h.OnRawContent != nil {
		if err := h.OnRawContent(content); err != nil {
			return OutcomeEmpty, models.MarkdownDocument{}, fmt.Errorf("failed to save raw content: %w", err)
		}
	}

	doc, err := converter.Convert(content)
	if err != nil {
		return OutcomeEmpty, models.MarkdownDocument{}, fmt.Errorf("failed to convert content: %w", err)
	}

	sink := h.Sink
	if sink == nil {
		sink = write
	}
	if err := sink(doc.Content); err != nil {
		return OutcomeEmpty, models.MarkdownDocument{}, fmt.Errorf("failed to write content: %w", err)
	}

	return OutcomeConverted, doc, nil
}
