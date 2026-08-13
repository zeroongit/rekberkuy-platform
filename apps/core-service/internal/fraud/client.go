package fraud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"rekberkuy/core-service/internal/domain"
)

// This package is an ADAPTER that implements the domain.FraudClient port.
// The real implementation calls the Python microservice (backend-ai / Groq).

// ---------------------------------------------------------------------------
// Stub (safe default for development before backend-ai is ready)
// ---------------------------------------------------------------------------

type fraudClientStub struct{}

// NewFraudClientStub returns a FraudClient that always considers a
// transaction safe (score 0.1, isSafe true). Used when backend-ai is not active.
func NewFraudClientStub() domain.FraudClient {
	return &fraudClientStub{}
}

func (f *fraudClientStub) AnalyzeTransactionRisk(ctx context.Context, userID string, amount int64) (float64, bool, error) {
	return 0.1, true, nil
}

// ---------------------------------------------------------------------------
// HTTP client (calls the backend-ai Python FastAPI)
// ---------------------------------------------------------------------------

// fraudHTTPClient calls the /api/v1/fraud/score endpoint on backend-ai.
//
// backend-ai is a VERIFICATION service: it returns only a score (+ reason).
// This adapter owns the DECISION: it computes isSafe by comparing the score to
// the platform's own UnsafeThreshold. On any error reaching backend-ai, the
// failOpen policy decides (fail-closed = refuse the release, the money-safe default).
type fraudHTTPClient struct {
	baseURL          string
	client           *http.Client
	unsafeThreshold  float64 // score >= threshold -> unsafe (core-service's decision)
	failOpen         bool     // true = assume safe when the service is unreachable (availability);
	//        false (default) = return an error so callers refuse the release (money-safety).
}

// NewFraudHTTPClient creates an HTTP-based fraud adapter to backend-ai.
// baseURL example: "http://localhost:8081". unsafeThreshold is the platform's own
// decision threshold (score >= threshold -> unsafe). failOpen=false is the safe
// default for an escrow platform (refuse release when fraud screening is unavailable).
func NewFraudHTTPClient(baseURL string, timeout time.Duration, failOpen bool, unsafeThreshold float64) domain.FraudClient {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &fraudHTTPClient{
		baseURL:         baseURL,
		client:          &http.Client{Timeout: timeout},
		unsafeThreshold: unsafeThreshold,
		failOpen:        failOpen,
	}
}

type fraudAnalyzeRequest struct {
	UserID string `json:"user_id"`
	Amount int64  `json:"amount"`
}

// fraudAnalyzeResponse mirrors backend-ai's verification payload (score + reason only).
// is_safe is intentionally absent — the decision is core-service's.
type fraudAnalyzeResponse struct {
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

func (c *fraudHTTPClient) AnalyzeTransactionRisk(ctx context.Context, userID string, amount int64) (float64, bool, error) {
	reqBody := fraudAnalyzeRequest{UserID: userID, Amount: amount}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return 0, false, fmt.Errorf("failed to marshal fraud body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/fraud/score", bytes.NewReader(body))
	if err != nil {
		return 0, false, fmt.Errorf("failed to create fraud request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		if c.failOpen {
			// Availability mode: assume safe when the fraud service is unreachable.
			return 0.1, true, nil
		}
		return 0, false, fmt.Errorf("fraud service unreachable (fail-closed): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		if c.failOpen {
			return 0.1, true, nil
		}
		return 0, false, fmt.Errorf("fraud service returned status %d (fail-closed)", resp.StatusCode)
	}

	var out fraudAnalyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		if c.failOpen {
			return 0.1, true, nil
		}
		return 0, false, fmt.Errorf("failed to decode fraud response (fail-closed): %w", err)
	}

	// THE DECISION LIVES HERE, NOT IN backend-ai: score >= threshold -> unsafe.
	isSafe := out.Score < c.unsafeThreshold
	return out.Score, isSafe, nil
}
