package usecase

import (
	"context"
	"fmt"
	"rekberkuy/core-service/internal/domain"
)


// TopUpWallet menangani logika validasi bisnis sebelum saldo diproses ke Midtrans
func (u *UserUsecase) TopUpWallet(ctx context.Context, userID string, amount int64) (*domain.RekberPayTransaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("nominal top-up harus lebih besar dari Rp 0")
	}

	descMsg := fmt.Sprintf("Permintaan top-up saldo via Midtrans sebesar Rp %d", amount)
	
	// Buat draf log transaksi dengan status PENDING sebelum dibayar oleh user
	initTx := &domain.RekberPayTransaction{
		WalletID:    userID,
		Type:        "TOP_UP",
		Status:      "PENDING", // Menunggu konfirmasi webhook Midtrans nanti
		Amount:      amount,
		Description: &descMsg,
	}

	// Panggil repository untuk mencatat draf transaksi secara aman di database
	err := u.walletRepo.UpdateBalanceTx(ctx, initTx, 0) // Delta 0 karena saldo baru bertambah pas webhook 'success'
	if err != nil {
		return nil, fmt.Errorf("gagal menginisialisasi transaksi top-up: %w", err)
	}

	return initTx, nil
}