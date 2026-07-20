package identity

import (
	"encoding/json"
	"net/http"
	"time"

	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
	"github.com/testra/testra/apps/api/internal/shared/middleware"
)

type Handler struct {
	service       *Service
	jwtExpiry     time.Duration
	refreshExpiry time.Duration
}

func NewHandler(service *Service, jwtExpiry time.Duration, refreshExpiry time.Duration) *Handler {
	return &Handler{
		service:       service,
		jwtExpiry:     jwtExpiry,
		refreshExpiry: refreshExpiry,
	}
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
		apihttp.MapError(w, err)
		return
	}

	h.setAuthCookies(w, r, result.Token, result.RefreshToken)
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
		apihttp.MapError(w, err)
		return
	}

	h.setAuthCookies(w, r, result.Token, result.RefreshToken)
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
		apihttp.MapError(w, err)
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
		apihttp.MapError(w, err)
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
		apihttp.MapError(w, err)
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
		apihttp.MapError(w, err)
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
		apihttp.MapError(w, err)
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
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"status": "password_reset"})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken := ""
	if cookieToken, ok := middleware.RefreshTokenFromCookie(r); ok {
		refreshToken = cookieToken
	} else {
		var req refreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
			return
		}
		refreshToken = req.RefreshToken
	}

	result, err := h.service.RefreshToken(r.Context(), RefreshTokenInput{
		RefreshToken: refreshToken,
	})
	if err != nil {
		middleware.ClearAuthCookies(w, r)
		apihttp.MapError(w, err)
		return
	}

	h.setAuthCookies(w, r, result.Token, result.RefreshToken)
	apihttp.JSON(w, http.StatusOK, authResponse{
		Token:        result.Token,
		RefreshToken: result.RefreshToken,
		User:         mapUserResponse(result.User),
	})
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken := ""
	if cookieToken, ok := middleware.RefreshTokenFromCookie(r); ok {
		refreshToken = cookieToken
	} else {
		var req logoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.ClearAuthCookies(w, r)
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
			return
		}
		refreshToken = req.RefreshToken
	}

	if err := h.service.Logout(r.Context(), refreshToken); err != nil {
		middleware.ClearAuthCookies(w, r)
		apihttp.MapError(w, err)
		return
	}

	middleware.ClearAuthCookies(w, r)
	apihttp.JSON(w, http.StatusOK, map[string]any{"status": "logged_out"})
}

func (h *Handler) LogoutAllDevices(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	if err := h.service.LogoutAllDevices(r.Context(), userID); err != nil {
		apihttp.MapError(w, err)
		return
	}

	middleware.ClearAuthCookies(w, r)
	apihttp.JSON(w, http.StatusOK, map[string]any{"status": "logged_out_all"})
}

func (h *Handler) CSRF(w http.ResponseWriter, r *http.Request) {
	token, err := middleware.GenerateCSRFToken()
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL", "failed to generate csrf token")
		return
	}
	middleware.SetCSRFCookie(w, r, token, middleware.DefaultCSRFCookieMaxAge())
	apihttp.JSON(w, http.StatusOK, map[string]any{"csrf_token": token})
}

func (h *Handler) setAuthCookies(w http.ResponseWriter, r *http.Request, accessToken, refreshToken string) {
	middleware.SetAccessTokenCookie(w, r, accessToken, int(h.jwtExpiry.Seconds()))
	middleware.SetRefreshTokenCookie(w, r, refreshToken, int(h.refreshExpiry.Seconds()))
}
