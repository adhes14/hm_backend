package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// mockStaffRepository implements repository.StaffRepository for testing.
type mockStaffRepository struct {
	// Data store
	staffByID       map[uuid.UUID]*domain.Staff
	staffByUsername map[string]*domain.Staff
	// Counts
	createCalls               int
	updatePasswordCalls       int
	updatePasswordClearCalls  int
	updateStaffCalls          int
	countActiveAdminsCalls    int
	countActiveAdminsVal      int
	countActiveAdminsErr      error
	getByIDErr                error
	// Error injection
	updatePasswordErr error
	updateStaffErr    error
	getByIDNotFound   bool
}

func newMockStaffRepository() *mockStaffRepository {
	return &mockStaffRepository{
		staffByID:       make(map[uuid.UUID]*domain.Staff),
		staffByUsername: make(map[string]*domain.Staff),
		countActiveAdminsVal: 1,
	}
}

func (m *mockStaffRepository) Create(ctx context.Context, staff *domain.Staff) error {
	m.createCalls++
	if staff.ID == uuid.Nil {
		staff.ID = uuid.New()
	}
	m.staffByID[staff.ID] = staff
	m.staffByUsername[staff.Username] = staff
	return nil
}

func (m *mockStaffRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Staff, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.getByIDNotFound {
		return nil, domain.ErrUserNotFound
	}
	staff, ok := m.staffByID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return staff, nil
}

func (m *mockStaffRepository) GetByUsername(ctx context.Context, username string) (*domain.Staff, error) {
	staff, ok := m.staffByUsername[username]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return staff, nil
}

func (m *mockStaffRepository) List(ctx context.Context) ([]domain.Staff, error) {
	var list []domain.Staff
	for _, s := range m.staffByID {
		list = append(list, *s)
	}
	return list, nil
}

