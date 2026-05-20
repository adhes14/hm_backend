package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrMissingSecret = errors.New("JWT_SECRET environment variable is not set")
)

type Claims struct {
	StaffID  uuid.UUID        `json:"user_id"`
	Username string           `json:"username"`
	Role     domain.StaffRole `json:"role"`
	FullName string           `json:"full_name"`
	jwt.RegisteredClaims
}

func getSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("mi_secreto"), nil
	}
	return []byte(secret), nil
}

func GenerateToken(staffID uuid.UUID, username string, role domain.StaffRole, fullName string) (string, error) {
	secret, err := getSecret()
	if err != nil {
		return "", err
	}

	claims := Claims{
		StaffID:  staffID,
		Username: username,
		Role:     role,
		FullName: fullName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ValidateToken(tokenString string) (*Claims, error) {
	secret, err := getSecret()
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
