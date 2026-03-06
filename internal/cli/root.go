// Package cli implements the command-line interface.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stn1slv/md-paste/internal/clipboard"
	"github.com/stn1slv/md-paste/internal/converter"
	"github.com/stn1slv/md-paste/internal/errors"
	"github.com/stn1slv/md-paste/internal/models"
)

var (
	rootCmd = &cobra.Command{
		Use:   "md-paste",
		Short: "Convert rich text on the macOS clipboard to Markdown",
		Long: `md-paste reads HTML from the macOS clipboard and converts it to Markdown.
By default, it writes the converted Markdown back to the clipboard.`,
		Example: `  # Convert clipboard HTML to Markdown and save it back to clipboard
  md-paste

  # Convert clipboard HTML to Markdown and print it to stdout
  md-paste --stdout
  md-paste -s

  # Pipe the converted Markdown to another command
  md-paste -s | grep "TODO"`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runPaste,
	}

	// Flags
	stdoutFlag bool
)

func init() {
	rootCmd.Flags().BoolVarP(&stdoutFlag, "stdout", "s", false, "Print converted Markdown to stdout instead of clipboard")
}

// Execute is the main entry point for the CLI.
func Execute() error {
	return rootCmd.Execute()
}

func runPaste(cmd *cobra.Command, _ []string) error {
	content, err := clipboard.Read()
	if err != nil {
		return errors.Wrap(err, "failed to read clipboard")
	}

	if content.ContentType == models.ContentTypeNone {
		// Silence-on-Empty: FR-003.1 says exit silently and perform no write.
		return nil
	}

	doc, err := converter.Convert(content)
	if err != nil {
		return errors.Wrap(err, "failed to convert content")
	}

	if stdoutFlag {
		fmt.Fprintln(cmd.OutOrStdout(), doc.Content)
		return nil
	}

	if err := clipboard.WriteMarkdown(doc.Content); err != nil {
		return errors.Wrap(err, "failed to write to clipboard")
	}

	// Silence-on-Success
	return nil
}
