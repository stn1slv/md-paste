// Package converter provides HTML to Markdown conversion logic.
package converter

import (
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/stn1slv/md-paste/internal/errors"
	"github.com/stn1slv/md-paste/internal/models"
)

const emptyDomain = ""

// Convert takes clipboard content and converts it to a Markdown document.
func Convert(content models.ClipboardContent) (models.MarkdownDocument, error) {
	if content.ContentType == models.ContentTypeNone || (content.RawHTML == "" && content.PlainText == "") {
		return models.MarkdownDocument{}, errors.New("no content to convert")
	}

	if content.ContentType == models.ContentTypePlainText {
		// Plain text is technically a subset of Markdown, but we might just return it as is.
		return models.MarkdownDocument{
			Content:    content.PlainText,
			SourceType: models.ContentTypePlainText,
		}, nil
	}

	converter := htmltomarkdown.NewConverter(emptyDomain, true, nil)

	// Some copied HTML might be wrapped heavily.
	// html-to-markdown handles standard tags well.
	markdown, err := converter.ConvertString(content.RawHTML)
	if err != nil {
		return models.MarkdownDocument{}, errors.Wrap(err, "failed to convert HTML to Markdown")
	}

	return models.MarkdownDocument{
		Content:    strings.TrimSpace(markdown),
		SourceType: models.ContentTypeHTML,
	}, nil
}
