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
		       bt.id, bt.name, bt.prefix
		FROM beds b
		LEFT JOIN bed_types bt ON b.bed_type_id = bt.id
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
		err := rows.Scan(&b.ID, &b.Number, &b.IsActive, &b.CurrentAdmissionID,
			&bt.ID, &bt.Name, &bt.Prefix)
		if err != nil {
			return nil, err
		}
		b.BedType = &bt
		beds = append(beds, b)
	}
	return beds, rows.Err()
}

func (r *bedRepo) GetByID(ctx context.Context, id int) (*domain.Bed, error) {
	query := `
		SELECT b.id, b.number, b.is_active, b.current_admission_id,
		       bt.id, bt.name, bt.prefix
		FROM beds b
		LEFT JOIN bed_types bt ON b.bed_type_id = bt.id
		WHERE b.id = $1`

	var b domain.Bed
	var bt domain.BedType
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.Number, &b.IsActive, &b.CurrentAdmissionID,
		&bt.ID, &bt.Name, &bt.Prefix)
	if err != nil {
		return nil, err
	}
	b.BedType = &bt
	return &b, nil
}

func (r *bedRepo) UpdateCurrentAdmission(ctx context.Context, bedID int, admissionID *uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE beds SET current_admission_id = $1 WHERE id = $2",
		admissionID, bedID)
	return err
}