package kyc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"rekberkuy/core-service/internal/domain"
)

// This package is an ADAPTER that implements the domain.KYCClient port.
// The real implementation calls the Python microservice (backend-ai / Groq).
//
// Semantics differ from the fraud adapter: the score returned here is a
// CONFIDENCE reference (higher = more likely genuine) that the reviewing
// admin sees. The AI NEVER decides approval — the admin approve/reject flow
// in core-service does. Callers must therefore treat errors from this port
// as "no AI reference available", not as a reason to reject the submission.

// ---------------------------------------------------------------------------
// Stub (safe default for development before backend-ai is ready)
// ---------------------------------------------------------------------------

type kycClientStub struct{}

// NewKYCClientStub returns a KYCClient that reports "no reference" — it errors
// with a sentinel so the usecase stores no AI score (dev/test env).
func NewKYCClientStub() domain.KYCClient {
	return &kycClientStub{}
}

func (k *kycClientStub) VerifyIdentity(ctx context.Context, userID string, idCardURL string, selfieURL string, targetRole domain.UserRole) (float64, string, error) {
	return 0, "", fmt.Errorf("kyc ai verification not configured (stub)")
}

// ---------------------------------------------------------------------------
// HTTP client (calls the backend-ai Python FastAPI)
// ---------------------------------------------------------------------------

type kycHTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewKYCHTTPClient creates an HTTP-based KYC adapter to backend-ai.
// baseURL example: "http://localhost:8081".
func NewKYCHTTPClient(baseURL string, timeout time.Duration) domain.KYCClient {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &kycHTTPClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
	}
}

// kycVerifyRequest mirrors backend-ai's KYCRequest Pydantic model
// (backend-ai/main.py). Must stay field-compatible with it.
type kycVerifyRequest struct {
	UserID     string `json:"user_id"`
	IDCardURL  string `json:"id_card_url"`
	SelfieURL  string `json:"selfie_url"`
	TargetRole string `json:"target_role"`
}

// kycVerifyResponse mirrors backend-ai's KYCResponse (score + reason only).
// There is deliberately no verdict field — the decision is the admin's.
type kycVerifyResponse struct {
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

func (c *kycHTTPClient) VerifyIdentity(ctx context.Context, userID string, idCardURL string, selfieURL string, targetRole domain.UserRole) (float64, string, error) {
	reqBody := kycVerifyRequest{
		UserID:     userID,
		IDCardURL:  idCardURL,
		SelfieURL:  selfieURL,
		TargetRole: string(targetRole),
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return 0, "", fmt.Errorf("failed to marshal kyc body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/kyc/verify", bytes.NewReader(body))
	if err != nil {
		return 0, "", fmt.Errorf("failed to create kyc verify request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("kyc ai service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, "", fmt.Errorf("kyc ai service returned status %d", resp.StatusCode)
	}

	var out kycVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, "", fmt.Errorf("failed to decode kyc verify response: %w", err)
	}

	// Clamp: backend-ai already clamps to [0,1], but never trust a remote score
	// with money-adjacent flows — clamp again locally.
	if out.Score < 0 {
		out.Score = 0
	} else if out.Score > 1 {
		out.Score = 1
	}
	return out.Score, out.Reason, nil
}
