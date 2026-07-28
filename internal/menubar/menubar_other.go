//go:build !darwin

package menubar

import "errors"

var errUnsupported = errors.New("unsupported platform: the menu bar app is only available on macOS")

// Run reports that the menu bar app is unavailable on non-macOS platforms.
func Run(_ Config) error {
	return errUnsupported
}
