package domain

import "context"

// This file defines the Unit of Work contract: a single transactional boundary
// that encompasses all repositories so that cross-table mutations are ACID.
//
// Previously the transactional boundary was patched into WalletRepository.ExecuteInTransaction
// but the transaction & finance repositories were called outside that *sql.Tx -> ACID leaked.
// Now the boundary is held at a single point: UnitOfWork.Do.

// TxStores groups all repositories that may participate in a single database
// transaction unit. When UnitOfWork.Do runs, each field is filled with a
// repository instance bound to the same *sql.Tx so that all mutations
// commit/rollback together (Serializable Isolation).
type TxStores struct {
	Users        UserRepository
	Wallets      WalletRepository
	Transactions TransactionRepository
	Finance      FinanceRepository
	Disputes     DisputeRepository
	KYC          KYCRepository
}

// UnitOfWork is the cross-repository transactional boundary contract.
// The concrete implementation (see repository package) guarantees fn() runs
// inside a single database transaction: commit if fn returns nil,
// rollback if fn returns an error.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context, stores TxStores) error) error
}
