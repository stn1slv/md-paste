//go:build darwin

package menubar

import (
	"context"
	"log/slog"
	"os/exec"
)

// openConfigFile opens the config file with the user's default handler so they
// can edit it. `open` returns quickly; the editor runs detached.
func openConfigFile(path string) error {
	//nolint:gosec // path is the app's own config file path
	cmd := exec.CommandContext(context.Background(), "open", path)
	if err := cmd.Start(); err != nil {
		return err
	}
	// Reap the child. The menu bar app is resident, so an unwaited `open` would
	// leave a zombie behind on every use of this menu item.
	go func() {
		if err := cmd.Wait(); err != nil {
			slog.Error("the config file handler exited with an error", "path", path, "error", err)
		}
	}()
	return nil
}
