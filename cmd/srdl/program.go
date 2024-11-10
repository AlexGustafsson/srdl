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

func program(args []string) error {
	commandLine := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	commandLine.Usage = printUsage
	commandLine.Parse(args)

	url := commandLine.Arg(0)
	if url == "" {
		commandLine.Usage()
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	programID, err := sr.DefaultClient.GetProgramID(ctx, url)
	if err == sr.ErrNotFound {
		fmt.Fprintf(os.Stderr, "Program page not found")
		return err
	} else if err == sr.ErrProgramIDNotFound {
		fmt.Fprintf(os.Stderr, "Cannot identify program id")
		return err
	} else if err != nil {
		return err
	}

	program, err := sr.DefaultClient.GetProgram(ctx, programID)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	return encoder.Encode(program)
}
