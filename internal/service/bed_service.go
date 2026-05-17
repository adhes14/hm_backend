package service

import (
	"context"

	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/repository"
)

type BedService struct {
	bedRepo       repository.BedRepository
	patientRepo   repository.PatientRepository
	admissionRepo repository.AdmissionRepository
}

func NewBedService(bedRepo repository.BedRepository, patientRepo repository.PatientRepository, admissionRepo repository.AdmissionRepository) *BedService {
	return &BedService{
		bedRepo:       bedRepo,
		patientRepo:   patientRepo,
		admissionRepo: admissionRepo,
	}
}

// GetAllBeds returns all beds with their current status
func (s *BedService) GetAllBeds(ctx context.Context) ([]domain.Bed, error) {
	return s.bedRepo.GetAll(ctx)
}

// GetBed returns a single bed by ID
func (s *BedService) GetBed(ctx context.Context, id int) (*domain.Bed, error) {
	return s.bedRepo.GetByID(ctx, id)
}

// CreateBed creates a new bed
func (s *BedService) CreateBed(ctx context.Context, bed *domain.Bed) error {
	return s.bedRepo.CreateBed(ctx, bed)
}

// UpdateBed updates an existing bed
func (s *BedService) UpdateBed(ctx context.Context, bed *domain.Bed) error {
	return s.bedRepo.UpdateBed(ctx, bed)
}

// DeleteBed deletes a bed only if it's not occupied
func (s *BedService) DeleteBed(ctx context.Context, id int) error {
	// Check if bed has an active admission
	admission, err := s.admissionRepo.GetByBedID(ctx, id)
	if err != nil {
		return err
	}
	if admission != nil && admission.Status == domain.AdmissionStatusActive {
		return domain.ErrBedOccupied
	}
	return s.bedRepo.DeleteBed(ctx, id)
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