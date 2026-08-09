package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
)

type Config struct {
	HTTPAddress          string
	ReadHeaderTimeout    time.Duration
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	IdleTimeout          time.Duration
	DatabaseURL          string
	Auth                 auth.Config
	InternalServiceToken string
	PuppeteerInternalURL string
}

func Load() (Config, error) {
	sessionTTL, err := envDuration("JWT_TTL", 24*time.Hour)

	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddress:       env("HTTP_ADDRESS", ":8090"),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		Auth: auth.Config{
			JWTSecret:  os.Getenv("JWT_SECRET"),
			SessionTTL: sessionTTL,
		},
		InternalServiceToken: os.Getenv("INTERNAL_SERVICE_TOKEN"),
		PuppeteerInternalURL: env("PUPPETEER_INTERNAL_URL", "http://puppeteer:8091"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	if len(cfg.Auth.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must be at least 32 characters")
	}

	if len(cfg.InternalServiceToken) < 32 {
		return Config{}, errors.New("INTERNAL_SERVICE_TOKEN must be at least 32 characters")
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
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
