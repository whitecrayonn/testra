package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/jwt"
	"github.com/testra/testra/apps/api/internal/shared/password"
)

type Service struct {
	repo      Repository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewService(repo Repository, jwtSecret string, jwtExpiry time.Duration) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
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
	User  User
	Token string
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
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

	return AuthResult{User: user, Token: token}, nil
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

	return AuthResult{User: *user, Token: token}, nil
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

	return rawToken, nil
}

type ResetPasswordInput struct {
	Token       string
	NewPassword string
}

func (s *Service) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
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
