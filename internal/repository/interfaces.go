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
	// CRUD methods for beds
	CreateBed(ctx context.Context, bed *domain.Bed) error
	UpdateBed(ctx context.Context, bed *domain.Bed) error
	DeleteBed(ctx context.Context, id int) error
	CountByBedTypeID(ctx context.Context, bedTypeID int) (int, error)
}

type BedTypeRepository interface {
	Create(ctx context.Context, bt *domain.BedType) error
	GetByID(ctx context.Context, id int) (*domain.BedType, error)
	GetAll(ctx context.Context) ([]domain.BedType, error)
	Update(ctx context.Context, bt *domain.BedType) error
	Delete(ctx context.Context, id int) error
	CountByID(ctx context.Context, id int) (int, error)
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
	GetActiveByPatientID(ctx context.Context, patientID uuid.UUID) (*domain.Admission, error)
	Discharge(ctx context.Context, id uuid.UUID) error
	GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Admission, error)
	GetByBedID(ctx context.Context, bedID int) (*domain.Admission, error)
}

type ClinicalLogRepository interface {
	Create(ctx context.Context, tx pgx.Tx, log *domain.ClinicalLog) error
	ListByAdmission(ctx context.Context, admissionID uuid.UUID) ([]domain.ClinicalLog, error)
	CountByAdmission(ctx context.Context, tx pgx.Tx, admissionID uuid.UUID) (int, error)
}

type StaffRepository interface {
	Create(ctx context.Context, staff *domain.Staff) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Staff, error)
	GetByUsername(ctx context.Context, username string) (*domain.Staff, error)
	List(ctx context.Context) ([]domain.Staff, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
}

type SettingsRepository interface {
	GetByKey(ctx context.Context, key string) (*domain.SystemSetting, error)
	GetAll(ctx context.Context) ([]domain.SystemSetting, error)
	Update(ctx context.Context, key string, value string) error
}

type SSETicketRepository interface {
	Create(ctx context.Context, ticket *domain.SSETicket) error
	ValidateAndConsume(ctx context.Context, ticketStr string) (*domain.SSETicket, error)
	DeleteExpired(ctx context.Context) error
}