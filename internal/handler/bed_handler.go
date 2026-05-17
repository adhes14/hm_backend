package handler

import (
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

// GET /api/v1/beds
func (h *BedHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	beds, err := h.bedService.GetAllBeds(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch beds")
		return
	}

	writeJSON(w, http.StatusOK, beds)
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