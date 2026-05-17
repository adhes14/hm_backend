package repository

import (
	"context"
	"errors"
	"time"

	"github.com/hospital_management/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type admissionRepo struct {
	pool *pgxpool.Pool
}

func NewAdmissionRepository(pool *pgxpool.Pool) AdmissionRepository {
	return &admissionRepo{pool: pool}
}

func (r *admissionRepo) Create(ctx context.Context, a *domain.Admission) error {
	query := `
		INSERT INTO admissions (patient_id, bed_id, status, event_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		a.PatientID, a.BedID, a.Status, a.EventType,
	).Scan(&a.ID, &a.CreatedAt)
}

func (r *admissionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Admission, error) {
	query := `
		SELECT id, patient_id, bed_id, status, event_type, event_at,
		       next_control_at, estimated_discharge_at, created_at, discharged_at
		FROM admissions WHERE id = $1`

	var a domain.Admission
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.PatientID, &a.BedID, &a.Status, &a.EventType,
		&a.EventAt, &a.NextControlAt, &a.EstimatedDischargeAt,
		&a.CreatedAt, &a.DischargedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAdmissionNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *admissionRepo) GetActiveByBedID(ctx context.Context, bedID int) (*domain.Admission, error) {
	query := `
		SELECT id, patient_id, bed_id, status, event_type, event_at,
		       next_control_at, estimated_discharge_at, created_at, discharged_at
		FROM admissions
		WHERE bed_id = $1 AND status = 'active'`

	var a domain.Admission
	err := r.pool.QueryRow(ctx, query, bedID).Scan(
		&a.ID, &a.PatientID, &a.BedID, &a.Status, &a.EventType,
		&a.EventAt, &a.NextControlAt, &a.EstimatedDischargeAt,
		&a.CreatedAt, &a.DischargedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No active admission for this bed
		}
		return nil, err
	}
	return &a, nil
}

func (r *admissionRepo) Discharge(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	tag, err := r.pool.Exec(ctx,
		"UPDATE admissions SET status = 'discharged', discharged_at = $1 WHERE id = $2 AND status = 'active'",
		now, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAdmissionNotActive
	}
	return nil
}

func (r *admissionRepo) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Admission, error) {
	query := `
		SELECT id, patient_id, bed_id, status, event_type, event_at,
		       next_control_at, estimated_discharge_at, created_at, discharged_at
		FROM admissions WHERE id = $1 FOR UPDATE`

	var a domain.Admission
	err := tx.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.PatientID, &a.BedID, &a.Status, &a.EventType,
		&a.EventAt, &a.NextControlAt, &a.EstimatedDischargeAt,
		&a.CreatedAt, &a.DischargedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAdmissionNotFound
		}
		return nil, err
	}
	return &a, nil
}