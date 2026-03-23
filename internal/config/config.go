package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	Environment string
	OrgSlug     string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Environment: getEnv("ENVIRONMENT", "development"),
		OrgSlug:     getEnv("ORG_SLUG", "demo"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
