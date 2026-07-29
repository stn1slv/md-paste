//go:build windows

package menubar

import _ "embed"

// Windows tray icons in .ico format (systray.SetIcon expects ICO bytes on
// Windows). Colored fills so they read on both light and dark taskbars;
// iconCheck is shown briefly after a successful conversion.
var (
	//go:embed icons/icon-normal.ico
	iconNormal []byte

	//go:embed icons/icon-check.ico
	iconCheck []byte
)
