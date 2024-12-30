package mp4

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileAllocate(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "test")

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
	require.NoError(t, err)

	_, err = file.WriteString("HelloWorld")
	require.NoError(t, err)

	// Allocate a gap between "Hello" and "World"
	mp4 := File{file}
	require.NoError(t, mp4.Allocate(5, 2))

	// Assert that offset is kept, even when it's after where the gap was created
	offset, err := file.Seek(0, io.SeekCurrent)
	require.NoError(t, err)
	assert.Equal(t, offset, int64(12))

	// Write in the gap
	_, err = file.WriteAt([]byte(", "), 5)
	require.NoError(t, err)

	// Assert its contents after sync
	require.NoError(t, file.Close())
	actualContents, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, []byte("Hello, World"), actualContents)
}
