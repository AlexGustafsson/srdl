package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

// Config contains global configuration.
type Config struct {
	// Output is the default path to the directory where srdl-sub will output its
	// files.
	Output string `yaml:"output"`
	// LogLevel is a a string representation of the log level to use.
	// Either debug, info, warn or error.
	LogLevel string `yaml:"logLevel"`
	// Presets maps presets by a unique id.
	Presets map[string]Preset `yaml:"presets"`
}

// SlogLogLevel returns the [slog.Level] that maps to the configured log level.
// If no value is set, [slog.LevelInfo] is returned.
func (c Config) SlogLogLevel() (slog.Level, error) {
	if c.LogLevel == "" {
		return slog.LevelInfo, nil
	}

	switch c.LogLevel {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid log level")
	}
}

// Preset defines a set of parameters influencing how a program is processed.
type Preset struct {
	// Output is the default path to the directory where srdl-sub will output its
	// files.
	Output string `yaml:"output"`
	// DownloadRange is the maximum age of epsiodes to consider for download.
	DownloadRange time.Duration `yaml:"downloadRange"`
	// Retention is the maximum age of files and directories in the output
	// directory before they are removed.
	Retention time.Duration `yaml:"retention"`
	// Throttling contains throttling configuration.
	Throttling Throttling `yaml:"throttling"`
}

// Apply returns a preset that is described by p and overridden by other.
func (p Preset) Apply(other Preset) Preset {
	if other.DownloadRange != 0 {
		p.DownloadRange = other.DownloadRange
	}

	if other.Retention != 0 {
		p.Retention = other.Retention
	}

	p.Throttling = p.Throttling.Apply(other.Throttling)

	return p
}

// Throttling contains throttling configuration.
type Throttling struct {
	// DownloadDelay is the delay before downloading an episode.
	DownloadDelay time.Duration `yaml:"perDownload"`
	// EpisodeDelay is the delay before processing an episode.
	EpisodeDelay time.Duration `yaml:"perEpisode"`
	// SubscriptionDelay is the delay before processing a subscription.
	SubscriptionDelay time.Duration `yaml:"perSubscription"`
	// MaxDownloadsPerProgram is the maxmimum number of downloads / episodes to
	// process per program.
	MaxDownloadsPerProgram int `yaml:"maxDownloadsPerProgram"`
}

// Apply returns a preset that is described by p and overridden by other.
func (t Throttling) Apply(other Throttling) Throttling {
	if other.MaxDownloadsPerProgram > 0 {
		t.MaxDownloadsPerProgram = other.MaxDownloadsPerProgram
	}

	if other.DownloadDelay > 0 {
		t.DownloadDelay = other.DownloadDelay
	}

	if other.EpisodeDelay > 0 {
		t.EpisodeDelay = other.EpisodeDelay
	}

	if other.SubscriptionDelay > 0 {
		t.SubscriptionDelay = other.SubscriptionDelay
	}

	return t
}

// Subscription contains configuration for the subscription of a specific
// program.
type Subscription struct {
	// ProgramID is the unique id of the program to subscribe to.
	ProgramID int `yaml:"programId"`
	// Artist is the name of the "artist" directory that is created in the
	// designated output directory.
	Artist string `yaml:"artist"`
	// Album is then name of the "album" directory that is created in the "artist"
	// directory.
	Album string `yaml:"album"`
	// Presets references all presets to use.
	Presets []string `yaml:"presets"`
}

// readYamlFromFile parses a YAML file from path into v.
func readYamlFromFile(path string, v any) error {
	file, err := os.Open(path)
	if err != nil {
		slog.Error("Failed to open file", slog.String("path", path), slog.Any("error", err))
		return err
	}

	defer file.Close()

	decoder := yaml.NewDecoder(file, yaml.Strict())
	if err := decoder.Decode(v); err != nil {
		slog.Error("Failed to parse file", slog.String("path", path), slog.Any("error", err))
		return err
	}

	return nil
}
