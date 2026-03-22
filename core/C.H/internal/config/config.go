package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port            string
	DatabaseURL     string
	SessionTTLHours int
	PasswordPepper  string
	AllowedOrigin   string
}

func Load() (Config, error) {
	cfg := Config{
		Port:           getOrDefault("PORT", "8088"),
		DatabaseURL:    os.Getenv("AMD_DATABASE_URL"),
		PasswordPepper: getOrDefault("PASSWORD_PEPPER", "dev-pepper"),
		AllowedOrigin:  getOrDefault("ALLOWED_ORIGIN", "http://localhost:3000"),
	}

	ttlRaw := getOrDefault("SESSION_TTL_HOURS", "168")
	ttl, err := strconv.Atoi(ttlRaw)
	if err != nil || ttl <= 0 {
		return Config{}, fmt.Errorf("invalid SESSION_TTL_HOURS: %s", ttlRaw)
	}
	cfg.SessionTTLHours = ttl

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("AMD_DATABASE_URL is required")
	}

	return cfg, nil
}

func getOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
