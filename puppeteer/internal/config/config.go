package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type Config struct {
	DatabaseURL              string
	HTTPAddress              string
	PollInterval             time.Duration
	InternalServiceToken     string
	TaskDefinitionsConfig    string
	LeaderboardRewardsConfig string
}

func Load() (Config, error) {
	pollInterval, err := durationEnv("PUPPETEER_POLL_INTERVAL", time.Minute)

	if err != nil {
		return Config{}, err
	}

	if pollInterval < time.Second {
		return Config{}, errors.New("PUPPETEER_POLL_INTERVAL must be at least 1s")
	}

	cfg := Config{
		DatabaseURL:              os.Getenv("DATABASE_URL"),
		HTTPAddress:              stringEnv("PUPPETEER_HTTP_ADDRESS", ":8091"),
		PollInterval:             pollInterval,
		InternalServiceToken:     os.Getenv("INTERNAL_SERVICE_TOKEN"),
		TaskDefinitionsConfig:    stringEnv("TASK_DEFINITIONS_CONFIG", "../config/task_definitions.yaml"),
		LeaderboardRewardsConfig: stringEnv("LEADERBOARD_REWARDS_CONFIG", "../config/leaderboard_rewards.yaml"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	if len(cfg.InternalServiceToken) < 32 {
		return Config{}, errors.New("INTERNAL_SERVICE_TOKEN must be at least 32 characters")
	}

	return cfg, nil
}

func stringEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)

	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)

	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}

	return parsed, nil
}
