package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	RedisURL    string

	DBMaxOpenConns int
	DBMaxIdleConns int
	DBLogSQL       bool
	DBLogColorful  bool

	AdminSessionTTL   time.Duration
	AdminCookieSecure bool

	LogLevel  string
	LogFormat string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	maxOpen, err := getEnvInt("DB_MAX_OPEN_CONNS", 25)
	if err != nil {
		return nil, err
	}
	maxIdle, err := getEnvInt("DB_MAX_IDLE_CONNS", 5)
	if err != nil {
		return nil, err
	}
	logSQL, err := getEnvBool("DB_LOG_SQL", false)
	if err != nil {
		return nil, err
	}
	sessionTTL, err := getEnvDuration("ADMIN_SESSION_TTL", 8*time.Hour)
	if err != nil {
		return nil, err
	}
	cookieSecure, err := getEnvBool("ADMIN_COOKIE_SECURE", false)
	if err != nil {
		return nil, err
	}
	logFormat := getEnv("LOG_FORMAT", "text")
	cfg := &Config{
		HTTPAddr:          getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		RedisURL:          os.Getenv("REDIS_URL"),
		DBMaxOpenConns:    maxOpen,
		DBMaxIdleConns:    maxIdle,
		DBLogSQL:          logSQL,
		DBLogColorful:     strings.EqualFold(logFormat, "text"),
		AdminSessionTTL:   sessionTTL,
		AdminCookieSecure: cookieSecure,
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		LogFormat:         logFormat,
	}
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if cfg.RedisURL == "" {
		return nil, errors.New("REDIS_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0, got %d", key, parsed)
	}
	return parsed, nil
}

func getEnvBool(key string, fallback bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return parsed, nil
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration such as 8h or 30m: %w", key, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0, got %s", key, parsed)
	}
	return parsed, nil
}
