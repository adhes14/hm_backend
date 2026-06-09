package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type staffRepository struct {
	db *pgxpool.Pool
}

func NewStaffRepository(db *pgxpool.Pool) StaffRepository {
	return &staffRepository{db: db}
}

func (r *staffRepository) Create(ctx context.Context, staff *domain.Staff) error {
	query := `
		INSERT INTO staff (id, full_name, role, username, password_hash, is_active, must_change_password)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	if staff.ID == uuid.Nil {
		staff.ID = uuid.New()
	}

	err := r.db.QueryRow(ctx, query,
		staff.ID,
		staff.FullName,
		staff.Role,
		staff.Username,
		staff.PasswordHash,
		staff.IsActive,
		staff.MustChangePassword,
	).Scan(&staff.ID)

	if err != nil {
		return err
	}
	return nil
}

func (r *staffRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Staff, error) {
	query := `
		SELECT id, full_name, role, username, password_hash, is_active, must_change_password
		FROM staff
		WHERE id = $1
	`

	var staff domain.Staff
	err := r.db.QueryRow(ctx, query, id).Scan(
		&staff.ID,
		&staff.FullName,
		&staff.Role,
		&staff.Username,
		&staff.PasswordHash,
		&staff.IsActive,
		&staff.MustChangePassword,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &staff, nil
}

func (r *staffRepository) GetByUsername(ctx context.Context, username string) (*domain.Staff, error) {
	query := `
		SELECT id, full_name, role, username, password_hash, is_active, must_change_password
		FROM staff
		WHERE username = $1
	`

	var staff domain.Staff
	err := r.db.QueryRow(ctx, query, username).Scan(
		&staff.ID,
		&staff.FullName,
		&staff.Role,
		&staff.Username,
		&staff.PasswordHash,
		&staff.IsActive,
		&staff.MustChangePassword,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &staff, nil
}

func (r *staffRepository) List(ctx context.Context) ([]domain.Staff, error) {
	query := `
		SELECT id, full_name, role, username, is_active, must_change_password
		FROM staff
		ORDER BY full_name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staffList []domain.Staff
	for rows.Next() {
		var staff domain.Staff
		err := rows.Scan(
			&staff.ID,
			&staff.FullName,
			&staff.Role,
			&staff.Username,
			&staff.IsActive,
			&staff.MustChangePassword,
		)
		if err != nil {
			return nil, err
		}
		staffList = append(staffList, staff)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return staffList, nil
}

func (r *staffRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	query := `
		UPDATE staff
		SET password_hash = $1
		WHERE id = $2
	`

	cmdTag, err := r.db.Exec(ctx, query, hash, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *staffRepository) UpdateStaff(ctx context.Context, id uuid.UUID, fullName string, role domain.StaffRole) (*domain.Staff, error) {
	query := `
		UPDATE staff
		SET full_name = $1, role = $2
		WHERE id = $3
		RETURNING id, full_name, role, username, is_active, must_change_password
	`

	var staff domain.Staff
	err := r.db.QueryRow(ctx, query, fullName, role, id).Scan(
		&staff.ID,
		&staff.FullName,
		&staff.Role,
		&staff.Username,
		&staff.IsActive,
		&staff.MustChangePassword,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &staff, nil
}

func (r *staffRepository) CountActiveAdmins(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM staff WHERE role = 'admin' AND is_active = true`

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *staffRepository) UpdatePasswordAndClearFlag(ctx context.Context, id uuid.UUID, hash string) error {
	query := `
		UPDATE staff
		SET password_hash = $1, must_change_password = false
		WHERE id = $2
	`

	cmdTag, err := r.db.Exec(ctx, query, hash, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *staffRepository) UpdatePasswordAndSetFlag(ctx context.Context, id uuid.UUID, hash string) error {
	query := `
		UPDATE staff
		SET password_hash = $1, must_change_password = true
		WHERE id = $2
	`

	cmdTag, err := r.db.Exec(ctx, query, hash, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *staffRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	query := `
		UPDATE staff
		SET is_active = $1
		WHERE id = $2
	`

	cmdTag, err := r.db.Exec(ctx, query, active, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
