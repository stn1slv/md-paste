package clipboard

import (
	"os"

	"github.com/stn1slv/md-paste/internal/errors"
	"github.com/stn1slv/md-paste/internal/models"
)

// SaveRaw saves the raw clipboard content to a file.
// It prioritizes RawHTML over PlainText.
// It returns an error if the path is a directory or if the file cannot be written.
func SaveRaw(path string, content models.ClipboardContent) error {
	info, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to stat path '%s'", path)
		}
	} else if info.IsDir() {
		return errors.New("'%s' is a directory", path)
	}

	var data []byte
	switch {
	case content.RawHTML != "":
		data = []byte(content.RawHTML)
	case content.PlainText != "":
		data = []byte(content.PlainText)
	default:
		// This should not happen if the clipboard is checked for empty before calling SaveRaw
		return nil
	}

	//nolint:gosec // File is intended to be readable by others (0644)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return errors.Wrap(err, "failed to write file '%s'", path)
	}

	return nil
}
