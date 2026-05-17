package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type clinicalLogRepo struct {
	pool *pgxpool.Pool
}

func NewClinicalLogRepository(pool *pgxpool.Pool) ClinicalLogRepository {
	return &clinicalLogRepo{pool: pool}
}

func (r *clinicalLogRepo) Create(ctx context.Context, tx pgx.Tx, log *domain.ClinicalLog) error {
	query := `
		INSERT INTO clinical_logs (admission_id, created_by, created_at,
			pa_systolic, pa_diastolic, heart_rate, resp_rate, temperature, spo2,
			pinard_status, lochia_type, lochia_amount, lochia_odor, has_clots, notes)
		VALUES ($1, $2, NOW(), $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at`

	return tx.QueryRow(ctx, query,
		log.AdmissionID, log.CreatedBy,
		log.PaSystolic, log.PaDiastolic, log.HeartRate, log.RespRate,
		log.Temperature, log.Spo2, log.PinardStatus, log.LochiaType,
		log.LochiaAmount, log.LochiaOdor, log.HasClots, log.Notes,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *clinicalLogRepo) ListByAdmission(ctx context.Context, admissionID uuid.UUID) ([]domain.ClinicalLog, error) {
	query := `
		SELECT id, admission_id, created_by, created_at,
			pa_systolic, pa_diastolic, heart_rate, resp_rate, temperature, spo2,
			pinard_status, lochia_type, lochia_amount, lochia_odor, has_clots, notes
		FROM clinical_logs
		WHERE admission_id = $1
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, admissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.ClinicalLog
	for rows.Next() {
		var log domain.ClinicalLog
		err := rows.Scan(
			&log.ID, &log.AdmissionID, &log.CreatedBy, &log.CreatedAt,
			&log.PaSystolic, &log.PaDiastolic, &log.HeartRate, &log.RespRate,
			&log.Temperature, &log.Spo2, &log.PinardStatus, &log.LochiaType,
			&log.LochiaAmount, &log.LochiaOdor, &log.HasClots, &log.Notes,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *clinicalLogRepo) CountByAdmission(ctx context.Context, tx pgx.Tx, admissionID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM clinical_logs WHERE admission_id = $1`

	var count int
	err := tx.QueryRow(ctx, query, admissionID).Scan(&count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}