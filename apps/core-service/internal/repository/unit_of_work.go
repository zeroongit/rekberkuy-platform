package repository

import (
	"context"
	"database/sql"
	"fmt"

	"rekberkuy/core-service/internal/domain"
)

// sqlUnitOfWork is the domain.UnitOfWork implementation on top of database/sql.
// A single *sql.Tx (Serializable) encompasses all repositories participating
// in fn(), so that wallet + transaction + finance + profile mutations commit
// together or roll back together -> satisfying the ACID requirement of the escrow platform.
type sqlUnitOfWork struct {
	db *sql.DB
}

// NewUnitOfWork initializes the application's main transactional boundary.
func NewUnitOfWork(db *sql.DB) domain.UnitOfWork {
	return &sqlUnitOfWork{db: db}
}

func (uow *sqlUnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, stores domain.TxStores) error) error {
	tx, err := uow.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to begin unit of work: %w", err)
	}

	// Each repository is wrapped with the same *sql.Tx. The repo methods will
	// automatically use that tx (the r.tx != nil path), including SELECT ... FOR UPDATE.
	stores := domain.TxStores{
		Users:        &userRepository{db: uow.db, tx: tx},
		Wallets:      &walletRepository{db: uow.db, tx: tx},
		Transactions: &TransactionRepository{db: uow.db, tx: tx},
		Finance:      &financeRepository{db: uow.db, tx: tx},
		Disputes:     &disputeRepository{db: uow.db, tx: tx},
		KYC:          &kycRepository{db: uow.db, tx: tx},
	}

	if err := fn(ctx, stores); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit unit of work: %w", err)
	}
	return nil
}
