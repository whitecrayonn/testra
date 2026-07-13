package config

import (
	"os"
)

type Config struct {
	Env             string
	Port            string
	DatabaseURL     string
	MigrationsPath  string
	JWTSecret       string
	JWTExpiryHours  int
	RedisAddr       string
}

func Load() Config {
	return Config{
		Env:            getEnv("ENV", "development"),
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://testra:testra@localhost:5432/testra?sslmode=disable"),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "apps/api/migrations"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production"),
		JWTExpiryHours: 168, // 7 days
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
