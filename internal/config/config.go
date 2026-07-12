package config

import (
	"os"
	"strings"
)

type Config struct {
	HTTPAddr      string
	SQLitePath    string
	LogDir        string
	WebRoot       string
	SeedDatabase  bool
	JWTSecret     string
	AdminName     string
	AdminEmail    string
	AdminPassword string
}

func Load() Config {
	return Config{
		HTTPAddr:      envOrDefault("HTTP_ADDR", ":8080"),
		SQLitePath:    envOrDefault("SQLITE_PATH", "data/korean-learning.db"),
		LogDir:        envOrDefault("LOG_DIR", "logs"),
		WebRoot:       strings.TrimSpace(os.Getenv("WEB_ROOT")),
		SeedDatabase:  envBool("DB_SEED", true),
		JWTSecret:     envOrDefault("JWT_SECRET", "dev-secret-change-me"),
		AdminName:     envOrDefault("ADMIN_NAME", "Admin"),
		AdminEmail:    envOrDefault("ADMIN_EMAIL", "admin@korean.local"),
		AdminPassword: envOrDefault("ADMIN_PASSWORD", "admin123"),
	}
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func envBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	return value == "1" || value == "true" || value == "yes" || value == "on"
}
