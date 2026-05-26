package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/repository"
)

type SurgicalScheduleService struct {
	repo        repository.SurgicalScheduleRepository
	patientRepo repository.PatientRepository
}

func NewSurgicalScheduleService(
	repo repository.SurgicalScheduleRepository,
	patientRepo repository.PatientRepository,
) *SurgicalScheduleService {
	return &SurgicalScheduleService{
		repo:        repo,
		patientRepo: patientRepo,
	}
}

func (s *SurgicalScheduleService) CreateSchedule(
	ctx context.Context,
	patientID uuid.UUID,
	procedureType string,
	scheduledAt time.Time,
	preSurgicalDiagnosis string,
) (*domain.SurgicalSchedule, error) {
	// Verify patient exists
	patient, err := s.patientRepo.GetByID(ctx, patientID)
	if err != nil {
		return nil, domain.ErrPatientNotFound
	}

	// Verify patient is not already admitted
	if patient.IsAdmitted {
		return nil, domain.ErrPatientAlreadyAdmitted
	}

	schedule := &domain.SurgicalSchedule{
		PatientID:            patientID,
		ProcedureType:        procedureType,
		ScheduledAt:          scheduledAt,
		PreSurgicalDiagnosis: preSurgicalDiagnosis,
	}

	err = s.repo.Create(ctx, schedule)
	if err != nil {
		return nil, err
	}

	// Fetch full details (includes patient name)
	return s.repo.GetByID(ctx, schedule.ID)
}

func (s *SurgicalScheduleService) GetByID(ctx context.Context, id uuid.UUID) (*domain.SurgicalSchedule, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SurgicalScheduleService) GetByPatientID(ctx context.Context, patientID uuid.UUID) (*domain.SurgicalSchedule, error) {
	return s.repo.GetByPatientID(ctx, patientID)
}

func (s *SurgicalScheduleService) GetSchedulesByMonth(ctx context.Context, year int, month int) ([]domain.SurgicalSchedule, error) {
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("invalid month: %d", month)
	}

	// Calculate range: from 1st of month 00:00:00 to 1st of next month 00:00:00
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	return s.repo.GetByDateRange(ctx, from, to)
}

func (s *SurgicalScheduleService) GetSchedulesByDate(ctx context.Context, dateStr string) ([]domain.SurgicalSchedule, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, must be YYYY-MM-DD: %w", err)
	}

	// Range of that specific day
	from := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 1)

	return s.repo.GetByDateRange(ctx, from, to)
}

func (s *SurgicalScheduleService) UpdateSchedule(
	ctx context.Context,
	id uuid.UUID,
	procedureType string,
	scheduledAt time.Time,
	preSurgicalDiagnosis string,
) (*domain.SurgicalSchedule, error) {
	// Retrieve existing
	schedule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update values
	schedule.ProcedureType = procedureType
	schedule.ScheduledAt = scheduledAt
	schedule.PreSurgicalDiagnosis = preSurgicalDiagnosis

	err = s.repo.Update(ctx, schedule)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *SurgicalScheduleService) DeleteSchedule(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
