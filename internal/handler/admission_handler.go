package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/service"
	"github.com/google/uuid"
)

type AdmissionHandler struct {
	admissionService *service.AdmissionService
}

func NewAdmissionHandler(admissionService *service.AdmissionService) *AdmissionHandler {
	return &AdmissionHandler{admissionService: admissionService}
}

type createAdmissionRequest struct {
	PatientID string `json:"patient_id"`
	BedID     int    `json:"bed_id"`
}

// POST /api/v1/admissions
func (h *AdmissionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createAdmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid patient_id")
		return
	}

	admission, err := h.admissionService.CreateAdmission(r.Context(), patientID, req.BedID)
	if err != nil {
		switch err {
		case domain.ErrPatientNotFound:
			writeError(w, http.StatusNotFound, "patient not found")
		case domain.ErrBedNotFound:
			writeError(w, http.StatusNotFound, "bed not found")
		case domain.ErrBedNotAvailable:
			writeError(w, http.StatusConflict, "bed is not available")
		default:
			writeError(w, http.StatusInternalServerError, "failed to create admission")
		}
		return
	}

	writeJSON(w, http.StatusCreated, admission)
}

// PUT /api/v1/admissions/{id}/discharge
func (h *AdmissionHandler) Discharge(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid admission id")
		return
	}

	if err := h.admissionService.DischargeAdmission(r.Context(), id); err != nil {
		switch err {
		case domain.ErrAdmissionNotFound:
			writeError(w, http.StatusNotFound, "admission not found")
		case domain.ErrAdmissionNotActive:
			writeError(w, http.StatusConflict, "admission is not active")
		default:
			writeError(w, http.StatusInternalServerError, "failed to discharge admission")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "discharged"})
}