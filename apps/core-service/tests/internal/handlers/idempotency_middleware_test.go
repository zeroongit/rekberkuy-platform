package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"rekberkuy/core-service/internal/delivery/handlers"
	"rekberkuy/core-service/internal/domain"
)

// mockIdemRepo for testing the idempotency middleware.
type mockIdemRepo struct {
	existing  *domain.IdempotencyRecord
	isNew     bool
	saved     bool
	savedBody []byte
}

func (m *mockIdemRepo) CheckOrLock(ctx context.Context, rec *domain.IdempotencyRecord) (*domain.IdempotencyRecord, bool, error) {
	if m.isNew {
		return rec, true, nil
	}
	return m.existing, false, nil
}
func (m *mockIdemRepo) SaveResponse(ctx context.Context, id string, status int, body []byte) error {
	m.saved = true
	m.savedBody = body
	return nil
}

func newIdemRouter(repo domain.IdempotencyRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(handlers.IdempotencyMiddleware(repo))
	r.POST("/pay", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})
	return r
}

func TestIdempotencyMiddleware_NoKeyPassesThrough(t *testing.T) {
	repo := &mockIdemRepo{isNew: true}
	r := newIdemRouter(repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/pay", bytes.NewReader([]byte(`{}`)))
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
	if repo.saved {
		t.Error("without key, response must not be persisted to store")
	}
}

func TestIdempotencyMiddleware_NewRequestPersistsResponse(t *testing.T) {
	repo := &mockIdemRepo{isNew: true}
	r := newIdemRouter(repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/pay", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Idempotency-Key", "key-1")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
	if !repo.saved {
		t.Error("new request (with key) must persist response body")
	}
}

func TestIdempotencyMiddleware_DuplicateReplaysCachedResponse(t *testing.T) {
	cachedBody := []byte(`{"cached":true}`)
	repo := &mockIdemRepo{
		isNew:    false,
		existing: &domain.IdempotencyRecord{ID: "key-1", ResponseStatus: 201, ResponseBody: cachedBody},
	}
	r := newIdemRouter(repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/pay", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Idempotency-Key", "key-1")
	r.ServeHTTP(w, req)

	if w.Code != 201 {
		t.Errorf("status = %d, want 201 (from cache)", w.Code)
	}
	if w.Header().Get("X-Cache-Idempotency") != "true" {
		t.Error("X-Cache-Idempotency header must be true on replay")
	}
	if !bytes.Equal(w.Body.Bytes(), cachedBody) {
		t.Errorf("body = %s, want replay cache %s", w.Body.String(), cachedBody)
	}
}

func TestIdempotencyMiddleware_InFlightDuplicateReturns409(t *testing.T) {
	// Race window: the original request locked the key (placeholder row, status 0)
	// but has not finished + persisted its response yet. A concurrent duplicate must
	// NOT get a malformed empty 200 and must NOT double-execute the handler.
	handlerRan := false
	repo := &mockIdemRepo{
		isNew:    false,
		existing: &domain.IdempotencyRecord{ID: "key-1", ResponseStatus: 0, ResponseBody: nil},
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(handlers.IdempotencyMiddleware(repo))
	r.POST("/pay", func(c *gin.Context) { handlerRan = true; c.Status(http.StatusCreated) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/pay", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Idempotency-Key", "key-1")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409 (in-flight duplicate)", w.Code)
	}
	if handlerRan {
		t.Error("handler must NOT execute for an in-flight duplicate (no double-spending)")
	}
}
