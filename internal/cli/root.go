// Package cli implements the command-line interface.
package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/stn1slv/md-paste/internal/clipboard"
	"github.com/stn1slv/md-paste/internal/models"
	"github.com/stn1slv/md-paste/internal/service"
)

// Build-time metadata, populated via -ldflags by goreleaser. See .goreleaser.yaml.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	rootCmd = &cobra.Command{
		Use:   "md-paste",
		Short: "Convert rich text on the clipboard to Markdown",
		Long: `md-paste reads HTML from the system clipboard and converts it to Markdown.
By default, it writes the converted Markdown back to the clipboard.`,
		Example: `  # Convert clipboard HTML to Markdown and save it back to clipboard
  md-paste

  # Convert clipboard HTML to Markdown and print it to stdout
  md-paste --stdout
  md-paste -s

  # Pipe the converted Markdown to another command
  md-paste -s | grep "TODO"`,
		Version:       fmt.Sprintf("%s (commit %s, built %s)", version, commit, date),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runPaste,
	}

	// Flags
	stdoutFlag  bool
	saveRawFlag string

	// Injected dependencies for testing
	clipboardRead  = clipboard.Read
	clipboardWrite = clipboard.WriteMarkdown
)

func init() {
	rootCmd.Flags().BoolVarP(&stdoutFlag, "stdout", "s", false, "Print converted Markdown to stdout instead of clipboard")
	rootCmd.Flags().StringVarP(&saveRawFlag, "save-raw", "r", "", "File path where raw clipboard data will be saved")
	rootCmd.SetVersionTemplate("md-paste {{.Version}}\n")
}

// Execute is the main entry point for the CLI.
func Execute() error {
	return rootCmd.Execute()
}

func runPaste(cmd *cobra.Command, _ []string) error {
	var hooks service.Hooks

	// --save-raw: persist the raw clipboard content before conversion.
	if saveRawFlag != "" {
		path := saveRawFlag
		hooks.OnRawContent = func(content models.ClipboardContent) error {
			return clipboard.SaveRaw(path, content)
		}
	}

	// --stdout: send the Markdown to stdout instead of the clipboard.
	if stdoutFlag {
		out := cmd.OutOrStdout()
		hooks.Sink = func(md string) error {
			return printToStdout(out, md)
		}
	}

	// Silence-on-Empty and Silence-on-Success are handled by service.Convert:
	// it writes nothing on empty clipboards and returns no message on success.
	_, _, err := service.Convert(clipboardRead, clipboardWrite, hooks)
	return err
}

func printToStdout(out io.Writer, content string) error {
	_, err := fmt.Fprintln(out, content)
	return err
}
