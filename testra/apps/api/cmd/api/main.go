package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/testra/testra/apps/api/internal/shared/config"
	"github.com/testra/testra/apps/api/internal/shared/db"
	"github.com/testra/testra/apps/api/internal/shared/server"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	router := server.New(server.Config{
		DB:                  database,
		JWTSecret:           cfg.JWTSecret,
		JWTExpiry:           time.Duration(cfg.JWTExpiryMinutes) * time.Minute,
		RefreshExpiryDays:   cfg.RefreshExpiryDays,
		RefreshAbsoluteDays: cfg.RefreshAbsoluteDays,
		RedisAddr:           cfg.RedisAddr,
		SMTPHost:            cfg.SMTPHost,
		SMTPPort:            cfg.SMTPPort,
		SMTPFrom:            cfg.SMTPFrom,
		CORSAllowedOrigins:  cfg.CORSAllowedOrigins,
		IdempotencyKeyTTL:   time.Duration(cfg.IdempotencyKeyTTLMinutes) * time.Minute,
		MLServiceURL:        cfg.MLServiceURL,
		StripeSecretKey:     cfg.StripeSecretKey,
		StripePriceID:       cfg.StripePriceID,
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Testra API server listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
