package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type bedRepo struct {
	pool *pgxpool.Pool
}

func NewBedRepository(pool *pgxpool.Pool) BedRepository {
	return &bedRepo{pool: pool}
}

func (r *bedRepo) GetAll(ctx context.Context) ([]domain.Bed, error) {
	query := `
		SELECT b.id, b.number, b.is_active, b.current_admission_id,
		       bt.id, bt.name, bt.prefix, bt.requires_postpartum_followup,
		       p.full_name, a.next_control_at, a.estimated_discharge_at, a.event_type,
		       COALESCE((SELECT COUNT(*) FROM clinical_logs WHERE admission_id = a.id), 0),
		       b.ward_id, w.name, w.description
		FROM beds b
		LEFT JOIN bed_types bt ON b.bed_type_id = bt.id
		LEFT JOIN admissions a ON b.current_admission_id = a.id
		LEFT JOIN patients p ON a.patient_id = p.id
		LEFT JOIN wards w ON b.ward_id = w.id
		ORDER BY b.number`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	beds := make([]domain.Bed, 0)
	for rows.Next() {
		var b domain.Bed
		var bt domain.BedType
		var w domain.Ward
		err := rows.Scan(&b.ID, &b.Number, &b.IsActive, &b.CurrentAdmissionID,
			&bt.ID, &bt.Name, &bt.Prefix, &bt.RequiresPostpartumFollowup,
			&b.CurrentPatientName, &b.NextControlAt, &b.EstimatedDischargeAt, &b.EventType,
			&b.ControlCount, &b.WardID, &w.Name, &w.Description)
		if err != nil {
			return nil, err
		}
		b.BedType = &bt
		w.ID = b.WardID
		b.Ward = &w
		beds = append(beds, b)
	}
	return beds, rows.Err()
}

func (r *bedRepo) GetByID(ctx context.Context, id int) (*domain.Bed, error) {
	query := `
		SELECT b.id, b.number, b.is_active, b.current_admission_id,
		       bt.id, bt.name, bt.prefix, bt.requires_postpartum_followup,
		       p.full_name, a.next_control_at, a.estimated_discharge_at, a.event_type,
		       COALESCE((SELECT COUNT(*) FROM clinical_logs WHERE admission_id = a.id), 0),
		       b.ward_id, w.name, w.description
		FROM beds b
		LEFT JOIN bed_types bt ON b.bed_type_id = bt.id
		LEFT JOIN admissions a ON b.current_admission_id = a.id
		LEFT JOIN patients p ON a.patient_id = p.id
		LEFT JOIN wards w ON b.ward_id = w.id
		WHERE b.id = $1`

	var b domain.Bed
	var bt domain.BedType
	var w domain.Ward
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.Number, &b.IsActive, &b.CurrentAdmissionID,
		&bt.ID, &bt.Name, &bt.Prefix, &bt.RequiresPostpartumFollowup,
		&b.CurrentPatientName, &b.NextControlAt, &b.EstimatedDischargeAt, &b.EventType,
		&b.ControlCount, &b.WardID, &w.Name, &w.Description)
	if err != nil {
		return nil, err
	}
	b.BedType = &bt
	w.ID = b.WardID
	b.Ward = &w
	return &b, nil
}

func (r *bedRepo) UpdateCurrentAdmission(ctx context.Context, bedID int, admissionID *uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE beds SET current_admission_id = $1 WHERE id = $2",
		admissionID, bedID)
	return err
}

func (r *bedRepo) CreateBed(ctx context.Context, bed *domain.Bed) error {
	query := `
		INSERT INTO beds (number, bed_type_id, is_active, ward_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	return r.pool.QueryRow(ctx, query, bed.Number, bed.BedType.ID, bed.IsActive, bed.WardID).Scan(&bed.ID)
}

func (r *bedRepo) UpdateBed(ctx context.Context, bed *domain.Bed) error {
	query := `
		UPDATE beds
		SET number = $1, bed_type_id = $2, ward_id = $3
		WHERE id = $4`

	_, err := r.pool.Exec(ctx, query, bed.Number, bed.BedType.ID, bed.WardID, bed.ID)
	return err
}

func (r *bedRepo) DeleteBed(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM beds WHERE id = $1", id)
	return err
}

func (r *bedRepo) CountByBedTypeID(ctx context.Context, bedTypeID int) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM beds WHERE bed_type_id = $1", bedTypeID).Scan(&count)
	return count, err
}

func (r *bedRepo) CountByWardID(ctx context.Context, wardID int) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM beds WHERE ward_id = $1", wardID).Scan(&count)
	return count, err
}