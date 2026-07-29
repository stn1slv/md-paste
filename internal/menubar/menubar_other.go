//go:build !darwin && !windows

package menubar

import "errors"

var errUnsupported = errors.New("unsupported platform: the menu bar app is only available on macOS and Windows")

// Run reports that the menu bar app is unavailable on this platform.
func Run(_ Config) error {
	return errUnsupported
}
