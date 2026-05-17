package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

type BedRepository interface {
	GetAll(ctx context.Context) ([]domain.Bed, error)
	GetByID(ctx context.Context, id int) (*domain.Bed, error)
	UpdateCurrentAdmission(ctx context.Context, bedID int, admissionID *uuid.UUID) error
}

type PatientRepository interface {
	Create(ctx context.Context, p *domain.Patient) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Patient, error)
	GetByBedID(ctx context.Context, bedID int) (*domain.Patient, error)
	Search(ctx context.Context, query string) ([]domain.Patient, error)
	List(ctx context.Context, page, limit int) ([]domain.Patient, int, error)
	Update(ctx context.Context, p *domain.Patient) error
}

type AdmissionRepository interface {
	Create(ctx context.Context, a *domain.Admission) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Admission, error)
	GetActiveByBedID(ctx context.Context, bedID int) (*domain.Admission, error)
	Discharge(ctx context.Context, id uuid.UUID) error
	GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Admission, error)
}

type ClinicalLogRepository interface {
	Create(ctx context.Context, tx pgx.Tx, log *domain.ClinicalLog) error
	ListByAdmission(ctx context.Context, admissionID uuid.UUID) ([]domain.ClinicalLog, error)
	CountByAdmission(ctx context.Context, tx pgx.Tx, admissionID uuid.UUID) (int, error)
}