package cli

import (
	"github.com/spf13/cobra"

	"github.com/stn1slv/md-paste/internal/menubar"
)

var menubarCmd = &cobra.Command{
	Use:   "menubar",
	Short: "Run md-paste as a macOS menu bar or Windows tray app",
	Long: `menubar runs md-paste as a resident macOS menu bar or Windows tray app.

Click the icon and choose "Convert clipboard to Markdown" to convert the
current clipboard content in place, or press the global shortcut. The menu
also offers a "Launch at login" toggle. This command is supported on macOS
and Windows.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		return menubar.Run(menubar.Config{Version: version, Commit: commit, Date: date})
	},
}

func init() {
	rootCmd.AddCommand(menubarCmd)
}
