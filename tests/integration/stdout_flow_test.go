//go:build darwin

package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stn1slv/md-paste/internal/clipboard"
	"github.com/stn1slv/md-paste/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:revive // Integration test requires linear state setup
func TestStdoutFlow(t *testing.T) {
	if os.Getenv("MD_PASTE_E2E") == "" {
		t.Skip("Skipping clipboard-mutating test; set MD_PASTE_E2E=1 to run")
	}

	// Save clipboard state to restore after the test
	originalContent, err := clipboard.Read()
	require.NoError(t, err)

	if originalContent.RawHTML != "" {
		t.Skip("Skipping test to avoid destructive cleanup of rich HTML clipboard state")
	} else if originalContent.PlainText == "" && originalContent.ContentType != models.ContentTypeNone {
		t.Skip("Skipping test to avoid destructive cleanup of non-text clipboard state")
	}

	t.Cleanup(func() {
		if originalContent.PlainText != "" {
			_ = clipboard.WriteMarkdown(originalContent.PlainText)
		} else {
			_ = clipboard.Clear()
		}
	})

	// 1. Build the binary to a temporary directory so we can test it directly
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "md-paste")
	//nolint:gosec // Testing execution of built binary
	cmdBuild := exec.CommandContext(t.Context(), "go", "build", "-o", binPath, "../../cmd/md-paste")
	err = cmdBuild.Run()
	require.NoError(t, err, "failed to build binary")

	// 2. Set up the clipboard
	htmlContent := "<h1>Integration Test</h1><p>Testing stdout flag.</p>"

	// Since we don't have a direct WriteHTML in our clipboard package (it only writes Markdown),
	// we will use plain text to test the flow, as our cli doesn't care whether it was HTML or plain text
	// for the stdout flag logic.
	err = clipboard.WriteMarkdown(htmlContent) // Write it as plain text to the clipboard
	require.NoError(t, err)

	// 3. Run the binary with --stdout
	//nolint:gosec // Testing execution of built binary
	cmdRun := exec.CommandContext(t.Context(), binPath, "--stdout")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmdRun.Stdout = &stdout
	cmdRun.Stderr = &stderr
	err = cmdRun.Run()
	require.NoError(t, err, "stderr: %s", stderr.String())

	// 4. Verify output and clipboard state
	output := strings.TrimSpace(stdout.String())
	assert.Equal(t, htmlContent, output, "stderr: %s", stderr.String()) // It should just output the plain text we wrote

	// Ensure the clipboard remained completely unchanged!
	afterStdoutContent, err := clipboard.Read()
	require.NoError(t, err)
	assert.Equal(t, htmlContent, afterStdoutContent.PlainText)

	// 5. Run the binary without --stdout (normal flow)
	err = clipboard.WriteMarkdown("<html><b>Test2</b></html>")
	require.NoError(t, err)

	//nolint:gosec // Testing execution of built binary
	cmdRunNormal := exec.CommandContext(t.Context(), binPath)
	var stdoutNormal bytes.Buffer
	cmdRunNormal.Stdout = &stdoutNormal
	err = cmdRunNormal.Run()
	require.NoError(t, err)

	assert.Empty(t, stdoutNormal.String(), "normal flow should be silent")

	// 6. Test combined --save-raw and --stdout
	rawFile := filepath.Join(binDir, "raw.html")
	rawHTML := "<html><body><h1>Combined</h1></body></html>"
	err = clipboard.WriteMarkdown(rawHTML)
	require.NoError(t, err)

	//nolint:gosec // Testing execution of built binary
	cmdCombined := exec.CommandContext(t.Context(), binPath, "--stdout", "--save-raw", rawFile)
	var stdoutCombined bytes.Buffer
	var stderrCombined bytes.Buffer
	cmdCombined.Stdout = &stdoutCombined
	cmdCombined.Stderr = &stderrCombined
	err = cmdCombined.Run()
	require.NoErrorf(t, err, "combined command failed, stderr: %s", stderrCombined.String())

	assert.Contains(t, stdoutCombined.String(), "Combined")
	//nolint:gosec // Integration test reads from known temporary path
	data, err := os.ReadFile(rawFile)
	require.NoError(t, err)
	assert.Equal(t, rawHTML, string(data))
}
