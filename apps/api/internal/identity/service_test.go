package identity

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/jwt"
	"github.com/testra/testra/apps/api/internal/shared/password"
)

type fakeRepo struct {
	users           map[uuid.UUID]*User
	usersByEmail    map[string]*User
	resetTokens     map[string]*PasswordResetToken
	refreshTokens   map[string]*RefreshToken
	mfaUpdates      int
	passwordUpdates int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		users:         make(map[uuid.UUID]*User),
		usersByEmail:  make(map[string]*User),
		resetTokens:   make(map[string]*PasswordResetToken),
		refreshTokens: make(map[string]*RefreshToken),
	}
}

func (r *fakeRepo) Create(ctx context.Context, user *User) error {
	r.users[user.ID] = user
	r.usersByEmail[user.Email] = user
	return nil
}

func (r *fakeRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	u, ok := r.usersByEmail[email]
	if !ok {
		return nil, sharederrors.ErrNotFound
	}
	return u, nil
}

func (r *fakeRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, sharederrors.ErrNotFound
	}
	return u, nil
}

func (r *fakeRepo) UpdateMFA(ctx context.Context, userID uuid.UUID, secret string, enabled bool) error {
	u, ok := r.users[userID]
	if !ok {
		return sharederrors.ErrNotFound
	}
	u.MFASecret = secret
	u.MFAEnabled = enabled
	r.mfaUpdates++
	return nil
}

func (r *fakeRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	u, ok := r.users[userID]
	if !ok {
		return sharederrors.ErrNotFound
	}
	u.Password = passwordHash
	r.passwordUpdates++
	return nil
}

func (r *fakeRepo) CreateResetToken(ctx context.Context, token *PasswordResetToken) error {
	r.resetTokens[token.TokenHash] = token
	return nil
}

func (r *fakeRepo) GetResetTokenByHash(ctx context.Context, hash string) (*PasswordResetToken, error) {
	t, ok := r.resetTokens[hash]
	if !ok {
		return nil, sharederrors.ErrNotFound
	}
	return t, nil
}

func (r *fakeRepo) MarkResetTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	for _, t := range r.resetTokens {
		if t.ID == tokenID {
			now := time.Now().UTC()
			t.UsedAt = &now
			return nil
		}
	}
	return sharederrors.ErrNotFound
}

func (r *fakeRepo) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	r.refreshTokens[token.TokenHash] = token
	return nil
}

func (r *fakeRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error) {
	t, ok := r.refreshTokens[hash]
	if !ok {
		return nil, sharederrors.ErrNotFound
	}
	return t, nil
}

func (r *fakeRepo) RevokeRefreshToken(ctx context.Context, tokenID uuid.UUID, replacedBy uuid.UUID) error {
	for _, t := range r.refreshTokens {
		if t.ID == tokenID {
			now := time.Now().UTC()
			t.RevokedAt = &now
			t.ReplacedBy = &replacedBy
			return nil
		}
	}
	return sharederrors.ErrNotFound
}

func (r *fakeRepo) RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) error {
	for _, t := range r.refreshTokens {
		if t.FamilyID == familyID {
			now := time.Now().UTC()
			t.RevokedAt = &now
		}
	}
	return nil
}

func (r *fakeRepo) RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	for _, t := range r.refreshTokens {
		if t.UserID == userID {
			now := time.Now().UTC()
			t.RevokedAt = &now
		}
	}
	return nil
}

func newTestService(repo *fakeRepo) *Service {
	tm, err := jwt.NewTestManager("test-issuer", "test-audience")
	if err != nil {
		panic(err)
	}
	return NewService(repo, tm, 15*time.Minute, 30*24*time.Hour, 90*24*time.Hour, SMTPConfig{})
}

