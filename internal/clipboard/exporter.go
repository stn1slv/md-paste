package clipboard

import (
	"fmt"
	"os"

	"github.com/stn1slv/md-paste/internal/atomicfile"
	"github.com/stn1slv/md-paste/internal/models"
)

// rawFileMode keeps exported clipboard data readable by its owner only. The
// clipboard routinely holds passwords, tokens and other private content, so it
// must not land on disk world-readable.
const rawFileMode = 0o600

// SaveRaw saves the raw clipboard content to a file.
// It prioritizes RawHTML over PlainText.
// It returns an error if the path is a directory or if the file cannot be written.
// The write is atomic and replaces the destination rather than writing through
// it, so an interrupted run cannot truncate a previous export and a symlink at
// the path is not followed.
func SaveRaw(path string, content models.ClipboardContent) error {
	info, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat path %q: %w", path, err)
		}
	} else if info.IsDir() {
		return fmt.Errorf("%q is a directory", path)
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

	if err := atomicfile.Write(path, data, rawFileMode); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}

	return nil
}
