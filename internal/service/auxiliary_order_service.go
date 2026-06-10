package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/repository"
)

type AuxiliaryOrderService struct {
	repo       repository.AuxiliaryOrderRepository
	sseService SSEService
}

func NewAuxiliaryOrderService(repo repository.AuxiliaryOrderRepository, sseService SSEService) *AuxiliaryOrderService {
	return &AuxiliaryOrderService{
		repo:       repo,
		sseService: sseService,
	}
}

func (s *AuxiliaryOrderService) Create(ctx context.Context, admissionID uuid.UUID, category domain.OrderCategory, description string, staffID *uuid.UUID) (*domain.AuxiliaryOrder, error) {
	order := &domain.AuxiliaryOrder{
		AdmissionID: admissionID,
		Category:    category,
		Description: description,
		Status:      domain.OrderStatusPending,
		CreatedBy:   staffID,
		UpdatedBy:   staffID,
	}

	err := s.repo.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	s.sseService.Broadcast(domain.SSEEvent{
		Type: "orders_updated",
		Data: map[string]interface{}{
			"action": "created",
		},
	})

	return order, nil
}

func (s *AuxiliaryOrderService) ListByAdmission(ctx context.Context, admissionID uuid.UUID) ([]domain.AuxiliaryOrder, error) {
	return s.repo.GetByAdmission(ctx, admissionID)
}

func (s *AuxiliaryOrderService) ListPending(ctx context.Context) ([]domain.AuxiliaryOrder, error) {
	return s.repo.GetAllPending(ctx)
}

func (s *AuxiliaryOrderService) UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus, result string, staffID *uuid.UUID) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Validate status transition
	// Pending -> Done -> Reported
	var valid bool
	switch order.Status {
	case domain.OrderStatusPending:
		valid = (status == domain.OrderStatusDone)
	case domain.OrderStatusDone:
		valid = (status == domain.OrderStatusReported)
	default:
		valid = false
	}

	if !valid {
		return domain.ErrInvalidStatusTransition
	}

	err = s.repo.UpdateStatus(ctx, id, status, result, staffID)
	if err != nil {
		return err
	}

	s.sseService.Broadcast(domain.SSEEvent{
		Type: "orders_updated",
		Data: map[string]interface{}{
			"action": "status_updated",
		},
	})

	return nil
}

func (s *AuxiliaryOrderService) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	s.sseService.Broadcast(domain.SSEEvent{
		Type: "orders_updated",
		Data: map[string]interface{}{
			"action": "deleted",
		},
	})

	return nil
}
