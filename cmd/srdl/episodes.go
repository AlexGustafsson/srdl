package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/AlexGustafsson/srdl/internal/sr"
)

func episodes(args []string) error {
	commandLine := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	programID := commandLine.Int("program-id", 0, "Program ID")
	page := commandLine.Int("page", 1, "Page number")
	pageSize := commandLine.Int("page-size", 30, "Page size")
	commandLine.Usage = printUsage
	commandLine.Parse(args)

	if *programID == 0 {
		commandLine.Usage()
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	listOptions := &sr.ListEpisodesInProgramOptions{
		Page:     *page,
		PageSize: *pageSize,
	}
	episodes, err := sr.DefaultClient.ListEpisodesInProgram(ctx, *programID, listOptions)
	if err == sr.ErrNotFound {
		fmt.Fprintf(os.Stderr, "Program not found")
		return err
	} else if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	return encoder.Encode(episodes)
}
