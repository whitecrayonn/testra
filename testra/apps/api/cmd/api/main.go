package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/testra/testra/apps/api/internal/shared/config"
	"github.com/testra/testra/apps/api/internal/shared/db"
	"github.com/testra/testra/apps/api/internal/shared/server"
)

func main() {
	cfg := config.Load()

	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	tokenManager, err := cfg.JWTManager()
	if err != nil {
		log.Fatalf("failed to load jwt manager: %v", err)
	}

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	router := server.New(server.Config{
		DB:                  database,
		JWTManager:          tokenManager,
		JWTExpiry:           time.Duration(cfg.JWTExpiryMinutes) * time.Minute,
		RefreshExpiryDays:   cfg.RefreshExpiryDays,
		RefreshAbsoluteDays: cfg.RefreshAbsoluteDays,
		RedisAddr:           cfg.RedisAddr,
		SMTPHost:            cfg.SMTPHost,
		SMTPPort:            cfg.SMTPPort,
		SMTPFrom:            cfg.SMTPFrom,
		SMTPUsername:        cfg.SMTPUsername,
		SMTPPasswordSecret:  cfg.SMTPPasswordSecret,
		SecretProvider:      cfg.SecretProvider(),
		CORSAllowedOrigins:  cfg.CORSAllowedOrigins,
		IdempotencyKeyTTL:   time.Duration(cfg.IdempotencyKeyTTLMinutes) * time.Minute,
		MLServiceURL:        cfg.MLServiceURL,
		StripeSecretKey:     cfg.StripeSecretKey,
		StripePriceID:       cfg.StripePriceID,
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Testra API server listening on %s", addr)
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down API server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("API server stopped")
}
