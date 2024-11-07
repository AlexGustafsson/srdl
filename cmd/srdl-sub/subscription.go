package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

func processSubscription(ctx context.Context, config Config, subscription Subscription, log *slog.Logger) error {
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

	log.Debug("Waiting before proceeding with processing subscription", slog.Duration("delay", appliedConfig.Throttling.SubscriptionDelay))
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(appliedConfig.Throttling.SubscriptionDelay):
	}

	log.Info("Processing subscription")

	if err := processProgram(ctx, subscription, appliedConfig, log); err != nil {
		if err != ctx.Err() {
			log.Error("Failed to process program", slog.Any("error", err))
		}
		return err
	}

	return nil
}
