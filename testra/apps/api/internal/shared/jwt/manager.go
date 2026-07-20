package jwt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Manager signs and verifies RS256 access tokens, validates issuer/audience,
// exposes a JWKS, and supports key rotation through a thread-safe key set.
type Manager struct {
	issuer   string
	audience string

	mu               sync.RWMutex
	signingKey       *keyPair
	verificationKeys map[string]*keyPair
}

type keyPair struct {
	kid     string
	private crypto.Signer
	public  crypto.PublicKey
	thumb   string // SHA-256 thumbprint of public key bytes, used as fallback kid
}

// NewManager creates a token manager from PEM-encoded key material.
// signingPEM is the current RSA private key used for signing.
// verificationPEMs is a map of kid -> PEM-encoded public keys. The map must
// contain the public key that corresponds to signingPEM under the same kid.
// If a verification PEM is a private key, its public key is extracted.
func NewManager(issuer, audience string, signingPEM []byte, verificationPEMs map[string][]byte) (*Manager, error) {
	signing, err := loadPrivateKey(signingPEM)
	if err != nil {
		return nil, fmt.Errorf("invalid signing key: %w", err)
	}

	verificationKeys := make(map[string]*keyPair, len(verificationPEMs))
	for kid, pemBytes := range verificationPEMs {
		pub, err := loadPublicKey(pemBytes)
		if err != nil {
			return nil, fmt.Errorf("invalid verification key %q: %w", kid, err)
		}
		kp := &keyPair{kid: kid, public: pub}
		kp.thumb, err = keyThumbprint(pub)
		if err != nil {
			return nil, err
		}
		verificationKeys[kid] = kp
	}

	// Ensure the signing key is present in the verification set.
	signingThumb, err := keyThumbprint(signing.Public())
	if err != nil {
		return nil, err
	}

	var signingKid string
	for kid, kp := range verificationKeys {
		if kp.thumb == signingThumb {
			signingKid = kid
			kp.private = signing
			break
		}
	}
	if signingKid == "" {
		// If the caller didn't include the signing public key, add it under a
		// generated kid derived from the thumbprint.
		signingKid = "auto-" + signingThumb[:12]
		verificationKeys[signingKid] = &keyPair{
			kid:     signingKid,
			private: signing,
			public:  signing.Public(),
			thumb:   signingThumb,
		}
	}

	if issuer == "" || audience == "" {
		return nil, fmt.Errorf("issuer and audience must be set")
	}

	return &Manager{
		issuer:           issuer,
		audience:         audience,
		signingKey:       verificationKeys[signingKid],
		verificationKeys: verificationKeys,
	}, nil
}

// NewTestManager returns a manager backed by a freshly generated RSA key pair.
func NewTestManager(issuer, audience string) (*Manager, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	pubPEM := pemEncodePublic(&privateKey.PublicKey)
	privPEM := pemEncodePrivate(privateKey)
	return NewManager(issuer, audience, privPEM, map[string][]byte{"test": pubPEM})
}

// Rotate replaces the active signing key. The previous signing key and all
// known verification keys stay in the verification set so tokens issued before
// rotation continue to be accepted until they expire. verificationPEMs is used
// to add new public keys; keys whose public key is already known are ignored so
// that historical kid labels (and therefore previously issued tokens) remain
// valid.
func (m *Manager) Rotate(signingPEM []byte, verificationPEMs map[string][]byte) error {
	m.mu.RLock()
	oldKeys := m.verificationKeys
	m.mu.RUnlock()

	combined := make(map[string][]byte, len(oldKeys)+len(verificationPEMs))
	thumbprints := make(map[string]string, len(oldKeys)+len(verificationPEMs))

	for kid, kp := range oldKeys {
		pubDER, err := x509.MarshalPKIXPublicKey(kp.public)
		if err != nil {
			return fmt.Errorf("marshal public key %q: %w", kid, err)
		}
		combined[kid] = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
		thumbprints[kp.thumb] = kid
	}

	for kid, data := range verificationPEMs {
		pub, err := loadPublicKey(data)
		if err != nil {
			return fmt.Errorf("invalid verification key %q: %w", kid, err)
		}
		thumb, err := keyThumbprint(pub)
		if err != nil {
			return err
		}
		if _, known := thumbprints[thumb]; known {
			continue
		}
		combined[kid] = data
		thumbprints[thumb] = kid
	}

	next, err := NewManager(m.issuer, m.audience, signingPEM, combined)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.signingKey = next.signingKey
	m.verificationKeys = next.verificationKeys
	return nil
}

