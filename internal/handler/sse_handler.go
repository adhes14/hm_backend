package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hospital_management/backend/internal/middleware"
	"github.com/hospital_management/backend/internal/service"
)

type SSEHandler struct {
	sseService service.SSEService
}

func NewSSEHandler(sseService service.SSEService) *SSEHandler {
	return &SSEHandler{sseService: sseService}
}

func (h *SSEHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ticket, err := h.sseService.CreateTicket(r.Context(), claims.Username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create ticket: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"ticket": ticket})
}

func (h *SSEHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	ticketStr := r.URL.Query().Get("ticket")
	if ticketStr == "" {
		http.Error(w, "missing sse ticket", http.StatusBadRequest)
		return
	}

	ticket, err := h.sseService.ValidateTicket(r.Context(), ticketStr)
	if err != nil {
		http.Error(w, "invalid or expired sse ticket", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch, cleanup := h.sseService.RegisterClient(ticket.Username)
	defer cleanup()

	// Send connection acknowledgement
	fmt.Fprintf(w, "event: connected\ndata: {\"username\":\"%s\"}\n\n", ticket.Username)
	flusher.Flush()

	keepAliveTicker := time.NewTicker(15 * time.Second)
	defer keepAliveTicker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-keepAliveTicker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case event, ok := <-ch:
			if !ok {
				return
			}
			dataBytes, err := json.Marshal(event.Data)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(dataBytes))
			flusher.Flush()
		}
	}
}
