package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/middleware"
	"github.com/hospital_management/backend/internal/service"
)

type AuxiliaryOrderHandler struct {
	service *service.AuxiliaryOrderService
}

func NewAuxiliaryOrderHandler(s *service.AuxiliaryOrderService) *AuxiliaryOrderHandler {
	return &AuxiliaryOrderHandler{service: s}
}

type CreateOrderRequest struct {
	Category    string `json:"category"`
	Description string `json:"description"`
}

func (h *AuxiliaryOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	admissionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid admission ID")
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	staffID := claims.StaffID

	order, err := h.service.Create(r.Context(), admissionID, domain.OrderCategory(req.Category), req.Description, &staffID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (h *AuxiliaryOrderHandler) ListByAdmission(w http.ResponseWriter, r *http.Request) {
	admissionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid admission ID")
		return
	}

	orders, err := h.service.ListByAdmission(r.Context(), admissionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if orders == nil {
		orders = []domain.AuxiliaryOrder{}
	}

	writeJSON(w, http.StatusOK, orders)
}

func (h *AuxiliaryOrderHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.ListPending(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if orders == nil {
		orders = []domain.AuxiliaryOrder{}
	}

	writeJSON(w, http.StatusOK, orders)
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

func (h *AuxiliaryOrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	orderID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	var req UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	staffID := claims.StaffID

	err = h.service.UpdateStatus(r.Context(), orderID, domain.OrderStatus(req.Status), &staffID)
	if err != nil {
		if err == domain.ErrOrderNotFound {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "status updated"})
}

func (h *AuxiliaryOrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orderID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	err = h.service.Delete(r.Context(), orderID)
	if err != nil {
		if err == domain.ErrOrderNotFound {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
