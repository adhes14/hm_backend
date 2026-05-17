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

type PatientHandler struct {
	patientService *service.PatientService
}

func NewPatientHandler(patientService *service.PatientService) *PatientHandler {
	return &PatientHandler{patientService: patientService}
}

type createPatientRequest struct {
	IdentityNumber   string          `json:"identity_number"`
	FullName         string          `json:"full_name"`
	BirthDate        string          `json:"birth_date"` // ISO 8601 format
	ObstetricHistory json.RawMessage `json:"obstetric_history"`
}

type updatePatientRequest struct {
	IdentityNumber   string          `json:"identity_number"`
	FullName         string          `json:"full_name"`
	BirthDate        string          `json:"birth_date"`
	ObstetricHistory json.RawMessage `json:"obstetric_history"`
}

type PaginatedResponse struct {
	Data       []domain.Patient `json:"data"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

// POST /api/v1/patients
func (h *PatientHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createPatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Parse birth_date
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid birth_date format, use YYYY-MM-DD")
		return
	}

	patient, err := h.patientService.CreatePatient(
		r.Context(), req.IdentityNumber, req.FullName, birthDate, req.ObstetricHistory)
	if err != nil {
		switch err {
		case domain.ErrPatientExists:
			writeError(w, http.StatusConflict, "patient with this identity number already exists")
		default:
			writeError(w, http.StatusBadRequest, "failed to create patient")
		}
		return
	}

	writeJSON(w, http.StatusCreated, patient)
}

// GET /api/v1/patients?page=1&limit=10
func (h *PatientHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	patients, total, err := h.patientService.ListPatients(r.Context(), page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list patients")
		return
	}

	if patients == nil {
		patients = []domain.Patient{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:       patients,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// GET /api/v1/patients/search?q=
func (h *PatientHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	patients, err := h.patientService.SearchPatients(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to search patients")
		return
	}

	if patients == nil {
		patients = []domain.Patient{}
	}

	writeJSON(w, http.StatusOK, patients)
}

// GET /api/v1/patients/{id}
func (h *PatientHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := parseUUID(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid patient id")
		return
	}

	patient, err := h.patientService.GetPatientByID(r.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrPatientNotFound:
			writeError(w, http.StatusNotFound, "patient not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to retrieve patient")
		}
		return
	}

	writeJSON(w, http.StatusOK, patient)
}

// PUT /api/v1/patients/{id}
func (h *PatientHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := parseUUID(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid patient id")
		return
	}

	var req updatePatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid birth_date format, use YYYY-MM-DD")
		return
	}

	patient, err := h.patientService.UpdatePatient(
		r.Context(), id, req.IdentityNumber, req.FullName, birthDate, req.ObstetricHistory)
	if err != nil {
		switch err {
		case domain.ErrPatientNotFound:
			writeError(w, http.StatusNotFound, "patient not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update patient")
		}
		return
	}

	writeJSON(w, http.StatusOK, patient)
}