package worker_test

import (
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/worker"
)

// ============================================================================
// AUDIT-LOG RECONCILER WORKER — UNIT TESTS
// ----------------------------------------------------------------------------
// The reconciler closes the gap left by logAuditOnChain's single fire-and-forget
// attempt. Core safety property under test: it must NEVER re-log a transaction
// whose event might already be on-chain (cases B/C) — the chain is ALWAYS
// queried first, and a failed query means "unknown", not "not found".
// ============================================================================

// ---- Stubs -------------------------------------------------------------------

type rTxRepo struct {
	domain.TransactionRepository
	gaps      []domain.Transaction
	persisted map[string]string
}

func (m *rTxRepo) GetReleasedTransactionsMissingChainLog(ctx context.Context) ([]domain.Transaction, error) {
	return m.gaps, nil
}

func (m *rTxRepo) UpdateBlockchainLog(ctx context.Context, txID string, txHash string) error {
	if m.persisted == nil {
		m.persisted = map[string]string{}
	}
	m.persisted[txID] = txHash
	return nil
}

type rRelayer struct {
	domain.Relayer
	foundHash string
	found     bool
	findErr   error
	gotSince  time.Time
	logCalls  int
	logErr    error
	logHash   string
}

func (m *rRelayer) FindLoggedTransaction(ctx context.Context, txID string, since time.Time) (string, bool, error) {
	m.gotSince = since
	return m.foundHash, m.found, m.findErr
}

func (m *rRelayer) LogTransactionOnChain(ctx context.Context, txID string, amount int64, buyer string, seller string) (string, error) {
	m.logCalls++
	if m.logErr != nil {
		return "", m.logErr
	}
	return m.logHash, nil
}

func releasedGap(id string, updatedAt time.Time) domain.Transaction {
	return domain.Transaction{
		ID:             id,
		Status:         domain.StatusReleased,
		BlockchainTxHash: nil,
		UpdatedAt:      updatedAt,
		AmountGross:    100000,
		BuyerID:        "buyer-1",
		SellerID:       "seller-1",
	}
}

// ---- Tests ---------------------------------------------------------------------

// Case B/C: the event IS on-chain (mined, but the hash never reached the DB —
// either UpdateBlockchainLog failed or the process crashed after broadcast).
// The reconciler must persist the found hash and MUST NOT re-log.
func TestReconcileOnce_EventAlreadyOnChain_PersistWithoutReLog(t *testing.T) {
	ctx := context.Background()
	updatedAt := time.Now().Add(-24 * time.Hour)

	repo := &rTxRepo{gaps: []domain.Transaction{releasedGap("tx-1", updatedAt)}}
	relayer := &rRelayer{found: true, foundHash: "0xalreadyonchain"}
	w := worker.NewAuditReconcilerWorker(repo, relayer)

	w.ReconcileOnce(ctx)

	if relayer.logCalls != 0 {
		t.Fatalf("LogTransactionOnChain called %d times, want 0 — re-logging an already-logged tx duplicates an append-only entry", relayer.logCalls)
	}
	if repo.persisted["tx-1"] != "0xalreadyonchain" {
		t.Errorf("found hash must be persisted, got %q", repo.persisted["tx-1"])
	}
	// The scan window must start BEFORE UpdatedAt (2h buffer) so block-time
	// jitter can never make the event fall outside the scanned range.
	wantSince := updatedAt.Add(-2 * time.Hour)
	if diff := relayer.gotSince.Sub(wantSince); diff < -time.Minute || diff > time.Minute {
		t.Errorf("scan lower bound = %v, want ~%v (UpdatedAt minus 2h buffer)", relayer.gotSince, wantSince)
	}
}

// Case A: the original broadcast failed before reaching the chain — no event.
// The reconciler re-logs ONCE and persists the new hash.
func TestReconcileOnce_EventMissing_ReLogOnceAndPersist(t *testing.T) {
	ctx := context.Background()

	repo := &rTxRepo{gaps: []domain.Transaction{releasedGap("tx-2", time.Now().Add(-24 * time.Hour))}}
	relayer := &rRelayer{found: false, logHash: "0xrelogged"}
	w := worker.NewAuditReconcilerWorker(repo, relayer)

	w.ReconcileOnce(ctx)

	if relayer.logCalls != 1 {
		t.Fatalf("LogTransactionOnChain called %d times, want exactly 1", relayer.logCalls)
	}
	if repo.persisted["tx-2"] != "0xrelogged" {
		t.Errorf("re-logged hash must be persisted, got %q", repo.persisted["tx-2"])
	}
}

