// Package menubar implements the resident status-bar presentation layer: the
// macOS menu bar app and the Windows system tray app. It runs a status-bar app
// whose menu triggers the shared clipboard-to-Markdown pipeline. Platforms other
// than macOS and Windows compile a stub that reports the feature is unsupported,
// mirroring the internal/clipboard platform pattern.
package menubar

// Config carries build-time metadata shown in the menu (e.g. the About item).
type Config struct {
	Version string
	Commit  string
	Date    string
}
