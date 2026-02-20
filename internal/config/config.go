package config

import (
	"fmt"
	"os"
)

type Config struct {
	TelegramBotToken string
	DBURL            string
	OllamaBaseURL    string
	OllamaAPIToken   string

	ModelStrategist string
	ModelFinancier  string
	ModelAuditor    string
	ModelAnalyst    string
	ModelModerator  string
}

func Load() (*Config, error) {
	cfg := &Config{
		TelegramBotToken: lookupEnvOrDefault("TELEGRAM_BOT_TOKEN", ""),
		DBURL:            lookupEnvOrDefault("DB_URL", ""),
		OllamaBaseURL:    lookupEnvOrDefault("OLLAMA_BASE_URL", "http://localhost:11434/v1"),
		OllamaAPIToken:   lookupEnvOrDefault("OLLAMA_API_TOKEN", ""),

		ModelStrategist: lookupEnvOrDefault("MODEL_STRATEGIST", "llama3:8b"),
		ModelFinancier:  lookupEnvOrDefault("MODEL_FINANCIER", "gemma2:9b"),
		ModelAuditor:    lookupEnvOrDefault("MODEL_AUDITOR", "mistral:7b"),
		ModelAnalyst:    lookupEnvOrDefault("MODEL_ANALYST", "qwen2.5:7b"),
		ModelModerator:  lookupEnvOrDefault("MODEL_MODERATOR", "llama3.1:8b"),
	}

	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.DBURL == "" {
		return nil, fmt.Errorf("DB_URL is required")
	}

	return cfg, nil
}

func lookupEnvOrDefault(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}
