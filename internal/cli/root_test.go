package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmdFlags(t *testing.T) {
	// Reset flag before testing
	stdoutFlag = false

	cmd := rootCmd
	// Only parse flags to avoid executing the actual clipboard logic
	err := cmd.ParseFlags([]string{"--stdout"})
	require.NoError(t, err)

	assert.True(t, stdoutFlag)
}
