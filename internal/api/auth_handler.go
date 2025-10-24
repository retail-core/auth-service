package api

import (
	"encoding/json"
	"net/http"
	"github.com/retail-core/auth-service/internal/auth"
	"github.com/retail-core/auth-service/internal/common"
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
	var req RegisterRequest

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, common.ErrBadRequest)
        return
    }

    if err := common.ValidateStruct(req); err != nil {
        WriteError(w, err)
        return
    }

    userId, err := h.service.Register(r.Context(), req.Username, req.Email, req.Password, req.Role, req.TenantID)
    if err != nil {
        WriteError(w, common.ErrInternal)
        return
    }

	res := RegisterResponse{UserID: userId}
	WriteJson(w, http.StatusCreated, res)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, common.ErrBadRequest)
		return
	}

	if err := common.ValidateStruct(req); err != nil {
		WriteError(w, err)
		return
	}

	token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		WriteError(w, err)
		return
	}

	res := LoginResponse{AccessToken: token, RefreshToken: "refresh_token"}
	WriteJson(w, http.StatusOK, res)
}

func(h *AuthHandler) Verify (w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, common.ErrBadRequest)
		return
	}

	if err := common.ValidateStruct(req); err != nil {
		WriteError(w, err)
		return
	}

	if err := h.service.Verify(r.Context(), req.Email, req.OTP);  err != nil {
		WriteError(w, common.ErrInvalidVerificationCredential)
		return
	}

	res := VerifyResponse{Success: true}
	WriteJson(w, http.StatusOK, res)
}

