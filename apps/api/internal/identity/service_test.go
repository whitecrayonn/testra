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
	denylistedJTIs  map[string]time.Time
	mfaUpdates      int
	passwordUpdates int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		users:          make(map[uuid.UUID]*User),
		usersByEmail:   make(map[string]*User),
		resetTokens:    make(map[string]*PasswordResetToken),
		refreshTokens:  make(map[string]*RefreshToken),
		denylistedJTIs: make(map[string]time.Time),
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

func (r *fakeRepo) IncrementFailedLoginAttempts(ctx context.Context, userID uuid.UUID) (int, error) {
	u, ok := r.users[userID]
	if !ok {
		return 0, sharederrors.ErrNotFound
	}
	u.FailedLoginAttempts++
	return u.FailedLoginAttempts, nil
}

func (r *fakeRepo) LockAccount(ctx context.Context, userID uuid.UUID, until time.Time) error {
	u, ok := r.users[userID]
	if !ok {
		return sharederrors.ErrNotFound
	}
	u.LockedUntil = &until
	return nil
}

func (r *fakeRepo) ResetFailedLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	u, ok := r.users[userID]
	if !ok {
		return sharederrors.ErrNotFound
	}
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
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

func (r *fakeRepo) DenylistAccessToken(ctx context.Context, jti string, userID uuid.UUID, expiresAt time.Time) error {
	if jti == "" {
		return nil
	}
	r.denylistedJTIs[jti] = expiresAt
	return nil
}

func (r *fakeRepo) IsAccessTokenDenylisted(ctx context.Context, jti string) (bool, error) {
	_, ok := r.denylistedJTIs[jti]
	return ok, nil
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

// TestLoginRepeatedFailuresLockout is a regression test for SBL-022: repeated
// wrong-password attempts must lock the account, independent of MFA.
func TestLoginRepeatedFailuresLockout(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "login-lockout@test.com", "TestPass123!@#")

	for i := 0; i < loginMaxAttempts; i++ {
		_, err := svc.Login(context.Background(), LoginInput{
			Email:    user.Email,
			Password: "WrongPassword!",
		})
		if err != sharederrors.ErrInvalidCredential {
			t.Fatalf("attempt %d: expected ErrInvalidCredential, got %v", i+1, err)
		}
	}

	// The account is now locked. The response must stay indistinguishable
	// from an ordinary wrong-password failure (ErrInvalidCredential, not a
	// distinct ErrTooManyRequests/429), even with the correct password,
	// so a caller can't enumerate "this email exists and is locked" versus
	// "invalid credentials". The lock itself is verified via the repo below.
	if user.LockedUntil == nil {
		t.Fatal("expected account to be locked in the repository")
	}
	_, err := svc.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "TestPass123!@#",
	})
	if err != sharederrors.ErrInvalidCredential {
		t.Fatalf("expected ErrInvalidCredential once locked (uniform response), got %v", err)
	}
}

// TestLoginLockoutExpiryClearsCounter is a regression test: once a lockout
// window has passed, a single further mistake must not immediately re-lock
// the account for another window. It must behave like a fresh account.
func TestLoginLockoutExpiryClearsCounter(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "expired-lockout@test.com", "TestPass123!@#")

	// Simulate an already-expired lockout left over from a previous window.
	user.FailedLoginAttempts = loginMaxAttempts
	past := time.Now().UTC().Add(-time.Minute)
	user.LockedUntil = &past

	// One more wrong password should behave like attempt 1 of a fresh
	// window (still ErrInvalidCredential), not re-trigger a lock.
	if _, err := svc.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "WrongPassword!",
	}); err != sharederrors.ErrInvalidCredential {
		t.Fatalf("expected ErrInvalidCredential, got %v", err)
	}
	if user.LockedUntil != nil {
		t.Fatal("expected account not to be re-locked after only one failure past an expired window")
	}

	// The correct password should now succeed outright.
	if _, err := svc.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "TestPass123!@#",
	}); err != nil {
		t.Fatalf("expected successful login after expired lockout cleared, got %v", err)
	}
}

// TestResetPasswordClearsLockout is a regression test: resetting a password
// is the normal self-service escape from a lockout and must clear it, not
// leave the account locked despite the user proving account ownership.
func TestResetPasswordClearsLockout(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "locked-reset@test.com", "TestPass123!@#")

	future := time.Now().UTC().Add(10 * time.Minute)
	user.FailedLoginAttempts = loginMaxAttempts
	user.LockedUntil = &future

	rawToken, _ := generateResetToken()
	hash := hashToken(rawToken)
	repo.resetTokens[hash] = &PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().UTC().Add(30 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}

	if err := svc.ResetPassword(context.Background(), ResetPasswordInput{
		Token:       rawToken,
		NewPassword: "NewPass123!@#",
	}); err != nil {
		t.Fatalf("reset password: %v", err)
	}

	if user.LockedUntil != nil {
		t.Fatal("expected lockout to be cleared by password reset")
	}
	if user.FailedLoginAttempts != 0 {
		t.Fatalf("expected failed attempts cleared, got %d", user.FailedLoginAttempts)
	}

	if _, err := svc.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "NewPass123!@#",
	}); err != nil {
		t.Fatalf("expected login with new password to succeed, got %v", err)
	}
}

