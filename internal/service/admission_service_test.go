package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// mockAdmissionRepo implements repository.AdmissionRepository for testing.
type mockAdmissionRepo struct {
	ListDischargedByBedIDWithDetailsFunc func(ctx context.Context, bedID int, from, to *time.Time, page, limit int) ([]domain.AdmissionWithDetails, int, error)
	// Track whether the repo was called (for the from_after_to test)
	called bool
}

// Compile-time interface check
var _ repository.AdmissionRepository = (*mockAdmissionRepo)(nil)

func (m *mockAdmissionRepo) Create(ctx context.Context, a *domain.Admission) error { return nil }
func (m *mockAdmissionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Admission, error) {
	return nil, domain.ErrAdmissionNotFound
}
func (m *mockAdmissionRepo) GetActiveByBedID(ctx context.Context, bedID int) (*domain.Admission, error) {
	return nil, nil
}
func (m *mockAdmissionRepo) GetActiveByPatientID(ctx context.Context, patientID uuid.UUID) (*domain.Admission, error) {
	return nil, nil
}
func (m *mockAdmissionRepo) Discharge(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockAdmissionRepo) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Admission, error) {
	return nil, nil
}
func (m *mockAdmissionRepo) GetByBedID(ctx context.Context, bedID int) (*domain.Admission, error) { return nil, nil }
func (m *mockAdmissionRepo) UpdateDiagnosis(ctx context.Context, id uuid.UUID, diagnosis string, updatedBy uuid.UUID) error {
	return nil
}
func (m *mockAdmissionRepo) UpdateTreatment(ctx context.Context, id uuid.UUID, treatment string) error {
	return nil
}
func (m *mockAdmissionRepo) GetAllByPatientID(ctx context.Context, patientID uuid.UUID) ([]domain.Admission, error) {
	return nil, nil
}
func (m *mockAdmissionRepo) ListDischargedByBedIDWithDetails(
	ctx context.Context, bedID int, from, to *time.Time, page, limit int,
) ([]domain.AdmissionWithDetails, int, error) {
	m.called = true
	if m.ListDischargedByBedIDWithDetailsFunc != nil {
		return m.ListDischargedByBedIDWithDetailsFunc(ctx, bedID, from, to, page, limit)
	}
	return []domain.AdmissionWithDetails{}, 0, nil
}

func TestListDischargedByBedID_NoRows(t *testing.T) {
	mock := &mockAdmissionRepo{
		ListDischargedByBedIDWithDetailsFunc: func(ctx context.Context, bedID int, from, to *time.Time, page, limit int) ([]domain.AdmissionWithDetails, int, error) {
			return []domain.AdmissionWithDetails{}, 0, nil
		},
	}

	svc := &AdmissionService{
		admissionRepo: mock,
	}

	admissions, total, err := svc.ListDischargedByBedID(context.Background(), 1, nil, nil, 1, 10)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(admissions) != 0 {
		t.Errorf("expected empty slice, got %d items", len(admissions))
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
}

func TestListDischargedByBedID_WithDateRange(t *testing.T) {
	now := time.Now()
	mockAdmissions := []domain.AdmissionWithDetails{
		{
			Admission: domain.Admission{
				ID:     uuid.New(),
				Status: domain.AdmissionStatusDischarged,
				BedID:  1,
			},
			PatientName:      "Test Patient",
			PatientDNI:       "12345678",
			ClinicalLogCount: 5,
		},
		{
			Admission: domain.Admission{
				ID:     uuid.New(),
				Status: domain.AdmissionStatusDischarged,
				BedID:  1,
			},
			PatientName:      "Test Patient 2",
			PatientDNI:       "87654321",
			ClinicalLogCount: 3,
		},
	}

	mock := &mockAdmissionRepo{
		ListDischargedByBedIDWithDetailsFunc: func(ctx context.Context, bedID int, from, to *time.Time, page, limit int) ([]domain.AdmissionWithDetails, int, error) {
			return mockAdmissions, 2, nil
		},
	}

	from := now.Add(-48 * time.Hour)
	to := now

	svc := &AdmissionService{
		admissionRepo: mock,
	}

	admissions, total, err := svc.ListDischargedByBedID(context.Background(), 1, &from, &to, 1, 10)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(admissions) != 2 {
		t.Errorf("expected 2 admissions, got %d", len(admissions))
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if admissions[0].PatientName != "Test Patient" {
		t.Errorf("expected 'Test Patient', got %q", admissions[0].PatientName)
	}
	if admissions[0].ClinicalLogCount != 5 {
		t.Errorf("expected clinical_log_count 5, got %d", admissions[0].ClinicalLogCount)
	}
}

func TestListDischargedByBedID_FromAfterTo(t *testing.T) {
	mock := &mockAdmissionRepo{}

	svc := &AdmissionService{
		admissionRepo: mock,
	}

	from := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	_, _, err := svc.ListDischargedByBedID(context.Background(), 1, &from, &to, 1, 10)
	if !errors.Is(err, domain.ErrInvalidDateRange) {
		t.Errorf("expected ErrInvalidDateRange, got %v", err)
	}
	if mock.called {
		t.Errorf("expected repo NOT to be called when from > to")
	}
}

func TestListDischargedByBedID_PageLimitClamping(t *testing.T) {
	var capturedPage, capturedLimit int
	mock := &mockAdmissionRepo{
		ListDischargedByBedIDWithDetailsFunc: func(ctx context.Context, bedID int, from, to *time.Time, page, limit int) ([]domain.AdmissionWithDetails, int, error) {
			capturedPage = page
			capturedLimit = limit
			return []domain.AdmissionWithDetails{}, 0, nil
		},
	}

	svc := &AdmissionService{
		admissionRepo: mock,
	}

	// page=0 should be clamped to 1; limit=999 should be clamped to 10
	_, _, err := svc.ListDischargedByBedID(context.Background(), 1, nil, nil, 0, 999)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if capturedPage != 1 {
		t.Errorf("expected page to be clamped to 1, got %d", capturedPage)
	}
	if capturedLimit != 10 {
		t.Errorf("expected limit to be clamped to 10, got %d", capturedLimit)
	}
}
