package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/service"
)

type WardHandler struct {
	wardService *service.WardService
}

func NewWardHandler(wardService *service.WardService) *WardHandler {
	return &WardHandler{wardService: wardService}
}

type createWardRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateWardRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// POST /api/v1/wards
func (h *WardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createWardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	ward := &domain.Ward{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.wardService.CreateWard(r.Context(), ward); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create ward")
		return
	}

	writeJSON(w, http.StatusCreated, ward)
}

// GET /api/v1/wards
func (h *WardHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	wards, err := h.wardService.GetAllWards(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch wards")
		return
	}

	if wards == nil {
		wards = []domain.Ward{}
	}

	writeJSON(w, http.StatusOK, wards)
}

// GET /api/v1/wards/{id}
func (h *WardHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ward id")
		return
	}

	ward, err := h.wardService.GetWard(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "ward not found")
		return
	}

	writeJSON(w, http.StatusOK, ward)
}

// PUT /api/v1/wards/{id}
func (h *WardHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ward id")
		return
	}

	var req updateWardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ward, err := h.wardService.GetWard(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "ward not found")
		return
	}

	if req.Name != nil {
		ward.Name = *req.Name
	}
	if req.Description != nil {
		ward.Description = *req.Description
	}

	if err := h.wardService.UpdateWard(r.Context(), ward); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update ward")
		return
	}

	writeJSON(w, http.StatusOK, ward)
}

// DELETE /api/v1/wards/{id}
func (h *WardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ward id")
		return
	}

	err = h.wardService.DeleteWard(r.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrWardNotEmpty:
			writeError(w, http.StatusConflict, "ward has assigned beds")
		case domain.ErrWardNotFound:
			writeError(w, http.StatusNotFound, "ward not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete ward")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
