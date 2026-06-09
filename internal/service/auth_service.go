package service

import (
	"context"
	"crypto/rand"
	"math/big"

	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/auth"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	staffRepo repository.StaffRepository
}

func NewAuthService(staffRepo repository.StaffRepository) *AuthService {
	return &AuthService{staffRepo: staffRepo}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, *domain.Staff, error) {
	staff, err := s.staffRepo.GetByUsername(ctx, username)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return "", nil, domain.ErrInvalidCredentials
		}
		return "", nil, err
	}

	if !staff.IsActive {
		return "", nil, domain.ErrUserInactive
	}

	err = bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(password))
	if err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(staff.ID, staff.Username, staff.Role, staff.FullName, staff.MustChangePassword)
	if err != nil {
		return "", nil, err
	}

	return token, staff, nil
}

type CreateStaffInput struct {
	FullName string           `json:"full_name"`
	Role     domain.StaffRole `json:"role"`
	Username string           `json:"username"`
	Password string           `json:"password"`
}

func (s *AuthService) CreateStaff(ctx context.Context, input *CreateStaffInput) (string, error) {
	tempPassword := input.Password
	if tempPassword == "" {
		tempPassword = generateTempPassword()
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	staff := &domain.Staff{
		FullName:           input.FullName,
		Role:               input.Role,
		Username:           input.Username,
		PasswordHash:       string(hash),
		IsActive:           true,
		MustChangePassword: true,
	}

	if err := s.staffRepo.Create(ctx, staff); err != nil {
		return "", err
	}

	return tempPassword, nil
}

func (s *AuthService) ListStaff(ctx context.Context) ([]domain.Staff, error) {
	return s.staffRepo.List(ctx)
}

// ChangeMyPassword updates the current user's password and returns the updated staff.
// For forced flow (must_change_password=true), currentPassword is ignored.
// For self-service flow (must_change_password=false), currentPassword is required and verified.
func (s *AuthService) ChangeMyPassword(ctx context.Context, staffID uuid.UUID, currentPassword, newPassword string) (*domain.Staff, error) {
	staff, err := s.staffRepo.GetByID(ctx, staffID)
	if err != nil {
		return nil, err
	}

	// Self-service flow: must verify current password
	if !staff.MustChangePassword {
		if currentPassword == "" {
			return nil, domain.ErrInvalidCredentials
		}
		if err := bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(currentPassword)); err != nil {
			return nil, domain.ErrInvalidCredentials
		}
	}

	// Validate new password against policy
	if valErr := ValidatePassword(newPassword); valErr != nil {
		return nil, valErr
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Forced flow: clear the flag after password change
	if staff.MustChangePassword {
		if err := s.staffRepo.UpdatePasswordAndClearFlag(ctx, staffID, string(hash)); err != nil {
			return nil, err
		}
	} else {
		// Self-service flow: just update password, keep flag as-is
		if err := s.staffRepo.UpdatePassword(ctx, staffID, string(hash)); err != nil {
			return nil, err
		}
	}

	// Return updated staff (must_change_password is now false for forced flow)
	return s.staffRepo.GetByID(ctx, staffID)
}

func (s *AuthService) ChangePassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.staffRepo.UpdatePassword(ctx, id, string(hash))
}

// UpdateStaffInput is the request body for updating a staff member's full_name and role.
type UpdateStaffInput struct {
	FullName string           `json:"full_name"`
	Role     domain.StaffRole `json:"role"`
}

// UpdateStaff updates a staff member's full_name and role with guardrails.
// Guardrails:
//   - A user cannot demote themselves (change their own role from admin to non-admin).
//   - The last active admin cannot be demoted or deactivated.
func (s *AuthService) UpdateStaff(ctx context.Context, id uuid.UUID, input *UpdateStaffInput, requestingStaffID uuid.UUID) (*domain.Staff, error) {
	// Self-demotion check: requesting user cannot change their own role to non-admin
	if id == requestingStaffID && input.Role != domain.RoleAdmin {
		return nil, domain.ErrCannotDemoteSelf
	}

	// If the target is currently an admin and is being demoted, check last-admin guardrail
	target, err := s.staffRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if target.Role == domain.RoleAdmin && input.Role != domain.RoleAdmin {
		count, err := s.staffRepo.CountActiveAdmins(ctx)
		if err != nil {
			return nil, err
		}
		if count <= 1 {
			return nil, domain.ErrCannotRemoveLastAdmin
		}
	}

	return s.staffRepo.UpdateStaff(ctx, id, input.FullName, input.Role)
}

// ResetStaffPassword resets a staff member's password and forces a password change on
// next login (must_change_password is set to true). If newPassword is empty, a random
// compliant password is generated. Returns the temporary password.
func (s *AuthService) ResetStaffPassword(ctx context.Context, id uuid.UUID, newPassword string) (string, error) {
	tempPassword := newPassword
	if tempPassword == "" {
		tempPassword = generateTempPassword()
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	if err := s.staffRepo.UpdatePasswordAndSetFlag(ctx, id, string(hash)); err != nil {
		return "", err
	}

	return tempPassword, nil
}

func (s *AuthService) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return s.staffRepo.SetActive(ctx, id, active)
}

// generateTempPassword creates a random 8-character password that satisfies the
// password policy (≥6 chars, ≥1 uppercase, ≥1 lowercase, ≥1 digit).
func generateTempPassword() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for {
		b := make([]byte, 8)
		for i := range b {
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
			b[i] = chars[n.Int64()]
		}
		pwd := string(b)
		if ValidatePassword(pwd) == nil {
			return pwd
		}
	}
}