func seedUser(repo *fakeRepo, email, plainPass string) *User {
	hash, _ := password.Hash(plainPass)
	user := &User{
		ID:        uuid.New(),
		Email:     email,
		Password:  hash,
		Name:      "Test User",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	repo.users[user.ID] = user
	repo.usersByEmail[email] = user
	return user
}

func TestLoginWithMFAEnabled_RequiresCode(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "mfa@test.com", "TestPass123!@#")
	user.MFAEnabled = true
	user.MFASecret = "JBSWY3DPEHPK3PXP"

	_, err := svc.Login(context.Background(), LoginInput{
		Email:    "mfa@test.com",
		Password: "TestPass123!@#",
		MFACode:  "",
	})
	if err != sharederrors.ErrMFARequired {
		t.Fatalf("expected ErrMFARequired, got %v", err)
	}
}

func TestLoginWithMFAEnabled_WrongCode(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "mfa2@test.com", "TestPass123!@#")
	user.MFAEnabled = true

	key, _ := totp.Generate(totp.GenerateOpts{Issuer: "Testra", AccountName: "mfa2@test.com"})
	user.MFASecret = key.Secret()

	_, err := svc.Login(context.Background(), LoginInput{
		Email:    "mfa2@test.com",
		Password: "TestPass123!@#",
		MFACode:  "000000",
	})
	if err != sharederrors.ErrInvalidCredential {
		t.Fatalf("expected ErrInvalidCredential, got %v", err)
	}
}

func TestLoginWithMFAEnabled_ValidCode(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "mfa3@test.com", "TestPass123!@#")
	user.MFAEnabled = true

	key, _ := totp.Generate(totp.GenerateOpts{Issuer: "Testra", AccountName: "mfa3@test.com"})
	user.MFASecret = key.Secret()

	code, _ := totp.GenerateCode(user.MFASecret, time.Now())

	_, err := svc.Login(context.Background(), LoginInput{
		Email:    "mfa3@test.com",
		Password: "TestPass123!@#",
		MFACode:  code,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestMFARepeatedFailuresLockout(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "lockout@test.com", "TestPass123!@#")
	user.MFAEnabled = true
	key, _ := totp.Generate(totp.GenerateOpts{Issuer: "Testra", AccountName: user.Email})
	user.MFASecret = key.Secret()

	// 5 consecutive wrong MFA codes should lock the account.
	for i := 0; i < mfaMaxAttempts; i++ {
		_, err := svc.Login(context.Background(), LoginInput{
			Email:    user.Email,
			Password: "TestPass123!@#",
			MFACode:  "000000",
		})
		if err != sharederrors.ErrInvalidCredential {
			t.Fatalf("attempt %d: expected ErrInvalidCredential, got %v", i+1, err)
		}
	}

	_, err := svc.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "TestPass123!@#",
		MFACode:  "000000",
	})
	if err != sharederrors.ErrTooManyRequests {
		t.Fatalf("expected ErrTooManyRequests after max attempts, got %v", err)
	}
}

func TestVerifyMFARepeatedFailuresLockout(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "verify-lockout@test.com", "TestPass123!@#")
	key, _ := totp.Generate(totp.GenerateOpts{Issuer: "Testra", AccountName: user.Email})
	user.MFASecret = key.Secret()

	for i := 0; i < mfaMaxAttempts; i++ {
		err := svc.VerifyMFA(context.Background(), user.ID, "000000")
		if err != sharederrors.ErrInvalidCredential {
			t.Fatalf("attempt %d: expected ErrInvalidCredential, got %v", i+1, err)
		}
	}

	err := svc.VerifyMFA(context.Background(), user.ID, "000000")
	if err != sharederrors.ErrTooManyRequests {
		t.Fatalf("expected ErrTooManyRequests, got %v", err)
	}
}

func TestLoginWithoutMFA(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	seedUser(repo, "plain@test.com", "TestPass123!@#")

	_, err := svc.Login(context.Background(), LoginInput{
		Email:    "plain@test.com",
		Password: "TestPass123!@#",
	})
	if err != nil {
		t.Fatalf("expected nil error for non-mfa login, got %v", err)
	}
}

