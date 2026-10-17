// Package converter provides HTML to Markdown conversion logic.
package converter

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/stn1slv/md-paste/internal/models"
	"golang.org/x/net/html"
)

const emptyDomain = ""

// Convert takes clipboard content and converts it to a Markdown document.
func Convert(content models.ClipboardContent) (models.MarkdownDocument, error) {
	if content.ContentType == models.ContentTypeNone || (content.RawHTML == "" && content.PlainText == "") {
		return models.MarkdownDocument{}, errors.New("no content to convert")
	}

	// 1. If HTML is available, it always wins (FR-011). The plain-text table
	// heuristic must not override it: the plain text of a <pre> block or of
	// aligned prose often looks tabular, and using it would drop the HTML.
	if content.RawHTML != "" {
		return performStandardHTMLConversion(content.RawHTML)
	}

	// 2. Try layout-aware text table extraction for plain-text-only content
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

// newMarkdownConverter returns an HTML-to-Markdown converter with the GFM
// extensions md-paste supports (strikethrough, task lists) and inline-style
// normalization. Tables are handled by a custom rule in
// performStandardHTMLConversion.
func newMarkdownConverter() *htmltomarkdown.Converter {
	converter := htmltomarkdown.NewConverter(emptyDomain, true, nil)
	converter.AddRules(strikethroughRule, taskListRule)
	converter.Before(normalizeInlineStyles)
	return converter
}

// The GFM rules below are written here instead of using the library's plugin
// package, which would also pull in gopkg.in/yaml.v2 for an unrelated plugin.

var strikethroughRule = htmltomarkdown.Rule{
	Filter: []string{"del", "s", "strike"},
	Replacement: func(content string, selec *goquery.Selection, _ *htmltomarkdown.Options) *string {
		content = strings.TrimSpace(content)
		if content == "" {
			return &content
		}
		content = htmltomarkdown.AddSpaceIfNessesary(selec, "~~"+content+"~~")
		return &content
	},
}

var taskListRule = htmltomarkdown.Rule{
	Filter: []string{"input"},
	Replacement: func(_ string, selec *goquery.Selection, _ *htmltomarkdown.Options) *string {
		if selec.AttrOr("type", "") != "checkbox" || !selec.Parent().Is("li") {
			return nil
		}
		box := "[ ]"
		if _, checked := selec.Attr("checked"); checked {
			box = "[x]"
		}
		// The label text usually brings its own leading space.
		next := selec.Nodes[0].NextSibling
		if next == nil || next.Type != html.TextNode || next.Data == "" || !unicode.IsSpace(rune(next.Data[0])) {
			box += " "
		}
		return &box
	},
}

func performStandardHTMLConversion(rawHTML string) (models.MarkdownDocument, error) {
	converter := newMarkdownConverter()

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
		return models.MarkdownDocument{}, fmt.Errorf("failed to convert HTML to Markdown: %w", err)
	}

	return models.MarkdownDocument{
		Content:    strings.TrimSpace(markdown),
		SourceType: models.ContentTypeHTML,
	}, nil
}
