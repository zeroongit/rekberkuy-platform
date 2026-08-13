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

// fraudHTTPClient calls the /fraud/analyze endpoint on backend-ai.
type fraudHTTPClient struct {
	baseURL  string
	apiKey   string
	client   *http.Client
	failOpen bool // true = assume safe when the service is unreachable (availability);
	//        false (default) = return an error so callers refuse the release (money-safety).
}

// NewFraudHTTPClient creates an HTTP-based fraud adapter to backend-ai.
// baseURL example: "http://localhost:8081". failOpen=false is the safe default
// for an escrow platform (refuse release when fraud screening is unavailable).
func NewFraudHTTPClient(baseURL, apiKey string, timeout time.Duration, failOpen bool) domain.FraudClient {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &fraudHTTPClient{
		baseURL:  baseURL,
		apiKey:   apiKey,
		client:   &http.Client{Timeout: timeout},
		failOpen: failOpen,
	}
}

type fraudAnalyzeRequest struct {
	UserID string `json:"user_id"`
	Amount int64  `json:"amount"`
}

type fraudAnalyzeResponse struct {
	Score  float64 `json:"score"`
	IsSafe bool    `json:"is_safe"`
}

func (c *fraudHTTPClient) AnalyzeTransactionRisk(ctx context.Context, userID string, amount int64) (float64, bool, error) {
	reqBody := fraudAnalyzeRequest{UserID: userID, Amount: amount}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return 0, false, fmt.Errorf("failed to marshal fraud body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/fraud/analyze", bytes.NewReader(body))
	if err != nil {
		return 0, false, fmt.Errorf("failed to create fraud request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

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
	return out.Score, out.IsSafe, nil
}
