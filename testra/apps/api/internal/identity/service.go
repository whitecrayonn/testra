package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/jwt"
	"github.com/testra/testra/apps/api/internal/shared/password"
	"github.com/testra/testra/apps/api/internal/shared/validation"
)

type SMTPConfig struct {
	Host string
	Port string
	From string
}

type Service struct {
	repo            Repository
	jwtSecret       string
	jwtExpiry       time.Duration
	refreshExpiry   time.Duration
	refreshAbsolute time.Duration
	smtp            SMTPConfig
}

func NewService(repo Repository, jwtSecret string, jwtExpiry time.Duration, refreshExpiry time.Duration, refreshAbsolute time.Duration, smtpCfg SMTPConfig) *Service {
	return &Service{
		repo:            repo,
		jwtSecret:       jwtSecret,
		jwtExpiry:       jwtExpiry,
		refreshExpiry:   refreshExpiry,
		refreshAbsolute: refreshAbsolute,
		smtp:            smtpCfg,
	}
}

type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

type LoginInput struct {
	Email    string
	Password string
	MFACode  string
}

type AuthResult struct {
	User         User
	Token        string
	RefreshToken string
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	if !validation.IsValidEmail(input.Email) {
		return AuthResult{}, sharederrors.ErrInvalidInput
	}
	if !validation.IsValidName(input.Name) {
		return AuthResult{}, sharederrors.ErrInvalidInput
	}
	if err := validatePassword(input.Password); err != nil {
		return AuthResult{}, err
	}

	existing, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil && err != sharederrors.ErrNotFound {
		return AuthResult{}, err
	}
	if existing != nil {
		return AuthResult{}, sharederrors.ErrConflict
	}

	hash, err := password.Hash(input.Password)
	if err != nil {
		return AuthResult{}, err
	}

	user := User{
		ID:        uuid.New(),
		Email:     input.Email,
		Password:  hash,
		Name:      input.Name,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, &user); err != nil {
		return AuthResult{}, err
	}

	token, err := jwt.Sign(user.ID, user.Email, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return AuthResult{}, err
	}

	refreshToken, err := s.issueRefreshToken(ctx, user.ID)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{User: user, Token: token, RefreshToken: refreshToken}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return AuthResult{}, sharederrors.ErrInvalidCredential
	}

	if !password.Verify(input.Password, user.Password) {
		return AuthResult{}, sharederrors.ErrInvalidCredential
	}

	if user.MFAEnabled {
		if input.MFACode == "" {
			return AuthResult{}, sharederrors.ErrMFARequired
		}
		valid := totp.Validate(input.MFACode, user.MFASecret)
		if !valid {
			return AuthResult{}, sharederrors.ErrInvalidCredential
		}
	}

	token, err := jwt.Sign(user.ID, user.Email, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return AuthResult{}, err
	}

	refreshToken, err := s.issueRefreshToken(ctx, user.ID)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{User: *user, Token: token, RefreshToken: refreshToken}, nil
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

type MFASetupResult struct {
	Secret string
	QRCode string
}

func (s *Service) SetupMFA(ctx context.Context, userID uuid.UUID) (MFASetupResult, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return MFASetupResult{}, err
	}

	if user.MFAEnabled {
		return MFASetupResult{}, sharederrors.ErrConflict
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Testra",
		AccountName: user.Email,
	})
	if err != nil {
		return MFASetupResult{}, err
	}

	if err := s.repo.UpdateMFA(ctx, userID, key.Secret(), false); err != nil {
		return MFASetupResult{}, err
	}

	return MFASetupResult{
		Secret: key.Secret(),
		QRCode: key.URL(),
	}, nil
}

func (s *Service) VerifyMFA(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.MFAEnabled {
		return sharederrors.ErrConflict
	}

	if user.MFASecret == "" {
		return sharederrors.ErrInvalidInput
	}

	valid := totp.Validate(code, user.MFASecret)
	if !valid {
		return sharederrors.ErrInvalidCredential
	}

	return s.repo.UpdateMFA(ctx, userID, user.MFASecret, true)
}

func (s *Service) DisableMFA(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.MFAEnabled {
		return sharederrors.ErrInvalidInput
	}

	if code != "" {
		valid := totp.Validate(code, user.MFASecret)
		if !valid {
			return sharederrors.ErrInvalidCredential
		}
	}

	return s.repo.UpdateMFA(ctx, userID, "", false)
}

