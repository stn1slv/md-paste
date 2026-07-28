// Package menubar implements the macOS menu bar presentation layer. It runs a
// resident status-bar app whose menu triggers the shared clipboard-to-Markdown
// pipeline. Non-darwin builds compile a stub that reports the feature is
// unsupported, mirroring the internal/clipboard platform pattern.
package menubar

// Config carries build-time metadata shown in the menu (e.g. the About item).
type Config struct {
	Version string
	Commit  string
	Date    string
}
