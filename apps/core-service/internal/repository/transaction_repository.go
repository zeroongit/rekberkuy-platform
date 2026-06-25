package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"rekberkuy/core-service/internal/domain"
)

type transactionRepository struct {
	db *sql.DB
	tx *sql.Tx
}

// NewTransactionRepository menginisialisasi adapter database untuk transaksi escrow
func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

// CreateTransaction menyimpan data master transaksi escrow murni IDR ke database Supabase
func (r *transactionRepository) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	query := `
		INSERT INTO transactions (
			id, buyer_id, seller_id, type, status, amount_base, shipping_fee, 
			service_fee, midtrans_fee, amount_gross, amount_net, 
			midtrans_order_id, idempotency_key, payment_method, 
			blockchain_tx_hash, blockchain_logged_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
	`

	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query,
			tx.ID, tx.BuyerID, tx.SellerID, tx.Type, tx.Status, tx.AmountBase, tx.ShippingFee,
			tx.ServiceFee, tx.MidtransFee, tx.AmountGross, tx.AmountNet,
			tx.MidtransOrderID, tx.IdempotencyKey, tx.PaymentMethod,
			tx.BlockchainTxHash, tx.BlockchainLoggedAt,
		)
	} else {
		_, err = r.db.ExecContext(ctx, query,
			tx.ID, tx.BuyerID, tx.SellerID, tx.Type, tx.Status, tx.AmountBase, tx.ShippingFee,
			tx.ServiceFee, tx.MidtransFee, tx.AmountGross, tx.AmountNet,
			tx.MidtransOrderID, tx.IdempotencyKey, tx.PaymentMethod,
			tx.BlockchainTxHash, tx.BlockchainLoggedAt,
		)
	}

	if err != nil {
		return fmt.Errorf("gagal membuat data transaksi escrow: %w", err)
	}
	return nil
}

// GetTransactionByID mengambil detail data transaksi lengkap dengan pengunci baris (FOR UPDATE)
func (r *transactionRepository) GetTransactionByID(ctx context.Context, id string) (*domain.Transaction, error) {
	query := `
		SELECT id, buyer_id, seller_id, type, status, amount_base, shipping_fee, 
		       service_fee, midtrans_fee, amount_gross, amount_net, 
		       midtrans_order_id, idempotency_key, payment_method, 
		       blockchain_tx_hash, blockchain_logged_at, created_at, updated_at
		FROM transactions 
		WHERE id = $1
	`
	
	if r.tx != nil {
		query += " FOR UPDATE"
	}

	var tx domain.Transaction
	var err error

	if r.tx != nil {
		err = r.tx.QueryRowContext(ctx, query, id).Scan(
			&tx.ID, &tx.BuyerID, &tx.SellerID, &tx.Type, &tx.Status, &tx.AmountBase, &tx.ShippingFee,
			&tx.ServiceFee, &tx.MidtransFee, &tx.AmountGross, &tx.AmountNet,
			&tx.MidtransOrderID, &tx.IdempotencyKey, &tx.PaymentMethod,
			&tx.BlockchainTxHash, &tx.BlockchainLoggedAt, &tx.CreatedAt, &tx.UpdatedAt,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, id).Scan(
			&tx.ID, &tx.BuyerID, &tx.SellerID, &tx.Type, &tx.Status, &tx.AmountBase, &tx.ShippingFee,
			&tx.ServiceFee, &tx.MidtransFee, &tx.AmountGross, &tx.AmountNet,
			&tx.MidtransOrderID, &tx.IdempotencyKey, &tx.PaymentMethod,
			&tx.BlockchainTxHash, &tx.BlockchainLoggedAt, &tx.CreatedAt, &tx.UpdatedAt,
		)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("transaksi dengan id %s tidak ditemukan", id)
		}
		return nil, err
	}

	return &tx, nil
}

// UpdateTransactionStatus memproses transisi State Machine (Universal State Machine)
func (r *transactionRepository) UpdateTransactionStatus(ctx context.Context, id string, status domain.TransactionStatus) error {
	query := `
		UPDATE transactions 
		SET status = $1, updated_at = NOW() 
		WHERE id = $2
	`

	var err error
	if r.tx != nil {
		_, err = r.tx.ExecContext(ctx, query, status, id)
	} else {
		_, err = r.db.ExecContext(ctx, query, status, id)
	}

	if err != nil {
		return fmt.Errorf("gagal memperbarui status transaksi: %w", err)
	}
	return nil
}