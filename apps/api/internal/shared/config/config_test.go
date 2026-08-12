package config

import "testing"

func TestValidate_NonProductionIsLenient(t *testing.T) {
	cfg := Config{
		Env:         "development",
		DatabaseURL: "postgres://testra:testra@localhost:5432/testra?sslmode=disable",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error in development, got %v", err)
	}
}

func TestValidate_ProductionRejectsMissingJWTKeyFiles(t *testing.T) {
	cfg := Config{
		Env:         "production",
		DatabaseURL: "postgres://user:strongpass@db:5432/testra?sslmode=require",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing JWT key files in production")
	}
}

func TestValidate_ProductionRejectsMissingIssuerOrAudience(t *testing.T) {
	cfg := Config{
		Env:               "production",
		JWTPrivateKeyFile: "/etc/testra/jwt.key",
		JWTPublicKeyFiles: "/etc/testra/jwt.pub",
		DatabaseURL:       "postgres://user:strongpass@db:5432/testra?sslmode=require",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing issuer or audience in production")
	}
}

func TestValidate_ProductionRejectsExampleDatabaseCredentials(t *testing.T) {
	cfg := Config{
		Env:               "production",
		JWTIssuer:         "https://api.testra.io",
		JWTAudience:       "testra-api",
		JWTPrivateKeyFile: "/etc/testra/jwt.key",
		JWTPublicKeyFiles: "/etc/testra/jwt.pub",
		DatabaseURL:       "postgres://testra:testra@localhost:5432/testra?sslmode=require",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error for example database credentials in production")
	}
}

func TestValidate_ProductionRejectsDisabledTLS(t *testing.T) {
	cfg := Config{
		Env:               "production",
		JWTIssuer:         "https://api.testra.io",
		JWTAudience:       "testra-api",
		JWTPrivateKeyFile: "/etc/testra/jwt.key",
		JWTPublicKeyFiles: "/etc/testra/jwt.pub",
		DatabaseURL:       "postgres://user:strongpass@db:5432/testra?sslmode=disable",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error for disabled TLS in production")
	}
}

func TestValidate_ProductionAcceptsStrongConfig(t *testing.T) {
	cfg := Config{
		Env:               "production",
		JWTIssuer:         "https://api.testra.io",
		JWTAudience:       "testra-api",
		JWTPrivateKeyFile: "/etc/testra/jwt.key",
		JWTPublicKeyFiles: "/etc/testra/jwt.pub",
		DatabaseURL:       "postgres://user:strongpass@db:5432/testra?sslmode=require",
		WebBaseURL:        "https://app.testra.io",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid production config, got %v", err)
	}
}

func validProductionConfig() Config {
	return Config{
		Env:               "production",
		JWTIssuer:         "https://api.testra.io",
		JWTAudience:       "testra-api",
		JWTPrivateKeyFile: "/etc/testra/jwt.key",
		JWTPublicKeyFiles: "/etc/testra/jwt.pub",
		DatabaseURL:       "postgres://user:strongpass@db:5432/testra?sslmode=require",
		WebBaseURL:        "https://app.testra.io",
	}
}

func TestValidate_ProductionRejectsMissingWebBaseURL(t *testing.T) {
	cfg := validProductionConfig()
	cfg.WebBaseURL = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing WEB_BASE_URL in production")
	}
}

func TestValidate_ProductionRejectsLocalhostWebBaseURL(t *testing.T) {
	cfg := validProductionConfig()
	cfg.WebBaseURL = "https://localhost:3000"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for a localhost WEB_BASE_URL in production")
	}
}

func TestValidate_ProductionRejectsInsecureWebBaseURL(t *testing.T) {
	cfg := validProductionConfig()
	cfg.WebBaseURL = "http://app.testra.io"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for a non-https WEB_BASE_URL in production")
	}
}
