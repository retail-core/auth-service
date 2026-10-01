package api

import (
	"encoding/json"
	"net/http"

	"github.com/retail-core/auth-service/internal/auth"
	"github.com/retail-core/auth-service/internal/common"
	"github.com/retail-core/auth-service/internal/dtos"
)

type AuthHandler struct {
	service auth.Service
}

func NewAuthHandler(s auth.Service) *AuthHandler {
	return &AuthHandler{
		service: s,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dtos.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, common.ErrBadRequest)
		return
	}

	if err := common.ValidateStruct(req); err != nil {
		WriteError(w, err)
		return
	}

	req.Sanitize() // make email lowercase and trim spaces

	msg, err := h.service.Register(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}

	res := dtos.RegisterResponse{Message: msg}
	WriteJson(w, http.StatusCreated, res)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dtos.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, common.ErrBadRequest)
		return
	}

	if err := common.ValidateStruct(req); err != nil {
		WriteError(w, err)
		return
	}

	req.Sanitize() // make email lowercase and trim spaces

	token, rToken, user, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		WriteError(w, err)
		return
	}

	res := dtos.LoginResponse{AccessToken: token, RefreshToken: rToken, User: dtos.LUser{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}}

	http.SetCookie(w, &http.Cookie{
    Name:     "access_token",
    Value:    res.AccessToken,
    Path:     "/",
    HttpOnly: true,
    Secure:   false, // true in production (HTTPS)
    SameSite: http.SameSiteLaxMode,
    MaxAge:   604800, // 1 hour
})

http.SetCookie(w, &http.Cookie{
    Name:     "refresh_token",
    Value:    res.RefreshToken,
    Path:     "/",
    HttpOnly: true,
    Secure:   false,
    SameSite: http.SameSiteLaxMode,
    MaxAge:   30 * 24 * 3600,
})

	WriteJson(w, http.StatusOK, res)
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var req dtos.VerifyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, common.ErrBadRequest)
		return
	}

	if err := common.ValidateStruct(req); err != nil {
		WriteError(w, err)
		return
	}

	req.Sanitize() // make email lowercase and trim spaces

	if err := h.service.Verify(r.Context(), req.Email, req.OTP); err != nil {
		WriteError(w, err)
		return
	}

	res := dtos.VerifyResponse{Success: true}
	WriteJson(w, http.StatusOK, res)
}

func (h *AuthHandler) ResendOTP(w http.ResponseWriter, r *http.Request) {
	var req dtos.ResendOTPRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, common.ErrBadRequest)
		return
	}

	if err := common.ValidateStruct(req); err != nil {
		WriteError(w, err)
		return
	}

	req.Sanitize() // make email lowercase and trim spaces

	if err := h.service.ResendOTP(r.Context(), req.Email); err != nil {
		WriteError(w, err)
		return
	}

	res := dtos.ResendOTPResponse{Success: true}
	WriteJson(w, http.StatusOK, res)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dtos.RefreshTokenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, common.ErrBadRequest)
		return
	}

	if err := common.ValidateStruct(req); err != nil {
		WriteError(w, err)
		return
	}

	newAccessToken, newRefreshToken, err := h.service.GenerateTokens(r.Context(), req.RefreshToken)
	if err != nil {
		WriteError(w, err)
		return
	}

	res := dtos.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}
	WriteJson(w, http.StatusOK, res)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req dtos.ResetPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, common.ErrBadRequest)
		return
	}

	req.Sanitize() // make email lowercase and trim spaces

	if err := common.ValidateStruct(req); err != nil {
		WriteError(w, err)
		return
	}

	if err := h.service.ResetPassword(r.Context(), req.Email, req.NewPassword); err != nil {
		WriteError(w, err)
		return
	}

	res := map[string]string{"message": "Password reset successful"}
	WriteJson(w, http.StatusOK, res)
}
