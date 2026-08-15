package kyc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"rekberkuy/core-service/internal/domain"
)

// TestKYCHTTPClient_Success verifies the request shape sent to backend-ai
// (must mirror its KYCRequest model) and that the response is passed through
// as a score+reason REFERENCE — no verdict is computed here.
func TestKYCHTTPClient_Success(t *testing.T) {
	var gotBody kycVerifyRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/kyc/verify" {
			t.Errorf("path = %s, want /api/v1/kyc/verify", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", ct)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		// backend-ai returns score + reason ONLY (no verdict — it does not decide).
		_ = json.NewEncoder(w).Encode(kycVerifyResponse{Score: 0.91, Reason: "same-person"})
	}))
	defer srv.Close()

	c := NewKYCHTTPClient(srv.URL, 2*time.Second)
	score, reason, err := c.VerifyIdentity(context.Background(), "user-1", "https://cdn/ktp.jpg", "https://cdn/selfie.jpg", domain.RoleVerifiedMerchant)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if score != 0.91 {
		t.Errorf("score = %v, want 0.91", score)
	}
	if reason != "same-person" {
		t.Errorf("reason = %s, want same-person", reason)
	}
	if gotBody.UserID != "user-1" || gotBody.IDCardURL != "https://cdn/ktp.jpg" || gotBody.SelfieURL != "https://cdn/selfie.jpg" {
		t.Errorf("request body mismatch: %+v", gotBody)
	}
	if gotBody.TargetRole != "VERIFIED_MERCHANT" {
		t.Errorf("target_role = %s, want VERIFIED_MERCHANT", gotBody.TargetRole)
	}
}

// TestKYCHTTPClient_ClampsOutOfRangeScore proves a misbehaving remote score is
// clamped locally — never trust a remote score unconditionally.
func TestKYCHTTPClient_ClampsOutOfRangeScore(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(kycVerifyResponse{Score: 7.5})
	}))
	defer srv.Close()

	c := NewKYCHTTPClient(srv.URL, 2*time.Second)
	score, _, err := c.VerifyIdentity(context.Background(), "user-1", "a", "b", domain.RoleVerifiedVendor)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if score != 1 {
		t.Errorf("score = %v, want clamped to 1", score)
	}
}

// TestKYCHTTPClient_Error503: backend-ai returns 503 (Groq down) — the adapter
// errors. Callers treat this as "no AI reference", never as a rejection.
func TestKYCHTTPClient_Error503(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := NewKYCHTTPClient(srv.URL, 2*time.Second)
	_, _, err := c.VerifyIdentity(context.Background(), "user-1", "a", "b", domain.RoleVerifiedMerchant)
	if err == nil {
		t.Fatal("expected a non-nil error when backend-ai returns 503")
	}
}

// TestKYCHTTPClient_Unreachable: connection failure returns an error
// (reference-only failure, the submission itself stays valid upstream).
func TestKYCHTTPClient_Unreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("should not be called")
	}))
	srv.Close()

	c := NewKYCHTTPClient(srv.URL, 1*time.Second)
	_, _, err := c.VerifyIdentity(context.Background(), "user-1", "a", "b", domain.RoleVerifiedMerchant)
	if err == nil {
		t.Fatal("expected a non-nil error when backend-ai is unreachable")
	}
}

// TestKYCStub_ReportsUnavailable ensures the stub always reports
// "no reference available" instead of inventing a score.
func TestKYCStub_ReportsUnavailable(t *testing.T) {
	c := NewKYCClientStub()
	_, _, err := c.VerifyIdentity(context.Background(), "u", "a", "b", domain.RoleVerifiedMerchant)
	if err == nil {
		t.Fatal("stub must return an error signalling 'no AI reference'")
	}
}
