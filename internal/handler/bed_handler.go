package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/service"
)

type BedHandler struct {
	bedService       *service.BedService
	admissionService *service.AdmissionService
}

// NewBedHandler creates a BedHandler with the given services.
// The second parameter (admissionService) was added for the bed history endpoint.
func NewBedHandler(bedService *service.BedService, admissionService *service.AdmissionService) *BedHandler {
	return &BedHandler{bedService: bedService, admissionService: admissionService}
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

type bedAdmissionsPaginatedResponse struct {
	Data       []domain.AdmissionWithDetails `json:"data"`
	Total      int                           `json:"total"`
	Page       int                           `json:"page"`
	Limit      int                           `json:"limit"`
	TotalPages int                           `json:"total_pages"`
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

// GET /api/v1/beds/{id}/admissions — returns paginated discharged admissions for a bed
func (h *BedHandler) ListBedAdmissions(w http.ResponseWriter, r *http.Request) {
	// Parse bed ID from path
	idStr := chi.URLParam(r, "id")
	bedID, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bed id")
		return
	}

	// Verify bed exists
	if _, err := h.bedService.GetBed(r.Context(), bedID); err != nil {
		writeError(w, http.StatusNotFound, "bed not found")
		return
	}

	// Parse query params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var fromTime, toTime *time.Time
	arLocation, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		// Fallback to UTC if tzdata is missing (extremely unlikely)
		arLocation = time.UTC
	}

	fromStr := r.URL.Query().Get("from")
	if fromStr != "" {
		parsed, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
			return
		}
		t := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, arLocation)
		fromTime = &t
	}

	toStr := r.URL.Query().Get("to")
	if toStr != "" {
		parsed, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
			return
		}
		t := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999_000_000, arLocation)
		toTime = &t
	}

	admissions, total, err := h.admissionService.ListDischargedByBedID(r.Context(), bedID, fromTime, toTime, page, limit)
	if err != nil {
		switch err {
		case domain.ErrInvalidDateRange:
			writeError(w, http.StatusBadRequest, "from date must not exceed to date")
		default:
			writeError(w, http.StatusInternalServerError, "failed to fetch bed history")
		}
		return
	}

	if admissions == nil {
		admissions = []domain.AdmissionWithDetails{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	writeJSON(w, http.StatusOK, bedAdmissionsPaginatedResponse{
		Data:       admissions,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}