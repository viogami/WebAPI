package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
		DatabaseURL:    getFirstEnv("CH_API_DATABASE_URL", "AMD_DATABASE_URL"),
		PasswordPepper: getFirstEnv("CH_API_PASSWORD_PEPPER", "PASSWORD_PEPPER"),
		AllowedOrigin:  getFirstEnvOrDefault("http://localhost:3000", "CH_API_ALLOWED_ORIGIN", "ALLOWED_ORIGIN"),
	}

	ttlRaw := getOrDefault("SESSION_TTL_HOURS", "168")
	ttl, err := strconv.Atoi(ttlRaw)
	if err != nil || ttl <= 0 {
		return Config{}, fmt.Errorf("invalid SESSION_TTL_HOURS: %s", ttlRaw)
	}
	cfg.SessionTTLHours = ttl

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("CH_API_DATABASE_URL is required")
	}
	if cfg.PasswordPepper == "" {
		return Config{}, fmt.Errorf("CH_API_PASSWORD_PEPPER is required")
	}

	return cfg, nil
}

func getOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getFirstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func getFirstEnvOrDefault(fallback string, keys ...string) string {
	if value := getFirstEnv(keys...); value != "" {
		return value
	}
	return fallback
}
