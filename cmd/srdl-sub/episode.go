package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/AlexGustafsson/srdl/internal/httputil"
	"github.com/AlexGustafsson/srdl/internal/mp4"
	"github.com/AlexGustafsson/srdl/internal/sr"
)

// processEpisode processes a single episode.
// Returns whether or not the episode was downloaded (since episodes can be
// processed but not downloaded if they're already downloaded).
func processEpisode(ctx context.Context, episode sr.Episode, config Preset, outputPath string, log *slog.Logger) (bool, error) {
	log = log.With(slog.Int("episode", episode.ID))
	log.Debug("Processing episode")

	if config.Retention > 0 && time.Since(episode.PublishDate.Time) > config.DownloadRange {
		log.Debug("Skipping old episode", slog.Time("publishDate", episode.PublishDate.Time))
		return false, nil
	}

	if episode.Broadcast == nil {
		log.Warn("No broadcast available for the episode")
		return false, fmt.Errorf("no broadcast")
	}

	if len(episode.Broadcast.Files) == 0 {
		log.Warn("No files available for any broadcast of the episode")
		return false, fmt.Errorf("no broadcast files")
	}

	outputPath = filepath.Join(outputPath, episode.Title+".m4a")
	log = log.With("outputPath", outputPath)

	// Try to download the episode's image
	if err := httputil.DownloadIfNotExist(ctx, filepath.Join(outputPath, episode.Title), episode.ImageURL); err != nil {
		log.Warn("Failed to download episode image", slog.Any("error", err))
		// Fallthrough
	}

	// Check if episode audio file already exists
	_, err := os.Stat(outputPath)
	if err == nil {
		log.Debug("Skipping episode that is already downloaded")
		return false, nil
	} else if !os.IsNotExist(err) {
		log.Error("Failed to identify if the episode is already downloaded", slog.Any("error", err))
		return false, err
	}

	if config.Throttling.DownloadDelay > 0 {
		log.Debug("Waiting before proceeding with download", slog.Duration("delay", config.Throttling.DownloadDelay))
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(config.Throttling.DownloadDelay):
		}
	}

	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Error("Failed to create output file", slog.Any("error", err))
		return false, err
	}
	defer file.Close()

	episodeFile, err := httputil.Download(ctx, episode.Broadcast.Files[0].URL)
	if err != nil {
		log.Error("Failed to download file", slog.Any("error", err))
		return false, err
	}
	defer episodeFile.Close()

	if _, err := io.Copy(file, episodeFile); err != nil {
		log.Error("Failed to download file", slog.Any("error", err))
		return false, err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Warn("Failed to process metadata", slog.Any("error", err))
		// Ignore the error as it's not critical, but don't continue further
		return true, nil
	}

	meta := mp4.Metadata{
		Title:       episode.Title,
		Album:       episode.Program.Name,
		Description: episode.Description,
		Released:    episode.PublishDate.Time,
	}

	if err := meta.Write(file); err != nil {
		log.Warn("Failed to write metadata", slog.Any("error", err))
		// Ignore the error as it's not critical, but don't continue further
		return true, nil
	}

	return true, nil
}
