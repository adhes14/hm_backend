package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/middleware"
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
	PatientID          string `json:"patient_id"`
	BedID              int    `json:"bed_id"`
	AdmissionDiagnosis string `json:"admission_diagnosis"`
}

type updateDiagnosisRequest struct {
	CurrentDiagnosis string `json:"current_diagnosis"`
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

	admission, err := h.admissionService.CreateAdmission(r.Context(), patientID, req.BedID, req.AdmissionDiagnosis)
	if err != nil {
		switch err {
		case domain.ErrPatientNotFound:
			writeError(w, http.StatusNotFound, "patient not found")
		case domain.ErrBedNotFound:
			writeError(w, http.StatusNotFound, "bed not found")
		case domain.ErrBedNotAvailable:
			writeError(w, http.StatusConflict, "bed is not available")
		case domain.ErrPatientAlreadyAdmitted:
			writeError(w, http.StatusConflict, "patient is already admitted")
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

// PUT /api/v1/admissions/{id}/diagnosis
func (h *AdmissionHandler) UpdateDiagnosis(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid admission id")
		return
	}

	var req updateDiagnosisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	staffID := claims.StaffID

	admission, err := h.admissionService.UpdateDiagnosis(r.Context(), id, req.CurrentDiagnosis, staffID)
	if err != nil {
		switch err {
		case domain.ErrAdmissionNotFound:
			writeError(w, http.StatusNotFound, "admission not found")
		case domain.ErrAdmissionNotActive:
			writeError(w, http.StatusConflict, "admission is not active")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update diagnosis")
		}
		return
	}

	writeJSON(w, http.StatusOK, admission)
}

// GET /api/v1/admissions/{id}
func (h *AdmissionHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid admission id")
		return
	}

	admission, err := h.admissionService.GetByID(r.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrAdmissionNotFound:
			writeError(w, http.StatusNotFound, "admission not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to get admission")
		}
		return
	}

	writeJSON(w, http.StatusOK, admission)
}

type registerEventRequest struct {
	EventType string `json:"event_type"`
}

// PUT /api/v1/admissions/{id}/event
func (h *AdmissionHandler) RegisterEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid admission id")
		return
	}

	var req registerEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.EventType != "parto" && req.EventType != "cesarea" {
		writeError(w, http.StatusBadRequest, "invalid event_type")
		return
	}

	eventType := domain.EventType(req.EventType)
	admission, err := h.admissionService.RegisterEvent(r.Context(), id, eventType)
	if err != nil {
		switch err {
		case domain.ErrAdmissionNotFound:
			writeError(w, http.StatusNotFound, "admission not found")
		case domain.ErrAdmissionNotActive:
			writeError(w, http.StatusConflict, "admission is not active")
		case domain.ErrEventAlreadyRegistered:
			writeError(w, http.StatusConflict, "event already registered")
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, admission)
}