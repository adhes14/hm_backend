package service

import (
	"unicode"

	"github.com/hospital_management/backend/internal/domain"
)

// ValidatePassword checks that password meets the strength rules:
// - at least 6 characters
// - at least 1 uppercase letter
// - at least 1 lowercase letter
// - at least 1 digit
//
// Returns nil if valid, or a *domain.PasswordValidationError with the list of failed rules.
func ValidatePassword(password string) *domain.PasswordValidationError {
	var details []domain.PasswordRuleDetail

	if len(password) < 6 {
		details = append(details, domain.PasswordRuleDetail{
			Rule:  "min_length",
			Value: 6,
		})
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}

	if !hasUpper {
		details = append(details, domain.PasswordRuleDetail{
			Rule:  "requires_uppercase",
			Value: true,
		})
	}
	if !hasLower {
		details = append(details, domain.PasswordRuleDetail{
			Rule:  "requires_lowercase",
			Value: true,
		})
	}
	if !hasDigit {
		details = append(details, domain.PasswordRuleDetail{
			Rule:  "requires_digit",
			Value: true,
		})
	}

	if len(details) == 0 {
		return nil
	}

	return &domain.PasswordValidationError{
		Details: details,
	}
}
