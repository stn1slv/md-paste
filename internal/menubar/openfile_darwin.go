//go:build darwin

package menubar

import (
	"context"
	"os/exec"
)

// openConfigFile opens the config file with the user's default handler so they
// can edit it. `open` returns quickly; the editor runs detached.
func openConfigFile(path string) error {
	//nolint:gosec // path is the app's own config file path
	return exec.CommandContext(context.Background(), "open", path).Start()
}
