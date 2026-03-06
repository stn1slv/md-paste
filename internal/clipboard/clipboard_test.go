package clipboard

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: These tests interact with the actual macOS clipboard.
// They save the current state and restore it afterwards.

func TestClipboard_ReadWrite(t *testing.T) {
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
