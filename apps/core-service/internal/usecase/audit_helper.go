package usecase

import (
	"context"
	"log"
	"time"

	"rekberkuy/core-service/internal/domain"
)

// logAuditOnChain records a completed (RELEASED) transaction to the blockchain
// as an audit log (best-effort, asynchronous) then stores the tx hash in the DB.
//
// Called AFTER the database transaction commits, not inside it, because
// on-chain confirmation is slow (seconds) and must not hold a Serializable lock.
// Uses a separate (detached) context so it is not cancelled when the HTTP
// request completes. Failures are logged; a separate reconciler can retry.
func logAuditOnChain(relayer domain.Relayer, txRepo domain.TransactionRepository, txID, buyerID, sellerID string, amount int64) {
	if relayer == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		txHash, err := relayer.LogTransactionOnChain(ctx, txID, amount, buyerID, sellerID)
		if err != nil {
			log.Printf("[RELAYER] failed to audit-log transaction %s: %v", txID, err)
			return
		}
		if err := txRepo.UpdateBlockchainLog(ctx, txID, txHash); err != nil {
			log.Printf("[RELAYER] transaction %s recorded on-chain (%s) but failed to save hash: %v", txID, txHash, err)
		}
	}()
}
