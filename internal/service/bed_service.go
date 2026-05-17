package service

import (
	"context"

	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/repository"
)

type BedService struct {
	bedRepo     repository.BedRepository
	patientRepo repository.PatientRepository
}

func NewBedService(bedRepo repository.BedRepository, patientRepo repository.PatientRepository) *BedService {
	return &BedService{
		bedRepo:     bedRepo,
		patientRepo: patientRepo,
	}
}

// GetAllBeds returns all beds with their current status
func (s *BedService) GetAllBeds(ctx context.Context) ([]domain.Bed, error) {
	return s.bedRepo.GetAll(ctx)
}

// GetBedPatient returns the patient assigned to a bed (for occupied beds)
func (s *BedService) GetBedPatient(ctx context.Context, bedID int) (*domain.Patient, error) {
	bed, err := s.bedRepo.GetByID(ctx, bedID)
	if err != nil {
		return nil, domain.ErrBedNotFound
	}

	if bed.IsAvailable() {
		return nil, domain.ErrBedNotFound // Bed is available, no patient
	}

	return s.patientRepo.GetByBedID(ctx, bedID)
}