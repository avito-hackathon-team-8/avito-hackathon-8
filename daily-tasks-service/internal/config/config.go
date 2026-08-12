package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/daily-tasks-service/internal/database"
)

const LeaderboardInterval = 10 * time.Minute

type Config struct {
	DatabaseURL              string
	HTTPAddress              string
	ReadHeaderTimeout        time.Duration
	ReadTimeout              time.Duration
	WriteTimeout             time.Duration
	IdleTimeout              time.Duration
	PollInterval             time.Duration
	InternalServiceToken     string
	TaskDefinitionsConfig    string
	LeaderboardRewardsConfig string
	DatabasePool             database.PoolConfig
}

func Load() (Config, error) {
	pollInterval, err := durationEnv("DAILY_TASKS_POLL_INTERVAL", LeaderboardInterval)

	if err != nil {
		return Config{}, err
	}

	if pollInterval < time.Second {
		return Config{}, errors.New("DAILY_TASKS_POLL_INTERVAL must be at least 1s")
	}

	maxOpen, err := intEnv("DB_MAX_OPEN_CONNS", 25)

	if err != nil {
		return Config{}, err
	}

	maxIdle, err := intEnv("DB_MAX_IDLE_CONNS", 10)

	if err != nil {
		return Config{}, err
	}

	lifetime, err := durationEnv("DB_CONN_MAX_LIFETIME", 30*time.Minute)

	if err != nil {
		return Config{}, err
	}

	idleTime, err := durationEnv("DB_CONN_MAX_IDLE_TIME", 5*time.Minute)

	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		DatabaseURL:              os.Getenv("DATABASE_URL"),
		HTTPAddress:              stringEnv("DAILY_TASKS_HTTP_ADDRESS", ":8091"),
		ReadHeaderTimeout:        5 * time.Second,
		ReadTimeout:              15 * time.Second,
		WriteTimeout:             30 * time.Second,
		IdleTimeout:              60 * time.Second,
		PollInterval:             pollInterval,
		InternalServiceToken:     os.Getenv("INTERNAL_SERVICE_TOKEN"),
		TaskDefinitionsConfig:    stringEnv("TASK_DEFINITIONS_CONFIG", "../config/task_definitions.yaml"),
		LeaderboardRewardsConfig: stringEnv("LEADERBOARD_REWARDS_CONFIG", "../config/leaderboard_rewards.yaml"),
		DatabasePool:             database.PoolConfig{MaxOpenConns: maxOpen, MaxIdleConns: maxIdle, ConnMaxLifetime: lifetime, ConnMaxIdleTime: idleTime},
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	if len(cfg.InternalServiceToken) < 32 {
		return Config{}, errors.New("INTERNAL_SERVICE_TOKEN must be at least 32 characters")
	}

	if cfg.DatabasePool.MaxOpenConns < 1 || cfg.DatabasePool.MaxIdleConns < 0 || cfg.DatabasePool.MaxIdleConns > cfg.DatabasePool.MaxOpenConns {
		return Config{}, errors.New("database pool limits are invalid")
	}

	return cfg, nil
}

func intEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)

	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)

	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}

	return parsed, nil
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
