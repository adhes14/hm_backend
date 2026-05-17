package repository

import (
	"context"

	"github.com/hospital_management/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type bedTypeRepo struct {
	pool *pgxpool.Pool
}

func NewBedTypeRepository(pool *pgxpool.Pool) BedTypeRepository {
	return &bedTypeRepo{pool: pool}
}

func (r *bedTypeRepo) Create(ctx context.Context, bt *domain.BedType) error {
	query := `
		INSERT INTO bed_types (name, prefix, requires_postpartum_followup)
		VALUES ($1, $2, $3)
		RETURNING id`

	return r.pool.QueryRow(ctx, query, bt.Name, bt.Prefix, bt.RequiresPostpartumFollowup).Scan(&bt.ID)
}

func (r *bedTypeRepo) GetByID(ctx context.Context, id int) (*domain.BedType, error) {
	query := `
		SELECT id, name, prefix, requires_postpartum_followup
		FROM bed_types
		WHERE id = $1`

	var bt domain.BedType
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&bt.ID, &bt.Name, &bt.Prefix, &bt.RequiresPostpartumFollowup)
	if err != nil {
		return nil, err
	}
	return &bt, nil
}

func (r *bedTypeRepo) GetAll(ctx context.Context) ([]domain.BedType, error) {
	query := `
		SELECT id, name, prefix, requires_postpartum_followup
		FROM bed_types
		ORDER BY id`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bedTypes []domain.BedType
	for rows.Next() {
		var bt domain.BedType
		err := rows.Scan(&bt.ID, &bt.Name, &bt.Prefix, &bt.RequiresPostpartumFollowup)
		if err != nil {
			return nil, err
		}
		bedTypes = append(bedTypes, bt)
	}
	return bedTypes, rows.Err()
}

func (r *bedTypeRepo) Update(ctx context.Context, bt *domain.BedType) error {
	query := `
		UPDATE bed_types
		SET name = $1, prefix = $2, requires_postpartum_followup = $3
		WHERE id = $4`

	_, err := r.pool.Exec(ctx, query, bt.Name, bt.Prefix, bt.RequiresPostpartumFollowup, bt.ID)
	return err
}

func (r *bedTypeRepo) Delete(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM bed_types WHERE id = $1", id)
	return err
}

func (r *bedTypeRepo) CountByID(ctx context.Context, id int) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM beds WHERE bed_type_id = $1", id).Scan(&count)
	return count, err
}