// Package converter provides HTML to Markdown conversion logic.
package converter

import (
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/stn1slv/md-paste/internal/errors"
	"github.com/stn1slv/md-paste/internal/models"
)

const emptyDomain = ""

// Convert takes clipboard content and converts it to a Markdown document.
func Convert(content models.ClipboardContent) (models.MarkdownDocument, error) {
	if content.ContentType == models.ContentTypeNone || (content.RawHTML == "" && content.PlainText == "") {
		return models.MarkdownDocument{}, errors.New("no content to convert")
	}

	// 1. If HTML is available, use the standard conversion with our custom table rule
	if content.RawHTML != "" {
		hasTable := strings.Contains(strings.ToLower(content.RawHTML), "<table")
		if !hasTable && content.PlainText != "" {
			if doc, ok := tryTextTableConversion(content.PlainText, content.ContentType); ok {
				return doc, nil
			}
		}
		return performStandardHTMLConversion(content.RawHTML)
	}

	// 2. Try layout-aware text table extraction as a secondary fallback
	if content.PlainText != "" {
		if doc, ok := tryTextTableConversion(content.PlainText, content.ContentType); ok {
			return doc, nil
		}
	}

	// 3. Perform standard plain text fallback
	return models.MarkdownDocument{
		Content:    content.PlainText,
		SourceType: models.ContentTypePlainText,
	}, nil
}

func tryTextTableConversion(plainText string, originalType models.ContentType) (models.MarkdownDocument, bool) {
	if table, ok := ExtractTableFromText(plainText); ok {
		return models.MarkdownDocument{
			Content:    RenderTable(table),
			SourceType: originalType,
		}, true
	}
	return models.MarkdownDocument{}, false
}

func performStandardHTMLConversion(rawHTML string) (models.MarkdownDocument, error) {
	converter := htmltomarkdown.NewConverter(emptyDomain, true, nil)

	// Add custom rule for tables to use our high-fidelity extraction.
	// This ensures that documents with text AND multiple tables are handled perfectly,
	// rather than truncating the whole document to just the first table.
	converter.AddRules(htmltomarkdown.Rule{
		Filter: []string{"table"},
		Replacement: func(_ string, selec *goquery.Selection, _ *htmltomarkdown.Options) *string {
			if len(selec.Nodes) == 0 {
				return nil
			}
			tableNode := selec.Nodes[0]
			tableModel := ParseTable(tableNode)
			if len(tableModel.Rows) == 0 {
				return nil // Fallback to default conversion if our parser finds nothing
			}
			md := RenderTable(tableModel)
			// Add blank lines around the table to ensure it renders correctly in Markdown
			res := "\n\n" + md + "\n\n"
			return &res
		},
	})

	markdown, err := converter.ConvertString(rawHTML)
	if err != nil {
		return models.MarkdownDocument{}, errors.Wrap(err, "failed to convert HTML to Markdown")
	}

	return models.MarkdownDocument{
		Content:    strings.TrimSpace(markdown),
		SourceType: models.ContentTypeHTML,
	}, nil
}
