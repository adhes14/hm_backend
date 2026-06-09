package repository

import (
	"context"
	"errors"
	"fmt"
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
		INSERT INTO admissions (patient_id, bed_id, status, event_type, admission_diagnosis, current_diagnosis)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		a.PatientID, a.BedID, a.Status, a.EventType, a.AdmissionDiagnosis, a.CurrentDiagnosis,
	).Scan(&a.ID, &a.CreatedAt)
}

func (r *admissionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Admission, error) {
	query := `
		SELECT a.id, a.patient_id, a.bed_id, a.status, a.event_type, a.event_at,
		       a.next_control_at, a.estimated_discharge_at, a.created_at, a.discharged_at,
		       a.admission_diagnosis, a.current_diagnosis, a.current_diagnosis_updated_by,
		       COALESCE(s.full_name, '') as current_diagnosis_updated_by_name
		FROM admissions a
		LEFT JOIN staff s ON a.current_diagnosis_updated_by = s.id
		WHERE a.id = $1`

	var a domain.Admission
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.PatientID, &a.BedID, &a.Status, &a.EventType,
		&a.EventAt, &a.NextControlAt, &a.EstimatedDischargeAt,
		&a.CreatedAt, &a.DischargedAt, &a.AdmissionDiagnosis, &a.CurrentDiagnosis,
		&a.CurrentDiagnosisUpdatedBy, &a.CurrentDiagnosisUpdatedByName)
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
		SELECT a.id, a.patient_id, a.bed_id, a.status, a.event_type, a.event_at,
		       a.next_control_at, a.estimated_discharge_at, a.created_at, a.discharged_at,
		       a.admission_diagnosis, a.current_diagnosis, a.current_diagnosis_updated_by,
		       COALESCE(s.full_name, '') as current_diagnosis_updated_by_name
		FROM admissions a
		LEFT JOIN staff s ON a.current_diagnosis_updated_by = s.id
		WHERE a.bed_id = $1 AND a.status = 'active'`

	var a domain.Admission
	err := r.pool.QueryRow(ctx, query, bedID).Scan(
		&a.ID, &a.PatientID, &a.BedID, &a.Status, &a.EventType,
		&a.EventAt, &a.NextControlAt, &a.EstimatedDischargeAt,
		&a.CreatedAt, &a.DischargedAt, &a.AdmissionDiagnosis, &a.CurrentDiagnosis,
		&a.CurrentDiagnosisUpdatedBy, &a.CurrentDiagnosisUpdatedByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No active admission for this bed
		}
		return nil, err
	}
	return &a, nil
}

func (r *admissionRepo) GetActiveByPatientID(ctx context.Context, patientID uuid.UUID) (*domain.Admission, error) {
	query := `
		SELECT a.id, a.patient_id, a.bed_id, a.status, a.event_type, a.event_at,
		       a.next_control_at, a.estimated_discharge_at, a.created_at, a.discharged_at,
		       a.admission_diagnosis, a.current_diagnosis, a.current_diagnosis_updated_by,
		       COALESCE(s.full_name, '') as current_diagnosis_updated_by_name
		FROM admissions a
		LEFT JOIN staff s ON a.current_diagnosis_updated_by = s.id
		WHERE a.patient_id = $1 AND a.status = 'active'`

	var a domain.Admission
	err := r.pool.QueryRow(ctx, query, patientID).Scan(
		&a.ID, &a.PatientID, &a.BedID, &a.Status, &a.EventType,
		&a.EventAt, &a.NextControlAt, &a.EstimatedDischargeAt,
		&a.CreatedAt, &a.DischargedAt, &a.AdmissionDiagnosis, &a.CurrentDiagnosis,
		&a.CurrentDiagnosisUpdatedBy, &a.CurrentDiagnosisUpdatedByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No active admission for this patient
		}
		return nil, err
	}
	return &a, nil
}

func (r *admissionRepo) Discharge(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
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

func (r *admissionRepo) UpdateDiagnosis(ctx context.Context, id uuid.UUID, diagnosis string, updatedBy uuid.UUID) error {
	query := `
		UPDATE admissions
		SET current_diagnosis = $1, current_diagnosis_updated_by = $2
		WHERE id = $3 AND status = 'active'`

	tag, err := r.pool.Exec(ctx, query, diagnosis, updatedBy, id)
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
		SELECT a.id, a.patient_id, a.bed_id, a.status, a.event_type, a.event_at,
		       a.next_control_at, a.estimated_discharge_at, a.created_at, a.discharged_at,
		       a.admission_diagnosis, a.current_diagnosis, a.current_diagnosis_updated_by,
		       COALESCE(s.full_name, '') as current_diagnosis_updated_by_name
		FROM admissions a
		LEFT JOIN staff s ON a.current_diagnosis_updated_by = s.id
		WHERE a.id = $1 FOR UPDATE OF a`

	var a domain.Admission
	err := tx.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.PatientID, &a.BedID, &a.Status, &a.EventType,
		&a.EventAt, &a.NextControlAt, &a.EstimatedDischargeAt,
		&a.CreatedAt, &a.DischargedAt, &a.AdmissionDiagnosis, &a.CurrentDiagnosis,
		&a.CurrentDiagnosisUpdatedBy, &a.CurrentDiagnosisUpdatedByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAdmissionNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *admissionRepo) GetByBedID(ctx context.Context, bedID int) (*domain.Admission, error) {
	query := `
		SELECT a.id, a.patient_id, a.bed_id, a.status, a.event_type, a.event_at,
		       a.next_control_at, a.estimated_discharge_at, a.created_at, a.discharged_at,
		       a.admission_diagnosis, a.current_diagnosis, a.current_diagnosis_updated_by,
		       COALESCE(s.full_name, '') as current_diagnosis_updated_by_name
		FROM admissions a
		LEFT JOIN staff s ON a.current_diagnosis_updated_by = s.id
		WHERE a.bed_id = $1`

	var a domain.Admission
	err := r.pool.QueryRow(ctx, query, bedID).Scan(
		&a.ID, &a.PatientID, &a.BedID, &a.Status, &a.EventType,
		&a.EventAt, &a.NextControlAt, &a.EstimatedDischargeAt,
		&a.CreatedAt, &a.DischargedAt, &a.AdmissionDiagnosis, &a.CurrentDiagnosis,
		&a.CurrentDiagnosisUpdatedBy, &a.CurrentDiagnosisUpdatedByName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *admissionRepo) GetAllByPatientID(ctx context.Context, patientID uuid.UUID) ([]domain.Admission, error) {
	query := `
		SELECT a.id, a.patient_id, a.bed_id, a.status, a.event_type, a.event_at,
		       a.next_control_at, a.estimated_discharge_at, a.created_at, a.discharged_at,
		       a.admission_diagnosis, a.current_diagnosis, a.current_diagnosis_updated_by,
		       COALESCE(s.full_name, '') as current_diagnosis_updated_by_name
		FROM admissions a
		LEFT JOIN staff s ON a.current_diagnosis_updated_by = s.id
		WHERE a.patient_id = $1
		ORDER BY a.created_at DESC`

	rows, err := r.pool.Query(ctx, query, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var admissions []domain.Admission
	for rows.Next() {
		var a domain.Admission
		err := rows.Scan(
			&a.ID, &a.PatientID, &a.BedID, &a.Status, &a.EventType,
			&a.EventAt, &a.NextControlAt, &a.EstimatedDischargeAt,
			&a.CreatedAt, &a.DischargedAt, &a.AdmissionDiagnosis, &a.CurrentDiagnosis,
			&a.CurrentDiagnosisUpdatedBy, &a.CurrentDiagnosisUpdatedByName)
		if err != nil {
			return nil, err
		}
		admissions = append(admissions, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return admissions, nil
}

func (r *admissionRepo) ListDischargedByBedIDWithDetails(
	ctx context.Context,
	bedID int,
	from, to *time.Time,
	page, limit int,
) ([]domain.AdmissionWithDetails, int, error) {
	offset := (page - 1) * limit

	// Build dynamic WHERE clause and args
	where := "WHERE a.bed_id = $1 AND a.status = 'discharged'"
	countArgs := []interface{}{bedID}
	queryArgs := []interface{}{bedID}
	argIdx := 2

	if from != nil {
		where += fmt.Sprintf(" AND a.discharged_at >= $%d", argIdx)
		countArgs = append(countArgs, *from)
		queryArgs = append(queryArgs, *from)
		argIdx++
	}
	if to != nil {
		where += fmt.Sprintf(" AND a.discharged_at <= $%d", argIdx)
		countArgs = append(countArgs, *to)
		queryArgs = append(queryArgs, *to)
		argIdx++
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM admissions a " + where
	err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	dataQuery := fmt.Sprintf(`
		SELECT a.id, a.patient_id, a.bed_id, a.status, a.event_type, a.event_at,
		       a.next_control_at, a.estimated_discharge_at, a.created_at, a.discharged_at,
		       a.admission_diagnosis, a.current_diagnosis, a.current_diagnosis_updated_by,
		       COALESCE(s.full_name, '') as current_diagnosis_updated_by_name,
		       COALESCE(p.full_name, '') as patient_name,
		       COALESCE(p.identity_number, '') as patient_dni,
		       (SELECT COUNT(*) FROM clinical_logs cl WHERE cl.admission_id = a.id) as clinical_log_count
		FROM admissions a
		LEFT JOIN staff s ON a.current_diagnosis_updated_by = s.id
		LEFT JOIN patients p ON p.id = a.patient_id
		%s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.pool.Query(ctx, dataQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var admissions []domain.AdmissionWithDetails
	for rows.Next() {
		var a domain.AdmissionWithDetails
		err := rows.Scan(
			&a.ID, &a.PatientID, &a.BedID, &a.Status, &a.EventType,
			&a.EventAt, &a.NextControlAt, &a.EstimatedDischargeAt,
			&a.CreatedAt, &a.DischargedAt, &a.AdmissionDiagnosis, &a.CurrentDiagnosis,
			&a.CurrentDiagnosisUpdatedBy, &a.CurrentDiagnosisUpdatedByName,
			&a.PatientName, &a.PatientDNI, &a.ClinicalLogCount)
		if err != nil {
			return nil, 0, err
		}
		admissions = append(admissions, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if admissions == nil {
		admissions = []domain.AdmissionWithDetails{}
	}

	return admissions, total, nil
}