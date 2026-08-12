package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/database"
)

type Config struct {
	HTTPAddress           string
	ReadHeaderTimeout     time.Duration
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	IdleTimeout           time.Duration
	DatabaseURL           string
	Auth                  auth.Config
	InternalServiceToken  string
	DailyTasksInternalURL string
	LevelRewardsConfig    string
	ShopItemsConfig       string
	ShopImagesDir         string
	PetStateInternalURL   string
	KafkaBrokers          []string
	PetStateKafkaTopic    string
	APIServiceInstanceID  string
	DatabasePool          database.PoolConfig
}

func Load() (Config, error) {
	sessionTTL, err := envDuration("JWT_TTL", 24*time.Hour)

	if err != nil {
		return Config{}, err
	}
	pool, err := databasePool()
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
		InternalServiceToken:  os.Getenv("INTERNAL_SERVICE_TOKEN"),
		DailyTasksInternalURL: env("DAILY_TASKS_INTERNAL_URL", "http://daily-tasks-service:8091"),
		LevelRewardsConfig:    env("LEVEL_REWARDS_CONFIG", "../config/level_rewards.yaml"),
		ShopItemsConfig:       env("SHOP_ITEMS_CONFIG", "../config/shop_items.yaml"),
		ShopImagesDir:         env("SHOP_IMAGES_DIR", "../shop-images"),
		PetStateInternalURL:   env("PET_STATE_INTERNAL_URL", "http://pet-state-service:8092"),
		KafkaBrokers:          strings.Split(env("KAFKA_BROKERS", "kafka:9092"), ","),
		PetStateKafkaTopic:    env("PET_STATE_KAFKA_TOPIC", "pet-state-events-v1"),
		APIServiceInstanceID:  env("API_SERVICE_INSTANCE_ID", hostname()),
		DatabasePool:          pool,
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

func databasePool() (database.PoolConfig, error) {
	maxOpen, err := envInt("DB_MAX_OPEN_CONNS", 25)

	if err != nil {
		return database.PoolConfig{}, err
	}

	maxIdle, err := envInt("DB_MAX_IDLE_CONNS", 10)

	if err != nil {
		return database.PoolConfig{}, err
	}

	lifetime, err := envDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute)

	if err != nil {
		return database.PoolConfig{}, err
	}

	idleTime, err := envDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute)

	if err != nil {
		return database.PoolConfig{}, err
	}

	if maxOpen < 1 || maxIdle < 0 || maxIdle > maxOpen {
		return database.PoolConfig{}, errors.New("database pool limits are invalid")
	}

	return database.PoolConfig{MaxOpenConns: maxOpen, MaxIdleConns: maxIdle, ConnMaxLifetime: lifetime, ConnMaxIdleTime: idleTime}, nil
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

func hostname() string {
	value, err := os.Hostname()

	if err != nil || value == "" {
		return "local"
	}

	return value
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
