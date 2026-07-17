package config

import (
	"os"
	"strconv"
)

type Config struct {
	Env                      string
	Port                     string
	DatabaseURL              string
	MigrationsPath           string
	JWTSecret                string
	JWTExpiryMinutes         int
	RefreshExpiryDays        int
	RefreshAbsoluteDays      int
	RedisAddr                string
	SMTPHost                 string
	SMTPPort                 string
	SMTPFrom                 string
	CORSAllowedOrigins       string
	IdempotencyKeyTTLMinutes int
}

func Load() Config {
	return Config{
		Env:                      getEnv("ENV", "development"),
		Port:                     getEnv("PORT", "8080"),
		DatabaseURL:              getEnv("DATABASE_URL", "postgres://testra:testra@localhost:5432/testra?sslmode=disable"),
		MigrationsPath:           getEnv("MIGRATIONS_PATH", "apps/api/migrations"),
		JWTSecret:                getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production"),
		JWTExpiryMinutes:         getEnvInt("JWT_EXPIRY_MINUTES", 15),
		RefreshExpiryDays:        getEnvInt("REFRESH_EXPIRY_DAYS", 30),
		RefreshAbsoluteDays:      getEnvInt("REFRESH_ABSOLUTE_DAYS", 90),
		RedisAddr:                getEnv("REDIS_ADDR", "localhost:6379"),
		SMTPHost:                 getEnv("SMTP_HOST", "localhost"),
		SMTPPort:                 getEnv("SMTP_PORT", "1025"),
		SMTPFrom:                 getEnv("SMTP_FROM", "noreply@testra.local"),
		CORSAllowedOrigins:       getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		IdempotencyKeyTTLMinutes: getEnvInt("IDEMPOTENCY_KEY_TTL_MINUTES", 1440),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
