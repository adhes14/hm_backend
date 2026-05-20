package repository

import (
	"context"
	"errors"

	"github.com/hospital_management/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type settingsRepository struct {
	db *pgxpool.Pool
}

func NewSettingsRepository(db *pgxpool.Pool) SettingsRepository {
	return &settingsRepository{db: db}
}

func (r *settingsRepository) GetByKey(ctx context.Context, key string) (*domain.SystemSetting, error) {
	query := `
		SELECT key, value, COALESCE(description, '')
		FROM system_settings
		WHERE key = $1
	`
	var setting domain.SystemSetting
	err := r.db.QueryRow(ctx, query, key).Scan(
		&setting.Key,
		&setting.Value,
		&setting.Description,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("setting not found")
		}
		return nil, err
	}
	return &setting, nil
}

func (r *settingsRepository) GetAll(ctx context.Context) ([]domain.SystemSetting, error) {
	query := `
		SELECT key, value, COALESCE(description, '')
		FROM system_settings
		ORDER BY key ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []domain.SystemSetting
	for rows.Next() {
		var setting domain.SystemSetting
		if err := rows.Scan(&setting.Key, &setting.Value, &setting.Description); err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}
	return settings, rows.Err()
}

func (r *settingsRepository) Update(ctx context.Context, key string, value string) error {
	query := `
		INSERT INTO system_settings (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = $2
	`
	_, err := r.db.Exec(ctx, query, key, value)
	return err
}