// TestLoginSuccessResetsFailedAttempts is a regression test for SBL-022: a
// successful login must clear the failed-attempt counter so occasional typos
// don't accumulate toward a future lockout.
func TestLoginSuccessResetsFailedAttempts(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "reset-attempts@test.com", "TestPass123!@#")

	for i := 0; i < loginMaxAttempts-1; i++ {
		_, err := svc.Login(context.Background(), LoginInput{
			Email:    user.Email,
			Password: "WrongPassword!",
		})
		if err != sharederrors.ErrInvalidCredential {
			t.Fatalf("attempt %d: expected ErrInvalidCredential, got %v", i+1, err)
		}
	}

	if _, err := svc.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "TestPass123!@#",
	}); err != nil {
		t.Fatalf("expected successful login, got %v", err)
	}

	if user.FailedLoginAttempts != 0 {
		t.Fatalf("expected failed attempts reset to 0, got %d", user.FailedLoginAttempts)
	}

	// A subsequent run of near-max failures should not carry over and should
	// not lock the account, since the counter was reset above.
	for i := 0; i < loginMaxAttempts-1; i++ {
		_, err := svc.Login(context.Background(), LoginInput{
			Email:    user.Email,
			Password: "WrongPassword!",
		})
		if err != sharederrors.ErrInvalidCredential {
			t.Fatalf("post-reset attempt %d: expected ErrInvalidCredential, got %v", i+1, err)
		}
	}
	if _, err := svc.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "TestPass123!@#",
	}); err != nil {
		t.Fatalf("expected login to still succeed (not locked), got %v", err)
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

func TestPasswordResetLink_UsesConfiguredWebBaseURL(t *testing.T) {
	repo := newFakeRepo()
	tm, err := jwt.NewTestManager("test-issuer", "test-audience")
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}
	svc := NewService(repo, tm, 15*time.Minute, 30*24*time.Hour, 90*24*time.Hour, SMTPConfig{
		WebBaseURL: "https://app.testra.example/",
	})

	link := svc.passwordResetLink("sometoken")
	want := "https://app.testra.example/reset-password?token=sometoken"
	if link != want {
		t.Fatalf("expected %q, got %q", want, link)
	}
}

func TestPasswordResetLink_DefaultsWhenUnconfigured(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	link := svc.passwordResetLink("sometoken")
	want := "http://localhost:3000/reset-password?token=sometoken"
	if link != want {
		t.Fatalf("expected %q, got %q", want, link)
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

// TestResetPassword_RevokesExistingSessions is a regression test for
// SBL-023: resetting a password must invalidate any refresh tokens issued
// before the reset, so a stolen session can't survive a password reset.
func TestResetPassword_RevokesExistingSessions(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "session-reset@test.com", "TestPass123!@#")

	loginResult, err := svc.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "TestPass123!@#",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	rawToken, _ := generateResetToken()
	hash := hashToken(rawToken)
	repo.resetTokens[hash] = &PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().UTC().Add(30 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}

	if err := svc.ResetPassword(context.Background(), ResetPasswordInput{
		Token:       rawToken,
		NewPassword: "NewPass123!@#",
	}); err != nil {
		t.Fatalf("reset password: %v", err)
	}

	if _, err := svc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: loginResult.RefreshToken}); err != sharederrors.ErrTokenRevoked {
		t.Fatalf("expected ErrTokenRevoked for session predating password reset, got %v", err)
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

func TestRevokeAccessToken(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)
	user := seedUser(repo, "revoke-access@test.com", "TestPass123!@#")

	result, err := svc.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "TestPass123!@#",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	claims, err := svc.tokenManager.Parse(result.Token)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}
	if claims.ID == "" {
		t.Fatal("expected access token to carry a non-empty jti")
	}

	denylisted, err := repo.IsAccessTokenDenylisted(context.Background(), claims.ID)
	if err != nil {
		t.Fatalf("check denylist before revoke: %v", err)
	}
	if denylisted {
		t.Fatal("expected token not to be denylisted before revocation")
	}

	if err := svc.RevokeAccessToken(context.Background(), result.Token); err != nil {
		t.Fatalf("revoke access token: %v", err)
	}

	denylisted, err = repo.IsAccessTokenDenylisted(context.Background(), claims.ID)
	if err != nil {
		t.Fatalf("check denylist after revoke: %v", err)
	}
	if !denylisted {
		t.Fatal("expected token to be denylisted after revocation")
	}
}

func TestRevokeAccessToken_EmptyTokenIsNoop(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	if err := svc.RevokeAccessToken(context.Background(), ""); err != nil {
		t.Fatalf("expected no error revoking an empty token, got %v", err)
	}
	if len(repo.denylistedJTIs) != 0 {
		t.Fatal("expected no denylist entries for an empty token")
	}
}

func TestRevokeAccessToken_MalformedTokenIsNoop(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	if err := svc.RevokeAccessToken(context.Background(), "not-a-real-jwt"); err != nil {
		t.Fatalf("expected no error revoking a malformed token, got %v", err)
	}
	if len(repo.denylistedJTIs) != 0 {
		t.Fatal("expected no denylist entries for a malformed token")
	}
}
