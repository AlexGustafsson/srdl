package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))

	configFilePath := flag.String("config", "", "Config file path")
	subscriptionsFilePath := flag.String("subscriptions", "", "Subscriptions file path")

	flag.Parse()

	if *configFilePath == "" || *subscriptionsFilePath == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*configFilePath, *subscriptionsFilePath); err != nil {
		os.Exit(1)
	}
}

func run(configFilePath string, subscriptionsFilePath string) error {
	var config Config
	if err := readYamlFromFile(configFilePath, &config); err != nil {
		return err
	}

	var subscriptions map[string]Subscription
	if err := readYamlFromFile(subscriptionsFilePath, &subscriptions); err != nil {
		return err
	}

	for subscriptionID, subscription := range subscriptions {
		log := slog.With(slog.String("subscription", subscriptionID), slog.Int("programId", subscription.ProgramID))
		log.Info("Processing subscription")

		appliedConfig := Preset{
			Output: config.Output,
		}
		for _, presetName := range subscription.Presets {
			preset, ok := config.Presets[presetName]
			if !ok {
				log.Error("No such preset", slog.String("preset", presetName))
				return fmt.Errorf("preset not found")
			}

			appliedConfig = appliedConfig.Apply(preset)
		}

		log.Debug("Resolved config", slog.Any("appliedConfig", appliedConfig))

		if err := processProgram(context.Background(), subscription, appliedConfig, log); err != nil {
			log.Warn("Failed to process program")
			continue
		}
	}

	return nil
}
