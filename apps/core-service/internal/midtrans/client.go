package midtrans

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"rekberkuy/core-service/internal/domain"
)

// This package is an ADAPTER that implements the domain.MidtransClient port.
// It calls the Midtrans Snap API (server-to-server, Basic Auth with Server Key).

const (
	snapEndpointSandbox = "https://app.sandbox.midtrans.com/snap/v1/transactions"
	snapEndpointProd    = "https://app.midtrans.com/snap/v1/transactions"
)

type snapClient struct {
	serverKey string
	snapURL   string
	http      *http.Client
}

// NewSnapClient creates the Midtrans adapter. The "production" environment uses the
// production URL; otherwise sandbox. Used when cfg.MidtransEnabled().
func NewSnapClient(serverKey, environment string, timeout time.Duration) domain.MidtransClient {
	url := snapEndpointSandbox
	if environment == "production" {
		url = snapEndpointProd
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &snapClient{
		serverKey: serverKey,
		snapURL:   url,
		http:      &http.Client{Timeout: timeout},
	}
}

// snapAPIRequest is the body structure for a request to the Midtrans Snap API.
type snapAPIRequest struct {
	TransactionDetails transactionDetails `json:"transaction_details"`
	CustomerDetails    customerDetails    `json:"customer_details,omitempty"`
}

type transactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

type customerDetails struct {
	FirstName string `json:"first_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

type snapAPIResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
	// Error fields (when Midtrans rejects)
	ErrorMessage string `json:"error_message,omitempty"`
}

// CreateSnapTransaction creates a Snap transaction and returns the token + redirect URL.
func (c *snapClient) CreateSnapTransaction(ctx context.Context, req domain.SnapRequest) (domain.SnapResult, error) {
	body := snapAPIRequest{
		TransactionDetails: transactionDetails{
			OrderID:     req.OrderID,
			GrossAmount: req.GrossAmount,
		},
		CustomerDetails: customerDetails{
			FirstName: req.CustomerName,
			Email:     req.CustomerEmail,
			Phone:     req.CustomerPhone,
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return domain.SnapResult{}, fmt.Errorf("failed to marshal snap request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.snapURL, bytes.NewReader(raw))
	if err != nil {
		return domain.SnapResult{}, fmt.Errorf("failed to create snap request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	// Basic Auth with server key (username: password = serverKey:empty)
	httpReq.SetBasicAuth(c.serverKey, "")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return domain.SnapResult{}, fmt.Errorf("failed to call midtrans snap API: %w", err)
	}
	defer resp.Body.Close()

	var out snapAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return domain.SnapResult{}, fmt.Errorf("failed to decode snap response (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode >= 400 || out.Token == "" {
		return domain.SnapResult{}, fmt.Errorf("midtrans snap rejected request (status %d): %s", resp.StatusCode, out.ErrorMessage)
	}
	return domain.SnapResult{Token: out.Token, RedirectURL: out.RedirectURL}, nil
}

// VerifySignature validates the Midtrans webhook SignatureKey.
// Official formula: SHA512(order_id + status_code + gross_amount + server_key).
func (c *snapClient) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	payload := orderID + statusCode + grossAmount + c.serverKey
	sum := sha512.Sum512([]byte(payload))
	return hex.EncodeToString(sum[:]) == signatureKey
}
