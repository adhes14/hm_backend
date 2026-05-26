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

type surgicalScheduleRepo struct {
	pool *pgxpool.Pool
}

func NewSurgicalScheduleRepository(pool *pgxpool.Pool) SurgicalScheduleRepository {
	return &surgicalScheduleRepo{pool: pool}
}

func (r *surgicalScheduleRepo) Create(ctx context.Context, s *domain.SurgicalSchedule) error {
	query := `
		INSERT INTO surgical_schedules (patient_id, procedure_type, scheduled_at, pre_surgical_diagnosis)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		s.PatientID, s.ProcedureType, s.ScheduledAt, s.PreSurgicalDiagnosis,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		// Check for unique key constraint violation
		// postgres error code '23505' is unique_violation
		var pgErr interface {
			SQLState() string
		}
		if errors.As(err, &pgErr) && pgErr.SQLState() == "23505" {
			return domain.ErrPatientAlreadyScheduled
		}
		return err
	}
	return nil
}

func (r *surgicalScheduleRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.SurgicalSchedule, error) {
	query := `
		SELECT s.id, s.patient_id, p.full_name as patient_name, s.procedure_type, 
		       s.scheduled_at, s.pre_surgical_diagnosis, s.created_at, s.updated_at
		FROM surgical_schedules s
		JOIN patients p ON s.patient_id = p.id
		WHERE s.id = $1`

	var s domain.SurgicalSchedule
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.PatientID, &s.PatientName, &s.ProcedureType,
		&s.ScheduledAt, &s.PreSurgicalDiagnosis, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrScheduleNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *surgicalScheduleRepo) GetByPatientID(ctx context.Context, patientID uuid.UUID) (*domain.SurgicalSchedule, error) {
	query := `
		SELECT s.id, s.patient_id, p.full_name as patient_name, s.procedure_type, 
		       s.scheduled_at, s.pre_surgical_diagnosis, s.created_at, s.updated_at
		FROM surgical_schedules s
		JOIN patients p ON s.patient_id = p.id
		WHERE s.patient_id = $1`

	var s domain.SurgicalSchedule
	err := r.pool.QueryRow(ctx, query, patientID).Scan(
		&s.ID, &s.PatientID, &s.PatientName, &s.ProcedureType,
		&s.ScheduledAt, &s.PreSurgicalDiagnosis, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil, nil when no schedule exists for the patient
		}
		return nil, err
	}
	return &s, nil
}

func (r *surgicalScheduleRepo) GetByDateRange(ctx context.Context, from, to time.Time) ([]domain.SurgicalSchedule, error) {
	query := `
		SELECT s.id, s.patient_id, p.full_name as patient_name, s.procedure_type, 
		       s.scheduled_at, s.pre_surgical_diagnosis, s.created_at, s.updated_at
		FROM surgical_schedules s
		JOIN patients p ON s.patient_id = p.id
		WHERE s.scheduled_at >= $1 AND s.scheduled_at < $2
		ORDER BY s.scheduled_at ASC`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schedules := make([]domain.SurgicalSchedule, 0)
	for rows.Next() {
		var s domain.SurgicalSchedule
		err := rows.Scan(
			&s.ID, &s.PatientID, &s.PatientName, &s.ProcedureType,
			&s.ScheduledAt, &s.PreSurgicalDiagnosis, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

func (r *surgicalScheduleRepo) Update(ctx context.Context, s *domain.SurgicalSchedule) error {
	query := `
		UPDATE surgical_schedules
		SET procedure_type = $1, scheduled_at = $2, pre_surgical_diagnosis = $3, updated_at = NOW()
		WHERE id = $4`

	result, err := r.pool.Exec(ctx, query, s.ProcedureType, s.ScheduledAt, s.PreSurgicalDiagnosis, s.ID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrScheduleNotFound
	}
	return nil
}

func (r *surgicalScheduleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM surgical_schedules WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrScheduleNotFound
	}
	return nil
}

func (r *surgicalScheduleRepo) DeleteByPatientID(ctx context.Context, tx pgx.Tx, patientID uuid.UUID) error {
	query := `DELETE FROM surgical_schedules WHERE patient_id = $1`
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, query, patientID)
	} else {
		_, err = r.pool.Exec(ctx, query, patientID)
	}
	return err
}
