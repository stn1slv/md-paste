//go:build windows

package menubar

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

const (
	runKeyPath   = `Software\Microsoft\Windows\CurrentVersion\Run`
	runValueName = "md-paste"
)

// formatRunValue wraps the executable path in double quotes so a path containing
// spaces (e.g. under Program Files) is parsed as a single argument by the shell.
func formatRunValue(exe string) string {
	return `"` + exe + `"`
}

// loginItemEnabled reports whether the HKCU Run entry points at the current
// executable. Requiring a match (not just presence) means a stale entry left by
// a moved or updated binary reads as disabled and is self-healed the next time
// the user toggles it on. If the executable path cannot be resolved, it falls
// back to treating any present entry as enabled.
func loginItemEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer func() { _ = k.Close() }()
	val, _, err := k.GetStringValue(runValueName)
	if err != nil {
		return false
	}
	exe, err := os.Executable()
	if err != nil {
		return true
	}
	return val == formatRunValue(exe)
}

// enableLoginItem adds an HKCU Run entry that starts the tray at login. It points
// at the currently running executable (the windowsgui md-paste-tray.exe), which
// defaults to the menu bar, so no console window appears at startup.
func enableLoginItem() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open Run key: %w", err)
	}
	defer func() { _ = k.Close() }()
	if err := k.SetStringValue(runValueName, formatRunValue(exe)); err != nil {
		return fmt.Errorf("failed to write Run value: %w", err)
	}
	return nil
}

// disableLoginItem removes the HKCU Run entry. A missing key or value is treated
// as already disabled.
func disableLoginItem() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to open Run key: %w", err)
	}
	defer func() { _ = k.Close() }()
	if err := k.DeleteValue(runValueName); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("failed to delete Run value: %w", err)
	}
	return nil
}
