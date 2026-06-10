package repository

import (
	"context"

	"github.com/hospital_management/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type wardRepo struct {
	pool *pgxpool.Pool
}

func NewWardRepository(pool *pgxpool.Pool) WardRepository {
	return &wardRepo{pool: pool}
}

func (r *wardRepo) Create(ctx context.Context, w *domain.Ward) error {
	query := `
		INSERT INTO wards (name, description)
		VALUES ($1, $2)
		RETURNING id`

	return r.pool.QueryRow(ctx, query, w.Name, w.Description).Scan(&w.ID)
}

func (r *wardRepo) GetByID(ctx context.Context, id int) (*domain.Ward, error) {
	query := `
		SELECT id, name, description
		FROM wards
		WHERE id = $1`

	var w domain.Ward
	err := r.pool.QueryRow(ctx, query, id).Scan(&w.ID, &w.Name, &w.Description)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *wardRepo) GetAll(ctx context.Context) ([]domain.Ward, error) {
	query := `
		SELECT w.id, w.name, w.description, COALESCE(COUNT(b.id), 0)
		FROM wards w
		LEFT JOIN beds b ON w.id = b.ward_id
		GROUP BY w.id, w.name, w.description
		ORDER BY w.name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wards []domain.Ward
	for rows.Next() {
		var w domain.Ward
		err := rows.Scan(&w.ID, &w.Name, &w.Description, &w.BedCount)
		if err != nil {
			return nil, err
		}
		wards = append(wards, w)
	}
	return wards, rows.Err()
}

func (r *wardRepo) Update(ctx context.Context, w *domain.Ward) error {
	query := `
		UPDATE wards
		SET name = $1, description = $2
		WHERE id = $3`

	_, err := r.pool.Exec(ctx, query, w.Name, w.Description, w.ID)
	return err
}

func (r *wardRepo) Delete(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM wards WHERE id = $1", id)
	return err
}

func (r *wardRepo) CountByID(ctx context.Context, id int) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM wards WHERE id = $1", id).Scan(&count)
	return count, err
}
