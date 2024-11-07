package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Output  string            `yaml:"output"`
	Presets map[string]Preset `yaml:"presets"`
}

type Preset struct {
	Output        string        `yaml:"output"`
	DownloadRange time.Duration `yaml:"downloadRange"`
	Retention     time.Duration `yaml:"retention"`
	Throttling    Throttling    `yaml:"throttling"`
}

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

type Throttling struct {
	DownloadDelay          time.Duration `yaml:"perDownload"`
	EpisodeDelay           time.Duration `yaml:"perEpisode"`
	SubscriptionDelay      time.Duration `yaml:"perSubscription"`
	MaxDownloadsPerProgram int           `yaml:"maxDownloadsPerProgram"`
}

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

type Subscription struct {
	ProgramID int      `yaml:"programId"`
	Artist    string   `yaml:"artist"`
	Album     string   `yaml:"album"`
	Presets   []string `yaml:"presets"`
}

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
