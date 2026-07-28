//go:build darwin

package menubar

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const launchAgentLabel = "com.stn1slv.md-paste"

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>menubar</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>ProcessType</key>
	<string>Background</string>
</dict>
</plist>
`

// renderPlist builds the LaunchAgent plist that starts `<exe> menubar` at login.
// The executable path is XML-escaped so paths containing characters such as
// '&' produce a valid plist.
func renderPlist(exe string) string {
	var escaped strings.Builder
	// xml.EscapeText only fails if the writer fails; strings.Builder never does.
	_ = xml.EscapeText(&escaped, []byte(exe))
	return fmt.Sprintf(plistTemplate, launchAgentLabel, escaped.String())
}

func plistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", launchAgentLabel+".plist"), nil
}

// loginItemEnabled reports whether the LaunchAgent plist is installed.
func loginItemEnabled() bool {
	path, err := plistPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// enableLoginItem installs a per-user LaunchAgent so md-paste starts at login.
// This approach is signing-independent, unlike SMAppService which needs a
// Developer ID-signed, notarized app bundle.
func enableLoginItem() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}
	path, err := plistPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}
	// Writing the plist is enough: launchd loads ~/Library/LaunchAgents at login.
	// We deliberately do NOT bootstrap/start it now, because the menu bar app is
	// already running (that is how this toggle was clicked); starting a second
	// instance would create a duplicate menu bar icon.
	if err := os.WriteFile(path, []byte(renderPlist(exe)), 0o600); err != nil {
		return fmt.Errorf("failed to write launch agent: %w", err)
	}
	return nil
}

// disableLoginItem unloads and removes the LaunchAgent.
func disableLoginItem() error {
	path, err := plistPath()
	if err != nil {
		return err
	}
	_ = bootoutLaunchAgent(path)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove launch agent: %w", err)
	}
	return nil
}

func bootoutLaunchAgent(path string) error {
	return runLaunchctl("bootout", guiDomain(), path)
}

func guiDomain() string {
	return "gui/" + strconv.Itoa(os.Getuid())
}

func runLaunchctl(args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Args are constructed internally (fixed subcommands plus a plist path
	// derived from os.UserHomeDir), never from untrusted input.
	//nolint:gosec // controlled arguments, no shell interpolation
	out, err := exec.CommandContext(ctx, "launchctl", args...).CombinedOutput()
	if err != nil {
		if msg := strings.TrimSpace(string(out)); msg != "" {
			return fmt.Errorf("launchctl %s: %w: %s", strings.Join(args, " "), err, msg)
		}
		return fmt.Errorf("launchctl %s: %w", strings.Join(args, " "), err)
	}
	return nil
}
