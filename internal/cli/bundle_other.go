//go:build !darwin

package cli

// isBundleLaunch is always false off macOS; there is no .app bundle.
func isBundleLaunch() bool { return false }
