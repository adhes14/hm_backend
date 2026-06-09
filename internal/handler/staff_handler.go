package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/middleware"
	"github.com/hospital_management/backend/internal/service"
)

type StaffHandler struct {
	authService *service.AuthService
}

func NewStaffHandler(authService *service.AuthService) *StaffHandler {
	return &StaffHandler{authService: authService}
}

func (h *StaffHandler) List(w http.ResponseWriter, r *http.Request) {
	staff, err := h.authService.ListStaff(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	if staff == nil {
		staff = []domain.Staff{}
	}

	writeJSON(w, http.StatusOK, staff)
}

func (h *StaffHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.CreateStaffInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tempPassword, err := h.authService.CreateStaff(r.Context(), &input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"status":             "created",
		"temporary_password": tempPassword,
	})
}

func (h *StaffHandler) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var input service.UpdateStaffInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	updated, err := h.authService.UpdateStaff(r.Context(), id, &input, claims.StaffID)
	if err != nil {
		switch err {
		case domain.ErrCannotDemoteSelf:
			writeError(w, http.StatusConflict, "cannot_demote_self")
		case domain.ErrCannotRemoveLastAdmin:
			writeError(w, http.StatusConflict, "cannot_remove_last_admin")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update user")
		}
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

type resetPasswordRequest struct {
	Password string `json:"password"`
}

func (h *StaffHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Body is optional; if no body, use empty string for auto-generated password
		req.Password = ""
	}

	tempPassword, err := h.authService.ResetStaffPassword(r.Context(), id, req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reset password")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"temporary_password": tempPassword,
	})
}

type changePasswordRequest struct {
	Password string `json:"password"`
}

func (h *StaffHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.authService.ChangePassword(r.Context(), id, req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update password")
		return
	}

	w.WriteHeader(http.StatusOK)
}

type setActiveRequest struct {
	IsActive bool `json:"is_active"`
}

func (h *StaffHandler) SetActive(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req setActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.authService.SetActive(r.Context(), id, req.IsActive)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update status")
		return
	}

	w.WriteHeader(http.StatusOK)
}
