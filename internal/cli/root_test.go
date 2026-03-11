package cli

import (
	"bytes"
	"testing"

	"github.com/stn1slv/md-paste/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmdFlags(t *testing.T) {
	// Reset flags after testing
	t.Cleanup(func() { 
		stdoutFlag = false 
		saveRawFlag = ""
	})

	// Reset pflag values before parsing, as ParseFlags accumulates
	require.NoError(t, rootCmd.Flags().Set("stdout", "false"))
	require.NoError(t, rootCmd.Flags().Set("save-raw", ""))

	cmd := rootCmd
	// Only parse flags to avoid executing the actual clipboard logic
	err := cmd.ParseFlags([]string{"--stdout", "--save-raw", "test.html"})
	require.NoError(t, err)

	assert.True(t, stdoutFlag)
	assert.Equal(t, "test.html", saveRawFlag)
}

func TestTableConversionIntegration(t *testing.T) {
	// Save and restore dependencies
	oldRead := clipboardRead
	oldWrite := clipboardWrite
	oldStdout := stdoutFlag
	oldSaveRaw := saveRawFlag
	oldOut := rootCmd.OutOrStdout()
	t.Cleanup(func() {
		clipboardRead = oldRead
		clipboardWrite = oldWrite
		stdoutFlag = oldStdout
		saveRawFlag = oldSaveRaw
		rootCmd.SetOut(oldOut)
	})

	// Set up mock content: A simple HTML table
	mockHTML := "<table><tr><th>H1</th></tr><tr><td>D1</td></tr></table>"
	clipboardRead = func() (models.ClipboardContent, error) {
		return models.ClipboardContent{
			RawHTML:     mockHTML,
			ContentType: models.ContentTypeHTML,
		}, nil
	}

	// Case 1: --stdout
	stdoutFlag = true
	var outBuf bytes.Buffer
	rootCmd.SetOut(&outBuf)

	err := runPaste(rootCmd, []string{})
	require.NoError(t, err)

	expectedMarkdown := "| H1 |\n| --- |\n| D1 |"
	assert.Contains(t, outBuf.String(), expectedMarkdown)

	// Case 2: normal (to clipboard)
	stdoutFlag = false
	var capturedMarkdown string
	clipboardWrite = func(md string) error {
		capturedMarkdown = md
		return nil
	}

	err = runPaste(rootCmd, []string{})
	require.NoError(t, err)
	assert.Equal(t, expectedMarkdown, capturedMarkdown)
}
