package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/AlexGustafsson/srdl/internal/sr"
)

func main() {
	episodeID := flag.Int("episode-id", 0, "ID of the episode to download")
	output := flag.String("output", "", "Output file")
	flag.Parse()

	episode, err := sr.DefaultClient.GetEpisode(context.Background(), *episodeID)
	if err != nil {
		slog.Error("Failed to get episode", slog.Any("error", err))
		os.Exit(1)
	}

	if *output == "" {
		*output = episode.Title + ".m4a"
	}

	if episode.Broadcast == nil {
		slog.Error("No broadcast available for the episode")
		os.Exit(1)
	}

	// TODO: Check with melodikrysset if there's multiple episodes
	if len(episode.Broadcast.Files) == 0 {
		slog.Error("No files available for the episode")
		os.Exit(1)
	}

	image, err := download(context.Background(), episode.ImageURL)
	if err == nil {
		defer image.Close()
	} else {
		slog.Warn("Failed to download image")
	}

	episodeFile, err := download(context.Background(), episode.Broadcast.Files[0].URL)
	if err != nil {
		slog.Error("Failed to download file", slog.Any("error", err))
		os.Exit(1)
	}
	defer episodeFile.Close()

	file, err := os.OpenFile(*output, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Failed to create output file", slog.Any("error", err))
		os.Exit(1)
	}

	if _, err := io.Copy(file, episodeFile); err != nil {
		slog.Error("Failed to download file", slog.Any("error", err))
		os.Exit(1)
	}
}

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
