//go:build windows

package menubar

import (
	"errors"
	"fmt"
	"os"
	"strings"

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

// startupCommand is the command line stored in the Run key. It explicitly passes
// the "menubar" subcommand so autostart starts the tray regardless of which
// binary enabled it: the windowsgui md-paste-tray.exe (no console window) or the
// console md-paste.exe run as `md-paste menubar`. Relying on the tray build's
// injected default command alone would make the console binary start a one-shot
// conversion at login instead.
func startupCommand(exe string) string {
	return formatRunValue(exe) + " menubar"
}

// loginItemEnabled reports whether the HKCU Run entry points at the current
// executable. Requiring a match (not just presence) means a stale entry left by
// a moved or updated binary reads as disabled and is self-healed the next time
// the user toggles it on. If the executable path cannot be resolved, it falls
// back to treating any present entry as enabled. The comparison is
// case-insensitive because Windows paths are.
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
	return strings.EqualFold(val, startupCommand(exe))
}

// reconcileLoginItem repairs a stale Run entry: when an entry exists (by name)
// but points at a different path (e.g. after a WinGet update to a new versioned
// directory), it is rewritten to the current executable so login autostart keeps
// working without the user having to re-toggle it. A missing entry means the user
// disabled autostart and is left untouched.
func reconcileLoginItem() {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return
	}
	defer func() { _ = k.Close() }()
	val, _, err := k.GetStringValue(runValueName)
	if err != nil {
		return // not enabled; nothing to reconcile
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	if want := startupCommand(exe); !strings.EqualFold(val, want) {
		_ = k.SetStringValue(runValueName, want)
	}
}

// enableLoginItem adds an HKCU Run entry that starts the tray at login, running
// the current executable with the "menubar" subcommand. Launching the windowsgui
// md-paste-tray.exe this way shows no console window.
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
	if err := k.SetStringValue(runValueName, startupCommand(exe)); err != nil {
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
