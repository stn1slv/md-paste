// Package atomicfile writes files atomically. Content goes to a temporary file
// in the destination directory and is then renamed into place, so an
// interrupted or failed write can never leave a truncated file where a complete
// one used to be. Writing to a fresh temporary file also means an existing
// symlink at the destination is replaced rather than followed.
package atomicfile

import (
	"fmt"
	"os"
	"path/filepath"
)

// dirMode is the mode for directories Write creates. Every caller stores private
// user data, so the directory is owner-only.
const dirMode = 0o700

// Write replaces path with data, creating the parent directory if needed. perm
// is the mode of the resulting file.
func Write(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp*")
	if err != nil {
		return fmt.Errorf("failed to create a temporary file in %q: %w", dir, err)
	}
	tmpName := tmp.Name()
	// Leave nothing behind on any failure past this point. After a successful
	// rename the temporary file no longer exists and the removal is a no-op.
	defer func() { _ = os.Remove(tmpName) }()

	if err := writeAndClose(tmp, data, perm); err != nil {
		return err
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("failed to replace %q: %w", path, err)
	}
	return nil
}

// writeAndClose fills f and closes it, closing f on every error path too.
func writeAndClose(f *os.File, data []byte, perm os.FileMode) error {
	if err := f.Chmod(perm); err != nil {
		_ = f.Close()
		return fmt.Errorf("failed to set permissions on %q: %w", f.Name(), err)
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return fmt.Errorf("failed to write %q: %w", f.Name(), err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close %q: %w", f.Name(), err)
	}
	return nil
}
