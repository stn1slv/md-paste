//go:build windows

package menubar

import (
	"context"
	"log/slog"
	"os/exec"
)

// openConfigFile opens the config file with the user's default handler so they
// can edit it. `cmd /c start` returns immediately; the editor runs detached.
// The empty "" argument is start's window-title placeholder so a quoted path is
// not consumed as the title.
func openConfigFile(path string) error {
	//nolint:gosec // path is the app's own config file path
	cmd := exec.CommandContext(context.Background(), "cmd", "/c", "start", "", path)
	if err := cmd.Start(); err != nil {
		return err
	}
	// Reap the child. The tray app is resident, so an unwaited `cmd` would leak a
	// process handle on every use of this menu item.
	go func() {
		if err := cmd.Wait(); err != nil {
			slog.Error("the config file handler exited with an error", "path", path, "error", err)
		}
	}()
	return nil
}
