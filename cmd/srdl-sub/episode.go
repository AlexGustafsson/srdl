package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/AlexGustafsson/srdl/internal/mp4"
	"github.com/AlexGustafsson/srdl/internal/sr"
)

func processEpisode(ctx context.Context, episode sr.Episode, subscription Subscription, config Preset, log *slog.Logger) error {
	log = log.With(slog.Int("episode", episode.ID))
	log.Debug("Processing episode")

	if config.Retention > 0 && time.Since(episode.PublishDate.Time) > config.DownloadRange {
		log.Debug("Skipping old episode", slog.Time("publishDate", episode.PublishDate.Time))
		return nil
	}

	if episode.Broadcast == nil {
		log.Warn("No broadcast available for the episode")
		return fmt.Errorf("no broadcast")
	}

	if len(episode.Broadcast.Files) == 0 {
		log.Warn("No files available for any broadcast of the episode")
		return fmt.Errorf("no broadcast files")
	}

	if config.Throttling.DownloadDelay > 0 {
		log.Debug("Waiting before proceeding with download", slog.Duration("delay", config.Throttling.DownloadDelay))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(config.Throttling.DownloadDelay):
		}
	}

	outputPath := path.Join(config.Output, subscription.Artist, subscription.Album, episode.Title+".m4a")
	log = log.With("outputPath", outputPath)

	_, err := os.Stat(outputPath)
	if err == nil {
		log.Debug("Skipping epsiode that is already downloaded")
		return nil
	} else if err != nil && !os.IsNotExist(err) {
		log.Error("Failed to identify if the episode is already downloaded", slog.Any("error", err))
		return err
	}

	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Error("Failed to create output file", slog.Any("error", err))
		defer file.Close()
	}

	episodeFile, err := download(ctx, episode.Broadcast.Files[0].URL)
	if err != nil {
		log.Error("Failed to download file", slog.Any("error", err))
		return err
	}
	defer episodeFile.Close()

	if _, err := io.Copy(file, episodeFile); err != nil {
		log.Error("Failed to download file", slog.Any("error", err))
		return err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Warn("Failed to process metadata", slog.Any("error", err))
		// Ignore the error as it's not critical
		return nil
	}

	meta := mp4.Metadata{
		Title:       episode.Title,
		Album:       episode.Program.Name,
		Description: episode.Description,
		Released:    episode.PublishDate.Time,
	}

	if err := meta.Write(file); err != nil {
		log.Warn("Failed to write metadata", slog.Any("error", err))
		// Ignore the error as it's not critical
		return nil
	}

	return nil
}
