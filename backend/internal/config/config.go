// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName string
	AppEnv  string
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DBMaxConns int32
	DBMinConns int32

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	RequestLogEnabled      bool
	RequestLogToRedis      bool
	RequestLogToDB         bool
	RequestLogMaxEntries   int64
	RequestLogPayloadLimit int

	LogLevel     string
	StockLogFile string

	SeedEnabled   bool
	SeedItemCount int

	CORSAllowOrigins []string

	ShutdownTimeout time.Duration
}

// Load reads the .env file when present (env vars always win) and builds a Config.
func Load() *Config {
	// Missing .env is not an error: docker-compose injects the variables directly.
	_ = godotenv.Load()

	cfg := &Config{
		AppName: env("APP_NAME", "SISWA"),
		AppEnv:  env("APP_ENV", "development"),
		AppPort: env("APP_PORT", "3069"),

		DBHost:     env("DB_HOST", "localhost"),
		DBPort:     env("DB_PORT", "5432"),
		DBUser:     env("DB_USER", "siswa"),
		DBPassword: env("DB_PASSWORD", "siswa_secret"),
		DBName:     env("DB_NAME", "siswa"),
		DBSSLMode:  env("DB_SSLMODE", "disable"),
		DBMaxConns: int32(envInt("DB_MAX_CONNS", 20)),
		DBMinConns: int32(envInt("DB_MIN_CONNS", 2)),

		RedisAddr:     env("REDIS_ADDR", "localhost:6379"),
		RedisPassword: env("REDIS_PASSWORD", ""),
		RedisDB:       envInt("REDIS_DB", 0),

		RequestLogEnabled:      envBool("REQUEST_LOG_ENABLED", true),
		RequestLogToRedis:      envBool("REQUEST_LOG_TO_REDIS", true),
		RequestLogToDB:         envBool("REQUEST_LOG_TO_DB", true),
		RequestLogMaxEntries:   int64(envInt("REQUEST_LOG_MAX_ENTRIES", 1000)),
		RequestLogPayloadLimit: envInt("REQUEST_LOG_PAYLOAD_LIMIT", 2048),

		LogLevel:     env("LOG_LEVEL", "info"),
		StockLogFile: env("STOCK_LOG_FILE", "logs/stock_mutation.log"),

		SeedEnabled:   envBool("SEED_ENABLED", true),
		SeedItemCount: envInt("SEED_ITEM_COUNT", 1000),

		CORSAllowOrigins: envList("CORS_ALLOW_ORIGINS", []string{"http://localhost:4069"}),

		ShutdownTimeout: 10 * time.Second,
	}

	return cfg
}

// DatabaseURL returns the libpq/pgx connection string.
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.AppEnv, "production")
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v, err := strconv.Atoi(env(key, "")); err == nil {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if v, err := strconv.ParseBool(env(key, "")); err == nil {
		return v
	}
	return fallback
}

func envList(key string, fallback []string) []string {
	raw := env(key, "")
	if raw == "" {
		return fallback
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
