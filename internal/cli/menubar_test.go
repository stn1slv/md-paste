package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The Windows tray build injects cli.defaultCommand="menubar" via ldflags (see
// .goreleaser.yaml). Pin the subcommand name so renaming menubarCmd.Use cannot
// silently break the tray launch.
func TestMenubarCommandName(t *testing.T) {
	assert.Equal(t, "menubar", menubarCmd.Use)
}
