package handler

import (
	"encoding/json"
	"net/http"

	"github.com/hospital_management/backend/internal/service"
)

type SettingsHandler struct {
	sseService service.SSEService
}

func NewSettingsHandler(sseService service.SSEService) *SettingsHandler {
	return &SettingsHandler{sseService: sseService}
}

func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.sseService.GetSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get settings: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.sseService.UpdateSettings(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update settings: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "settings updated successfully"})
}
