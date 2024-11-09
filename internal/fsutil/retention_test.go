package fsutil

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveOldFiles(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, createDir(path.Join(root, "new"), time.Now()))
	require.NoError(t, createFile(path.Join(root, "new", "new1"), time.Now()))
	require.NoError(t, createFile(path.Join(root, "new", "new2"), time.Now()))
	require.NoError(t, createFile(path.Join(root, "new", "old1"), time.Now().Add(-2*time.Hour)))
	require.NoError(t, createFile(path.Join(root, "new", "old2"), time.Now().Add(-2*time.Hour)))

	require.NoError(t, createDir(path.Join(root, "old"), time.Now()))
	require.NoError(t, createFile(path.Join(root, "old", "new1"), time.Now()))
	require.NoError(t, createFile(path.Join(root, "old", "new2"), time.Now()))
	require.NoError(t, createFile(path.Join(root, "old", "old1"), time.Now().Add(-2*time.Hour)))
	require.NoError(t, createFile(path.Join(root, "old", "old2"), time.Now().Add(-2*time.Hour)))

	require.NoError(t, createDir(path.Join(root, "old2"), time.Now()))

	beforeDelete, err := tree(root)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"new",
		"new/new1",
		"new/new2",
		"new/old1",
		"new/old2",
		"old",
		"old/new1",
		"old/new2",
		"old/old1",
		"old/old2",
		"old2",
	}, beforeDelete)

	require.NoError(t, RemoveOldFiles(root, time.Now().Add(-1*time.Hour)))

	afterDelete, err := tree(root)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"new",
		"new/new1",
		"new/new2",
		"old",
		"old/new1",
		"old/new2",
		"old2",
	}, afterDelete)
}

func TestRemoveEmptyDirectories(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, createDir(path.Join(root, "new"), time.Now()))
	require.NoError(t, createFile(path.Join(root, "new", "new1"), time.Now()))
	require.NoError(t, createFile(path.Join(root, "new", "new2"), time.Now()))
	require.NoError(t, createFile(path.Join(root, "new", "old1"), time.Now().Add(-2*time.Hour)))
	require.NoError(t, createFile(path.Join(root, "new", "old2"), time.Now().Add(-2*time.Hour)))

	require.NoError(t, createDir(path.Join(root, "old"), time.Now()))
	require.NoError(t, createFile(path.Join(root, "old", "new1"), time.Now()))
	require.NoError(t, createFile(path.Join(root, "old", "new2"), time.Now()))
	require.NoError(t, createFile(path.Join(root, "old", "old1"), time.Now().Add(-2*time.Hour)))
	require.NoError(t, createFile(path.Join(root, "old", "old2"), time.Now().Add(-2*time.Hour)))

	require.NoError(t, createDir(path.Join(root, "old2"), time.Now()))

	beforeDelete, err := tree(root)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"new",
		"new/new1",
		"new/new2",
		"new/old1",
		"new/old2",
		"old",
		"old/new1",
		"old/new2",
		"old/old1",
		"old/old2",
		"old2",
	}, beforeDelete)

	require.NoError(t, RemoveEmptyDirectories(root))

	afterDelete, err := tree(root)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"new",
		"new/new1",
		"new/new2",
		"new/old1",
		"new/old2",
		"old",
		"old/new1",
		"old/new2",
		"old/old1",
		"old/old2",
	}, afterDelete)
}

func tree(root string) ([]string, error) {
	entries := make([]string, 0)
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if p == root {
			return nil
		}

		entries = append(entries, strings.TrimPrefix(p, root+"/"))
		return nil
	})

	sort.Strings(entries)

	return entries, nil
}

func createFile(path string, modTime time.Time) error {
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		return err
	}

	return os.Chtimes(path, time.Now(), modTime)
}

func createDir(path string, modTime time.Time) error {
	if err := os.Mkdir(path, os.ModePerm); err != nil {
		return err
	}

	return os.Chtimes(path, time.Now(), modTime)
}
