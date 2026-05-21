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

type AdmissionService struct {
	pool          *pgxpool.Pool
	admissionRepo repository.AdmissionRepository
	bedRepo       repository.BedRepository
	patientRepo   repository.PatientRepository
	sseService    SSEService
}

func NewAdmissionService(
	pool *pgxpool.Pool,
	admissionRepo repository.AdmissionRepository,
	bedRepo repository.BedRepository,
	patientRepo repository.PatientRepository,
	sseService SSEService,
) *AdmissionService {
	return &AdmissionService{
		pool:          pool,
		admissionRepo: admissionRepo,
		bedRepo:       bedRepo,
		patientRepo:   patientRepo,
		sseService:    sseService,
	}
}

// GetByID returns a single admission by ID
func (s *AdmissionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Admission, error) {
	admission, err := s.admissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrAdmissionNotFound
	}
	return admission, nil
}

// CreateAdmission assigns a patient to an available bed
// Uses a transaction to ensure atomicity
func (s *AdmissionService) CreateAdmission(ctx context.Context, patientID uuid.UUID, bedID int) (*domain.Admission, error) {
	// Verify patient exists
	if _, err := s.patientRepo.GetByID(ctx, patientID); err != nil {
		return nil, domain.ErrPatientNotFound
	}

	// Verify patient doesn't have an active admission
	activeAdmission, err := s.admissionRepo.GetActiveByPatientID(ctx, patientID)
	if err != nil {
		return nil, err
	}
	if activeAdmission != nil {
		return nil, domain.ErrPatientAlreadyAdmitted
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

	// Read sound setting
	sound := false
	settings, err := s.sseService.GetSettings(ctx)
	if err == nil {
		if val, ok := settings["sound_alert_patient_admitted"]; ok && val == "true" {
			sound = true
		}
	}

	// Broadcast bed update
	s.sseService.Broadcast(domain.SSEEvent{
		Type: "bed_updated",
		Data: map[string]interface{}{
			"bed_id": bedID,
			"action": "admitted",
			"sound":  sound,
		},
	})

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
	now := time.Now().UTC()
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

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	// Read sound setting
	sound := false
	settings, err := s.sseService.GetSettings(ctx)
	if err == nil {
		if val, ok := settings["sound_alert_patient_discharged"]; ok && val == "true" {
			sound = true
		}
	}

	// Broadcast bed update and clear alert
	s.sseService.Broadcast(domain.SSEEvent{
		Type: "bed_updated",
		Data: map[string]interface{}{
			"bed_id": admission.BedID,
			"action": "discharged",
			"sound":  sound,
		},
	})
	s.sseService.Broadcast(domain.SSEEvent{
		Type: "alert_cleared",
		Data: map[string]interface{}{
			"admission_id": admissionID.String(),
		},
	})

	return nil
}

// RegisterEvent registers a birth event (parto or cesarea) on an admission
func (s *AdmissionService) RegisterEvent(ctx context.Context, admissionID uuid.UUID, eventType domain.EventType) (*domain.Admission, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	admission, err := s.admissionRepo.GetByIDForUpdate(ctx, tx, admissionID)
	if err != nil {
		return nil, err
	}

	if admission.Status != domain.AdmissionStatusActive {
		return nil, domain.ErrAdmissionNotActive
	}

	if admission.EventAt != nil {
		return nil, domain.ErrEventAlreadyRegistered
	}

	if eventType != domain.EventTypeParto && eventType != domain.EventTypeCesarea {
		return nil, fmt.Errorf("invalid event_type: %s", eventType)
	}

	now := time.Now().UTC()
	var estimatedDischargeAt time.Time
	switch eventType {
	case domain.EventTypeParto:
		estimatedDischargeAt = now.Add(24 * time.Hour)
	case domain.EventTypeCesarea:
		estimatedDischargeAt = now.Add(48 * time.Hour)
	}

	// Note: RegisterEvent only sets estimated_discharge_at, NOT next_control_at
	// First clinical log is what starts the follow-up chain
	_, err = tx.Exec(ctx,
		"UPDATE admissions SET event_at = $1, estimated_discharge_at = $2, event_type = $3 WHERE id = $4",
		now, estimatedDischargeAt, eventType, admissionID)
	if err != nil {
		return nil, err
	}

	admission.EventAt = &now
	admission.EstimatedDischargeAt = &estimatedDischargeAt
	admission.EventType = eventType

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Broadcast bed update
	s.sseService.Broadcast(domain.SSEEvent{
		Type: "bed_updated",
		Data: map[string]interface{}{
			"bed_id": admission.BedID,
		},
	})

	return admission, nil
}