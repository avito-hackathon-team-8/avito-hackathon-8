package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/email"
)

type Config struct {
	HTTPAddress          string
	ReadHeaderTimeout    time.Duration
	DatabaseURL          string
	Auth                 auth.Config
	Email                email.Config
	InternalServiceToken string
	PuppeteerInternalURL string
}

func Load() (Config, error) {
	otpLength, err := envInt("OTP_LENGTH", 8)

	if err != nil {
		return Config{}, err
	}

	sessionTTL, err := envDuration("JWT_TTL", 24*time.Hour)

	if err != nil {
		return Config{}, err
	}

	otpTTL, err := envDuration("OTP_TTL", 5*time.Minute)

	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddress:       env("HTTP_ADDRESS", ":8090"),
		ReadHeaderTimeout: 5 * time.Second,
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		Auth: auth.Config{
			JWTSecret:  os.Getenv("JWT_SECRET"),
			SessionTTL: sessionTTL,
			OTPTTL:     otpTTL,
			OTPLength:  otpLength,
		},
		Email: email.Config{
			Mode:     env("EMAIL_MODE", "smtp"),
			From:     os.Getenv("EMAIL_FROM"),
			Host:     os.Getenv("SMTP_HOST"),
			Port:     env("SMTP_PORT", "587"),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
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

	if cfg.Auth.OTPLength < 6 || cfg.Auth.OTPLength > 10 {
		return Config{}, errors.New("OTP_LENGTH must be between 6 and 10")
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

func envInt(key string, fallback int) (int, error) {
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
