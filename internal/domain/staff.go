package domain

import (
	"errors"

	"github.com/google/uuid"
)

type StaffRole string

const (
	RoleHealthStaff StaffRole = "health_staff"
	RoleAdmin       StaffRole = "admin"
)

type Staff struct {
	ID           uuid.UUID `json:"id"`
	FullName     string    `json:"full_name"`
	Role         StaffRole `json:"role"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	IsActive     bool      `json:"is_active"`
}

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
)
