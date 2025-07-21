package main

import (
	"fmt"
	"os"
)

const usageTemplate = `usage: %[1]s <command> [options...] [args...]

commands:
- program
- download

examples:

%[1]s program <url>
%[1]s episodes -program-id 1234
%[1]s download -output file -episode-id 1234
`

func printUsage() {
	fmt.Fprintf(os.Stderr, usageTemplate, os.Args[0])
}

func main() {
	if len(os.Args) == 1 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	if command == "" {
		printUsage()
		os.Exit(1)
	}

	var err error
	switch command {
	case "program":
		err = program(os.Args[2:])
	case "episodes":
		err = episodes(os.Args[2:])
	case "download":
		err = download(os.Args[2:])
	default:
		err = fmt.Errorf("invalid command: %s", command)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