// Sign creates an access token for the given user.
func (m *Manager) Sign(userID uuid.UUID, email string, expiry time.Duration) (string, error) {
	m.mu.RLock()
	kp := m.signingKey
	m.mu.RUnlock()

	if kp == nil || kp.private == nil {
		return "", fmt.Errorf("no signing key configured")
	}

	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    m.issuer,
			Audience:  jwt.ClaimStrings{m.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kp.kid
	return token.SignedString(kp.private)
}

// Parse validates and returns the claims from an access token.
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	m.mu.RLock()
	keys := m.verificationKeys
	m.mu.RUnlock()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, fmt.Errorf("token missing kid header")
		}

		kp, ok := keys[kid]
		if !ok {
			return nil, fmt.Errorf("unknown kid: %s", kid)
		}
		return kp.public, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithAudience(m.audience))
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// JWKS returns the public verification keys as JWK objects suitable for a
// standard JWKS endpoint.
func (m *Manager) JWKS() []JWK {
	m.mu.RLock()
	keys := make([]*keyPair, 0, len(m.verificationKeys))
	for _, kp := range m.verificationKeys {
		keys = append(keys, kp)
	}
	m.mu.RUnlock()

	out := make([]JWK, 0, len(keys))
	for _, kp := range keys {
		jwk, err := publicKeyToJWK(kp.kid, kp.public)
		if err != nil {
			continue
		}
		out = append(out, jwk)
	}
	return out
}

// JWK is a minimal RSA public-key JWK.
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func publicKeyToJWK(kid string, pub crypto.PublicKey) (JWK, error) {
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return JWK{}, fmt.Errorf("unsupported public key type")
	}
	return JWK{
		Kty: "RSA",
		Kid: kid,
		Use: "sig",
		Alg: "RS256",
		N:   base64.RawURLEncoding.EncodeToString(rsaPub.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(serializeExponent(rsaPub.E)),
	}, nil
}

func serializeExponent(e int) []byte {
	// RSA public exponent e is typically 65537 (0x010001).
	buf := []byte{0, 0, 0, 0}
	for i := len(buf) - 1; i >= 0; i-- {
		buf[i] = byte(e & 0xff)
		e >>= 8
		if e == 0 {
			return buf[i:]
		}
	}
	return buf
}

func loadPrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
		return rsaKey, nil
	default:
		return nil, fmt.Errorf("unsupported PEM type %q", block.Type)
	}
}

func loadPublicKey(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	switch block.Type {
	case "RSA PUBLIC KEY":
		return x509.ParsePKCS1PublicKey(block.Bytes)
	case "PUBLIC KEY":
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA public key")
		}
		return rsaKey, nil
	case "PRIVATE KEY":
		privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := privateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
		return &rsaKey.PublicKey, nil
	default:
		return nil, fmt.Errorf("unsupported PEM type %q", block.Type)
	}
}

func keyThumbprint(pub crypto.PublicKey) (string, error) {
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("unsupported public key type")
	}
	der, err := x509.MarshalPKIXPublicKey(rsaPub)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(der)
	return base64.RawURLEncoding.EncodeToString(sum[:]), nil
}

func pemEncodePrivate(key *rsa.PrivateKey) []byte {
	der := x509.MarshalPKCS1PrivateKey(key)
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})
}

func pemEncodePublic(key *rsa.PublicKey) []byte {
	der, _ := x509.MarshalPKIXPublicKey(key)
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}

// JWKSResponse is the top-level structure returned by the JWKS endpoint.
type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

// Marshal returns the JSON JWKS document.
func (m *Manager) MarshalJWKS() ([]byte, error) {
	return json.Marshal(JWKSResponse{Keys: m.JWKS()})
}
