//go:build !darwin && !windows

package menubar

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunUnsupportedPlatform(t *testing.T) {
	err := Run(Config{Version: "test"})
	require.Error(t, err)
	assert.ErrorIs(t, err, errUnsupported)
}
