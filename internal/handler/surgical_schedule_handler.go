package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/service"
)

type SurgicalScheduleHandler struct {
	service *service.SurgicalScheduleService
}

func NewSurgicalScheduleHandler(service *service.SurgicalScheduleService) *SurgicalScheduleHandler {
	return &SurgicalScheduleHandler{service: service}
}

type createSurgicalScheduleRequest struct {
	PatientID            string    `json:"patient_id"`
	ProcedureType        string    `json:"procedure_type"`
	ScheduledAt          time.Time `json:"scheduled_at"`
	PreSurgicalDiagnosis string    `json:"pre_surgical_diagnosis"`
}

type updateSurgicalScheduleRequest struct {
	ProcedureType        string    `json:"procedure_type"`
	ScheduledAt          time.Time `json:"scheduled_at"`
	PreSurgicalDiagnosis string    `json:"pre_surgical_diagnosis"`
}

// POST /api/v1/surgical-schedules
func (h *SurgicalScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createSurgicalScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid patient_id")
		return
	}

	if req.ProcedureType == "" {
		writeError(w, http.StatusBadRequest, "procedure_type is required")
		return
	}

	if req.ScheduledAt.IsZero() {
		writeError(w, http.StatusBadRequest, "scheduled_at is required")
		return
	}

	schedule, err := h.service.CreateSchedule(r.Context(), patientID, req.ProcedureType, req.ScheduledAt, req.PreSurgicalDiagnosis)
	if err != nil {
		switch err {
		case domain.ErrPatientNotFound:
			writeError(w, http.StatusNotFound, "patient not found")
		case domain.ErrPatientAlreadyAdmitted:
			writeError(w, http.StatusConflict, "patient is already admitted")
		case domain.ErrPatientAlreadyScheduled:
			writeError(w, http.StatusConflict, "patient already has a surgical schedule")
		default:
			writeError(w, http.StatusInternalServerError, "failed to create surgical schedule")
		}
		return
	}

	writeJSON(w, http.StatusCreated, schedule)
}

// GET /api/v1/surgical-schedules?year=&month=
func (h *SurgicalScheduleHandler) ListByMonth(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	if yearStr == "" || monthStr == "" {
		writeError(w, http.StatusBadRequest, "year and month parameters are required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid year parameter")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid month parameter")
		return
	}

	schedules, err := h.service.GetSchedulesByMonth(r.Context(), year, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, schedules)
}

// GET /api/v1/surgical-schedules/date?date=YYYY-MM-DD
func (h *SurgicalScheduleHandler) ListByDate(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		writeError(w, http.StatusBadRequest, "date parameter is required")
		return
	}

	schedules, err := h.service.GetSchedulesByDate(r.Context(), dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, schedules)
}

// GET /api/v1/surgical-schedules/patient/{patientId}
func (h *SurgicalScheduleHandler) GetByPatientID(w http.ResponseWriter, r *http.Request) {
	patientIDStr := chi.URLParam(r, "patientId")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid patient id")
		return
	}

	schedule, err := h.service.GetByPatientID(r.Context(), patientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query surgical schedule")
		return
	}

	if schedule == nil {
		writeJSON(w, http.StatusOK, nil)
		return
	}

	writeJSON(w, http.StatusOK, schedule)
}

// PUT /api/v1/surgical-schedules/{id}
func (h *SurgicalScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid surgical schedule id")
		return
	}

	var req updateSurgicalScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ProcedureType == "" {
		writeError(w, http.StatusBadRequest, "procedure_type is required")
		return
	}

	if req.ScheduledAt.IsZero() {
		writeError(w, http.StatusBadRequest, "scheduled_at is required")
		return
	}

	schedule, err := h.service.UpdateSchedule(r.Context(), id, req.ProcedureType, req.ScheduledAt, req.PreSurgicalDiagnosis)
	if err != nil {
		switch err {
		case domain.ErrScheduleNotFound:
			writeError(w, http.StatusNotFound, "surgical schedule not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update surgical schedule")
		}
		return
	}

	writeJSON(w, http.StatusOK, schedule)
}

// DELETE /api/v1/surgical-schedules/{id}
func (h *SurgicalScheduleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid surgical schedule id")
		return
	}

	err = h.service.DeleteSchedule(r.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrScheduleNotFound:
			writeError(w, http.StatusNotFound, "surgical schedule not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete surgical schedule")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
