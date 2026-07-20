package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/testra/testra/apps/api/internal/shared/jwt"
	"github.com/testra/testra/apps/api/internal/shared/secrets"
	"github.com/testra/testra/apps/api/internal/shared/server"
)

// routeInfo holds one registered method+path pair from the chi router.
type routeInfo struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Fprintln(os.Stderr, "Usage: openapi\n\nPrints a YAML list of all routes registered in the chi router.")
		return
	}

	cfg := server.Config{
		DB:                  openTestDB(),
		JWTManager:          newTestJWTManager(),
		JWTExpiry:           0,
		RefreshExpiryDays:   30,
		RefreshAbsoluteDays: 365,
		RedisAddr:           "",
		SMTPHost:            "smtp.example.com",
		SMTPPort:            "587",
		SMTPFrom:            "noreply@example.com",
		SMTPUsername:        "user",
		SMTPPasswordSecret:  "secret",
		SecretProvider:      secrets.NewEnvProvider(),
		CORSAllowedOrigins:  "http://localhost:3000",
		IdempotencyKeyTTL:   0,
		MLServiceURL:        "",
		StripeSecretKey:     "",
		StripePriceID:       "",
	}

	handler := server.New(cfg)

	routes := collectRoutes(handler)

	fmt.Println("# OpenAPI route inventory generated from chi router")
	fmt.Println("routes:")
	for _, r := range routes {
		fmt.Printf("  - method: %s\n    path: %s\n", r.Method, r.Path)
	}
}

func openTestDB() *sql.DB {
	db, err := sql.Open("pgx", "postgres://testra:testra@localhost:5432/testra?sslmode=disable")
	if err != nil {
		// Fallback to an in-memory SQLite-like connection string? pgx does not support
		// SQLite, but opening without a reachable server is enough for route enumeration.
		// If this still fails, panic.
		fmt.Fprintf(os.Stderr, "open db: %v\n", err)
		os.Exit(1)
	}
	return db
}

func newTestJWTManager() *jwt.Manager {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate key: %v\n", err)
		os.Exit(1)
	}

	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	publicDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal public key: %v\n", err)
		os.Exit(1)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})

	m, err := jwt.NewManager("testra", "testra", privatePEM, map[string][]byte{"test": publicPEM})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create jwt manager: %v\n", err)
		os.Exit(1)
	}
	return m
}

func collectRoutes(handler http.Handler) []routeInfo {
	router, ok := handler.(*chi.Mux)
	if !ok {
		fmt.Fprintln(os.Stderr, "server handler is not a chi mux")
		os.Exit(1)
	}

	var routes []routeInfo
	err := chi.Walk(router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		routes = append(routes, routeInfo{Method: method, Path: route})
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "walk router: %v\n", err)
		os.Exit(1)
	}

	// Drop chi's default NotFound/MethodNotAllowed handlers from the walk output
	// and sort for deterministic output.
	filtered := routes[:0]
	for _, r := range routes {
		if strings.HasPrefix(r.Path, "/-") || strings.Contains(r.Path, "notfound") || strings.Contains(r.Path, "methodnotallowed") {
			continue
		}
		filtered = append(filtered, r)
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Path != filtered[j].Path {
			return filtered[i].Path < filtered[j].Path
		}
		return filtered[i].Method < filtered[j].Method
	})
	return filtered
}
