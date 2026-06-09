package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/auth"
	"github.com/hospital_management/backend/internal/domain"
)

// testHandler is a simple http.Handler that records whether it was called.
type testHandler struct {
	called bool
	status int
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called = true
	w.WriteHeader(h.status)
}

// withClaims returns a context with the given claims injected.
func withClaims(ctx context.Context, claims *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

// newRequestWithClaims creates an httptest request with claims in context.
func newRequestWithClaims(method, target string, claims *auth.Claims) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	if claims != nil {
		req = req.WithContext(withClaims(req.Context(), claims))
	}
	return req
}

func TestRequirePasswordChanged_FlagTrue_NonWhitelisted_Returns403(t *testing.T) {
	claims := &auth.Claims{
		StaffID:            uuid.New(),
		Username:           "testuser",
		Role:               domain.RoleHealthStaff,
		MustChangePassword: true,
	}

	handler := &testHandler{status: http.StatusOK}
	mw := RequirePasswordChanged(handler)

	req := newRequestWithClaims("GET", "/api/v1/beds", claims)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if handler.called {
		t.Error("handler was called when MustChangePassword=true on non-whitelisted path")
	}

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}

	expectedBody := `{"error":"password_change_required","message":"You must change your password before continuing"}`
	if rec.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, rec.Body.String())
	}
}

func TestRequirePasswordChanged_FlagTrue_Whitelisted_AllowsPassThrough(t *testing.T) {
	claims := &auth.Claims{
		StaffID:            uuid.New(),
		Username:           "testuser",
		Role:               domain.RoleHealthStaff,
		MustChangePassword: true,
	}

	handler := &testHandler{status: http.StatusOK}
	mw := RequirePasswordChanged(handler)

	req := newRequestWithClaims("POST", "/api/v1/auth/change-password", claims)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if !handler.called {
		t.Error("handler was not called on whitelisted path even though MustChangePassword=true")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestRequirePasswordChanged_FlagFalse_AllowsPassThrough(t *testing.T) {
	claims := &auth.Claims{
		StaffID:            uuid.New(),
		Username:           "testuser",
		Role:               domain.RoleHealthStaff,
		MustChangePassword: false,
	}

	handler := &testHandler{status: http.StatusOK}
	mw := RequirePasswordChanged(handler)

	req := newRequestWithClaims("GET", "/api/v1/beds", claims)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if !handler.called {
		t.Error("handler was not called when MustChangePassword=false")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestRequirePasswordChanged_NoClaims_AllowsPassThrough(t *testing.T) {
	// Missing claims defaults to false (safe behavior per spec)
	handler := &testHandler{status: http.StatusOK}
	mw := RequirePasswordChanged(handler)

	req := newRequestWithClaims("GET", "/api/v1/beds", nil)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if !handler.called {
		t.Error("handler was not called when no claims present (safe default)")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestRequirePasswordChanged_FlagTrue_MultipleNonWhitelistedPaths(t *testing.T) {
	claims := &auth.Claims{
		StaffID:            uuid.New(),
		Username:           "testuser",
		Role:               domain.RoleHealthStaff,
		MustChangePassword: true,
	}

	paths := []string{
		"/api/v1/beds",
		"/api/v1/users",
		"/api/v1/patients",
		"/api/v1/admissions",
		"/api/v1/orders",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			handler := &testHandler{status: http.StatusOK}
			mw := RequirePasswordChanged(handler)

			req := newRequestWithClaims("GET", path, claims)
			rec := httptest.NewRecorder()

			mw.ServeHTTP(rec, req)

			if handler.called {
				t.Errorf("handler was called for path %q when MustChangePassword=true", path)
			}
			if rec.Code != http.StatusForbidden {
				t.Errorf("path %q: expected 403, got %d", path, rec.Code)
			}
		})
	}
}

func TestWhitelist_ContainsChangePasswordEndpoint(t *testing.T) {
	// Verify that the whitelist explicitly contains the change-password endpoint
	path := "/api/v1/auth/change-password"
	if _, ok := passwordChangeWhitelist[path]; !ok {
		t.Errorf("whitelist does not contain %q", path)
	}
}

func TestAuthenticate_ValidToken_SetsClaims(t *testing.T) {
	// Generate a valid token
	token, err := auth.GenerateToken(
		uuid.New(),
		"testuser",
		domain.RoleHealthStaff,
		"Test User",
		false,
	)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Capture claims in the handler body — the request passed to next.ServeHTTP
	// has the new context with claims.
	var capturedClaims *auth.Claims
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = GetUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	mw := Authenticate(innerHandler)

	req := httptest.NewRequest("GET", "/api/v1/beds", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if capturedClaims == nil {
		t.Fatal("claims were not set in context after authentication")
	}
	if capturedClaims.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %q", capturedClaims.Username)
	}
}

func TestAuthenticate_NoHeader_Returns401(t *testing.T) {
	handler := &testHandler{status: http.StatusOK}
	mw := Authenticate(handler)

	req := httptest.NewRequest("GET", "/api/v1/beds", nil)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if handler.called {
		t.Error("handler was called when no Authorization header provided")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuthenticate_InvalidToken_Returns401(t *testing.T) {
	handler := &testHandler{status: http.StatusOK}
	mw := Authenticate(handler)

	req := httptest.NewRequest("GET", "/api/v1/beds", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if handler.called {
		t.Error("handler was called with invalid token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireRole_AdminRequired_AdminPasses(t *testing.T) {
	claims := &auth.Claims{
		StaffID:  uuid.New(),
		Username: "adminuser",
		Role:     domain.RoleAdmin,
	}

	handler := &testHandler{status: http.StatusOK}
	mw := RequireRole(domain.RoleAdmin)(handler)

	req := newRequestWithClaims("GET", "/api/v1/users", claims)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if !handler.called {
		t.Error("admin user was blocked by RequireRole(admin)")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireRole_AdminRequired_HealthStaffBlocked(t *testing.T) {
	claims := &auth.Claims{
		StaffID:  uuid.New(),
		Username: "healthstaff",
		Role:     domain.RoleHealthStaff,
	}

	handler := &testHandler{status: http.StatusOK}
	mw := RequireRole(domain.RoleAdmin)(handler)

	req := newRequestWithClaims("GET", "/api/v1/users", claims)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if handler.called {
		t.Error("health_staff was allowed through RequireRole(admin)")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestRequireRole_NilClaims_Returns401(t *testing.T) {
	handler := &testHandler{status: http.StatusOK}
	mw := RequireRole(domain.RoleAdmin)(handler)

	req := newRequestWithClaims("GET", "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if handler.called {
		t.Error("handler was called when no claims present")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