func TestSetupMFA_AlreadyEnabled(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "setup@test.com", "TestPass123!@#")
	user.MFAEnabled = true

	_, err := svc.SetupMFA(context.Background(), user.ID)
	if err != sharederrors.ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestSetupMFA_Success(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "setup2@test.com", "TestPass123!@#")

	result, err := svc.SetupMFA(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Secret == "" {
		t.Fatal("expected non-empty secret")
	}
	if result.QRCode == "" {
		t.Fatal("expected non-empty qr code")
	}
	if repo.mfaUpdates != 1 {
		t.Fatalf("expected 1 mfa update, got %d", repo.mfaUpdates)
	}
	if user.MFASecret == "" {
		t.Fatal("expected user to have mfa secret stored")
	}
	if user.MFAEnabled {
		t.Fatal("expected mfa to not be enabled yet")
	}
}

func TestVerifyMFA_NoSecretSetup(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "verify@test.com", "TestPass123!@#")

	err := svc.VerifyMFA(context.Background(), user.ID, "123456")
	if err != sharederrors.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestVerifyMFA_AlreadyEnabled(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "verify2@test.com", "TestPass123!@#")
	user.MFASecret = "JBSWY3DPEHPK3PXP"
	user.MFAEnabled = true

	err := svc.VerifyMFA(context.Background(), user.ID, "123456")
	if err != sharederrors.ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestVerifyMFA_Success(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "verify3@test.com", "TestPass123!@#")

	key, _ := totp.Generate(totp.GenerateOpts{Issuer: "Testra", AccountName: "verify3@test.com"})
	user.MFASecret = key.Secret()

	code, _ := totp.GenerateCode(user.MFASecret, time.Now())

	err := svc.VerifyMFA(context.Background(), user.ID, code)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !user.MFAEnabled {
		t.Fatal("expected mfa to be enabled after verification")
	}
}

func TestDisableMFA_NotEnabled(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "disable@test.com", "TestPass123!@#")

	err := svc.DisableMFA(context.Background(), user.ID, "")
	if err != sharederrors.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestDisableMFA_Success(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "disable2@test.com", "TestPass123!@#")
	user.MFAEnabled = true
	user.MFASecret = "JBSWY3DPEHPK3PXP"

	code, err := totp.GenerateCode(user.MFASecret, time.Now())
	if err != nil {
		t.Fatalf("generate totp: %v", err)
	}

	err = svc.DisableMFA(context.Background(), user.ID, code)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user.MFAEnabled {
		t.Fatal("expected mfa to be disabled")
	}
	if user.MFASecret != "" {
		t.Fatal("expected mfa secret to be cleared")
	}
}

func TestDisableMFA_MissingCode(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "disable3@test.com", "TestPass123!@#")
	user.MFAEnabled = true
	user.MFASecret = "JBSWY3DPEHPK3PXP"

	err := svc.DisableMFA(context.Background(), user.ID, "")
	if err != sharederrors.ErrMFARequired {
		t.Fatalf("expected ErrMFARequired, got %v", err)
	}
}

