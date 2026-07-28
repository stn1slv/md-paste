//go:build darwin

package cli

import (
	"os"
	"strings"
)

// isBundleLaunch reports whether the process is running from inside a macOS
// .app bundle (as opposed to a plain CLI binary on PATH).
func isBundleLaunch() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	return pathIsBundle(exe)
}

func pathIsBundle(exe string) bool {
	return strings.Contains(exe, ".app/Contents/MacOS/")
}
