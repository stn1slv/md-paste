//go:build windows

package menubar

import (
	"context"
	"os/exec"
)

// openConfigFile opens the config file with the user's default handler so they
// can edit it. `cmd /c start` returns immediately; the editor runs detached.
// The empty "" argument is start's window-title placeholder so a quoted path is
// not consumed as the title.
func openConfigFile(path string) error {
	//nolint:gosec // path is the app's own config file path
	return exec.CommandContext(context.Background(), "cmd", "/c", "start", "", path).Start()
}
