// Package errors provides error wrapping utilities.
package errors

import "fmt"

// Wrap wraps an error with a contextual message.
// It uses the standard fmt.Errorf with the %w verb to allow error unwrap.
func Wrap(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", msg, err)
}

// New creates a new error with the given message.
func New(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
