package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdmissionService struct {
	pool         *pgxpool.Pool
	admissionRepo repository.AdmissionRepository
	bedRepo      repository.BedRepository
	patientRepo  repository.PatientRepository
}

func NewAdmissionService(
	pool *pgxpool.Pool,
	admissionRepo repository.AdmissionRepository,
	bedRepo repository.BedRepository,
	patientRepo repository.PatientRepository,
) *AdmissionService {
	return &AdmissionService{
		pool:         pool,
		admissionRepo: admissionRepo,
		bedRepo:      bedRepo,
		patientRepo:  patientRepo,
	}
}

// CreateAdmission assigns a patient to an available bed
// Uses a transaction to ensure atomicity
func (s *AdmissionService) CreateAdmission(ctx context.Context, patientID uuid.UUID, bedID int) (*domain.Admission, error) {
	// Verify patient exists
	if _, err := s.patientRepo.GetByID(ctx, patientID); err != nil {
		return nil, domain.ErrPatientNotFound
	}

	// Verify bed exists
	bed, err := s.bedRepo.GetByID(ctx, bedID)
	if err != nil {
		return nil, domain.ErrBedNotFound
	}

	// Verify bed is available
	if !bed.IsAvailable() {
		return nil, domain.ErrBedNotAvailable
	}

	// Use transaction for atomic operation
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Create admission
	admission := &domain.Admission{
		PatientID: patientID,
		BedID:     bedID,
		Status:    domain.AdmissionStatusActive,
		EventType: domain.EventTypeNinguno,
	}

	// Insert admission within transaction
	insertQuery := `
		INSERT INTO admissions (patient_id, bed_id, status, event_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	err = tx.QueryRow(ctx, insertQuery,
		admission.PatientID, admission.BedID, admission.Status, admission.EventType,
	).Scan(&admission.ID, &admission.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Update bed's current_admission_id within same transaction
	updateQuery := `UPDATE beds SET current_admission_id = $1 WHERE id = $2`
	_, err = tx.Exec(ctx, updateQuery, admission.ID, bedID)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return admission, nil
}

// DischargeAdmission releases a bed and marks admission as discharged
// Uses a transaction to ensure atomicity
func (s *AdmissionService) DischargeAdmission(ctx context.Context, admissionID uuid.UUID) error {
	// Get admission
	admission, err := s.admissionRepo.GetByID(ctx, admissionID)
	if err != nil {
		return domain.ErrAdmissionNotFound
	}

	// Verify admission is active
	if admission.Status != domain.AdmissionStatusActive {
		return domain.ErrAdmissionNotActive
	}

	// Use transaction for atomic discharge
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update admission status
	now := time.Now()
	updateAdmission := `
		UPDATE admissions
		SET status = 'discharged', discharged_at = $1
		WHERE id = $2 AND status = 'active'`

	tag, err := tx.Exec(ctx, updateAdmission, now, admissionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAdmissionNotActive
	}

	// Clear bed's current_admission_id
	updateBed := `UPDATE beds SET current_admission_id = NULL WHERE id = $1`
	_, err = tx.Exec(ctx, updateBed, admission.BedID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}