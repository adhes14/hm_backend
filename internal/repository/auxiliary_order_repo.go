package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type auxiliaryOrderRepository struct {
	pool *pgxpool.Pool
}

func NewAuxiliaryOrderRepository(pool *pgxpool.Pool) AuxiliaryOrderRepository {
	return &auxiliaryOrderRepository{pool: pool}
}

func (r *auxiliaryOrderRepository) Create(ctx context.Context, order *domain.AuxiliaryOrder) error {
	query := `
		INSERT INTO auxiliary_orders (admission_id, category, description, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $4)
		RETURNING id, status, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		order.AdmissionID, order.Category, order.Description, order.CreatedBy,
	).Scan(&order.ID, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	return err
}

func (r *auxiliaryOrderRepository) GetByAdmission(ctx context.Context, admissionID uuid.UUID) ([]domain.AuxiliaryOrder, error) {
	query := `
		SELECT o.id, o.admission_id, o.category, o.description, o.status, o.result,
		       o.created_by, o.updated_by, o.created_at, o.updated_at,
		       COALESCE(c.full_name, ''), COALESCE(u.full_name, '')
		FROM auxiliary_orders o
		LEFT JOIN staff c ON o.created_by = c.id
		LEFT JOIN staff u ON o.updated_by = u.id
		WHERE o.admission_id = $1
		ORDER BY o.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, admissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.AuxiliaryOrder
	for rows.Next() {
		var o domain.AuxiliaryOrder
		err := rows.Scan(
			&o.ID, &o.AdmissionID, &o.Category, &o.Description, &o.Status, &o.Result,
			&o.CreatedBy, &o.UpdatedBy, &o.CreatedAt, &o.UpdatedAt,
			&o.CreatedByName, &o.UpdatedByName,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *auxiliaryOrderRepository) GetAllPending(ctx context.Context) ([]domain.AuxiliaryOrder, error) {
	query := `
		SELECT o.id, o.admission_id, o.category, o.description, o.status, o.result,
		       o.created_by, o.updated_by, o.created_at, o.updated_at,
		       COALESCE(c.full_name, ''), COALESCE(u.full_name, ''),
			   p.full_name, b.number, bt.prefix
		FROM auxiliary_orders o
		JOIN admissions a ON o.admission_id = a.id
		JOIN patients p ON a.patient_id = p.id
		JOIN beds b ON a.bed_id = b.id
		JOIN bed_types bt ON b.bed_type_id = bt.id
		LEFT JOIN staff c ON o.created_by = c.id
		LEFT JOIN staff u ON o.updated_by = u.id
		WHERE o.status IN ('pending', 'done') AND a.status = 'active'
		ORDER BY o.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.AuxiliaryOrder
	for rows.Next() {
		var o domain.AuxiliaryOrder
		err := rows.Scan(
			&o.ID, &o.AdmissionID, &o.Category, &o.Description, &o.Status, &o.Result,
			&o.CreatedBy, &o.UpdatedBy, &o.CreatedAt, &o.UpdatedAt,
			&o.CreatedByName, &o.UpdatedByName,
			&o.PatientName, &o.BedNumber, &o.BedPrefix,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *auxiliaryOrderRepository) GetByID(ctx context.Context, id int64) (*domain.AuxiliaryOrder, error) {
	query := `
		SELECT id, admission_id, category, description, status, result, 
		       created_by, updated_by, created_at, updated_at
		FROM auxiliary_orders
		WHERE id = $1
	`
	var o domain.AuxiliaryOrder
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.AdmissionID, &o.Category, &o.Description, &o.Status, &o.Result,
		&o.CreatedBy, &o.UpdatedBy, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	return &o, nil
}

func (r *auxiliaryOrderRepository) UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus, result string, updatedBy *uuid.UUID) error {
	query := `
		UPDATE auxiliary_orders 
		SET status = $1, result = $2, updated_by = $3, updated_at = $4
		WHERE id = $5
	`
	tag, err := r.pool.Exec(ctx, query, status, result, updatedBy, time.Now().UTC(), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

func (r *auxiliaryOrderRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM auxiliary_orders WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}
