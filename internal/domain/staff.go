package domain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type StaffRole string

const (
	RoleHealthStaff StaffRole = "health_staff"
	RoleAdmin       StaffRole = "admin"
)

type Staff struct {
	ID                 uuid.UUID `json:"id"`
	FullName           string    `json:"full_name"`
	Role               StaffRole `json:"role"`
	Username           string    `json:"username"`
	PasswordHash       string    `json:"-"`
	IsActive           bool      `json:"is_active"`
	MustChangePassword bool      `json:"must_change_password"`
}

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUsernameTaken          = errors.New("username already taken")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrUserInactive           = errors.New("user is inactive")
	ErrPasswordTooWeak        = errors.New("password does not meet strength requirements")
	ErrCannotDemoteSelf       = errors.New("cannot change your own role")
	ErrCannotRemoveLastAdmin  = errors.New("cannot remove the last active admin")
	ErrPasswordChangeRequired = errors.New("password change is required before continuing")
)

// PasswordRuleDetail describes a single failed password rule.
type PasswordRuleDetail struct {
	Rule  string      `json:"rule"`
	Value interface{} `json:"value"`
}

// PasswordValidationError carries the specific password rules that failed.
type PasswordValidationError struct {
	Details []PasswordRuleDetail `json:"details"`
}

func (e *PasswordValidationError) Error() string {
	rules := make([]string, len(e.Details))
	for i, d := range e.Details {
		rules[i] = d.Rule
	}
	return fmt.Sprintf("password validation failed: %s", strings.Join(rules, ", "))
}
