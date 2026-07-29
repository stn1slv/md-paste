// Package config loads and saves the resident app's user configuration: the
// global shortcut and the launch-at-login preference. The file is YAML at the
// per-user OS config directory. It is platform-neutral; the menu bar / tray app
// supplies platform-specific defaults (e.g. the default shortcut) when loading.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the on-disk user configuration.
type Config struct {
	// Hotkey is the global shortcut spec, e.g. "cmd+opt+m". Modifiers and the
	// key are joined by "+"; see internal/menubar for parsing.
	Hotkey string `yaml:"hotkey"`
	// LaunchAtLogin is the desired autostart state. It is the source of truth;
	// the app reconciles the OS mechanism (LaunchAgent / Run key) to match it.
	LaunchAtLogin bool `yaml:"launch_at_login"`
}

// Path returns the config file location: <os.UserConfigDir>/md-paste/config.yaml.
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve user config dir: %w", err)
	}
	return filepath.Join(dir, "md-paste", "config.yaml"), nil
}

// Load reads the config at path. When the file does not exist it returns
// defaults with existed=false so the caller can seed and create it. When the
// file exists, any empty field falls back to the corresponding default, so a
// partial file still yields a usable configuration.
func Load(path string, defaults Config) (cfg Config, existed bool, err error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is the app's own config file, not user-supplied input
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return defaults, false, nil
		}
		return Config{}, false, fmt.Errorf("failed to read config: %w", err)
	}

	cfg = defaults
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, true, fmt.Errorf("failed to parse config: %w", err)
	}
	if cfg.Hotkey == "" {
		cfg.Hotkey = defaults.Hotkey
	}
	return cfg, true, nil
}

// Save writes the config to path, creating the parent directory if needed.
func Save(path string, c Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}
