package fraud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestFraudHTTPClient_Success verifies decoding of the backend-ai response.
func TestFraudHTTPClient_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fraud/analyze" {
			t.Errorf("path = %s, want /fraud/analyze", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("authorization header = %q, want Bearer test-key", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(fraudAnalyzeResponse{Score: 0.87, IsSafe: false})
	}))
	defer srv.Close()

	c := NewFraudHTTPClient(srv.URL, "test-key", 2*time.Second, false)
	score, isSafe, err := c.AnalyzeTransactionRisk(context.Background(), "user-1", 500000)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if score != 0.87 {
		t.Errorf("score = %v, want 0.87", score)
	}
	if isSafe {
		t.Error("isSafe = true, want false (suspicious transaction)")
	}
}

// TestFraudHTTPClient_FailOpenMode: with failOpen=true, an unreachable service is assumed safe.
func TestFraudHTTPClient_FailOpenMode(t *testing.T) {
	// server that immediately closes the connection -> error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("should not be called")
	}))
	srv.Close() // close it so it becomes unreachable

	c := NewFraudHTTPClient(srv.URL, "key", 1*time.Second, true) // failOpen = true
	score, isSafe, err := c.AnalyzeTransactionRisk(context.Background(), "user-1", 1000)
	if err != nil {
		t.Fatalf("fail-open should return nil error, got: %v", err)
	}
	if !isSafe {
		t.Error("fail-open should return isSafe=true")
	}
	if score != 0.1 {
		t.Errorf("fail-open score = %v, want 0.1", score)
	}
}

// TestFraudHTTPClient_FailClosedMode: with failOpen=false (the safe default), an
// unreachable fraud service returns an error so the caller refuses the release.
func TestFraudHTTPClient_FailClosedMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("should not be called")
	}))
	srv.Close()

	c := NewFraudHTTPClient(srv.URL, "key", 1*time.Second, false) // failOpen = false (default)
	_, _, err := c.AnalyzeTransactionRisk(context.Background(), "user-1", 1000)
	if err == nil {
		t.Fatal("fail-closed should return a non-nil error when the service is unreachable")
	}
}

// TestFraudStub_DefaultSafe ensures the stub always considers transactions safe.
func TestFraudStub_DefaultSafe(t *testing.T) {
	c := NewFraudClientStub()
	_, isSafe, err := c.AnalyzeTransactionRisk(context.Background(), "u", 1)
	if err != nil {
		t.Fatalf("stub error: %v", err)
	}
	if !isSafe {
		t.Error("stub should return isSafe=true")
	}
}
