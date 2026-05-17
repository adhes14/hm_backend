package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/service"
)

type BedHandler struct {
	bedService *service.BedService
}

func NewBedHandler(bedService *service.BedService) *BedHandler {
	return &BedHandler{bedService: bedService}
}

type createBedRequest struct {
	Number     int `json:"number"`
	BedTypeID  int `json:"bed_type_id"`
	IsActive   bool `json:"is_active"`
}

type updateBedRequest struct {
	Number    *int `json:"number,omitempty"`
	BedTypeID *int `json:"bed_type_id,omitempty"`
}

// GET /api/v1/beds
func (h *BedHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	beds, err := h.bedService.GetAllBeds(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch beds")
		return
	}

	writeJSON(w, http.StatusOK, beds)
}

// GET /api/v1/beds/{id}
func (h *BedHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bed id")
		return
	}

	bed, err := h.bedService.GetBed(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "bed not found")
		return
	}

	writeJSON(w, http.StatusOK, bed)
}

// POST /api/v1/beds
func (h *BedHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createBedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Number <= 0 || req.BedTypeID <= 0 {
		writeError(w, http.StatusBadRequest, "number and bed_type_id are required")
		return
	}

	bed := &domain.Bed{
		Number:   req.Number,
		IsActive: req.IsActive,
	}
	bed.BedType = &domain.BedType{ID: req.BedTypeID}

	if err := h.bedService.CreateBed(r.Context(), bed); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create bed")
		return
	}

	writeJSON(w, http.StatusCreated, bed)
}

// PUT /api/v1/beds/{id}
func (h *BedHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bed id")
		return
	}

	var req updateBedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	bed, err := h.bedService.GetBed(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "bed not found")
		return
	}

	if req.Number != nil {
		bed.Number = *req.Number
	}
	if req.BedTypeID != nil {
		bed.BedType.ID = *req.BedTypeID
	}

	if err := h.bedService.UpdateBed(r.Context(), bed); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update bed")
		return
	}

	writeJSON(w, http.StatusOK, bed)
}

// DELETE /api/v1/beds/{id}
func (h *BedHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bed id")
		return
	}

	err = h.bedService.DeleteBed(r.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrBedOccupied:
			writeError(w, http.StatusConflict, "bed is occupied")
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete bed")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/beds/{id}/patient
func (h *BedHandler) GetPatient(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bed id")
		return
	}

	patient, err := h.bedService.GetBedPatient(r.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrBedNotFound:
			writeError(w, http.StatusNotFound, "bed not found or available")
		default:
			writeError(w, http.StatusInternalServerError, "failed to fetch patient")
		}
		return
	}

	writeJSON(w, http.StatusOK, patient)
}