type RequestPasswordResetInput struct {
	Email string
}

type ResetPasswordInput struct {
	Token       string
	NewPassword string
}

func (s *Service) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	if err := validatePassword(input.NewPassword); err != nil {
		return err
	}

	hash := hashToken(input.Token)

	token, err := s.repo.GetResetTokenByHash(ctx, hash)
	if err != nil {
		return sharederrors.ErrInvalidCredential
	}

	if token.UsedAt != nil {
		return sharederrors.ErrInvalidCredential
	}

	if time.Now().UTC().After(token.ExpiresAt) {
		return sharederrors.ErrInvalidCredential
	}

	newHash, err := password.Hash(input.NewPassword)
	if err != nil {
		return err
	}

	if err := s.repo.UpdatePassword(ctx, token.UserID, newHash); err != nil {
		return err
	}

	return s.repo.MarkResetTokenUsed(ctx, token.ID)
}

type RefreshTokenInput struct {
	RefreshToken string
}

func (s *Service) RefreshToken(ctx context.Context, input RefreshTokenInput) (AuthResult, error) {
	hash := hashToken(input.RefreshToken)

	stored, err := s.repo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return AuthResult{}, sharederrors.ErrInvalidCredential
	}

	if stored.RevokedAt != nil {
		_ = s.repo.RevokeRefreshTokenFamily(ctx, stored.FamilyID)
		return AuthResult{}, sharederrors.ErrTokenRevoked
	}

	if time.Now().UTC().After(stored.ExpiresAt) || time.Now().UTC().After(stored.AbsoluteExpiresAt) {
		return AuthResult{}, sharederrors.ErrTokenExpired
	}

	user, err := s.repo.GetByID(ctx, stored.UserID)
	if err != nil {
		return AuthResult{}, sharederrors.ErrInvalidCredential
	}

	accessToken, err := jwt.Sign(user.ID, user.Email, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return AuthResult{}, err
	}

	newRaw, err := s.issueRefreshTokenWithFamily(ctx, user.ID, stored.FamilyID)
	if err != nil {
		return AuthResult{}, err
	}

	_ = s.repo.RevokeRefreshToken(ctx, stored.ID, uuid.Nil)

	return AuthResult{User: *user, Token: accessToken, RefreshToken: newRaw}, nil
}

func (s *Service) issueRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	familyID := uuid.New()
	return s.issueRefreshTokenWithFamily(ctx, userID, familyID)
}

func (s *Service) issueRefreshTokenWithFamily(ctx context.Context, userID, familyID uuid.UUID) (string, error) {
	raw, err := generateRefreshToken()
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	token := &RefreshToken{
		ID:                uuid.New(),
		UserID:            userID,
		TokenHash:         hashToken(raw),
		FamilyID:          familyID,
		ExpiresAt:         now.Add(s.refreshExpiry),
		AbsoluteExpiresAt: now.Add(s.refreshAbsolute),
		CreatedAt:         now,
	}

	if err := s.repo.CreateRefreshToken(ctx, token); err != nil {
		return "", err
	}

	return raw, nil
}

func validatePassword(pw string) error {
	if len(pw) < 12 {
		return sharederrors.ErrInvalidInput
	}
	return nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, input RequestPasswordResetInput) (string, error) {
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		if err == sharederrors.ErrNotFound {
			return "", nil
		}
		return "", err
	}

	rawToken, err := generateResetToken()
	if err != nil {
		return "", err
	}

	hash := hashToken(rawToken)

	resetToken := &PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().UTC().Add(30 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.CreateResetToken(ctx, resetToken); err != nil {
		return "", err
	}

	if err := s.sendPasswordResetEmail(user.Email, rawToken); err != nil {
		return "", err
	}

	return rawToken, nil
}

func (s *Service) sendPasswordResetEmail(to, token string) error {
	if s.smtp.Host == "" {
		return nil
	}

	subject := "Testra — Password Reset"
	body := fmt.Sprintf("Use this token to reset your password: %s\nThis token expires in 30 minutes.", token)
	msg := strings.Join([]string{
		"From: " + s.smtp.From,
		"To: " + to,
		"Subject: " + subject,
		"",
		body,
	}, "\r\n")

	addr := s.smtp.Host + ":" + s.smtp.Port
	return smtp.SendMail(addr, nil, s.smtp.From, []string{to}, []byte(msg))
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return "rt_" + hex.EncodeToString(b), nil
}

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
