package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmdFlags(t *testing.T) {
	// Reset flags before testing
	stdoutFlag = false

	// Execute with --stdout
	cmd := rootCmd
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{"--stdout"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Since we mock or don't want to actually run the clipboard logic fully if it's empty,
	// if clipboard is empty, it shouldn't error, but it won't output.
	// Actually testing the flag parsing itself:
	assert.True(t, stdoutFlag)
}
