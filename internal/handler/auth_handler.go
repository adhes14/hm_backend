package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hospital_management/backend/internal/auth"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/middleware"
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

type changeMyPasswordRequest struct {
	CurrentPassword string `json:"current_password"` // optional (forced flow)
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

func (h *AuthHandler) ChangeMyPassword(w http.ResponseWriter, r *http.Request) {
	var req changeMyPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.NewPassword == "" || req.ConfirmPassword == "" {
		writeError(w, http.StatusBadRequest, "new_password and confirm_password are required")
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		writeError(w, http.StatusBadRequest, "new_password and confirm_password do not match")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	updatedStaff, err := h.authService.ChangeMyPassword(r.Context(), claims.StaffID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		switch e := err.(type) {
		case *domain.PasswordValidationError:
			writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
				"error":   "password_too_weak",
				"details": e.Details,
			})
		default:
			switch err {
			case domain.ErrInvalidCredentials:
				writeError(w, http.StatusUnauthorized, "invalid_current_password")
			default:
				fmt.Printf("ChangeMyPassword error: %v\n", err)
				writeError(w, http.StatusInternalServerError, "failed to change password")
			}
		}
		return
	}

	newToken, err := auth.GenerateToken(updatedStaff.ID, updatedStaff.Username, updatedStaff.Role, updatedStaff.FullName, updatedStaff.MustChangePassword)
	if err != nil {
		fmt.Printf("GenerateToken error: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token: newToken,
		User:  updatedStaff,
	})
}
