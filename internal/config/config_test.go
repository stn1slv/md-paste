package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMissingReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	defaults := Config{Hotkey: "cmd+opt+m", LaunchAtLogin: true}

	cfg, existed, err := Load(path, defaults)

	require.NoError(t, err)
	assert.False(t, existed)
	assert.Equal(t, defaults, cfg)
}

func TestLoadExistingParsesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("hotkey: ctrl+alt+v\nlaunch_at_login: true\n"), 0o600))

	cfg, existed, err := Load(path, Config{Hotkey: "cmd+opt+m"})

	require.NoError(t, err)
	assert.True(t, existed)
	assert.Equal(t, Config{Hotkey: "ctrl+alt+v", LaunchAtLogin: true}, cfg)
}

func TestLoadEmptyHotkeyFallsBackToDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("hotkey: \"\"\nlaunch_at_login: false\n"), 0o600))

	cfg, existed, err := Load(path, Config{Hotkey: "cmd+opt+m"})

	require.NoError(t, err)
	assert.True(t, existed)
	assert.Equal(t, "cmd+opt+m", cfg.Hotkey)
}

func TestLoadInvalidYAMLReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("hotkey: [unterminated"), 0o600))

	_, existed, err := Load(path, Config{})

	require.Error(t, err)
	assert.True(t, existed)
}

func TestSaveRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.yaml")
	want := Config{Hotkey: "ctrl+alt+m", LaunchAtLogin: true}

	require.NoError(t, Save(path, want))

	got, existed, err := Load(path, Config{})
	require.NoError(t, err)
	assert.True(t, existed)
	assert.Equal(t, want, got)
}

func TestPathEndsWithExpectedSuffix(t *testing.T) {
	path, err := Path()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("md-paste", "config.yaml"), filepath.Join(filepath.Base(filepath.Dir(path)), filepath.Base(path)))
}