func TestRequestPasswordReset_UserNotFound_ReturnsEmpty(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	token, err := svc.RequestPasswordReset(context.Background(), RequestPasswordResetInput{
		Email: "nonexistent@test.com",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if token != "" {
		t.Fatal("expected empty token for nonexistent user")
	}
}

func TestRequestPasswordReset_Success(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	seedUser(repo, "reset@test.com", "TestPass123!@#")

	token, err := svc.RequestPasswordReset(context.Background(), RequestPasswordResetInput{
		Email: "reset@test.com",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if len(repo.resetTokens) != 1 {
		t.Fatalf("expected 1 reset token stored, got %d", len(repo.resetTokens))
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	err := svc.ResetPassword(context.Background(), ResetPasswordInput{
		Token:       "invalidtoken",
		NewPassword: "NewPass123!@#",
	})
	if err != sharederrors.ErrInvalidCredential {
		t.Fatalf("expected ErrInvalidCredential, got %v", err)
	}
}

func TestResetPassword_ExpiredToken(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "expired@test.com", "TestPass123!@#")

	rawToken, _ := generateResetToken()
	hash := hashToken(rawToken)
	expired := time.Now().UTC().Add(-1 * time.Minute)
	repo.resetTokens[hash] = &PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: expired,
		CreatedAt: time.Now().UTC().Add(-31 * time.Minute),
	}

	err := svc.ResetPassword(context.Background(), ResetPasswordInput{
		Token:       rawToken,
		NewPassword: "NewPass123!@#",
	})
	if err != sharederrors.ErrInvalidCredential {
		t.Fatalf("expected ErrInvalidCredential, got %v", err)
	}
}

func TestResetPassword_AlreadyUsed(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "used@test.com", "TestPass123!@#")

	rawToken, _ := generateResetToken()
	hash := hashToken(rawToken)
	usedAt := time.Now().UTC().Add(-5 * time.Minute)
	repo.resetTokens[hash] = &PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().UTC().Add(25 * time.Minute),
		UsedAt:    &usedAt,
		CreatedAt: time.Now().UTC().Add(-5 * time.Minute),
	}

	err := svc.ResetPassword(context.Background(), ResetPasswordInput{
		Token:       rawToken,
		NewPassword: "NewPass123!@#",
	})
	if err != sharederrors.ErrInvalidCredential {
		t.Fatalf("expected ErrInvalidCredential, got %v", err)
	}
}

func TestResetPassword_Success(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "success@test.com", "TestPass123!@#")

	rawToken, _ := generateResetToken()
	hash := hashToken(rawToken)
	token := &PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().UTC().Add(30 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}
	repo.resetTokens[hash] = token

	err := svc.ResetPassword(context.Background(), ResetPasswordInput{
		Token:       rawToken,
		NewPassword: "NewPass123!@#",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repo.passwordUpdates != 1 {
		t.Fatalf("expected 1 password update, got %d", repo.passwordUpdates)
	}
	if token.UsedAt == nil {
		t.Fatal("expected token to be marked as used")
	}
}

func TestRefreshTokenReuse(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "reuse@test.com", "TestPass123!@#")

	result, err := svc.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "TestPass123!@#",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	// First refresh consumes the token and issues a replacement.
	_, err = svc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: result.RefreshToken})
	if err != nil {
		t.Fatalf("first refresh: %v", err)
	}

	// Reusing the original token must revoke the family.
	_, err = svc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: result.RefreshToken})
	if err != sharederrors.ErrTokenRevoked {
		t.Fatalf("expected ErrTokenRevoked on reuse, got %v", err)
	}
}

func TestLogout(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "logout@test.com", "TestPass123!@#")

	result, err := svc.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "TestPass123!@#",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if err := svc.Logout(context.Background(), result.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}

	if _, err := svc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: result.RefreshToken}); err != sharederrors.ErrTokenRevoked {
		t.Fatalf("expected ErrTokenRevoked after logout, got %v", err)
	}
}

func TestLogoutAllDevices(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "logoutall@test.com", "TestPass123!@#")

	result1, err := svc.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "TestPass123!@#",
	})
	if err != nil {
		t.Fatalf("login 1: %v", err)
	}

	_, err = svc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: result1.RefreshToken})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	// A new login on another device produces a separate family.
	result2, err := svc.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "TestPass123!@#",
	})
	if err != nil {
		t.Fatalf("login 2: %v", err)
	}

	if err := svc.LogoutAllDevices(context.Background(), user.ID); err != nil {
		t.Fatalf("logout all: %v", err)
	}

	if _, err := svc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: result2.RefreshToken}); err != sharederrors.ErrTokenRevoked {
		t.Fatalf("expected ErrTokenRevoked after logout all, got %v", err)
	}
}
