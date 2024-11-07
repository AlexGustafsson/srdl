package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
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

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)

		caught := 0

		for {
			select {
			case <-ctx.Done():
				close(sigint)
				return
			case <-sigint:
				caught++
				if caught == 1 {
					slog.Info("Caught signal, exiting gracefully")
					cancel()
				} else {
					slog.Info("Caught signal, exiting now")
					os.Exit(1)
				}
			}
		}
	}()

	if err := run(ctx, *configFilePath, *subscriptionsFilePath); err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context, configFilePath string, subscriptionsFilePath string) error {
	var config Config
	if err := readYamlFromFile(configFilePath, &config); err != nil {
		return err
	}

	var subscriptions map[string]Subscription
	if err := readYamlFromFile(subscriptionsFilePath, &subscriptions); err != nil {
		return err
	}

	// NOTE: Although all of the requests could be made parallel, let's keep them
	// synchronous as it acts as a natural rate limit to make sure the load is
	// fair
	for subscriptionID, subscription := range subscriptions {
		if err := ctx.Err(); err != nil {
			return err
		}

		log := slog.With(slog.String("subscription", subscriptionID), slog.Int("programId", subscription.ProgramID))
		if err := processSubscription(ctx, config, subscription, log); err != nil {
			if err != ctx.Err() {
				log.Error("Failed to process subscription", slog.Any("error", err))
			}
			continue
		}
	}

	return nil
}
