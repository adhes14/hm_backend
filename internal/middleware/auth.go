package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/hospital_management/backend/internal/auth"
	"github.com/hospital_management/backend/internal/domain"
)

type contextKey string

const (
	claimsContextKey contextKey = "claims"
)

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(roles ...domain.StaffRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserFromContext(r.Context())
			if claims == nil {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			hasRole := false
			for _, role := range roles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				writeError(w, http.StatusForbidden, "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// passwordChangeWhitelist stores paths that are exempt from RequirePasswordChanged.
// Use a set (map[string]struct{}) for O(1) lookup.
var passwordChangeWhitelist = map[string]struct{}{
	"/api/v1/auth/change-password": {},
}

// RequirePasswordChanged blocks requests when MustChangePassword is true,
// unless the request path is in the whitelist (e.g., the change-password endpoint itself).
// Missing claims defaults to false (safe behavior).
func RequirePasswordChanged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r.Context())
		if claims != nil && claims.MustChangePassword {
			if _, ok := passwordChangeWhitelist[r.URL.Path]; !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"password_change_required","message":"You must change your password before continuing"}`))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func GetUserFromContext(ctx context.Context) *auth.Claims {
	claims, ok := ctx.Value(claimsContextKey).(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}
