package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/service"
)

type ClinicalLogHandler struct {
	clinicalLogService *service.ClinicalLogService
	admissionService   *service.AdmissionService
}

func NewClinicalLogHandler(clinicalLogService *service.ClinicalLogService, admissionService *service.AdmissionService) *ClinicalLogHandler {
	return &ClinicalLogHandler{
		clinicalLogService: clinicalLogService,
		admissionService:   admissionService,
	}
}

type createClinicalLogRequest struct {
	PaSystolic   int16   `json:"pa_systolic"`
	PaDiastolic  int16   `json:"pa_diastolic"`
	HeartRate    int16   `json:"heart_rate"`
	RespRate     int16   `json:"resp_rate"`
	Temperature float32 `json:"temperature"`
	Spo2         int16   `json:"spo2"`
	PinardStatus bool    `json:"pinard_status"`
	LochiaType   int16   `json:"lochia_type"`
	LochiaAmount int16   `json:"lochia_amount"`
	LochiaOdor   bool    `json:"lochia_odor"`
	HasClots     bool    `json:"has_clots"`
	Notes        *string `json:"notes,omitempty"`
}

// POST /api/v1/admissions/{id}/clinical-logs
func (h *ClinicalLogHandler) Create(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	admissionID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid admission id")
		return
	}

	var req createClinicalLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := &service.CreateClinicalLogInput{
		PaSystolic:   req.PaSystolic,
		PaDiastolic:  req.PaDiastolic,
		HeartRate:    req.HeartRate,
		RespRate:     req.RespRate,
		Temperature:  req.Temperature,
		Spo2:         req.Spo2,
		PinardStatus: req.PinardStatus,
		LochiaType:   req.LochiaType,
		LochiaAmount: req.LochiaAmount,
		LochiaOdor:   req.LochiaOdor,
		HasClots:     req.HasClots,
		Notes:        req.Notes,
	}

	resp, err := h.clinicalLogService.CreateClinicalLog(r.Context(), admissionID, input)
	if err != nil {
		switch err {
		case domain.ErrAdmissionNotFound:
			writeError(w, http.StatusNotFound, "admission not found")
		case domain.ErrEventRequired:
			writeError(w, http.StatusBadRequest, "event must be registered before creating clinical logs")
		case domain.ErrControlWindowComplete:
			writeError(w, http.StatusConflict, "monitoring complete")
		case domain.ErrInvalidVitalSign:
			writeError(w, http.StatusBadRequest, err.Error())
		case domain.ErrNotesTooLong:
			writeError(w, http.StatusBadRequest, "notes exceeds 500 character limit")
		default:
			writeError(w, http.StatusInternalServerError, "failed to create clinical log")
		}
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// GET /api/v1/admissions/{id}/clinical-logs
func (h *ClinicalLogHandler) List(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	admissionID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid admission id")
		return
	}

	logs, err := h.clinicalLogService.ListByAdmission(r.Context(), admissionID)
	if err != nil {
		switch err {
		case domain.ErrAdmissionNotFound:
			writeError(w, http.StatusNotFound, "admission not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to list clinical logs")
		}
		return
	}

	if logs == nil {
		logs = []domain.ClinicalLog{}
	}

	writeJSON(w, http.StatusOK, logs)
}