//go:build windows

package menubar

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWinModifiersMask(t *testing.T) {
	mask, err := winModifiersMask([]string{"ctrl", "alt"})
	require.NoError(t, err)
	assert.Equal(t, uint32(modControl|modAlt), mask)

	_, err = winModifiersMask([]string{"ctrl", "bogus"})
	require.Error(t, err)

	_, err = winModifiersMask(nil)
	require.Error(t, err)
}

func TestWinKeyCodeDefault(t *testing.T) {
	mods, key, err := splitHotkey(defaultHotkey)
	require.NoError(t, err)
	assert.Equal(t, []string{"ctrl", "alt"}, mods)
	assert.Equal(t, uint32(0x4D), winKeyCodes[key]) // VK 'M'
	assert.Equal(t, uint32(0x70), winKeyCodes["f1"])
	assert.Equal(t, uint32(0x20), winKeyCodes["space"])
}
