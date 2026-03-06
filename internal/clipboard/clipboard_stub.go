//go:build !darwin

// Package clipboard provides fallback clipboard integration for non-macOS systems.
package clipboard

import (
	"github.com/stn1slv/md-paste/internal/errors"
	"github.com/stn1slv/md-paste/internal/models"
)

var errUnsupported = errors.New("unsupported platform: md-paste requires macOS")

// Read returns an error on non-macOS platforms.
func Read() (models.ClipboardContent, error) {
	return models.ClipboardContent{}, errUnsupported
}

// WriteMarkdown returns an error on non-macOS platforms.
func WriteMarkdown(text string) error {
	return errUnsupported
}

// Clear returns an error on non-macOS platforms.
func Clear() error {
	return errUnsupported
}
