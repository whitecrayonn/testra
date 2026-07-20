package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/testra/testra/apps/api/internal/shared/jwt"
	"github.com/testra/testra/apps/api/internal/shared/secrets"
)

type Config struct {
	Env                      string
	Port                     string
	DatabaseURL              string
	MigrationsPath           string
	JWTIssuer                string
	JWTAudience              string
	JWTPrivateKeyFile        string
	JWTPublicKeyFiles        string
	JWTExpiryMinutes         int
	RefreshExpiryDays        int
	RefreshAbsoluteDays      int
	RedisAddr                string
	SMTPHost                 string
	SMTPPort                 string
	SMTPFrom                 string
	SMTPUsername             string
	SMTPPasswordSecret       string
	CORSAllowedOrigins       string
	IdempotencyKeyTTLMinutes int
	MLServiceURL             string
	StripeSecretKey          string
	StripePriceID            string
}

func Load() Config {
	return Config{
		Env:                      getEnv("ENV", "development"),
		Port:                     getEnv("PORT", "8080"),
		DatabaseURL:              getEnv("DATABASE_URL", "postgres://testra:testra@localhost:5432/testra?sslmode=disable"),
		MigrationsPath:           getEnv("MIGRATIONS_PATH", "apps/api/migrations"),
		JWTIssuer:                getEnv("JWT_ISSUER", "testra"),
		JWTAudience:              getEnv("JWT_AUDIENCE", "testra-api"),
		JWTPrivateKeyFile:        getEnv("JWT_PRIVATE_KEY_FILE", ""),
		JWTPublicKeyFiles:        getEnv("JWT_PUBLIC_KEY_FILES", ""),
		JWTExpiryMinutes:         getEnvInt("JWT_EXPIRY_MINUTES", 15),
		RefreshExpiryDays:        getEnvInt("REFRESH_EXPIRY_DAYS", 30),
		RefreshAbsoluteDays:      getEnvInt("REFRESH_ABSOLUTE_DAYS", 90),
		RedisAddr:                getEnv("REDIS_ADDR", "localhost:6379"),
		SMTPHost:                 getEnv("SMTP_HOST", "localhost"),
		SMTPPort:                 getEnv("SMTP_PORT", "1025"),
		SMTPFrom:                 getEnv("SMTP_FROM", "noreply@testra.local"),
		SMTPUsername:             getEnv("SMTP_USERNAME", ""),
		SMTPPasswordSecret:       getEnv("SMTP_PASSWORD_SECRET", "SMTP_PASSWORD"),
		CORSAllowedOrigins:       getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		IdempotencyKeyTTLMinutes: getEnvInt("IDEMPOTENCY_KEY_TTL_MINUTES", 1440),
		MLServiceURL:             getEnv("ML_SERVICE_URL", ""),
		StripeSecretKey:          getEnv("STRIPE_SECRET_KEY", ""),
		StripePriceID:            getEnv("STRIPE_PRICE_ID", ""),
	}
}

// IsProduction reports whether the service is running in a production
// environment.
func (c Config) IsProduction() bool {
	env := strings.ToLower(strings.TrimSpace(c.Env))
	return env == "production" || env == "prod"
}

// Validate enforces production safety invariants. In non-production
// environments it is a no-op so local development remains frictionless.
// In production it fails fast when secrets are missing, weak, or left at
// their example defaults (Technical Debt Register item C4).
func (c Config) Validate() error {
	if !c.IsProduction() {
		return nil
	}

	if c.JWTPrivateKeyFile == "" {
		return fmt.Errorf("JWT_PRIVATE_KEY_FILE must be set in production")
	}
	if c.JWTPublicKeyFiles == "" {
		return fmt.Errorf("JWT_PUBLIC_KEY_FILES must be set in production")
	}
	if c.JWTIssuer == "" || c.JWTAudience == "" {
		return fmt.Errorf("JWT_ISSUER and JWT_AUDIENCE must be set in production")
	}

	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL must be set in production")
	}
	if strings.Contains(c.DatabaseURL, "testra:testra@") {
		return fmt.Errorf("DATABASE_URL uses the example credentials; provide real production credentials")
	}
	if strings.Contains(c.DatabaseURL, "sslmode=disable") {
		return fmt.Errorf("DATABASE_URL must not disable TLS (sslmode=disable) in production")
	}

	return nil
}

// JWTManager loads the configured PEM key material and returns a TokenManager.
// When no private key file is configured (typical for local development), an
// ephemeral RSA key pair is generated so the server can start without manual
// key provisioning.
func (c Config) JWTManager() (*jwt.Manager, error) {
	var privatePEM []byte
	publicPEMs := make(map[string][]byte)

	if c.JWTPrivateKeyFile == "" {
		// Dev-only ephemeral key. Production validation prevents this path.
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("generate dev jwt key: %w", err)
		}
		privatePEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
		pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			return nil, err
		}
		publicPEMs["dev"] = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	} else {
		var err error
		privatePEM, err = os.ReadFile(c.JWTPrivateKeyFile)
		if err != nil {
			return nil, fmt.Errorf("read JWT_PRIVATE_KEY_FILE: %w", err)
		}

		for _, entry := range strings.Split(c.JWTPublicKeyFiles, ",") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			kid := strings.TrimSuffix(filepath.Base(entry), filepath.Ext(entry))
			if idx := strings.Index(entry, "="); idx > 0 {
				kid = strings.TrimSpace(entry[:idx])
				entry = strings.TrimSpace(entry[idx+1:])
			}
			data, err := os.ReadFile(entry)
			if err != nil {
				return nil, fmt.Errorf("read public key %q: %w", entry, err)
			}
			publicPEMs[kid] = data
		}
	}

	return jwt.NewManager(c.JWTIssuer, c.JWTAudience, privatePEM, publicPEMs)
}

// SecretProvider returns the configured secrets backend. Currently this is a
// thin wrapper around environment variables; it can be swapped for a local
// secrets store or file-based secret manager without changing service code.
func (c Config) SecretProvider() secrets.Provider {
	return secrets.NewEnvProvider()
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
