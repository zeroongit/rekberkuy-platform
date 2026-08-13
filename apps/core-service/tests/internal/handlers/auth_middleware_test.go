package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"rekberkuy/core-service/internal/delivery/handlers"
	"rekberkuy/core-service/internal/domain"
)

// ============================================================================
// AUTH MIDDLEWARE — UNIT TESTS (RBAC + JWT validation)
// ============================================================================

const testJWTSecret = "test-secret-key"

// makeToken signs a valid JWT with a given role for test purposes.
func makeToken(t *testing.T, secret string, userID string, role domain.UserRole) string {
	t.Helper()
	claims := &domain.JWTCustomClaims{
		UserID:   userID,
		Username: "tester",
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "rekberkuy-test",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return s
}

// newTestRouter builds a minimal gin router with one protected route.
func newTestRouter(mw *handlers.AuthMiddleware) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", mw.RequireRole(domain.RoleUser), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": c.MustGet("user_id")})
	})
	return r
}

func doRequest(r http.Handler, authHeader string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	r := newTestRouter(handlers.NewAuthMiddleware(testJWTSecret))
	w := doRequest(r, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without header, got %d", w.Code)
	}
}

func TestAuthMiddleware_WrongFormat(t *testing.T) {
	r := newTestRouter(handlers.NewAuthMiddleware(testJWTSecret))
	w := doRequest(r, "Token abc123") // not "Bearer "
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong format, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	r := newTestRouter(handlers.NewAuthMiddleware(testJWTSecret))
	w := doRequest(r, "Bearer bukan-jwt-valid")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid token, got %d", w.Code)
	}
}

func TestAuthMiddleware_WrongSigningSecret(t *testing.T) {
	r := newTestRouter(handlers.NewAuthMiddleware(testJWTSecret))
	tok := makeToken(t, "different-secret", "user-1", domain.RoleUser)
	w := doRequest(r, "Bearer "+tok)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for token signed with different secret, got %d", w.Code)
	}
}

func TestAuthMiddleware_ForbiddenRole(t *testing.T) {
	mw := handlers.NewAuthMiddleware(testJWTSecret)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// This route is for ADMIN only
	r.GET("/admin", mw.RequireRole(domain.RoleAdmin), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// A USER-role token tries to access an ADMIN route
	tok := makeToken(t, testJWTSecret, "user-1", domain.RoleUser)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for unauthorized role, got %d", w.Code)
	}
}

func TestAuthMiddleware_Success(t *testing.T) {
	r := newTestRouter(handlers.NewAuthMiddleware(testJWTSecret))
	tok := makeToken(t, testJWTSecret, "user-uuid-1", domain.RoleUser)
	w := doRequest(r, "Bearer "+tok)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for valid token + matching role, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestAuthMiddleware_MultipleAllowedRoles(t *testing.T) {
	mw := handlers.NewAuthMiddleware(testJWTSecret)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/eo", mw.RequireRole(domain.RoleEventOrganizer, domain.RoleAdmin), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Admin allowed
	tok := makeToken(t, testJWTSecret, "admin-1", domain.RoleAdmin)
	req := httptest.NewRequest(http.MethodGet, "/eo", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("admin expected 200, got %d", w.Code)
	}

	// EO allowed
	tok = makeToken(t, testJWTSecret, "eo-1", domain.RoleEventOrganizer)
	req = httptest.NewRequest(http.MethodGet, "/eo", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("EO expected 200, got %d", w.Code)
	}
}
