package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/repository"
	"github.com/hospital_management/backend/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// stubBedService is a minimal stub for BedService used in handler tests
type stubBedService struct {
	getBedFunc func(ctx context.Context, id int) (*domain.Bed, error)
}

func newStubBedService() *stubBedService {
	return &stubBedService{
		getBedFunc: func(ctx context.Context, id int) (*domain.Bed, error) {
			return &domain.Bed{ID: id, Number: 1}, nil
		},
	}
}

// stubAdmissionService is a minimal stub for AdmissionService used in handler tests
type stubAdmissionService struct {
	listDischargedByBedIDFunc func(ctx context.Context, bedID int, from, to *time.Time, page, limit int) ([]domain.AdmissionWithDetails, int, error)
}

func (s *stubAdmissionService) ListDischargedByBedID(ctx context.Context, bedID int, from, to *time.Time, page, limit int) ([]domain.AdmissionWithDetails, int, error) {
	if s.listDischargedByBedIDFunc != nil {
		return s.listDischargedByBedIDFunc(ctx, bedID, from, to, page, limit)
	}
	return []domain.AdmissionWithDetails{}, 0, nil
}

// Make bedService.BedService and admissionService.AdmissionService work with the handler.
// We use a helper to construct a fake service.Service for the handler.
// Since BedHandler requires *service.BedService and *service.AdmissionService concretely,
// we create them using the real constructors with stub repositories.

// stubBedRepo is a minimal stub for BedRepository
type stubBedRepo struct {
	getByIDFunc func(ctx context.Context, id int) (*domain.Bed, error)
}

func (r *stubBedRepo) GetAll(ctx context.Context) ([]domain.Bed, error)                  { return nil, nil }
func (r *stubBedRepo) GetByID(ctx context.Context, id int) (*domain.Bed, error) {
	if r.getByIDFunc != nil {
		return r.getByIDFunc(ctx, id)
	}
	return &domain.Bed{ID: id, Number: 1}, nil
}
func (r *stubBedRepo) UpdateCurrentAdmission(ctx context.Context, bedID int, admissionID *uuid.UUID) error {
	return nil
}
func (r *stubBedRepo) CreateBed(ctx context.Context, bed *domain.Bed) error              { return nil }
func (r *stubBedRepo) UpdateBed(ctx context.Context, bed *domain.Bed) error              { return nil }
func (r *stubBedRepo) DeleteBed(ctx context.Context, id int) error                       { return nil }
func (r *stubBedRepo) CountByBedTypeID(ctx context.Context, bedTypeID int) (int, error)  { return 0, nil }

var _ repository.BedRepository = (*stubBedRepo)(nil)

// stubAdmissionRepo is a minimal stub for AdmissionRepository
type stubAdmissionRepo struct {
	listFunc func(ctx context.Context, bedID int, from, to *time.Time, page, limit int) ([]domain.AdmissionWithDetails, int, error)
}

func (r *stubAdmissionRepo) Create(ctx context.Context, a *domain.Admission) error { return nil }
func (r *stubAdmissionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Admission, error) {
	return nil, domain.ErrAdmissionNotFound
}
func (r *stubAdmissionRepo) GetActiveByBedID(ctx context.Context, bedID int) (*domain.Admission, error) {
	return nil, nil
}
func (r *stubAdmissionRepo) GetActiveByPatientID(ctx context.Context, patientID uuid.UUID) (*domain.Admission, error) {
	return nil, nil
}
func (r *stubAdmissionRepo) Discharge(ctx context.Context, id uuid.UUID) error { return nil }
func (r *stubAdmissionRepo) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Admission, error) {
	return nil, nil
}
func (r *stubAdmissionRepo) GetByBedID(ctx context.Context, bedID int) (*domain.Admission, error) { return nil, nil }
func (r *stubAdmissionRepo) UpdateDiagnosis(ctx context.Context, id uuid.UUID, diagnosis string, updatedBy uuid.UUID) error {
	return nil
}
func (r *stubAdmissionRepo) GetAllByPatientID(ctx context.Context, patientID uuid.UUID) ([]domain.Admission, error) {
	return nil, nil
}
func (r *stubAdmissionRepo) ListDischargedByBedIDWithDetails(ctx context.Context, bedID int, from, to *time.Time, page, limit int) ([]domain.AdmissionWithDetails, int, error) {
	if r.listFunc != nil {
		return r.listFunc(ctx, bedID, from, to, page, limit)
	}
	return []domain.AdmissionWithDetails{}, 0, nil
}

var _ repository.AdmissionRepository = (*stubAdmissionRepo)(nil)

