package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"
)

type Config struct {
	Environment           string
	HTTPAddr              string
	DatabaseURL           string
	RedisURL              string
	TelegramBotToken      string
	TelegramWebhookSecret string
	MiniAppURL            string
	LogLevel              slog.Level
	ReadHeaderTimeout     time.Duration
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	IdleTimeout           time.Duration
	ShutdownTimeout       time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Environment:           env("APP_ENV", "development"),
		HTTPAddr:              renderAddr(),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		RedisURL:              os.Getenv("REDIS_URL"),
		TelegramBotToken:      os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramWebhookSecret: os.Getenv("TELEGRAM_WEBHOOK_SECRET"),
		MiniAppURL:            os.Getenv("MINI_APP_URL"),
		ReadHeaderTimeout:     duration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
		ReadTimeout:           duration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:          duration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:           duration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout:       duration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}

	switch env("LOG_LEVEL", "info") {
	case "debug":
		cfg.LogLevel = slog.LevelDebug
	case "info":
		cfg.LogLevel = slog.LevelInfo
	case "warn":
		cfg.LogLevel = slog.LevelWarn
	case "error":
		cfg.LogLevel = slog.LevelError
	default:
		return Config{}, fmt.Errorf("LOG_LEVEL must be debug, info, warn, or error")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR cannot be empty")
	}
	if cfg.Environment != "development" && cfg.Environment != "test" && cfg.Environment != "production" {
		return Config{}, fmt.Errorf("APP_ENV must be development, test, or production")
	}
	return cfg, nil
}

func renderAddr() string {\n\tif port := os.Getenv("PORT"); port != "" {\n\t\treturn ":" + port\n\t}\n\treturn env("HTTP_ADDR", ":8080")\n}\n\nfunc env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func duration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
