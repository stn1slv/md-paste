//go:build darwin || windows

package menubar

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitHotkey(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		wantMods []string
		wantKey  string
	}{
		{"single modifier", "cmd+m", []string{"cmd"}, "m"},
		{"multiple modifiers", "cmd+opt+m", []string{"cmd", "opt"}, "m"},
		{"whitespace and case", " Cmd + Opt + M ", []string{"cmd", "opt"}, "m"},
		{"windows default", "ctrl+alt+m", []string{"ctrl", "alt"}, "m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mods, key, err := splitHotkey(tt.spec)
			require.NoError(t, err)
			assert.Equal(t, tt.wantMods, mods)
			assert.Equal(t, tt.wantKey, key)
		})
	}
}

func TestSplitHotkeyErrors(t *testing.T) {
	for _, spec := range []string{"m", "", "cmd+", "+m", "cmd++m"} {
		t.Run(spec, func(t *testing.T) {
			_, _, err := splitHotkey(spec)
			assert.Error(t, err)
		})
	}
}
