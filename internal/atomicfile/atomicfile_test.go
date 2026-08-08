package atomicfile_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stn1slv/md-paste/internal/atomicfile"
)

func TestWriteCreatesTheFileWithTheRequestedMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")

	require.NoError(t, atomicfile.Write(path, []byte("hello"), 0o600))

	data, err := os.ReadFile(path) //nolint:gosec // test-controlled path
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))

	if runtime.GOOS == "windows" {
		t.Skip("Windows does not model POSIX permission bits")
	}
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestWriteCreatesMissingParentDirectories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "deeper", "out.txt")

	require.NoError(t, atomicfile.Write(path, []byte("hello"), 0o600))

	data, err := os.ReadFile(path) //nolint:gosec // test-controlled path
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))
}

func TestWriteReplacesExistingContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")
	require.NoError(t, os.WriteFile(path, []byte("old and much longer"), 0o600))

	require.NoError(t, atomicfile.Write(path, []byte("new"), 0o600))

	data, err := os.ReadFile(path) //nolint:gosec // test-controlled path
	require.NoError(t, err)
	assert.Equal(t, "new", string(data))
}

// The destination is replaced, not written through: a symlink planted at the
// path must not redirect the write to its target.
func TestWriteDoesNotFollowASymlinkAtTheDestination(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks on Windows requires elevated privileges")
	}

	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "link.txt")
	require.NoError(t, os.WriteFile(target, []byte("untouched"), 0o600))
	require.NoError(t, os.Symlink(target, link))

	require.NoError(t, atomicfile.Write(link, []byte("secret"), 0o600))

	targetData, err := os.ReadFile(target) //nolint:gosec // test-controlled path
	require.NoError(t, err)
	assert.Equal(t, "untouched", string(targetData), "the write followed the symlink")

	linkData, err := os.ReadFile(link) //nolint:gosec // test-controlled path
	require.NoError(t, err)
	assert.Equal(t, "secret", string(linkData))
}

// A failed write must leave no temporary files in the destination directory.
func TestWriteLeavesNoTemporaryFileBehind(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	require.NoError(t, atomicfile.Write(path, []byte("hello"), 0o600))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "out.txt", entries[0].Name())
}

func TestWriteFailsWhenTheDirectoryIsNotWritable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod-based permission denial does not apply on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}

	dir := t.TempDir()
	//nolint:gosec // Directory mode: r-x is what makes the write fail
	require.NoError(t, os.Chmod(dir, 0o500))
	//nolint:gosec // Restore write access so t.TempDir can clean up
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	err := atomicfile.Write(filepath.Join(dir, "out.txt"), []byte("hello"), 0o600)
	assert.Error(t, err)
}
