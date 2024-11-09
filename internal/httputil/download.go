package httputil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	urlpkg "net/url"
	pathpkg "path"
)

// Download returns a reader for the file at url.
// It is the caller's responsibility to close the returned reader.
func Download(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return res.Body, nil
}

// DownloadIfNotExist writes the resource at url to the file at path if it does
// not already exist. If no extension is specified in path, the extension will
// be modified to mirror that of the resource at url.
func DownloadIfNotExist(ctx context.Context, path string, url string) error {
	if pathpkg.Ext(path) == "" {
		u, err := urlpkg.Parse(url)
		if err != nil {
			return err
		}

		path += pathpkg.Ext(u.Path)
	}

	stat, err := os.Stat(path)
	if err == nil && stat.Size() > 0 {
		return nil
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := Download(ctx, url)
	if err != nil {
		return err
	}
	defer content.Close()

	if _, err := io.Copy(file, content); err != nil {
		return err
	}

	return nil
}
