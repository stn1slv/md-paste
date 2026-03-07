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

	if content.RawHTML != "" {
		if doc, ok := tryHTMLTableConversion(content.RawHTML); ok {
			return doc, nil
		}
	}

	if content.ContentType == models.ContentTypePlainText {
		if doc, ok := tryTextTableConversion(content.PlainText); ok {
			return doc, nil
		}
		return models.MarkdownDocument{
			Content:    content.PlainText,
			SourceType: models.ContentTypePlainText,
		}, nil
	}

	return performStandardHTMLConversion(content.RawHTML)
}

func tryHTMLTableConversion(rawHTML string) (models.MarkdownDocument, bool) {
	if table, ok := ExtractTableFromHTML(rawHTML); ok {
		return models.MarkdownDocument{
			Content:    RenderTable(table),
			SourceType: models.ContentTypeHTML,
		}, true
	}
	return models.MarkdownDocument{}, false
}

func tryTextTableConversion(plainText string) (models.MarkdownDocument, bool) {
	if table, ok := ExtractTableFromText(plainText); ok {
		return models.MarkdownDocument{
			Content:    RenderTable(table),
			SourceType: models.ContentTypePlainText,
		}, true
	}
	return models.MarkdownDocument{}, false
}

func performStandardHTMLConversion(rawHTML string) (models.MarkdownDocument, error) {
	converter := htmltomarkdown.NewConverter(emptyDomain, true, nil)
	markdown, err := converter.ConvertString(rawHTML)
	if err != nil {
		return models.MarkdownDocument{}, errors.Wrap(err, "failed to convert HTML to Markdown")
	}

	return models.MarkdownDocument{
		Content:    strings.TrimSpace(markdown),
		SourceType: models.ContentTypeHTML,
	}, nil
}
