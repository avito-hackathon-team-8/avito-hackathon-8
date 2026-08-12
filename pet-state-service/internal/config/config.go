package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL          string
	HTTPAddress          string
	InternalServiceToken string
	KafkaBrokers         []string
	KafkaTopic           string
	DBMaxOpenConns       int
	DBMaxIdleConns       int
	DBConnMaxLifetime    time.Duration
	DBConnMaxIdleTime    time.Duration
}

func Load() (Config, error) {
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
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		HTTPAddress:          env("PET_STATE_HTTP_ADDRESS", ":8092"),
		InternalServiceToken: os.Getenv("INTERNAL_SERVICE_TOKEN"),
		KafkaBrokers:         strings.Split(env("KAFKA_BROKERS", "kafka:9092"), ","),
		KafkaTopic:           env("PET_STATE_KAFKA_TOPIC", "pet-state-events-v1"),
		DBMaxOpenConns:       maxOpen, DBMaxIdleConns: maxIdle, DBConnMaxLifetime: lifetime, DBConnMaxIdleTime: idleTime,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	if len(cfg.InternalServiceToken) < 32 {
		return Config{}, errors.New("INTERNAL_SERVICE_TOKEN must be at least 32 characters")
	}

	if cfg.DBMaxOpenConns < 1 || cfg.DBMaxIdleConns < 0 || cfg.DBMaxIdleConns > cfg.DBMaxOpenConns {
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

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
