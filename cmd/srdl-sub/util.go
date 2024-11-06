package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

func download(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return res.Body, nil
}

func downloadIfNotExist(ctx context.Context, path string, url string) error {
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

	content, err := download(ctx, url)
	if err != nil {
		return err
	}
	defer content.Close()

	if _, err := io.Copy(file, content); err != nil {
		return err
	}

	return nil
}
