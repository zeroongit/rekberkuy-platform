package domain

import "context"

// This file holds the PORTS (interface contracts) for external services.
// Ports are defined in the domain (not in the adapter package) so that the usecase
// only depends on the domain — following Clean Architecture rules:
//   "usecase/ may only import domain/".
// Concrete adapters (fraud/, relayer/) live outside the domain and implement
// the interfaces defined here.

// FraudClient is the transaction risk-analysis port (anti-fraud scoring).
// The real implementation calls a Python microservice (backend-ai via Groq).
type FraudClient interface {
	// AnalyzeTransactionRisk returns a risk score (0-1) and an isSafe flag.
	// isSafe=false indicates a suspicious transaction that should be held/frozen.
	AnalyzeTransactionRisk(ctx context.Context, userID string, amount int64) (score float64, isSafe bool, err error)
}

// Relayer is the gasless audit-log port to the Avalanche blockchain.
// The backend acts as a relayer: it pays the gas and executes the smart contract
// in the background. Users do not need a crypto wallet.
type Relayer interface {
	// LogTransactionOnChain records a completed transaction as an on-chain audit log
	// and returns the transaction hash to be stored in the DB.
	LogTransactionOnChain(ctx context.Context, txID string, amount int64, buyer string, seller string) (txHash string, err error)
}

// MidtransClient is the Midtrans payment-gateway port (Snap Token + webhook).
// The concrete implementation (see midtrans package) calls the Snap API.
type MidtransClient interface {
	// CreateSnapTransaction creates a Snap transaction on the Midtrans side and returns
	// a snap token + redirect URL for the user to complete payment.
	CreateSnapTransaction(ctx context.Context, req SnapRequest) (SnapResult, error)
	// VerifySignature validates the SignatureKey from a Midtrans notification webhook.
	// Formula: SHA512(order_id + status_code + gross_amount + server_key).
	VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool
}

// SnapRequest parameters for creating a Midtrans Snap transaction.
type SnapRequest struct {
	OrderID       string
	GrossAmount   int64
	CustomerName  string
	CustomerEmail string
	CustomerPhone string
}

// SnapResult output of Snap transaction creation.
type SnapResult struct {
	Token       string
	RedirectURL string
}

// MidtransNotification core payload from a Midtrans notification webhook.
type MidtransNotification struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"` // capture | settlement | pending | deny | expire | cancel
	PaymentType       string `json:"payment_type"`
	FraudStatus       string `json:"fraud_status"`   // accept | deny | challenge
	TransactionID     string `json:"transaction_id"` // Midtrans ID (stored into MidtransTopUpID)
}

// MidtransOrderPrefixes distinguishes the transaction type from the order_id.
const (
	OrderPrefixGoods   = "REKBERKUY-GOODS-"
	OrderPrefixService = "REKBERKUY-SERVICE-"
	OrderPrefixEvent   = "REKBERKUY-EVENT-"
	OrderPrefixTopUp   = "REKBERKUY-TOPUP-"
)
