//go:build darwin

package clipboard

import (
	"os"
	"testing"

	"github.com/stn1slv/md-paste/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: These tests interact with the actual macOS clipboard.
// They save the current state and restore it afterwards.

//nolint:revive // Test setup involves multiple state checks
func TestClipboard_ReadWrite(t *testing.T) {
	if os.Getenv("MD_PASTE_E2E") == "" {
		t.Skip("Skipping clipboard-mutating test; set MD_PASTE_E2E=1 to run")
	}

	// Save the current clipboard state to restore it afterwards
	originalContent, err := Read()
	require.NoError(t, err)

	if originalContent.PlainText == "" && originalContent.RawHTML != "" {
		t.Skip("Skipping test to avoid destructive cleanup of rich HTML-only clipboard state")
	} else if originalContent.PlainText == "" && originalContent.ContentType != models.ContentTypeNone {
		t.Skip("Skipping test to avoid destructive cleanup of non-text clipboard state")
	}

	t.Cleanup(func() {
		// Try to restore appropriately based on what was there
		switch {
		case originalContent.RawHTML != "":
			// We don't have a specific WriteHTML, but restoring plain text is better than nothing
			// Actually, just restoring the plain text for now.
			_ = WriteMarkdown(originalContent.PlainText)
		case originalContent.PlainText != "":
			_ = WriteMarkdown(originalContent.PlainText)
		default:
			_ = Clear()
		}
	})

	// Skip if not on macOS or no display environment? Actually, since it's a macOS specific app,
	// let's assume the test environment supports it.

	// 1. Test Writing and Reading Plain Text
	t.Run("PlainText", func(t *testing.T) {
		err := WriteMarkdown("Hello, Markdown!")
		require.NoError(t, err)

		content, err := Read()
		require.NoError(t, err)

		// The reader might prioritize HTML if something else left it there, but our WriteMarkdown
		// should clear the pasteboard and set plain text (or we might set HTML too? No, markdown is plain text).
		assert.Equal(t, "Hello, Markdown!", content.PlainText)
	})

	// 2. Test Empty Clipboard (can be hard to simulate perfectly without clearing it completely)
	t.Run("ClearAndRead", func(t *testing.T) {
		err := Clear()
		require.NoError(t, err)

		content, err := Read()
		require.NoError(t, err)
		assert.Empty(t, content.PlainText)
		assert.Empty(t, content.RawHTML)
	})
}
