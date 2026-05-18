package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string        `json:"token"`
	User  *domain.Staff `json:"user"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, user, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials, domain.ErrUserNotFound:
			writeError(w, http.StatusUnauthorized, "invalid username or password")
		case domain.ErrUserInactive:
			writeError(w, http.StatusForbidden, "user is inactive")
		default:
			fmt.Printf("Login error: %v\n", err)
			writeError(w, http.StatusInternalServerError, "failed to login")
		}
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token: token,
		User:  user,
	})
}
