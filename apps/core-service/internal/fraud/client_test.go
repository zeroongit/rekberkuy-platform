package fraud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestFraudHTTPClient_Success verifies that the decision (isSafe) is computed
// locally from the score using the configured threshold — backend-ai only sent a score.
func TestFraudHTTPClient_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/fraud/score" {
			t.Errorf("path = %s, want /api/v1/fraud/score", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "" {
			t.Errorf("adapter must not forward any credentials to backend-ai, got Authorization=%q", auth)
		}
		// backend-ai returns score + reason ONLY (no is_safe — it does not decide).
		_ = json.NewEncoder(w).Encode(fraudAnalyzeResponse{Score: 0.87, Reason: "very large amount"})
	}))
	defer srv.Close()

	c := NewFraudHTTPClient(srv.URL, 2*time.Second, false, 0.5)
	score, isSafe, err := c.AnalyzeTransactionRisk(context.Background(), "user-1", 500000)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if score != 0.87 {
		t.Errorf("score = %v, want 0.87", score)
	}
	if isSafe {
		t.Error("isSafe = true, want false (0.87 >= threshold 0.5 — core-service decided unsafe)")
	}
}

// TestFraudHTTPClient_BelowThresholdSafe proves core-service decides safe when
// the score is below ITS threshold.
func TestFraudHTTPClient_BelowThresholdSafe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(fraudAnalyzeResponse{Score: 0.2})
	}))
	defer srv.Close()

	c := NewFraudHTTPClient(srv.URL, 2*time.Second, false, 0.5)
	_, isSafe, err := c.AnalyzeTransactionRisk(context.Background(), "user-1", 1000)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if !isSafe {
		t.Error("isSafe = false, want true (0.2 < threshold 0.5)")
	}
}

// TestFraudHTTPClient_ThresholdIsOwnDecision proves the threshold is core-service's,
// not backend-ai's: a score of 0.4 is unsafe when core-service's threshold is 0.3,
// even though 0.4 would be safe under the typical 0.5.
func TestFraudHTTPClient_ThresholdIsOwnDecision(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(fraudAnalyzeResponse{Score: 0.4})
	}))
	defer srv.Close()

	c := NewFraudHTTPClient(srv.URL, 2*time.Second, false, 0.3) // stricter threshold
	_, isSafe, err := c.AnalyzeTransactionRisk(context.Background(), "user-1", 1000)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if isSafe {
		t.Error("isSafe = true, want false (0.4 >= core-service threshold 0.3)")
	}
}

// TestFraudHTTPClient_Error503FailClosed: backend-ai returns 503 (e.g. Groq down);
// with failOpen=false the adapter errors so the caller refuses the release.
func TestFraudHTTPClient_Error503FailClosed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := NewFraudHTTPClient(srv.URL, 2*time.Second, false, 0.5)
	_, _, err := c.AnalyzeTransactionRisk(context.Background(), "user-1", 1000)
	if err == nil {
		t.Fatal("fail-closed should return a non-nil error when backend-ai returns 503")
	}
}

// TestFraudHTTPClient_Error503FailOpen: same 503 but failOpen=true -> assume safe.
func TestFraudHTTPClient_Error503FailOpen(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := NewFraudHTTPClient(srv.URL, 2*time.Second, true, 0.5)
	_, isSafe, err := c.AnalyzeTransactionRisk(context.Background(), "user-1", 1000)
	if err != nil {
		t.Fatalf("fail-open should return nil error, got: %v", err)
	}
	if !isSafe {
		t.Error("fail-open should return isSafe=true on 503")
	}
}

// TestFraudHTTPClient_FailOpenMode: with failOpen=true, an unreachable service is assumed safe.
func TestFraudHTTPClient_FailOpenMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("should not be called")
	}))
	srv.Close() // unreachable

	c := NewFraudHTTPClient(srv.URL, 1*time.Second, true, 0.5)
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

	c := NewFraudHTTPClient(srv.URL, 1*time.Second, false, 0.5)
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