func buildTestBedHandler(admissionListFunc func(ctx context.Context, bedID int, from, to *time.Time, page, limit int) ([]domain.AdmissionWithDetails, int, error)) *BedHandler {
	stubBed := &stubBedRepo{}
	admissionRepo := &stubAdmissionRepo{listFunc: admissionListFunc}

	bedSvc := service.NewBedService(stubBed, nil, admissionRepo)
	admissionSvc := &service.AdmissionService{} // will be partial; we populate the repo field by direct struct access

	// We need to set admissionRepo on the admissionService. The AdmissionService struct has unexported fields.
	// Instead, use reflection or a constructor with nil args. Better: use a test constructor approach.
	// Since we can't access unexported fields, we construct a minimal real service.
	// Actually, AdmissionService only needs the repo to call. Let me just create it with all nil deps
	// and set the repo field directly. But it's unexported...
	// For the handler test, we stub the entire admissionService with a wrapper.

	// Let's use a different approach: create a wrapper service that delegates to our stub.
	// We'll use the real NewAdmissionService with nil pool and sseService but proper stub repos.
	_ = admissionSvc

	// Simpler approach: we create a type that wraps the handler's admissionService dependency
	// via embedding. But the handler holds *service.AdmissionService which is concrete.
	// Let's just make a full AdmissionService with stub repos.

	// Better yet: the AdmissionService just needs the admissionRepo.
	// We can create one with nil for unused fields since ListDischargedByBedID only uses admissionRepo.
	pool := (*pgxpool.Pool)(nil) // won't be used by ListDischargedByBedID
	admissionSvc2 := service.NewAdmissionService(pool, admissionRepo, nil, nil, nil, nil)

	handler := NewBedHandler(bedSvc, admissionSvc2)
	return handler
}

func TestListBedAdmissions_InvalidBedID(t *testing.T) {
	handler := buildTestBedHandler(nil)
	r := chi.NewRouter()
	r.Get("/beds/{id}/admissions", handler.ListBedAdmissions)

	req := httptest.NewRequest(http.MethodGet, "/beds/abc/admissions", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] != "invalid bed id" {
		t.Errorf("expected 'invalid bed id', got %q", body["error"])
	}
}

func TestListBedAdmissions_InvalidDateFormat(t *testing.T) {
	handler := buildTestBedHandler(nil)
	r := chi.NewRouter()
	r.Get("/beds/{id}/admissions", handler.ListBedAdmissions)

	req := httptest.NewRequest(http.MethodGet, "/beds/1/admissions?from=not-a-date", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] != "invalid date format, use YYYY-MM-DD" {
		t.Errorf("expected date format error, got %q", body["error"])
	}
}

func TestListBedAdmissions_ValidRequest(t *testing.T) {
	admissionID := uuid.New()
	handler := buildTestBedHandler(func(ctx context.Context, bedID int, from, to *time.Time, page, limit int) ([]domain.AdmissionWithDetails, int, error) {
		return []domain.AdmissionWithDetails{
			{
				Admission: domain.Admission{
					ID:     admissionID,
					Status: domain.AdmissionStatusDischarged,
					BedID:  1,
				},
				PatientName:      "Test Patient",
				PatientDNI:       "12345678",
				ClinicalLogCount: 8,
			},
		}, 1, nil
	})

	r := chi.NewRouter()
	r.Get("/beds/{id}/admissions", handler.ListBedAdmissions)

	req := httptest.NewRequest(http.MethodGet, "/beds/1/admissions?page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp bedAdmissionsPaginatedResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Page != 1 {
		t.Errorf("expected page 1, got %d", resp.Page)
	}
	if resp.Limit != 10 {
		t.Errorf("expected limit 10, got %d", resp.Limit)
	}
	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}
	if resp.TotalPages != 1 {
		t.Errorf("expected total_pages 1, got %d", resp.TotalPages)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data item, got %d", len(resp.Data))
	}
	if resp.Data[0].PatientName != "Test Patient" {
		t.Errorf("expected 'Test Patient', got %q", resp.Data[0].PatientName)
	}
	if resp.Data[0].ClinicalLogCount != 8 {
		t.Errorf("expected clinical_log_count 8, got %d", resp.Data[0].ClinicalLogCount)
	}
}

func TestListBedAdmissions_FromAfterToPropagates400(t *testing.T) {
	handler := buildTestBedHandler(func(ctx context.Context, bedID int, from, to *time.Time, page, limit int) ([]domain.AdmissionWithDetails, int, error) {
		return nil, 0, domain.ErrInvalidDateRange
	})

	r := chi.NewRouter()
	r.Get("/beds/{id}/admissions", handler.ListBedAdmissions)

	req := httptest.NewRequest(http.MethodGet, "/beds/1/admissions?from=2026-02-01&to=2026-01-01", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] != "from date must not exceed to date" {
		t.Errorf("expected 'from date must not exceed to date', got %q", body["error"])
	}
}
