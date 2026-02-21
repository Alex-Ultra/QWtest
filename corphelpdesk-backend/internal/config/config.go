package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	EncryptionKey  string
	JWTSecret      string
	TelegramBotToken string
	TelegramWebhookURL string
	Environment    string
}

func Load() *Config {
	// Load .env file if it exists
	godotenv.Load()

	return &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/corphelpdesk?sslmode=disable"),
		EncryptionKey:      getEnv("ENCRYPTION_KEY", ""),
		JWTSecret:          getEnv("JWT_SECRET", "default_secret_for_dev"),
		TelegramBotToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramWebhookURL: getEnv("TELEGRAM_WEBHOOK_URL", ""),
		Environment:        getEnv("ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}