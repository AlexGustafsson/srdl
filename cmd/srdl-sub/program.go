package main

import (
	"context"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/AlexGustafsson/srdl/internal/sr"
)

func processProgram(ctx context.Context, subscription Subscription, config Preset, log *slog.Logger) error {
	log.Debug("Processing program")

	outputPath := path.Join(config.Output, subscription.Artist, subscription.Album)
	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
		return err
	}

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
		if processed > config.Throttling.MaxDownloadsPerProgram {
			log.Debug("Skipping further processing as it would exceed maximum downloads per program")
			break
		}

		if config.Throttling.EpisodeDelay > 0 {
			log.Debug("Waiting before proceeding with processing episode", slog.Duration("delay", config.Throttling.EpisodeDelay))
			time.Sleep(config.Throttling.EpisodeDelay)
		}

		processEpisode(ctx, episode, subscription, config, log)
		processed++
	}

	if err := downloadIfNotExist(ctx, path.Join(outputPath, "cover.png"), program.ImageURL); err != nil {
		log.Warn("Failed to download cover image", slog.Any("error", err))
		// Fallthrough
	}

	if err := downloadIfNotExist(ctx, path.Join(outputPath, "backdrop.png"), program.ImageTemplateWideURL); err != nil {
		log.Warn("Failed to download backdrop image", slog.Any("error", err))
		// Fallthrough
	}

	// TODO: Delete old files according to config.Retention

	return nil
}
