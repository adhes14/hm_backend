package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/repository"
)

type PatientService struct {
	patientRepo repository.PatientRepository
}

func NewPatientService(patientRepo repository.PatientRepository) *PatientService {
	return &PatientService{
		patientRepo: patientRepo,
	}
}

// CreatePatient creates a new patient with validation
func (s *PatientService) CreatePatient(ctx context.Context, identityNumber, fullName string, birthDate time.Time, obstetricHistory json.RawMessage) (*domain.Patient, error) {
	// Validate required fields
	if identityNumber == "" || fullName == "" {
		return nil, domain.ErrPatientNotFound
	}

	// Validate obstetric_history is valid JSON
	if len(obstetricHistory) == 0 {
		obstetricHistory = json.RawMessage(`{}`)
	} else if !json.Valid(obstetricHistory) {
		return nil, domain.ErrPatientNotFound
	}

	patient := &domain.Patient{
		IdentityNumber:   identityNumber,
		FullName:         fullName,
		BirthDate:        birthDate,
		ObstetricHistory: obstetricHistory,
	}

	if err := s.patientRepo.Create(ctx, patient); err != nil {
		return nil, domain.ErrPatientExists
	}

	return patient, nil
}

// SearchPatients searches by identity_number (exact) or full_name (ILIKE)
func (s *PatientService) SearchPatients(ctx context.Context, query string) ([]domain.Patient, error) {
	if query == "" {
		return []domain.Patient{}, nil
	}
	return s.patientRepo.Search(ctx, query)
}

// GetPatientByID returns a patient by UUID
func (s *PatientService) GetPatientByID(ctx context.Context, id uuid.UUID) (*domain.Patient, error) {
	return s.patientRepo.GetByID(ctx, id)
}