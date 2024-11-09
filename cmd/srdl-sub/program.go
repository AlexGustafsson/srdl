package main

import (
	"context"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/AlexGustafsson/srdl/internal/fsutil"
	"github.com/AlexGustafsson/srdl/internal/httputil"
	"github.com/AlexGustafsson/srdl/internal/sr"
)

// processProgram processes a single program.
func processProgram(ctx context.Context, subscription Subscription, config Preset, log *slog.Logger) error {
	log.Debug("Processing program")

	outputPath := path.Join(config.Output, subscription.Artist, subscription.Album)
	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
		return err
	}
	log = log.With(slog.String("programOutput", outputPath))

	program, err := sr.DefaultClient.GetProgram(ctx, subscription.ProgramID)
	if err != nil {
		log.Error("Failed to get program", slog.Any("error", err))
		return err
	}

	// TODO: Paginate through all episodes?
	result, err := sr.DefaultClient.ListEpisodesInProgram(ctx, subscription.ProgramID, nil)
	if err != nil {
		log.Error("Failed to list episodes in program", slog.Any("error", err))
		return err
	}

	processed := 0
	for _, episode := range result.Episodes {
		if err := ctx.Err(); err != nil {
			return err
		}

		if processed > config.Throttling.MaxDownloadsPerProgram {
			log.Debug("Skipping further processing as it would exceed maximum downloads per program")
			break
		}

		if config.Throttling.EpisodeDelay > 0 {
			log.Debug("Waiting before proceeding with processing episode", slog.Duration("delay", config.Throttling.EpisodeDelay))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(config.Throttling.EpisodeDelay):
			}
		}

		if err := processEpisode(ctx, episode, subscription, config, log); err != nil {
			if err != ctx.Err() {
				log.Error("Failed to process episode", slog.Any("error", err))
			}
			continue
		}
		processed++
	}

	if err := httputil.DownloadIfNotExist(ctx, path.Join(outputPath, "cover"), program.ImageURL); err != nil {
		log.Warn("Failed to download cover image", slog.Any("error", err))
		// Fallthrough
	}

	if err := httputil.DownloadIfNotExist(ctx, path.Join(outputPath, "backdrop"), program.ImageTemplateWideURL); err != nil {
		log.Warn("Failed to download backdrop image", slog.Any("error", err))
		// Fallthrough
	}

	// Try to remove old files
	if config.Retention > 0 {
		maxAge := time.Now().Add(-config.Retention)
		log.Debug("Removing old files", slog.Time("maxAge", maxAge))
		if err := fsutil.RemoveOldFiles(outputPath, maxAge); err != nil {
			log.Warn("Failed to clean up old files", slog.Any("error", err))
			// Fallthrough
		}
	}

	// Try to remove empty directories
	log.Debug("Cleaning up empty directories")
	if err := fsutil.RemoveEmptyDirectories(outputPath); err != nil {
		log.Warn("Failed to clean empty directories", slog.Any("error", err))
		// Fallthrough
	}

	return nil
}
