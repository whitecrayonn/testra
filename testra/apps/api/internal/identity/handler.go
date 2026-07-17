package identity

import (
	"encoding/json"
	"net/http"

	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
	"github.com/testra/testra/apps/api/internal/shared/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	MFACode  string `json:"mfa_code"`
}

type authResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	User         userResponse `json:"user"`
}

type userResponse struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	MFAEnabled bool   `json:"mfa_enabled"`
}

func mapUserResponse(u User) userResponse {
	return userResponse{
		ID:         u.ID.String(),
		Email:      u.Email,
		Name:       u.Name,
		MFAEnabled: u.MFAEnabled,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	result, err := h.service.Register(r.Context(), RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, authResponse{
		Token:        result.Token,
		RefreshToken: result.RefreshToken,
		User:         mapUserResponse(result.User),
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	result, err := h.service.Login(r.Context(), LoginInput{
		Email:    req.Email,
		Password: req.Password,
		MFACode:  req.MFACode,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, authResponse{
		Token:        result.Token,
		RefreshToken: result.RefreshToken,
		User:         mapUserResponse(result.User),
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, mapUserResponse(*user))
}

type mfaSetupResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}

func (h *Handler) SetupMFA(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	result, err := h.service.SetupMFA(r.Context(), userID)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, mfaSetupResponse{
		Secret: result.Secret,
		QRCode: result.QRCode,
	})
}

type verifyMFARequest struct {
	Code string `json:"code"`
}

func (h *Handler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	var req verifyMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.service.VerifyMFA(r.Context(), userID, req.Code); err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"status": "mfa_enabled"})
}

type disableMFARequest struct {
	Code string `json:"code"`
}

func (h *Handler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	var req disableMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.service.DisableMFA(r.Context(), userID, req.Code); err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"status": "mfa_disabled"})
}

type requestPasswordResetRequest struct {
	Email string `json:"email"`
}

func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req requestPasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	_, err := h.service.RequestPasswordReset(r.Context(), RequestPasswordResetInput{
		Email: req.Email,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"status": "reset_email_sent"})
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.service.ResetPassword(r.Context(), ResetPasswordInput{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}); err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"status": "password_reset"})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	result, err := h.service.RefreshToken(r.Context(), RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, authResponse{
		Token:        result.Token,
		RefreshToken: result.RefreshToken,
		User:         mapUserResponse(result.User),
	})
}

func mapError(w http.ResponseWriter, err error) {
	switch err {
	case sharederrors.ErrConflict:
		apihttp.ErrorJSON(w, http.StatusConflict, "CONFLICT", err.Error())
	case sharederrors.ErrInvalidCredential:
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error())
	case sharederrors.ErrMFARequired:
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "MFA_REQUIRED", err.Error())
	case sharederrors.ErrNotFound:
		apihttp.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case sharederrors.ErrInvalidInput:
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
	case sharederrors.ErrTokenExpired:
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "TOKEN_EXPIRED", err.Error())
	case sharederrors.ErrTokenRevoked:
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "TOKEN_REVOKED", err.Error())
	case sharederrors.ErrTooManyRequests:
		apihttp.ErrorJSON(w, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", err.Error())
	default:
		apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
