package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClinicalLogService struct {
	pool            *pgxpool.Pool
	clinicalLogRepo repository.ClinicalLogRepository
	admissionRepo   repository.AdmissionRepository
}

func NewClinicalLogService(pool *pgxpool.Pool, clinicalLogRepo repository.ClinicalLogRepository, admissionRepo repository.AdmissionRepository) *ClinicalLogService {
	return &ClinicalLogService{
		pool:            pool,
		clinicalLogRepo: clinicalLogRepo,
		admissionRepo:   admissionRepo,
	}
}

type CreateClinicalLogInput struct {
	PaSystolic   int16
	PaDiastolic  int16
	HeartRate    int16
	RespRate     int16
	Temperature  float32
	Spo2         int16
	PinardStatus bool
	LochiaType   int16
	LochiaAmount int16
	LochiaOdor   bool
	HasClots     bool
	Notes        *string
}

type CreateClinicalLogResponse struct {
	Log           *domain.ClinicalLog `json:"log"`
	NextControlAt *time.Time          `json:"next_control_at"`
}

func (s *ClinicalLogService) CreateClinicalLog(ctx context.Context, admissionID uuid.UUID, input *CreateClinicalLogInput) (*CreateClinicalLogResponse, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	admission, err := s.admissionRepo.GetByIDForUpdate(ctx, tx, admissionID)
	if err != nil {
		return nil, err
	}

	if admission.EventAt == nil {
		return nil, domain.ErrEventRequired
	}

	count, err := s.clinicalLogRepo.CountByAdmission(ctx, tx, admissionID)
	if err != nil {
		return nil, err
	}

	if count >= 8 {
		return nil, domain.ErrControlWindowComplete
	}

	if err := validateVitalSigns(input); err != nil {
		return nil, err
	}

	nextControlAt := CalculateNextControlAt(count+1, time.Now())

	log := &domain.ClinicalLog{
		AdmissionID:  admissionID,
		PaSystolic:   input.PaSystolic,
		PaDiastolic:  input.PaDiastolic,
		HeartRate:    input.HeartRate,
		RespRate:     input.RespRate,
		Temperature:  input.Temperature,
		Spo2:         input.Spo2,
		PinardStatus: input.PinardStatus,
		LochiaType:   input.LochiaType,
		LochiaAmount: input.LochiaAmount,
		LochiaOdor:   input.LochiaOdor,
		HasClots:     input.HasClots,
		Notes:        input.Notes,
	}

	if err := s.clinicalLogRepo.Create(ctx, tx, log); err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, "UPDATE admissions SET next_control_at = $1 WHERE id = $2", nextControlAt, admissionID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &CreateClinicalLogResponse{
		Log:           log,
		NextControlAt: nextControlAt,
	}, nil
}

func (s *ClinicalLogService) ListByAdmission(ctx context.Context, admissionID uuid.UUID) ([]domain.ClinicalLog, error) {
	return s.clinicalLogRepo.ListByAdmission(ctx, admissionID)
}

// CalculateNextControlAt returns the next control time based on control count
func CalculateNextControlAt(controlCount int, now time.Time) *time.Time {
	switch {
	case controlCount >= 8:
		return nil
	case controlCount >= 5:
		t := now.Add(30 * time.Minute)
		return &t
	default: // 1-4
		t := now.Add(15 * time.Minute)
		return &t
	}
}

// validateVitalSigns validates all vital signs are within acceptable ranges
func validateVitalSigns(input *CreateClinicalLogInput) error {
	ranges := domain.DefaultVitalSignRanges()

	if float64(input.PaSystolic) < ranges.PaSystolic.Min || float64(input.PaSystolic) > ranges.PaSystolic.Max {
		return fmt.Errorf("%w: pa_systolic", domain.ErrInvalidVitalSign)
	}
	if float64(input.PaDiastolic) < ranges.PaDiastolic.Min || float64(input.PaDiastolic) > ranges.PaDiastolic.Max {
		return fmt.Errorf("%w: pa_diastolic", domain.ErrInvalidVitalSign)
	}
	if float64(input.HeartRate) < ranges.HeartRate.Min || float64(input.HeartRate) > ranges.HeartRate.Max {
		return fmt.Errorf("%w: heart_rate", domain.ErrInvalidVitalSign)
	}
	if float64(input.RespRate) < ranges.RespRate.Min || float64(input.RespRate) > ranges.RespRate.Max {
		return fmt.Errorf("%w: resp_rate", domain.ErrInvalidVitalSign)
	}
	if float64(input.Temperature) < ranges.Temperature.Min || float64(input.Temperature) > ranges.Temperature.Max {
		return fmt.Errorf("%w: temperature", domain.ErrInvalidVitalSign)
	}
	if float64(input.Spo2) < ranges.Spo2.Min || float64(input.Spo2) > ranges.Spo2.Max {
		return fmt.Errorf("%w: spo2", domain.ErrInvalidVitalSign)
	}

	if input.Notes != nil && len(*input.Notes) > 500 {
		return domain.ErrNotesTooLong
	}

	return nil
}