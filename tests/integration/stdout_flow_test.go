package integration

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"github.com/stn1slv/md-paste/internal/clipboard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const binPath = "../../bin/md-paste"

func TestStdoutFlow(t *testing.T) {
	// 1. Build the binary so we can test it directly
	cmdBuild := exec.Command("go", "build", "-o", binPath, "../../cmd/md-paste")
	err := cmdBuild.Run()
	require.NoError(t, err, "failed to build binary")

	// 2. Set up the clipboard
	htmlContent := "<h1>Integration Test</h1><p>Testing stdout flag.</p>"

	// Since we don't have a direct WriteHTML in our clipboard package (it only writes Markdown),
	// we will use plain text to test the flow, as our cli doesn't care whether it was HTML or plain text
	// for the stdout flag logic.
	err = clipboard.WriteMarkdown(htmlContent) // Write it as plain text to the clipboard
	require.NoError(t, err)

	// 3. Run the binary with --stdout
	cmdRun := exec.Command(binPath, "--stdout")
	var stdout bytes.Buffer
	cmdRun.Stdout = &stdout
	err = cmdRun.Run()
	require.NoError(t, err)

	// 4. Verify output
	output := strings.TrimSpace(stdout.String())
	assert.Equal(t, htmlContent, output) // It should just output the plain text we wrote

	// 5. Run the binary without --stdout (normal flow)
	err = clipboard.WriteMarkdown("<html><b>Test2</b></html>")
	require.NoError(t, err)

	cmdRunNormal := exec.Command(binPath)
	var stdoutNormal bytes.Buffer
	cmdRunNormal.Stdout = &stdoutNormal
	err = cmdRunNormal.Run()
	require.NoError(t, err)

	assert.Empty(t, stdoutNormal.String(), "normal flow should be silent")
}
