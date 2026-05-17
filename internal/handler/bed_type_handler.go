package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/service"
)

type BedTypeHandler struct {
	bedTypeService *service.BedTypeService
}

func NewBedTypeHandler(bedTypeService *service.BedTypeService) *BedTypeHandler {
	return &BedTypeHandler{bedTypeService: bedTypeService}
}

type createBedTypeRequest struct {
	Name                      string `json:"name"`
	Prefix                    string `json:"prefix"`
	RequiresPostpartumFollowup bool   `json:"requires_postpartum_followup"`
}

type updateBedTypeRequest struct {
	Name                      *string `json:"name,omitempty"`
	Prefix                    *string `json:"prefix,omitempty"`
	RequiresPostpartumFollowup *bool   `json:"requires_postpartum_followup,omitempty"`
}

// POST /api/v1/bed-types
func (h *BedTypeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createBedTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Prefix == "" {
		writeError(w, http.StatusBadRequest, "name and prefix are required")
		return
	}

	bt := &domain.BedType{
		Name:                      req.Name,
		Prefix:                    req.Prefix,
		RequiresPostpartumFollowup: req.RequiresPostpartumFollowup,
	}

	if err := h.bedTypeService.CreateBedType(r.Context(), bt); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create bed type")
		return
	}

	writeJSON(w, http.StatusCreated, bt)
}

// GET /api/v1/bed-types
func (h *BedTypeHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	bedTypes, err := h.bedTypeService.GetAllBedTypes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch bed types")
		return
	}

	if bedTypes == nil {
		bedTypes = []domain.BedType{}
	}

	writeJSON(w, http.StatusOK, bedTypes)
}

// GET /api/v1/bed-types/{id}
func (h *BedTypeHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bed type id")
		return
	}

	bt, err := h.bedTypeService.GetBedType(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "bed type not found")
		return
	}

	writeJSON(w, http.StatusOK, bt)
}

// PUT /api/v1/bed-types/{id}
func (h *BedTypeHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bed type id")
		return
	}

	var req updateBedTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	bt, err := h.bedTypeService.GetBedType(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "bed type not found")
		return
	}

	if req.Name != nil {
		bt.Name = *req.Name
	}
	if req.Prefix != nil {
		bt.Prefix = *req.Prefix
	}
	if req.RequiresPostpartumFollowup != nil {
		bt.RequiresPostpartumFollowup = *req.RequiresPostpartumFollowup
	}

	if err := h.bedTypeService.UpdateBedType(r.Context(), bt); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update bed type")
		return
	}

	writeJSON(w, http.StatusOK, bt)
}

// DELETE /api/v1/bed-types/{id}
func (h *BedTypeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bed type id")
		return
	}

	err = h.bedTypeService.DeleteBedType(r.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrBedTypeInUse:
			writeError(w, http.StatusConflict, "bed type has assigned beds")
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete bed type")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}