func (m *mockStaffRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	m.updatePasswordCalls++
	if m.updatePasswordErr != nil {
		return m.updatePasswordErr
	}
	staff, ok := m.staffByID[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	staff.PasswordHash = hash
	return nil
}

func (m *mockStaffRepository) UpdatePasswordAndClearFlag(ctx context.Context, id uuid.UUID, hash string) error {
	m.updatePasswordClearCalls++
	if m.updatePasswordErr != nil {
		return m.updatePasswordErr
	}
	staff, ok := m.staffByID[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	staff.PasswordHash = hash
	staff.MustChangePassword = false
	return nil
}

func (m *mockStaffRepository) UpdateStaff(ctx context.Context, id uuid.UUID, fullName string, role domain.StaffRole) (*domain.Staff, error) {
	m.updateStaffCalls++
	if m.updateStaffErr != nil {
		return nil, m.updateStaffErr
	}
	staff, ok := m.staffByID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	staff.FullName = fullName
	staff.Role = role
	return staff, nil
}

func (m *mockStaffRepository) CountActiveAdmins(ctx context.Context) (int, error) {
	m.countActiveAdminsCalls++
	if m.countActiveAdminsErr != nil {
		return 0, m.countActiveAdminsErr
	}
	return m.countActiveAdminsVal, nil
}

func (m *mockStaffRepository) UpdatePasswordAndSetFlag(ctx context.Context, id uuid.UUID, hash string) error {
	if m.updatePasswordErr != nil {
		return m.updatePasswordErr
	}
	staff, ok := m.staffByID[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	staff.PasswordHash = hash
	staff.MustChangePassword = true
	return nil
}

func (m *mockStaffRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	staff, ok := m.staffByID[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	staff.IsActive = active
	return nil
}

// Compile-time check that mock implements the interface.
var _ repository.StaffRepository = (*mockStaffRepository)(nil)

// ---------------------------------------------------------------------------
// Tests: ChangeMyPassword — Forced Flow (must_change_password = true)
// ---------------------------------------------------------------------------

func TestChangeMyPassword_ForcedFlow_NoCurrentPassword_Success(t *testing.T) {
	mock := newMockStaffRepository()
	svc := NewAuthService(mock)

	staffID := uuid.New()
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("OldPass1"), bcrypt.MinCost)

	staff := &domain.Staff{
		ID:                 staffID,
		Username:           "forceduser",
		FullName:           "Forced User",
		Role:               domain.RoleHealthStaff,
		PasswordHash:       string(oldHash),
		IsActive:           true,
		MustChangePassword: true,
	}
	mock.staffByID[staffID] = staff

	// Forced flow: no current_password needed, flag=true
	updatedStaff, err := svc.ChangeMyPassword(context.Background(), staffID, "", "NewPass1")
	if err != nil {
		t.Fatalf("ChangeMyPassword failed: %v", err)
	}

	// Verify returned staff has must_change_password cleared
	if updatedStaff == nil {
		t.Fatal("expected updated staff, got nil")
	}
	if updatedStaff.MustChangePassword {
		t.Error("returned staff has MustChangePassword=true, expected false")
	}

	// Verify password was updated
	if staff.PasswordHash == string(oldHash) {
		t.Error("password hash was not updated")
	}

	// Verify must_change_password flag was cleared
	if staff.MustChangePassword {
		t.Error("MustChangePassword flag was not cleared after forced password change")
	}

	// Verify UpdatePasswordAndClearFlag was called (not UpdatePassword)
	if mock.updatePasswordClearCalls != 1 {
		t.Errorf("expected 1 UpdatePasswordAndClearFlag call, got %d", mock.updatePasswordClearCalls)
	}
	if mock.updatePasswordCalls != 0 {
		t.Errorf("expected 0 UpdatePassword calls for forced flow, got %d", mock.updatePasswordCalls)
	}
}

func TestChangeMyPassword_ForcedFlow_RejectsWeakPassword(t *testing.T) {
	mock := newMockStaffRepository()
	svc := NewAuthService(mock)

	staffID := uuid.New()
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("OldPass1"), bcrypt.MinCost)

	staff := &domain.Staff{
		ID:                 staffID,
		Username:           "weakuser",
		FullName:           "Weak User",
		Role:               domain.RoleHealthStaff,
		PasswordHash:       string(oldHash),
		IsActive:           true,
		MustChangePassword: true,
	}
	mock.staffByID[staffID] = staff

	// Password "abc" fails all rules
	_, err := svc.ChangeMyPassword(context.Background(), staffID, "", "abc")
	if err == nil {
		t.Fatal("expected validation error for weak password, got nil")
	}

	valErr, ok := err.(*domain.PasswordValidationError)
	if !ok {
		t.Fatalf("expected *PasswordValidationError, got %T: %v", err, err)
	}

	if len(valErr.Details) == 0 {
		t.Error("expected Details to be populated")
	}

	// Verify password was NOT changed
	if staff.MustChangePassword != true {
		t.Error("MustChangePassword should still be true after validation failure")
	}
}

// ---------------------------------------------------------------------------
// Tests: ChangeMyPassword — Self-Service Flow (must_change_password = false)
// ---------------------------------------------------------------------------

func TestChangeMyPassword_SelfService_CorrectPassword_Success(t *testing.T) {
	mock := newMockStaffRepository()
	svc := NewAuthService(mock)

	staffID := uuid.New()
	originalPassword := "Correct1"
	oldHash, _ := bcrypt.GenerateFromPassword([]byte(originalPassword), bcrypt.MinCost)

	staff := &domain.Staff{
		ID:                 staffID,
		Username:           "selfservice",
		FullName:           "Self Service",
		Role:               domain.RoleHealthStaff,
		PasswordHash:       string(oldHash),
		IsActive:           true,
		MustChangePassword: false, // flag=false → self-service flow
	}
	mock.staffByID[staffID] = staff

	updatedStaff, err := svc.ChangeMyPassword(context.Background(), staffID, originalPassword, "NewPass2")
	if err != nil {
		t.Fatalf("ChangeMyPassword failed: %v", err)
	}

	// Verify returned staff
	if updatedStaff == nil {
		t.Fatal("expected updated staff, got nil")
	}

	// Verify password was updated
	if staff.PasswordHash == string(oldHash) {
		t.Error("password hash was not updated")
	}

	// Verify UpdatePassword was called (not UpdatePasswordAndClearFlag)
	if mock.updatePasswordCalls != 1 {
		t.Errorf("expected 1 UpdatePassword call for self-service flow, got %d", mock.updatePasswordCalls)
	}
	if mock.updatePasswordClearCalls != 0 {
		t.Errorf("expected 0 UpdatePasswordAndClearFlag calls, got %d", mock.updatePasswordClearCalls)
	}

	// Verify flag remains false
	if staff.MustChangePassword {
		t.Error("MustChangePassword should still be false after self-service change")
	}
}

func TestChangeMyPassword_SelfService_WrongCurrentPassword_ReturnsError(t *testing.T) {
	mock := newMockStaffRepository()
	svc := NewAuthService(mock)

	staffID := uuid.New()
	correctPassword := "Correct1"
	oldHash, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.MinCost)

	staff := &domain.Staff{
		ID:                 staffID,
		Username:           "wrongpw",
		FullName:           "Wrong PW",
		Role:               domain.RoleHealthStaff,
		PasswordHash:       string(oldHash),
		IsActive:           true,
		MustChangePassword: false,
	}
	mock.staffByID[staffID] = staff

	_, err := svc.ChangeMyPassword(context.Background(), staffID, "WrongPassword1", "NewPass3")
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	// Verify password was NOT changed
	if staff.PasswordHash != string(oldHash) {
		t.Error("password should not have been changed with wrong current password")
	}
}

func TestChangeMyPassword_SelfService_EmptyCurrentPassword_ReturnsError(t *testing.T) {
	mock := newMockStaffRepository()
	svc := NewAuthService(mock)

	staffID := uuid.New()
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("SomePass1"), bcrypt.MinCost)

	staff := &domain.Staff{
		ID:                 staffID,
		Username:           "emptypw",
		FullName:           "Empty PW",
		Role:               domain.RoleHealthStaff,
		PasswordHash:       string(oldHash),
		IsActive:           true,
		MustChangePassword: false, // flag=false → self-service flow
	}
	mock.staffByID[staffID] = staff

	_, err := svc.ChangeMyPassword(context.Background(), staffID, "", "NewPass4")
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials for empty current_password in self-service, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: UpdateStaff — Admin Guardrails
// ---------------------------------------------------------------------------

func TestUpdateStaff_SelfDemotion_Returns409(t *testing.T) {
	mock := newMockStaffRepository()
	svc := NewAuthService(mock)

	adminID := uuid.New()
	mock.staffByID[adminID] = &domain.Staff{
		ID:       adminID,
		Username: "admin1",
		Role:     domain.RoleAdmin,
		FullName: "Admin One",
		IsActive: true,
	}
	mock.countActiveAdminsVal = 2 // More than 1, so last-admin check would pass

	input := &UpdateStaffInput{
		FullName: "Admin One",
		Role:     domain.RoleHealthStaff, // Demoting to non-admin
	}

	_, err := svc.UpdateStaff(context.Background(), adminID, input, adminID) // Same user
	if err != domain.ErrCannotDemoteSelf {
		t.Fatalf("expected ErrCannotDemoteSelf, got %v", err)
	}

	if mock.updateStaffCalls != 0 {
		t.Error("UpdateStaff repo call should not have been made on self-demotion")
	}
}

func TestUpdateStaff_RemoveLastAdmin_Returns409(t *testing.T) {
	mock := newMockStaffRepository()
	svc := NewAuthService(mock)

	adminID := uuid.New()
	// Only one active admin in the system
	mock.countActiveAdminsVal = 1

	mock.staffByID[adminID] = &domain.Staff{
		ID:       adminID,
		Username: "lastadmin",
		Role:     domain.RoleAdmin,
		FullName: "Last Admin",
		IsActive: true,
	}

	// Different user doing the edit (so self-demotion check passes)
	requesterID := uuid.New()
	mock.staffByID[requesterID] = &domain.Staff{
		ID:       requesterID,
		Username: "otheruser",
		Role:     domain.RoleHealthStaff,
		FullName: "Other User",
		IsActive: true,
	}

	input := &UpdateStaffInput{
		FullName: "Last Admin",
		Role:     domain.RoleHealthStaff, // Demoting the last admin
	}

	_, err := svc.UpdateStaff(context.Background(), adminID, input, requesterID)
	if err != domain.ErrCannotRemoveLastAdmin {
		t.Fatalf("expected ErrCannotRemoveLastAdmin, got %v", err)
	}

	if mock.updateStaffCalls != 0 {
		t.Error("UpdateStaff repo call should not have been made on last-admin removal")
	}
}

func TestUpdateStaff_NormalUpdate_Success(t *testing.T) {
	mock := newMockStaffRepository()
	svc := NewAuthService(mock)

	targetID := uuid.New()
	mock.staffByID[targetID] = &domain.Staff{
		ID:       targetID,
		Username: "targetuser",
		Role:     domain.RoleHealthStaff,
		FullName: "Old Name",
		IsActive: true,
	}
	mock.countActiveAdminsVal = 2

	requesterID := uuid.New()
	mock.staffByID[requesterID] = &domain.Staff{
		ID:       requesterID,
		Username: "adminuser",
		Role:     domain.RoleAdmin,
		FullName: "Admin User",
		IsActive: true,
	}

	input := &UpdateStaffInput{
		FullName: "New Name",
		Role:     domain.RoleAdmin, // promoting to admin
	}

	updated, err := svc.UpdateStaff(context.Background(), targetID, input, requesterID)
	if err != nil {
		t.Fatalf("UpdateStaff failed: %v", err)
	}

	if mock.updateStaffCalls != 1 {
		t.Errorf("expected 1 UpdateStaff call, got %d", mock.updateStaffCalls)
	}

	if updated.FullName != "New Name" {
		t.Errorf("expected FullName 'New Name', got %q", updated.FullName)
	}
	if updated.Role != domain.RoleAdmin {
		t.Errorf("expected Role admin, got %q", updated.Role)
	}
}

func TestUpdateStaff_PromotionNotBlocked(t *testing.T) {
	mock := newMockStaffRepository()
	svc := NewAuthService(mock)

	targetID := uuid.New()
	mock.staffByID[targetID] = &domain.Staff{
		ID:       targetID,
		Username: "targetuser",
		Role:     domain.RoleHealthStaff,
		FullName: "Target",
		IsActive: true,
	}
	mock.countActiveAdminsVal = 1

	requesterID := uuid.New()
	mock.staffByID[requesterID] = &domain.Staff{
		ID:       requesterID,
		Username: "adminuser",
		Role:     domain.RoleAdmin,
		FullName: "Admin",
		IsActive: true,
	}

	input := &UpdateStaffInput{
		FullName: "Target",
		Role:     domain.RoleAdmin, // promotion, not demotion
	}

	_, err := svc.UpdateStaff(context.Background(), targetID, input, requesterID)
	if err != nil {
		t.Fatalf("promotion should not be blocked: %v", err)
	}

	if mock.updateStaffCalls != 1 {
		t.Errorf("expected UpdateStaff to be called, got %d calls", mock.updateStaffCalls)
	}
}

func TestUpdateStaff_AdminChangingOwnName_NotSelfDemotion(t *testing.T) {
	mock := newMockStaffRepository()
	svc := NewAuthService(mock)

	adminID := uuid.New()
	mock.staffByID[adminID] = &domain.Staff{
		ID:       adminID,
		Username: "admin",
		Role:     domain.RoleAdmin,
		FullName: "Old Admin Name",
		IsActive: true,
	}
	mock.countActiveAdminsVal = 2

	input := &UpdateStaffInput{
		FullName: "New Admin Name",
		Role:     domain.RoleAdmin, // same role, just changing name
	}

	_, err := svc.UpdateStaff(context.Background(), adminID, input, adminID)
	if err != nil {
		t.Fatalf("admin changing own name should succeed: %v", err)
	}

	if mock.updateStaffCalls != 1 {
		t.Errorf("expected UpdateStaff to be called, got %d calls", mock.updateStaffCalls)
	}
}
