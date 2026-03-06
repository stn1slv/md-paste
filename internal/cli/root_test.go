package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmdFlags(t *testing.T) {
	// Reset flag after testing
	t.Cleanup(func() { stdoutFlag = false })

	cmd := rootCmd
	// Only parse flags to avoid executing the actual clipboard logic
	err := cmd.ParseFlags([]string{"--stdout"})
	require.NoError(t, err)

	assert.True(t, stdoutFlag)
}
