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

// KYCClient is the identity-verification port (document + selfie checking).
// Like FraudClient, the real implementation calls backend-ai — but note the
// different semantics: the returned score is a CONFIDENCE reference for the
// reviewing admin. This port NEVER decides approval; the admin
// approve/reject flow in core-service does. An error from this port must not
// block the submission — the KYC row simply stays without an AI reference.
type KYCClient interface {
	// VerifyIdentity returns a verification confidence score (0-1, higher =
	// more likely genuine) plus a short reason from the AI verification service.
	VerifyIdentity(ctx context.Context, userID string, idCardURL string, selfieURL string, targetRole UserRole) (score float64, reason string, err error)
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

// DisbursementClient is the bank-payout port (Midtrans Payout/Disbursement
// product). This seam is PREPARED but not yet wired to a production caller:
// today the bank transfer completes out-of-band (ADR-0002 stance) and the
// ledger reconciles the real cost at disbursement confirmation (true-up).
// When Midtrans Payout credentials arrive, the HTTP adapter implements this
// port and the withdrawal flow can execute transfers in-system, reading the
// ACTUAL fee charged straight from the API response.
type DisbursementClient interface {
	// ExecuteDisbursement performs the bank transfer via the provider API and
	// reports the actual disbursement fee charged plus the provider reference.
	ExecuteDisbursement(ctx context.Context, req DisbursementRequest) (DisbursementResult, error)
	// EstimateDisbursementFee returns the current expected fee for a transfer.
	// The stub returns the configured estimate (MIDTRANS_DISBURSEMENT_FEE).
	EstimateDisbursementFee(ctx context.Context, req DisbursementRequest) (int64, error)
}

// DisbursementRequest describes a single bank payout.
type DisbursementRequest struct {
	WithdrawalID   string
	BankName       string
	AccountNumber  string
	AccountHolder  string
	Amount         int64
	ReferenceLabel string // free-form narration sent to the bank
}

// DisbursementResult reports the outcome of an executed disbursement.
type DisbursementResult struct {
	ReferenceID string // provider-side id (Midtrans payout id)
	Fee         int64  // ACTUAL fee charged — the source of truth for true-up
}
