// Package config loads application configuration from environment variables
// with sane defaults, so the service runs locally with zero setup and stays
// 12-factor friendly in containers.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppName         string
	Env             string // local | staging | production
	HTTPPort        int
	DatabaseURL     string
	ShutdownTimeout time.Duration
	RateLimitPerSec float64
}

func Load() (Config, error) {
	cfg := Config{
		AppName:         getenv("APP_NAME", "go-service-kit"),
		Env:             getenv("APP_ENV", "local"),
		DatabaseURL:     getenv("DATABASE_URL", "postgres://app:app@localhost:5432/app?sslmode=disable"),
		ShutdownTimeout: 10 * time.Second,
	}

	port, err := strconv.Atoi(getenv("HTTP_PORT", "8080"))
	if err != nil {
		return Config{}, fmt.Errorf("config: invalid HTTP_PORT: %w", err)
	}
	cfg.HTTPPort = port

	rate, err := strconv.ParseFloat(getenv("RATE_LIMIT_PER_SEC", "20"), 64)
	if err != nil {
		return Config{}, fmt.Errorf("config: invalid RATE_LIMIT_PER_SEC: %w", err)
	}
	cfg.RateLimitPerSec = rate

	return cfg, nil
}

func (c Config) IsLocal() bool { return c.Env == "local" }

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
