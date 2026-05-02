//go:build !darwin && !windows

// Package clipboard provides fallback clipboard integration for unsupported platforms.
package clipboard

import (
	"errors"

	"github.com/stn1slv/md-paste/internal/models"
)

var errUnsupported = errors.New("unsupported platform: md-paste requires macOS or Windows")

// Read returns an error on non-macOS platforms.
func Read() (models.ClipboardContent, error) {
	return models.ClipboardContent{}, errUnsupported
}

// WriteMarkdown returns an error on non-macOS platforms.
func WriteMarkdown(_ string) error {
	return errUnsupported
}

// Clear returns an error on non-macOS platforms.
func Clear() error {
	return errUnsupported
}
