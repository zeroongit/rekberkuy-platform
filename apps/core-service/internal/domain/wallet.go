package domain

import (
	"context"
	"time"
)

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
	UserID    string      `gorm:"type:uuid;primaryKey;not null" json:"user_id"`
	UserProfile UserProfile `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Balance   int64       `gorm:"type:bigint;not null;default:0" json:"balance"`
	IsFrozen  bool        `gorm:"type:boolean;not null;default:false" json:"is_frozen"`
	UpdatedAt time.Time   `gorm:"default:now()" json:"updated_at"`
}

type RekberPayTransaction struct {
	ID                     string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WalletID               string          `gorm:"type:uuid;not null;index" json:"wallet_id"`
	Wallet                 RekberPayWallet `gorm:"foreignKey:WalletID;constraint:OnDelete:CASCADE"`
	Type                   WalletTxType    `gorm:"type:varchar(50);not null" json:"type"`
	Status                 WalletTxStatus  `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	Amount                 int64           `gorm:"type:bigint;not null" json:"amount"`              
	AdminFee               int64           `gorm:"type:bigint;not null;default:0" json:"admin_fee"`           
	PlatformNetProfit      int64           `gorm:"type:bigint;not null;default:0" json:"platform_net_profit"` 
	ReferenceTransactionID *string         `gorm:"type:uuid" json:"reference_transaction_id,omitempty"`
	MidtransTopUpID        *string         `gorm:"type:varchar(255)" json:"midtrans_topup_id,omitempty"`
	Description            *string         `gorm:"type:text" json:"description,omitempty"`
	CreatedAt              time.Time       `gorm:"default:now()" json:"created_at"`
}

type WalletRepository interface {
	GetBalance(ctx context.Context, userID string) (*RekberPayWallet, error)
	CreateWallet(ctx context.Context, userID string) error
	UpdateBalanceTx(ctx context.Context, txRecord *RekberPayTransaction, amountModifier int64) error
	GetAllUsersForCRMEvaluation(ctx context.Context) ([]*UserProfile, error)
	ExecuteInTransaction(ctx context.Context, fn func(txRepo WalletRepository) error) error
	GetCRMLoyaltyByUserID(ctx context.Context, userID string) (*CRMLoyalty, error)
	UpdateCRMLoyalty(ctx context.Context, crmProfile *CRMLoyalty) error
	GetVendorAllocationsByTxID(ctx context.Context, transactionID string) ([]*EventVendorAllocation, error)
	CreateVendorPayoutRecord(ctx context.Context, payout *EventVendorPayout) error
	GetVendorPayoutByID(ctx context.Context, payoutID string) (*EventVendorPayout, error)
	UpdateVendorPayoutStatus(ctx context.Context, payoutID string, status string) error
}