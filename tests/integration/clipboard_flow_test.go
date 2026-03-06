package integration

import (
	"strings"
	"testing"

	"github.com/stn1slv/md-paste/internal/converter"
	"github.com/stn1slv/md-paste/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClipboardFlow(t *testing.T) {
	// This tests the flow for different "simulated" browser formats.
	// In a real integration test on macOS, we'd write these exact types to the pasteboard
	// using a custom CGO writer for test purposes, but since md-paste only reads them
	// and writes plain text, we will simulate the flow by directly calling Convert
	// with contents that represent Safari, Chrome, and Word.

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "Safari simulated HTML",
			// Safari often includes meta charset and body wrappers
			html:     "<html><head><meta charset=\"utf-8\"></head><body><b>Bold from Safari</b></body></html>",
			expected: "**Bold from Safari**",
		},
		{
			name: "Chrome simulated HTML",
			// Chrome often uses inline styles or generic wrappers
			html: "<meta charset='utf-8'><span style=\"font-weight: bold;\">Bold from Chrome</span>",
			// Note: html-to-markdown handles standard b/strong. inline styles might be dropped unless explicitly handled,
			// but we will test standard HTML for now as "Chrome". Actually chrome copies both b/strong.
			expected: "Bold from Chrome", // Without a strong tag, it's just text. Let's use <strong>
		},
		{
			name:     "Chrome simulated HTML with bold",
			html:     "<meta charset='utf-8'><strong style=\"font-weight: bold;\">Bold from Chrome</strong>",
			expected: "**Bold from Chrome**",
		},
		{
			name: "MS Word simulated HTML",
			// Word has complex XML namespaces and classes
			html:     "<html xmlns:o=\"urn:schemas-microsoft-com:office:office\"><body lang=EN-US style='tab-interval:36.0pt'><p class=MsoNormal><i>Italic from Word</i></p></body></html>",
			expected: "_Italic from Word_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := models.ClipboardContent{
				ContentType: models.ContentTypeHTML,
				RawHTML:     tt.html,
			}

			doc, err := converter.Convert(content)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, strings.TrimSpace(doc.Content))
		})
	}
}
