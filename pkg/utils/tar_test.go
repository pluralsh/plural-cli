package utils_test

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUntar(t *testing.T) {
	t.Run("extracts files inside destination", func(t *testing.T) {
		archive := createTar(t, "nested/file.txt", "content")
		dst := t.TempDir()

		require.NoError(t, utils.Untar(dst, archive))

		content, err := os.ReadFile(filepath.Join(dst, "nested", "file.txt"))
		require.NoError(t, err)
		assert.Equal(t, "content", string(content))
	})

	t.Run("rejects files outside destination", func(t *testing.T) {
		archive := createTar(t, "../outside.txt", "content")
		parent := t.TempDir()
		dst := filepath.Join(parent, "destination")

		err := utils.Untar(dst, archive)

		require.Error(t, err)
		assert.ErrorContains(t, err, "resolves outside destination")
		_, err = os.Stat(filepath.Join(parent, "outside.txt"))
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}

func createTar(t *testing.T, name, content string) *bytes.Reader {
	t.Helper()

	var archive bytes.Buffer
	w := tar.NewWriter(&archive)
	require.NoError(t, w.WriteHeader(&tar.Header{
		Name: name,
		Mode: 0o600,
		Size: int64(len(content)),
	}))
	_, err := w.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	return bytes.NewReader(archive.Bytes())
}
