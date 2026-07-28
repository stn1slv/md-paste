//go:build darwin

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathIsBundle(t *testing.T) {
	assert.True(t, pathIsBundle("/Applications/md-paste.app/Contents/MacOS/md-paste"))
	assert.False(t, pathIsBundle("/usr/local/bin/md-paste"))
	assert.False(t, pathIsBundle("/opt/homebrew/bin/md-paste"))
}
