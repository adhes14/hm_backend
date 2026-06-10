package service

import (
	"context"

	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/repository"
)

type WardService struct {
	wardRepo repository.WardRepository
	bedRepo  repository.BedRepository
}

func NewWardService(wardRepo repository.WardRepository, bedRepo repository.BedRepository) *WardService {
	return &WardService{
		wardRepo: wardRepo,
		bedRepo:  bedRepo,
	}
}

func (s *WardService) GetAllWards(ctx context.Context) ([]domain.Ward, error) {
	return s.wardRepo.GetAll(ctx)
}

func (s *WardService) GetWard(ctx context.Context, id int) (*domain.Ward, error) {
	return s.wardRepo.GetByID(ctx, id)
}

func (s *WardService) CreateWard(ctx context.Context, w *domain.Ward) error {
	return s.wardRepo.Create(ctx, w)
}

func (s *WardService) UpdateWard(ctx context.Context, w *domain.Ward) error {
	// Verify ward exists
	_, err := s.wardRepo.GetByID(ctx, w.ID)
	if err != nil {
		return domain.ErrWardNotFound
	}
	return s.wardRepo.Update(ctx, w)
}

func (s *WardService) DeleteWard(ctx context.Context, id int) error {
	// Verify ward has no beds
	bedCount, err := s.bedRepo.CountByWardID(ctx, id)
	if err != nil {
		return err
	}
	if bedCount > 0 {
		return domain.ErrWardNotEmpty
	}

	return s.wardRepo.Delete(ctx, id)
}