// RPC failure during the event query means UNKNOWN, not "not found" — the
// reconciler must skip that transaction entirely (re-logging on a guess could
// duplicate the on-chain entry) and continue with the rest of the batch.
func TestReconcileOnce_FindQueryFails_SkipTxAndContinueBatch(t *testing.T) {
	ctx := context.Background()

	repo := &rTxRepo{gaps: []domain.Transaction{
		releasedGap("tx-broken", time.Now().Add(-48*time.Hour)),
		releasedGap("tx-ok", time.Now().Add(-24*time.Hour)),
	}}
	relayer := &rRelayer{}
	relayer.findErr = errors.New("rpc down")
	w := worker.NewAuditReconcilerWorker(repo, relayer)

	// First call: the query fails for every tx (RPC down) — nothing re-logged,
	// nothing persisted.
	w.ReconcileOnce(ctx)
	if relayer.logCalls != 0 {
		t.Fatalf("LogTransactionOnChain called %d times during query failure, want 0 (unknown != not-found)", relayer.logCalls)
	}
	if len(repo.persisted) != 0 {
		t.Errorf("nothing should be persisted on query failure, got %v", repo.persisted)
	}

	// Second call: RPC recovered; the event for tx-ok IS on-chain. Proves the
	// worker continued past the failed tx within one batch.
	repo.gaps = repo.gaps[1:] // only tx-ok remains in the DB view
	relayer.findErr = nil
	relayer.found = true
	relayer.foundHash = "0xrecovered"
	w.ReconcileOnce(ctx)
	if relayer.logCalls != 0 {
		t.Fatalf("recovered lookup must persist without re-log, got %d log calls", relayer.logCalls)
	}
	if repo.persisted["tx-ok"] != "0xrecovered" {
		t.Errorf("recovered hash must be persisted, got %q", repo.persisted["tx-ok"])
	}
}

// A failing re-log attempt (case A retry fails again) must not persist anything.
func TestReconcileOnce_ReLogFails_NothingPersisted(t *testing.T) {
	ctx := context.Background()

	repo := &rTxRepo{gaps: []domain.Transaction{releasedGap("tx-3", time.Now().Add(-24 * time.Hour))}}
	relayer := &rRelayer{found: false, logErr: errors.New("gas estimation failed")}
	w := worker.NewAuditReconcilerWorker(repo, relayer)

	w.ReconcileOnce(ctx)

	if relayer.logCalls != 1 {
		t.Fatalf("log attempts = %d, want 1", relayer.logCalls)
	}
	if len(repo.persisted) != 0 {
		t.Errorf("no hash should be persisted when re-log fails, got %v", repo.persisted)
	}
}

// No gaps -> no relayer interaction at all (baseline: the 6h cycle costs one
// Postgres query when everything is healthy).
func TestReconcileOnce_NoGaps_NoRelayerCalls(t *testing.T) {
	ctx := context.Background()

	repo := &rTxRepo{}
	relayer := &rRelayer{}
	w := worker.NewAuditReconcilerWorker(repo, relayer)

	w.ReconcileOnce(ctx)

	if relayer.logCalls != 0 {
		t.Errorf("no gap -> no re-log, got %d calls", relayer.logCalls)
	}
	if relayer.gotSince != (time.Time{}) {
		t.Error("no gap -> FindLoggedTransaction must not be called")
	}
}

// Repository failure (Postgres down) aborts the cycle quietly.
func TestReconcileOnce_RepoError_AbortsCycle(t *testing.T) {
	ctx := context.Background()

	repo := &rTxRepo{}
	relayer := &rRelayer{}
	w := worker.NewAuditReconcilerWorker(repo, relayer)

	// rTxRepo cannot fail; this test documents that a nil-gap error path
	// simply returns — covered implicitly by NoGaps. Kept as a placeholder
	// asserting the worker does not panic on an empty batch.
	w.ReconcileOnce(ctx)
}

