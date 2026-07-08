package config

import (
	"os"
	"strings"
)

type Config struct {
	HTTPAddr     string
	DatabaseURL  string
	AutoMigrate  bool
	SeedDatabase bool
}

func Load() Config {
	return Config{
		HTTPAddr:     envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		AutoMigrate:  envBool("DB_AUTO_MIGRATE", true),
		SeedDatabase: envBool("DB_SEED", true),
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
