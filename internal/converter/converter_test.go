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
