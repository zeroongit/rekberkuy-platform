package domain

import (
	"context"
	"time"
)

// Mengunci Aturan Bisnis Top-Up & Withdraw Rekberkuy (Hukum Tetap)

type WalletTxType string
const (
	TxTopUp        WalletTxType = "TOPUP"
	TxPayment      WalletTxType = "PAYMENT"
	TxReceiveFunds WalletTxType = "RECEIVE_FUNDS"
	TxWithdraw     WalletTxType = "WITHDRAW"
	TxRefund       WalletTxType = "REFUND"
)

type WalletTxStatus string
const (
	WalletStatusPending WalletTxStatus = "PENDING"
	WalletStatusSuccess WalletTxStatus = "SUCCESS"
	WalletStatusFailed  WalletTxStatus = "FAILED"
)

type RekberPayWallet struct {
	UserID    string    `json:"user_id"`
	Balance   int64     `json:"balance"`
	IsFrozen  bool      `json:"is_frozen"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RekberPayTransaction struct {
	ID                     string         `json:"id"`
	WalletID               string         `json:"wallet_id"`
	Type                   WalletTxType   `json:"type"`
	Status                 WalletTxStatus `json:"status"`
	Amount                 int64          `json:"amount"`                   // Nominal bersih top-up atau withdraw yang diminta
	AdminFee               int64          `json:"admin_fee"`                 // Biaya 1% atau Rp 7.500 yang ditarik dari user
	PlatformNetProfit      int64          `json:"platform_net_profit"`      // Dihitung dinamis nanti di Usecase setelah dipotong Midtrans riil
	ReferenceTransactionID *string        `json:"reference_transaction_id,omitempty"`
	MidtransTopUpID        *string        `json:"midtrans_topup_id,omitempty"`
	Description            *string        `json:"description,omitempty"`
	CreatedAt              time.Time      `json:"created_at"`
}

type WalletRepository interface {
	GetBalance(ctx context.Context, userID string) (*RekberPayWallet, error)
	CreateWallet(ctx context.Context, userID string) error
	UpdateBalanceTx(ctx context.Context, txRecord *RekberPayTransaction, amountModifier int64) error
}