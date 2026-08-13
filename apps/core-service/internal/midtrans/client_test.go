package midtrans

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"rekberkuy/core-service/internal/domain"
)

// TestVerifySignature_Valid verifies the Midtrans SHA512 formula.
func TestVerifySignature_Valid(t *testing.T) {
	c := &snapClient{serverKey: "SBK-test-server-key"}
	orderID := "REKBERKUY-TOPUP-abc12345"
	statusCode := "200"
	grossAmount := "150000.00"
	payload := orderID + statusCode + grossAmount + c.serverKey
	sum := sha512.Sum512([]byte(payload))
	validSig := hex.EncodeToString(sum[:])

	if !c.VerifySignature(orderID, statusCode, grossAmount, validSig) {
		t.Error("valid signature should return true")
	}
	if c.VerifySignature(orderID, statusCode, grossAmount, "0xbad") {
		t.Error("invalid signature should return false")
	}
}

// TestCreateSnapTransaction_Success verifies parsing of the Snap API response.
func TestCreateSnapTransaction_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "SBK-server-key" || pass != "" {
			t.Errorf("basic auth = %s:%s, want SBK-server-key: (empty pass)", user, pass)
		}
		var body map[string]json.RawMessage
		_ = json.NewDecoder(r.Body).Decode(&body)
		if _, ok := body["transaction_details"]; !ok {
			t.Error("body should contain transaction_details")
		}
		_ = json.NewEncoder(w).Encode(snapAPIResponse{Token: "snap-abc", RedirectURL: "https://snap.example.com/snap-abc"})
	}))
	defer srv.Close()

	c := &snapClient{
		serverKey: "SBK-server-key",
		snapURL:   srv.URL,
		http:      &http.Client{Timeout: 2 * time.Second},
	}
	res, err := c.CreateSnapTransaction(context.Background(), domain.SnapRequest{
		OrderID:     "REKBERKUY-TOPUP-xyz",
		GrossAmount: 150000,
	})
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if res.Token != "snap-abc" || !strings.Contains(res.RedirectURL, "snap-abc") {
		t.Errorf("snap result mismatched: %+v", res)
	}
}

// TestCreateSnapTransaction_ServerError ensures an error is returned when Midtrans rejects.
func TestCreateSnapTransaction_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(snapAPIResponse{ErrorMessage: "order_id duplicated"})
	}))
	defer srv.Close()

	c := &snapClient{
		serverKey: "SBK-server-key",
		snapURL:   srv.URL,
		http:      &http.Client{Timeout: 2 * time.Second},
	}
	_, err := c.CreateSnapTransaction(context.Background(), domain.SnapRequest{
		OrderID:     "DUP",
		GrossAmount: 1000,
	})
	if err == nil {
		t.Fatal("expected error when Midtrans rejects (400)")
	}
}

// TestNewSnapClient_EnvironmentURL verifies the sandbox/production URL selection.
func TestNewSnapClient_EnvironmentURL(t *testing.T) {
	sb := NewSnapClient("k", "sandbox", time.Second).(*snapClient)
	if sb.snapURL != snapEndpointSandbox {
		t.Errorf("sandbox url = %s", sb.snapURL)
	}
	pr := NewSnapClient("k", "production", time.Second).(*snapClient)
	if pr.snapURL != snapEndpointProd {
		t.Errorf("prod url = %s", pr.snapURL)
	}
}
