package main

import (
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/joho/godotenv/autoload"

	"github.com/testra/testra/apps/api/internal/shared/config"
)

func main() {
	cfg := config.Load()

	dsn := cfg.DatabaseURL
	if idx := strings.Index(dsn, "://"); idx != -1 {
		dsn = "pgx://" + dsn[idx+3:]
	}

	m, err := migrate.New(
		"file://"+cfg.MigrationsPath,
		dsn,
	)
	if err != nil {
		log.Fatalf("failed to create migrator: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run migrations: %v", err)
	}

	log.Println("migrations applied successfully")
}
