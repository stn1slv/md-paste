package converter

import (
	"testing"

	"github.com/stn1slv/md-paste/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name     string
		input    models.ClipboardContent
		expected string
		wantErr  bool
	}{
		{
			name: "HTML Priority",
			input: models.ClipboardContent{
				ContentType: models.ContentTypeHTML,
				RawHTML:     "<h1>Hello</h1><p>World</p>",
				PlainText:   "Hello World",
			},
			expected: "# Hello\n\nWorld",
			wantErr:  false,
		},
		{
			name: "Plain Text Fallback",
			input: models.ClipboardContent{
				ContentType: models.ContentTypePlainText,
				RawHTML:     "",
				PlainText:   "Just plain text",
			},
			expected: "Just plain text",
			wantErr:  false,
		},
		{
			name: "Empty Content",
			input: models.ClipboardContent{
				ContentType: models.ContentTypeNone,
			},
			expected: "",
			wantErr:  true,
		},
		{
			name: "Complex HTML (Lists and Links)",
			input: models.ClipboardContent{
				ContentType: models.ContentTypeHTML,
				RawHTML:     "<ul><li><a href=\"https://example.com\">Link</a></li></ul>",
			},
			expected: "- [Link](https://example.com)",
			wantErr:  false,
		},
		{
			// The plain text of a <pre> block looks tabular; the HTML must win.
			name: "HTML pre block with aligned columns stays a code block",
			input: models.ClipboardContent{
				ContentType: models.ContentTypeHTML,
				RawHTML:     "<pre><code>NAME      READY   STATUS\nweb-1     1/1     Running\ndb-0      1/1     Running</code></pre>",
				PlainText:   "NAME      READY   STATUS\nweb-1     1/1     Running\ndb-0      1/1     Running",
			},
			expected: "```\nNAME      READY   STATUS\nweb-1     1/1     Running\ndb-0      1/1     Running\n```",
		},
		{
			name: "HTML aligned prose is not turned into a table",
			input: models.ClipboardContent{
				ContentType: models.ContentTypeHTML,
				RawHTML:     "<p>Name:  <b>John</b></p><p>Role:  Admin</p><p>Team:  Core</p>",
				PlainText:   "Name:  John\nRole:  Admin\nTeam:  Core",
			},
			expected: "Name: **John**\n\nRole: Admin\n\nTeam: Core",
		},
		{
			name: "Strikethrough",
			input: models.ClipboardContent{
				ContentType: models.ContentTypeHTML,
				RawHTML:     "<p><del>old</del> <s>gone</s> <strike>legacy</strike> kept</p>",
			},
			expected: "~~old~~ ~~gone~~ ~~legacy~~ kept",
		},
		{
			name: "Task list",
			input: models.ClipboardContent{
				ContentType: models.ContentTypeHTML,
				RawHTML:     `<ul><li><input type="checkbox" checked> done</li><li><input type="checkbox"> todo</li></ul>`,
			},
			expected: "- [x] done\n- [ ] todo",
		},
		{
			name: "Task list without a space before the label",
			input: models.ClipboardContent{
				ContentType: models.ContentTypeHTML,
				RawHTML:     `<ul><li><input type="checkbox">todo</li></ul>`,
			},
			expected: "- [ ] todo",
		},
		{
			name: "Inline styles inside table cells",
			input: models.ClipboardContent{
				ContentType: models.ContentTypeHTML,
				RawHTML:     `<table><tr><td>A</td><td>B</td></tr><tr><td><span style="font-weight:700">x</span></td><td><del>y</del></td></tr></table>`,
			},
			expected: "| A | B |\n| --- | --- |\n| **x** | ~~y~~ |",
		},
		{
			name: "Google Docs styles",
			input: models.ClipboardContent{
				ContentType: models.ContentTypeHTML,
				RawHTML: `<meta charset="utf-8"><b style="font-weight:normal;" id="docs-internal-guid-1">` +
					`<p dir="ltr"><span style="font-weight:400;">Plain</span> ` +
					`<span style="font-weight:700;">bold</span> ` +
					`<span style="font-style:italic;">italic</span> ` +
					`<span style="text-decoration:line-through;">struck</span></p></b>`,
			},
			expected: "Plain **bold** _italic_ ~~struck~~",
		},
		{
			name: "Google Docs heading spans are not bolded",
			input: models.ClipboardContent{
				ContentType: models.ContentTypeHTML,
				RawHTML:     `<b style="font-weight:normal;"><h2><span style="font-weight:700;">Title</span></h2></b>`,
			},
			expected: "## Title",
		},
		{
			name: "Bold tag without a normal weight stays bold",
			input: models.ClipboardContent{
				ContentType: models.ContentTypeHTML,
				RawHTML:     `<b style="color:red">red bold</b> <strong style="font-weight:bold">bold</strong>`,
			},
			expected: "**red bold** **bold**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Convert(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, doc.Content)
				assert.Equal(t, tt.input.ContentType, doc.SourceType)
			}
		})
	}
}
