//go:build windows

package menubar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatRunValue(t *testing.T) {
	// Windows paths keep single backslashes; only surrounding quotes are added so
	// a path with spaces is treated as a single argument.
	assert.Equal(t,
		`"C:\Program Files\md-paste\md-paste-tray.exe"`,
		formatRunValue(`C:\Program Files\md-paste\md-paste-tray.exe`),
	)
}

func TestRunKeyConstants(t *testing.T) {
	assert.Equal(t, `Software\Microsoft\Windows\CurrentVersion\Run`, runKeyPath)
	assert.Equal(t, "md-paste", runValueName)
}
