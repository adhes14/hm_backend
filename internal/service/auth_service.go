package service

import (
	"context"

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

	token, err := auth.GenerateToken(staff.ID, staff.Username, staff.Role, staff.FullName)
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

func (s *AuthService) CreateStaff(ctx context.Context, input *CreateStaffInput) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	staff := &domain.Staff{
		FullName:     input.FullName,
		Role:         input.Role,
		Username:     input.Username,
		PasswordHash: string(hash),
		IsActive:     true,
	}

	return s.staffRepo.Create(ctx, staff)
}

func (s *AuthService) ListStaff(ctx context.Context) ([]domain.Staff, error) {
	return s.staffRepo.List(ctx)
}

func (s *AuthService) ChangePassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.staffRepo.UpdatePassword(ctx, id, string(hash))
}

func (s *AuthService) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return s.staffRepo.SetActive(ctx, id, active)
}
