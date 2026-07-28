//go:build darwin

package menubar

import _ "embed"

// Menu bar template icons: black-on-transparent PNGs whose alpha channel drives
// the macOS status-bar rendering (auto light/dark). iconNormal is the resting
// state; iconCheck is shown briefly after a successful conversion.
var (
	//go:embed icons/icon-normal.png
	iconNormal []byte

	//go:embed icons/icon-check.png
	iconCheck []byte
)
