package worker

import (
	"context"
	"log"
	"time"

	"rekberkuy/core-service/internal/domain"
)

const (
	// auditReconcileInterval is deliberately sparse (6h): unlike auto-release,
	// closing an audit-log gap is not time-sensitive — the money already moved
	// correctly; only the immutable receipt is missing.
	auditReconcileInterval = 6 * time.Hour
	// chainScanLookback buffers the UpdatedAt→block-number conversion. Avalanche
	// block time is ~2s but not constant, so the chain scan starts 2h BEFORE the
	// release timestamp — scanning slightly too far back costs a few extra
	// eth_getLogs chunks, while starting too late could miss the event and tempt
	// a duplicate re-log of an append-only entry.
	chainScanLookback = 2 * time.Hour
	// reconcileTxTimeout bounds one transaction's chain scan (many chunks) so a
	// slow RPC cannot stall the whole cycle.
	reconcileTxTimeout = 10 * time.Minute
	// reconcileBatchPause spaces out per-transaction scans to stay gentle on
	// RPC provider rate limits.
	reconcileBatchPause = 250 * time.Millisecond

	// Stuck-gap observability: a gap that keeps being skipped must surface as
	// a distinct WARNING, not drown among per-cycle info logs. WARN when the
	// SAME transaction has been skipped this many consecutive cycles...
	warnAfterConsecutiveSkips = 3
	// ...or when the gap itself is this old, even if the skip count is still
	// low (e.g. scan spans too large to finish within the cycle timeout).
	warnStuckGapAge = 7 * 24 * time.Hour
)

// AuditReconcilerWorker closes the gap left by logAuditOnChain's single
// fire-and-forget attempt ("a separate reconciler can retry" — this is it).
//
// Why it must NEVER blindly re-log a NULL-hash transaction: the original
// failure has three indistinguishable-from-DB shapes —
//
//	A: broadcast failed before reaching the chain  -> event absent   -> safe to re-log
//	B: tx mined but UpdateBlockchainLog failed      -> event present -> must persist only
//	C: crash after broadcast, before any commit     -> event present -> must persist only
//
// B and C cannot be told apart from A without querying the chain, so the chain
// is ALWAYS queried first (FindLoggedTransaction by keccak256(txID) topic):
// event found → persist the found hash; definitively absent → re-log once; the
// query itself failed → UNKNOWN, skip (a wrong guess could permanently
// duplicate an append-only on-chain entry).
type AuditReconcilerWorker struct {
	txRepo   domain.TransactionRepository
	relayer  domain.Relayer
	ticker   *time.Ticker
	stopChan chan struct{}

	// skipStreaks counts CONSECUIVE skip cycles per transaction ID (in-memory;
	// resets on process restart and on any successful persist). Not a persisted
	// cursor — its only job is to make perpetually-stuck gaps observable via a
	// distinct WARNING log instead of silent infinite retries. ReconcileOnce is
	// expected to run single-threaded (one worker goroutine / manual ops run),
	// so no mutex guards this map.
	skipStreaks map[string]int
}

// NewAuditReconcilerWorker initializes the on-chain audit-log gap closer.
func NewAuditReconcilerWorker(tr domain.TransactionRepository, rel domain.Relayer) *AuditReconcilerWorker {
	return &AuditReconcilerWorker{
		txRepo:      tr,
		relayer:     rel,
		stopChan:    make(chan struct{}),
		skipStreaks: map[string]int{},
	}
}

func (w *AuditReconcilerWorker) Start(ctx context.Context) {
	w.ticker = time.NewTicker(auditReconcileInterval)
	log.Println("⛓️  RECONCILER: on-chain audit-log reconciler started (6h cycle)")

	go func() {
		for {
			select {
			case <-w.ticker.C:
				w.ReconcileOnce(ctx)
			case <-w.stopChan:
				log.Println("⛓️  RECONCILER: safely stopped.")
				return
			}
		}
	}()
}

// ReconcileOnce runs one reconciliation cycle. Exported so it can be triggered
// manually (ops) and exercised by tests without waiting for the ticker.
func (w *AuditReconcilerWorker) ReconcileOnce(ctx context.Context) {
	gaps, err := w.txRepo.GetReleasedTransactionsMissingChainLog(ctx)
	if err != nil {
		log.Printf("⛓️  RECONCILER ERROR: failed to scan for missing audit logs: %v", err)
		return
	}
	if len(gaps) == 0 {
		log.Println("⛓️  RECONCILER: no missing on-chain audit logs.")
		return
	}

	log.Printf("⛓️  RECONCILER: found %d released transactions missing their on-chain audit log.", len(gaps))
	for i, tx := range gaps {
		if i > 0 {
			select {
			case <-time.After(reconcileBatchPause):
			case <-ctx.Done():
				return
			}
		}
		w.reconcileOne(ctx, &tx)
	}
	log.Println("⛓️  RECONCILER: reconciliation cycle completed.")
}

