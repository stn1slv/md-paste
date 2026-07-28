//go:build darwin

package menubar

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderPlist(t *testing.T) {
	out := renderPlist("/Applications/md-paste.app/Contents/MacOS/md-paste-bin")

	assert.Contains(t, out, "<string>"+launchAgentLabel+"</string>")
	assert.Contains(t, out, "<string>/Applications/md-paste.app/Contents/MacOS/md-paste-bin</string>")
	assert.Contains(t, out, "<string>menubar</string>")
	assert.Contains(t, out, "<key>RunAtLoad</key>")
}

func TestGUIDomain(t *testing.T) {
	d := guiDomain()
	require.NotEmpty(t, d)
	assert.Regexp(t, `^gui/\d+$`, d)
}
