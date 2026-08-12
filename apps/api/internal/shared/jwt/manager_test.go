package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestManager_SignAndParse(t *testing.T) {
	m, err := NewTestManager("test-issuer", "test-audience")
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}

	userID := uuid.New()
	token, err := m.Sign(userID, "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	claims, err := m.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("user id mismatch")
	}
	if claims.Email != "test@example.com" {
		t.Fatalf("email mismatch")
	}
	if claims.Issuer != "test-issuer" {
		t.Fatalf("issuer mismatch: %s", claims.Issuer)
	}
	if len(claims.Audience) != 1 || claims.Audience[0] != "test-audience" {
		t.Fatalf("audience mismatch: %v", claims.Audience)
	}
}

func TestManager_Parse_RejectsWrongIssuer(t *testing.T) {
	m, err := NewTestManager("issuer-a", "audience-a")
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	token, err := m.Sign(uuid.New(), "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	other, err := NewTestManager("issuer-b", "audience-a")
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	if _, err := other.Parse(token); err == nil {
		t.Fatal("expected issuer mismatch error")
	}
}

func TestManager_Parse_RejectsWrongAudience(t *testing.T) {
	m, err := NewTestManager("issuer-a", "audience-a")
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	token, err := m.Sign(uuid.New(), "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	other, err := NewTestManager("issuer-a", "audience-b")
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	if _, err := other.Parse(token); err == nil {
		t.Fatal("expected audience mismatch error")
	}
}

func TestManager_Parse_RejectsExpiredToken(t *testing.T) {
	m, err := NewTestManager("test-issuer", "test-audience")
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	token, err := m.Sign(uuid.New(), "test@example.com", -time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := m.Parse(token); err == nil {
		t.Fatal("expected expired token error")
	}
}

func TestManager_Parse_UnknownKid(t *testing.T) {
	m, err := NewTestManager("test-issuer", "test-audience")
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	userID := uuid.New()
	token, err := m.Sign(userID, "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	other, _ := NewTestManager("test-issuer", "test-audience")
	if _, err := other.Parse(token); err == nil {
		t.Fatal("expected unknown kid error")
	}
}

func TestManager_Rotate_KeepsPreviousKeyValid(t *testing.T) {
	m1, err := NewTestManager("test-issuer", "test-audience")
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	token, err := m1.Sign(uuid.New(), "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	// Generate a brand-new key pair and rotate it in, keeping the old public key.
	newKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	newPrivatePEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(newKey)})

	oldJWKs := m1.JWKS()
	if len(oldJWKs) != 1 {
		t.Fatalf("expected 1 jwk, got %d", len(oldJWKs))
	}

	publicPEMs := map[string][]byte{
		"new": pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: func() []byte {
			b, _ := x509.MarshalPKIXPublicKey(&newKey.PublicKey)
			return b
		}()}),
		"old": func() []byte {
			b, _ := x509.MarshalPKIXPublicKey(m1.signingKey.public.(*rsa.PublicKey))
			return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: b})
		}(),
	}

	if err := m1.Rotate(newPrivatePEM, publicPEMs); err != nil {
		t.Fatalf("rotate: %v", err)
	}

	// Old token should still verify with the previous public key.
	if _, err := m1.Parse(token); err != nil {
		t.Fatalf("parse after rotation: %v", err)
	}

	// New tokens must be signed with the new key.
	newToken, err := m1.Sign(uuid.New(), "new@example.com", time.Hour)
	if err != nil {
		t.Fatalf("sign after rotation: %v", err)
	}
	if _, err := m1.Parse(newToken); err != nil {
		t.Fatalf("parse new token: %v", err)
	}

	jwks := m1.JWKS()
	if len(jwks) != 2 {
		t.Fatalf("expected 2 jwks after rotation, got %d", len(jwks))
	}
}

func TestManager_JWKS(t *testing.T) {
	m, err := NewTestManager("test-issuer", "test-audience")
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}

	jwks := m.JWKS()
	if len(jwks) != 1 {
		t.Fatalf("expected 1 jwk, got %d", len(jwks))
	}
	if jwks[0].Kty != "RSA" || jwks[0].Alg != "RS256" || jwks[0].Use != "sig" {
		t.Fatalf("unexpected jwk fields: %+v", jwks[0])
	}
	if jwks[0].Kid == "" {
		t.Fatal("expected kid")
	}
	if jwks[0].N == "" || jwks[0].E == "" {
		t.Fatal("expected n and e")
	}

	marshaled, err := m.MarshalJWKS()
	if err != nil {
		t.Fatalf("marshal jwks: %v", err)
	}
	var resp JWKSResponse
	if err := json.Unmarshal(marshaled, &resp); err != nil {
		t.Fatalf("unmarshal jwks: %v", err)
	}
	if len(resp.Keys) != 1 {
		t.Fatalf("expected 1 key in jwks response")
	}
}

func TestManager_Parse_RejectsHMAC(t *testing.T) {
	m, err := NewTestManager("test-issuer", "test-audience")
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	claims := Claims{
		UserID: uuid.New(),
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.New().String(),
			Issuer:    "test-issuer",
			Audience:  jwt.ClaimStrings{"test-audience"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("symmetric-secret"))
	if err != nil {
		t.Fatalf("sign hmac: %v", err)
	}
	if _, err := m.Parse(signed); err == nil || !strings.Contains(err.Error(), "unexpected signing method") {
		t.Fatalf("expected unexpected signing method error, got %v", err)
	}
}