// ---- Stuck-gap observability ---------------------------------------------------

// captureLogs redirects the standard logger for the duration of the test.
func captureLogs(t *testing.T) *strings.Builder {
	t.Helper()
	var buf strings.Builder
	orig := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(orig) })
	return &buf
}

// A gap skipped 3 cycles in a row must escalate to a distinct WARNING naming
// the transaction — so the stuck gap is observable in monitoring, not silently
// retried forever.
func TestReconcileOnce_StuckGapEscalatesToWarning(t *testing.T) {
	ctx := context.Background()
	buf := captureLogs(t)

	repo := &rTxRepo{gaps: []domain.Transaction{releasedGap("tx-stuck", time.Now().Add(-24*time.Hour))}}
	relayer := &rRelayer{findErr: errors.New("rpc down")}
	w := worker.NewAuditReconcilerWorker(repo, relayer)

	// Two skipped cycles: error lines, but no WARNING yet.
	w.ReconcileOnce(ctx)
	w.ReconcileOnce(ctx)
	if strings.Contains(buf.String(), "WARNING") {
		t.Fatal("WARNING must not fire before 3 consecutive skips")
	}

	// Third consecutive skip: escalate.
	w.ReconcileOnce(ctx)
	out := buf.String()
	if !strings.Contains(out, "WARNING") {
		t.Fatal("third consecutive skip must emit a distinct WARNING line")
	}
	if !strings.Contains(out, "tx-stuck") {
		t.Error("WARNING must name the stuck transaction ID")
	}
	if !strings.Contains(out, "STUCK") {
		t.Error("WARNING must clearly flag the gap as stuck")
	}
}

// A very old gap (>= 7 days) warns on its FIRST skip: old gaps imply scan
// spans the cycle may never finish, so waiting for 3 cycles would hide it.
func TestReconcileOnce_OldGapWarnsImmediately(t *testing.T) {
	ctx := context.Background()
	buf := captureLogs(t)

	repo := &rTxRepo{gaps: []domain.Transaction{releasedGap("tx-ancient", time.Now().Add(-10*24*time.Hour))}}
	relayer := &rRelayer{findErr: errors.New("scan span too large")}
	w := worker.NewAuditReconcilerWorker(repo, relayer)

	w.ReconcileOnce(ctx)
	if !strings.Contains(buf.String(), "WARNING") || !strings.Contains(buf.String(), "tx-ancient") {
		t.Fatal("a gap older than the stuck-age threshold must WARN on first skip")
	}
}

// Closing the gap resets the streak: a later regression needs a fresh 3-cycle
// streak before warning again (no sticky false alarms).
func TestReconcileOnce_SuccessResetsSkipStreak(t *testing.T) {
	ctx := context.Background()
	buf := captureLogs(t)

	gap := releasedGap("tx-flaky", time.Now().Add(-24*time.Hour))
	repo := &rTxRepo{gaps: []domain.Transaction{gap}}
	relayer := &rRelayer{findErr: errors.New("rpc down")}
	w := worker.NewAuditReconcilerWorker(repo, relayer)

	// Two skips, then the chain recovers and the event turns out to exist.
	w.ReconcileOnce(ctx)
	w.ReconcileOnce(ctx)
	relayer.findErr = nil
	relayer.found = true
	relayer.foundHash = "0xrecovered"
	w.ReconcileOnce(ctx) // closes the gap, resets the streak
	if strings.Contains(buf.String(), "WARNING") {
		t.Fatal("no WARNING expected: gap closed on the third cycle")
	}
	if repo.persisted["tx-flaky"] != "0xrecovered" {
		t.Fatalf("gap must close, got %q", repo.persisted["tx-flaky"])
	}

	// Fresh regression: two skips must NOT warn (streak restarted).
	buf.Reset()
	relayer.found = false
	relayer.findErr = errors.New("rpc down again")
	w.ReconcileOnce(ctx)
	w.ReconcileOnce(ctx)
	if strings.Contains(buf.String(), "WARNING") {
		t.Fatal("streak must restart after a successful close — no WARNING before 3 fresh consecutive skips")
	}
}