// reconcileOne closes a single transaction's audit-log gap.
func (w *AuditReconcilerWorker) reconcileOne(ctx context.Context, tx *domain.Transaction) {
	// Per-transaction timeout so one long scan cannot stall the whole cycle.
	txCtx, cancel := context.WithTimeout(ctx, reconcileTxTimeout)
	defer cancel()

	// ALWAYS query the chain before re-logging anything (cases B/C above).
	since := tx.UpdatedAt.Add(-chainScanLookback)
	txHash, found, err := w.relayer.FindLoggedTransaction(txCtx, tx.ID, since)
	if err != nil {
		// UNKNOWN, not "not found": skipping is the only safe move — a wrong
		// re-log would permanently duplicate the append-only on-chain entry.
		log.Printf("⛓️  RECONCILER ERROR: chain lookup for %s failed, skipping this cycle: %v", tx.ID, err)
		w.noteSkip(tx, "chain lookup failed: "+err.Error())
		return
	}

	if found {
		if err := w.txRepo.UpdateBlockchainLog(txCtx, tx.ID, txHash); err != nil {
			log.Printf("⛓️  RECONCILER ERROR: %s is already on-chain (%s) but persisting the hash failed: %v", tx.ID, txHash, err)
			w.noteSkip(tx, "persist of found hash failed: "+err.Error())
			return
		}
		w.noteClosed(tx)
		log.Printf("⛓️  RECONCILER SUCCESS: %s was already on-chain (%s); hash persisted without re-logging.", tx.ID, txHash)
		return
	}

	// Case A confirmed: the scan range (which provably covers everything since
	// before the release) contains no event — the broadcast never landed.
	newHash, err := w.relayer.LogTransactionOnChain(txCtx, tx.ID, tx.AmountGross, tx.BuyerID, tx.SellerID)
	if err != nil {
		log.Printf("⛓️  RECONCILER ERROR: re-logging %s failed, will retry next cycle: %v", tx.ID, err)
		w.noteSkip(tx, "re-log failed: "+err.Error())
		return
	}
	if err := w.txRepo.UpdateBlockchainLog(txCtx, tx.ID, newHash); err != nil {
		log.Printf("⛓️  RECONCILER ERROR: %s re-logged (%s) but persisting the hash failed: %v", tx.ID, newHash, err)
		w.noteSkip(tx, "persist of re-logged hash failed: "+err.Error())
		return
	}
	w.noteClosed(tx)
	log.Printf("⛓️  RECONCILER SUCCESS: %s re-logged on-chain (%s) and hash persisted.", tx.ID, newHash)
}

// noteSkip records one more consecutive skip for the transaction and escalates
// to a distinct WARNING (for monitoring/alerting greps) once the gap qualifies
// as stuck: >= warnAfterConsecutiveSkips consecutive cycles, or the gap is
// older than warnStuckGapAge (very old gaps imply scan spans the cycle can
// never finish — e.g. RPC range caps or persistent provider failures).
func (w *AuditReconcilerWorker) noteSkip(tx *domain.Transaction, reason string) {
	w.skipStreaks[tx.ID]++
	streak := w.skipStreaks[tx.ID]
	age := time.Since(tx.UpdatedAt)
	if streak >= warnAfterConsecutiveSkips || age >= warnStuckGapAge {
		log.Printf("⛓️  RECONCILER WARNING: transaction %s STUCK — skipped %d consecutive cycles (gap age %s; last reason: %s). It will keep retrying silently unless resolved: check scan span vs RPC block-range caps, provider rate limits, and chain lookup health.",
			tx.ID, streak, age.Truncate(time.Hour), reason)
	}
}

// noteClosed resets the skip streak after the gap is successfully closed, so a
// future regression for the same transaction needs a fresh streak to warn.
func (w *AuditReconcilerWorker) noteClosed(tx *domain.Transaction) {
	delete(w.skipStreaks, tx.ID)
}

func (w *AuditReconcilerWorker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
	close(w.stopChan)
}
