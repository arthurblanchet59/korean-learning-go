package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr                   string
	SQLitePath                 string
	LogDir                     string
	WebRoot                    string
	SeedDatabase               bool
	JWTSecret                  string
	AdminName                  string
	AdminEmail                 string
	AdminPassword              string
	AzureAIEndpoint            string
	AzureAIAPIKey              string
	AzureAIModel               string
	AzureAIEmbeddingEndpoint   string
	AzureAIEmbeddingAPIKey     string
	AzureAIEmbeddingModel      string
	AzureAIEmbeddingDimensions int
}

func Load() Config {
	// Local development can use .env; real environment variables keep priority.
	_ = godotenv.Load()

	return Config{
		HTTPAddr:                   envOrDefault("HTTP_ADDR", ":8080"),
		SQLitePath:                 envOrDefault("SQLITE_PATH", "data/korean-learning.db"),
		LogDir:                     envOrDefault("LOG_DIR", "logs"),
		WebRoot:                    strings.TrimSpace(os.Getenv("WEB_ROOT")),
		SeedDatabase:               envBool("DB_SEED", true),
		JWTSecret:                  envOrDefault("JWT_SECRET", "dev-secret-change-me"),
		AdminName:                  envOrDefault("ADMIN_NAME", "Admin"),
		AdminEmail:                 envOrDefault("ADMIN_EMAIL", "admin@korean.local"),
		AdminPassword:              envOrDefault("ADMIN_PASSWORD", "admin123"),
		AzureAIEndpoint:            strings.TrimSpace(os.Getenv("AZURE_AI_ENDPOINT")),
		AzureAIAPIKey:              strings.TrimSpace(os.Getenv("AZURE_AI_API_KEY")),
		AzureAIModel:               strings.TrimSpace(os.Getenv("AZURE_AI_MODEL")),
		AzureAIEmbeddingEndpoint:   strings.TrimSpace(os.Getenv("AZURE_AI_EMBEDDING_ENDPOINT")),
		AzureAIEmbeddingAPIKey:     strings.TrimSpace(os.Getenv("AZURE_AI_EMBEDDING_API_KEY")),
		AzureAIEmbeddingModel:      strings.TrimSpace(os.Getenv("AZURE_AI_EMBEDDING_MODEL")),
		AzureAIEmbeddingDimensions: envInt("AZURE_AI_EMBEDDING_DIMENSIONS", 1024),
	}
}

func (config Config) Validate() error {
	configured := 0
	for _, value := range []string{config.AzureAIEndpoint, config.AzureAIAPIKey, config.AzureAIModel} {
		if value != "" {
			configured++
		}
	}
	if configured != 0 && configured != 3 {
		return fmt.Errorf("AZURE_AI_ENDPOINT, AZURE_AI_API_KEY and AZURE_AI_MODEL must be configured together")
	}
	if config.AzureAIEmbeddingModel != "" {
		if config.EmbeddingEndpoint() == "" || config.EmbeddingAPIKey() == "" {
			return fmt.Errorf("an embedding endpoint and API key are required when AZURE_AI_EMBEDDING_MODEL is configured")
		}
		switch config.AzureAIEmbeddingDimensions {
		case 256, 512, 1024, 1536:
		default:
			return fmt.Errorf("AZURE_AI_EMBEDDING_DIMENSIONS must be 256, 512, 1024 or 1536")
		}
	} else if config.AzureAIEmbeddingEndpoint != "" || config.AzureAIEmbeddingAPIKey != "" {
		return fmt.Errorf("AZURE_AI_EMBEDDING_MODEL is required when embedding credentials are configured")
	}
	return nil
}

func (config Config) AzureAIEnabled() bool {
	return config.AzureAIEndpoint != "" && config.AzureAIAPIKey != "" && config.AzureAIModel != ""
}

func (config Config) RAGEnabled() bool {
	return config.AzureAIEnabled() &&
		config.AzureAIEmbeddingModel != "" &&
		config.EmbeddingEndpoint() != "" &&
		config.EmbeddingAPIKey() != ""
}

func (config Config) EmbeddingEndpoint() string {
	if config.AzureAIEmbeddingEndpoint != "" {
		return config.AzureAIEmbeddingEndpoint
	}
	return config.AzureAIEndpoint
}

func (config Config) EmbeddingAPIKey() string {
	if config.AzureAIEmbeddingAPIKey != "" {
		return config.AzureAIEmbeddingAPIKey
	}
	return config.AzureAIAPIKey
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

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
