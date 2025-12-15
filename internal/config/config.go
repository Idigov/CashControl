package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress string
	DatabaseURL   string
	Environment   string
	// UseDatabaseURL определяет, использовать ли DATABASE_URL или отдельные параметры
	// Автоматически определяется: если DATABASE_URL задан - true, иначе - false
	UseDatabaseURL bool
	DBHost           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBPort           string
	DBSSLMode        string
	JWTSecret        string
	TelegramBotToken string
}

func Load() (*Config, error) {
	// Загрузка .env файла
	_ = godotenv.Load()

	databaseURL := getEnv("DATABASE_URL", "")

	// Определяем формат подключения:
	// 1. Если задан USE_DATABASE_URL - используем его значение
	// 2. Иначе автоматически: если DATABASE_URL задан - true, иначе - false
	useDatabaseURL := getEnv("USE_DATABASE_URL", "")
	var useURL bool
	if useDatabaseURL != "" {
		useURL = useDatabaseURL == "true" || useDatabaseURL == "1"
	} else {
		// Автоматическое определение: если DATABASE_URL задан - используем его
		useURL = databaseURL != ""
	}

	cfg := &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":8000"),
		DatabaseURL:   databaseURL,
		Environment:   getEnv("ENVIRONMENT", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "cashcontrol"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", "secret"),

		UseDatabaseURL: useURL,
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", "8567102489:AAFACiJvXn4-DYXDFwhnQ1HhrlfJciGnxV8"),

	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("валидация конфигурации: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.ServerAddress == "" {
		return fmt.Errorf("SERVER_ADDRESS не может быть пустым")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
