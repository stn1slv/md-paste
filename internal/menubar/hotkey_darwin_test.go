//go:build darwin

package menubar

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCarbonModifiers(t *testing.T) {
	mask, err := carbonModifiers([]string{"cmd", "opt"})
	require.NoError(t, err)
	assert.Equal(t, uint32(0x0100|0x0800), mask)

	_, err = carbonModifiers([]string{"cmd", "bogus"})
	require.Error(t, err)

	_, err = carbonModifiers(nil)
	require.Error(t, err)
}

func TestMacKeyCodeDefault(t *testing.T) {
	mods, key, err := splitHotkey(defaultHotkey)
	require.NoError(t, err)
	assert.Equal(t, []string{"ctrl", "cmd", "opt"}, mods)
	assert.Equal(t, uint32(0x2E), macKeyCodes[key]) // kVK_ANSI_M
}
