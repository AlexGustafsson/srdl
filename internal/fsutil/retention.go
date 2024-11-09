package fsutil

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// RemoveOldFiles removes all files that have not been modified since maxAge.
func RemoveOldFiles(root string, maxAge time.Time) error {
	toRemove := make([]string, 0)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		info, err := d.Info()
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.ModTime().Compare(maxAge) > 0 {
			return nil
		}

		toRemove = append(toRemove, p)
		return nil
	})
	if err != nil {
		return err
	}

	for _, p := range toRemove {
		if err := os.Remove(p); err != nil {
			return err
		}
	}

	return err
}

// RemoveEmptyDirectories removes all empty directories under root.
func RemoveEmptyDirectories(root string) error {
	entries := make(map[string]int)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if p == root {
			return nil
		}

		if d.IsDir() {
			entries[p] = 0
		} else {
			entries[filepath.Dir(p)]++
		}

		return nil
	})
	if err != nil {
		return err
	}

	for dir, files := range entries {
		if files != 0 {
			continue
		}

		if err := os.Remove(dir); err != nil {
			return err
		}
	}

	return nil
}
