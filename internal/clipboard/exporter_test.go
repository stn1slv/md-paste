package clipboard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stn1slv/md-paste/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveRaw(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Save HTML content", func(t *testing.T) {
		path := filepath.Join(tmpDir, "test.html")
		content := models.ClipboardContent{
			RawHTML:     "<html><body><h1>Hello</h1></body></html>",
			PlainText:   "Hello",
			ContentType: models.ContentTypeHTML,
		}

		err := SaveRaw(path, content)
		require.NoError(t, err)

		//nolint:gosec // Test reads from known path
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, content.RawHTML, string(data))
	})

	t.Run("Save PlainText content as fallback", func(t *testing.T) {
		path := filepath.Join(tmpDir, "test.txt")
		content := models.ClipboardContent{
			RawHTML:     "",
			PlainText:   "Hello from PlainText",
			ContentType: models.ContentTypePlainText,
		}

		err := SaveRaw(path, content)
		require.NoError(t, err)

		//nolint:gosec // Test reads from known path
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, content.PlainText, string(data))
	})

	t.Run("Error if path is a directory", func(t *testing.T) {
		path := tmpDir
		content := models.ClipboardContent{
			RawHTML:     "html",
			PlainText:   "text",
			ContentType: models.ContentTypeHTML,
		}

		err := SaveRaw(path, content)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "is a directory")
	})

	t.Run("Error if path is unwritable", func(t *testing.T) {
		// Create a directory with no write permissions
		readonlyDir := filepath.Join(tmpDir, "readonly")
		//nolint:gosec // Readonly directory for test
		err := os.Mkdir(readonlyDir, 0o555)
		require.NoError(t, err)

		path := filepath.Join(readonlyDir, "test.html")
		content := models.ClipboardContent{
			RawHTML:     "html",
			PlainText:   "text",
			ContentType: models.ContentTypeHTML,
		}

		err = SaveRaw(path, content)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write file")
	})
}
