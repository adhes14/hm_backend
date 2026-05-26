package repository

import (
	"context"
	"errors"

	"github.com/hospital_management/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type patientRepo struct {
	pool *pgxpool.Pool
}

func NewPatientRepository(pool *pgxpool.Pool) PatientRepository {
	return &patientRepo{pool: pool}
}

func (r *patientRepo) Create(ctx context.Context, p *domain.Patient) error {
	query := `
		INSERT INTO patients (identity_number, full_name, birth_date, obstetric_history)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	return r.pool.QueryRow(ctx, query,
		p.IdentityNumber, p.FullName, p.BirthDate, p.ObstetricHistory,
	).Scan(&p.ID)
}

func (r *patientRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Patient, error) {
	query := `
		SELECT p.id, p.identity_number, p.full_name, p.birth_date, p.obstetric_history,
		       EXISTS (SELECT 1 FROM admissions a WHERE a.patient_id = p.id AND a.status = 'active') AS is_admitted,
		       (SELECT a.id FROM admissions a WHERE a.patient_id = p.id AND a.status = 'active' LIMIT 1) AS current_admission_id,
		       s.scheduled_at
		FROM patients p
		LEFT JOIN surgical_schedules s ON s.patient_id = p.id
		WHERE p.id = $1`

	var p domain.Patient
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.IdentityNumber, &p.FullName, &p.BirthDate, &p.ObstetricHistory,
		&p.IsAdmitted, &p.CurrentAdmissionID, &p.ScheduledAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPatientNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *patientRepo) GetByBedID(ctx context.Context, bedID int) (*domain.Patient, error) {
	query := `
		SELECT p.id, p.identity_number, p.full_name, p.birth_date, p.obstetric_history,
		       true AS is_admitted,
		       a.id AS current_admission_id
		FROM patients p
		JOIN admissions a ON a.patient_id = p.id
		WHERE a.bed_id = $1 AND a.status = 'active'`

	var p domain.Patient
	err := r.pool.QueryRow(ctx, query, bedID).Scan(
		&p.ID, &p.IdentityNumber, &p.FullName, &p.BirthDate, &p.ObstetricHistory,
		&p.IsAdmitted, &p.CurrentAdmissionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPatientNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *patientRepo) Search(ctx context.Context, query string) ([]domain.Patient, error) {
	sql := `
		SELECT p.id, p.identity_number, p.full_name, p.birth_date, p.obstetric_history,
		       EXISTS (SELECT 1 FROM admissions a WHERE a.patient_id = p.id AND a.status = 'active') AS is_admitted,
		       (SELECT a.id FROM admissions a WHERE a.patient_id = p.id AND a.status = 'active' LIMIT 1) AS current_admission_id,
		       s.scheduled_at
		FROM patients p
		LEFT JOIN surgical_schedules s ON s.patient_id = p.id
		WHERE p.identity_number = $1 OR p.full_name ILIKE '%' || $1 || '%'
		LIMIT 20`

	rows, err := r.pool.Query(ctx, sql, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	patients := make([]domain.Patient, 0)
	for rows.Next() {
		var p domain.Patient
		err := rows.Scan(&p.ID, &p.IdentityNumber, &p.FullName, &p.BirthDate, &p.ObstetricHistory,
			&p.IsAdmitted, &p.CurrentAdmissionID, &p.ScheduledAt)
		if err != nil {
			return nil, err
		}
		patients = append(patients, p)
	}
	return patients, rows.Err()
}

func (r *patientRepo) List(ctx context.Context, page, limit int) ([]domain.Patient, int, error) {
	offset := (page - 1) * limit

	// Get total count
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM patients").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	sql := `
		SELECT p.id, p.identity_number, p.full_name, p.birth_date, p.obstetric_history,
		       EXISTS (SELECT 1 FROM admissions a WHERE a.patient_id = p.id AND a.status = 'active') AS is_admitted,
		       (SELECT a.id FROM admissions a WHERE a.patient_id = p.id AND a.status = 'active' LIMIT 1) AS current_admission_id,
		       s.scheduled_at
		FROM patients p
		LEFT JOIN surgical_schedules s ON s.patient_id = p.id
		ORDER BY p.full_name ASC
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, sql, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	patients := make([]domain.Patient, 0)
	for rows.Next() {
		var p domain.Patient
		err := rows.Scan(&p.ID, &p.IdentityNumber, &p.FullName, &p.BirthDate, &p.ObstetricHistory,
			&p.IsAdmitted, &p.CurrentAdmissionID, &p.ScheduledAt)
		if err != nil {
			return nil, 0, err
		}
		patients = append(patients, p)
	}
	return patients, total, rows.Err()
}

func (r *patientRepo) Update(ctx context.Context, p *domain.Patient) error {
	sql := `
		UPDATE patients
		SET identity_number = $1, full_name = $2, birth_date = $3, obstetric_history = $4
		WHERE id = $5`

	result, err := r.pool.Exec(ctx, sql,
		p.IdentityNumber, p.FullName, p.BirthDate, p.ObstetricHistory, p.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrPatientNotFound
	}
	return nil
}