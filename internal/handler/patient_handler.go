package handler

import (
	"encoding/json"
	"net/http"
	"time"

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