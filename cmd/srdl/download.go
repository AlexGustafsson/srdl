package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexGustafsson/srdl/internal/httputil"
	"github.com/AlexGustafsson/srdl/internal/mp4"
	"github.com/AlexGustafsson/srdl/internal/sr"
)

func download(args []string) error {
	commandLine := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	episodeID := commandLine.Int("episode-id", 0, "Episode ID")
	output := commandLine.String("output", "", "Optional output file path")
	commandLine.Usage = printUsage
	commandLine.Parse(args)

	if *episodeID == 0 {
		commandLine.Usage()
		os.Exit(1)
	}

	episode, err := sr.DefaultClient.GetEpisode(context.Background(), *episodeID)
	if err != nil {
		return fmt.Errorf("failed to get episode: %w", err)
	}

	if *output == "" {
		*output = episode.Title + ".m4a"
	}

	if episode.Broadcast == nil {
		return fmt.Errorf("no broadcast available for the episode")
	}

	// TODO: Check with melodikrysset if there's multiple episodes
	if len(episode.Broadcast.Files) == 0 {
		return fmt.Errorf("no files available for the episode")
	}

	program, err := sr.DefaultClient.GetProgram(context.Background(), episode.Program.ID)
	if err != nil {
		return fmt.Errorf("failed to get program: %w", err)
	}

	episodeFile, err := httputil.Download(context.Background(), episode.Broadcast.Files[0].URL)
	if err != nil {
		return fmt.Errorf("Failed to download file: %w", err)
	}
	defer episodeFile.Close()

	file, err := os.OpenFile(*output, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}

	if _, err := io.Copy(file, episodeFile); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		slog.Warn("Failed to process metadata", slog.Any("error", err))
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
		slog.Warn("Failed to write metadata", slog.Any("error", err))
		// Ignore the error as it's not critical
	}

	err = httputil.DownloadIfNotExist(context.Background(), filepath.Join(filepath.Dir(*output), "cover"), program.ImageURL)
	if err != nil {
		slog.Warn("Failed to download cover image", slog.Any("error", err))
		// Ignore the error as it's not critical
	}

	err = httputil.DownloadIfNotExist(context.Background(), filepath.Join(filepath.Dir(*output), "backdrop"), program.ImageTemplateWideURL)
	if err != nil {
		slog.Warn("Failed to download backdrop image", slog.Any("error", err))
		// Ignore the error as it's not critical
	}

	err = httputil.DownloadIfNotExist(context.Background(), filepath.Join(filepath.Dir(*output), "backdrop"), program.ImageTemplateWideURL)
	if err != nil {
		slog.Warn("Failed to download episode image", slog.Any("error", err))
		// Ignore the error as it's not critical
	}

	episodeImagePath := filepath.Join(filepath.Dir(*output), strings.TrimSuffix(filepath.Base(*output), filepath.Ext(*output)))
	if err := httputil.DownloadIfNotExist(context.Background(), episodeImagePath, episode.ImageURL); err != nil {
		slog.Warn("Failed to download episode image", slog.Any("error", err))
		// Fallthrough
	}

	return nil
}
