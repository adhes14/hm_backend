package service

import (
	"context"

	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/repository"
)

type BedTypeService struct {
	bedTypeRepo repository.BedTypeRepository
}

func NewBedTypeService(bedTypeRepo repository.BedTypeRepository) *BedTypeService {
	return &BedTypeService{
		bedTypeRepo: bedTypeRepo,
	}
}

func (s *BedTypeService) CreateBedType(ctx context.Context, bt *domain.BedType) error {
	return s.bedTypeRepo.Create(ctx, bt)
}

func (s *BedTypeService) GetBedType(ctx context.Context, id int) (*domain.BedType, error) {
	return s.bedTypeRepo.GetByID(ctx, id)
}

func (s *BedTypeService) GetAllBedTypes(ctx context.Context) ([]domain.BedType, error) {
	return s.bedTypeRepo.GetAll(ctx)
}

func (s *BedTypeService) UpdateBedType(ctx context.Context, bt *domain.BedType) error {
	return s.bedTypeRepo.Update(ctx, bt)
}

func (s *BedTypeService) DeleteBedType(ctx context.Context, id int) error {
	count, err := s.bedTypeRepo.CountByID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrBedTypeInUse
	}
	return s.bedTypeRepo.Delete(ctx, id)
}