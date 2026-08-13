# Per-transaction on-chain logging instead of Merkle-tree batching

`TransactionLogger.sol` records every completed escrow transaction as its own `logTransaction` call (one relayer-paid tx per platform transaction), rather than batching many transactions into a single Merkle root commit. At current and near-term volume, the relayer's per-call gas cost (already minimised: event-only, `bytes32` hashes, calldata params, no dynamic storage) is cheaper to operate and reason about than a batching pipeline, and it keeps the audit trail directly and immediately verifiable on-chain — anyone can inspect a completed transaction's event the moment it happens, with no dependency on the backend to serve a Merkle proof.

## Considered

- **Merkle-tree batching (commit one root for N transactions)** — deferred. Reduces on-chain call count, but requires the backend to run a batching window, retain leaf data long-term to serve proofs on demand, and expose a proof-verification endpoint — a new trust dependency on the backend for something the current design gets "for free" from a public event. Worthwhile once gas spend at real volume justifies the added moving parts; not before.
- **Simple batch array (send N hashes in one call, no proof system)** — a cheaper middle ground than Merkle batching if/when call-count reduction alone becomes the goal, without adding proof-generation machinery. Not implemented now because per-transaction logging is not yet a measured cost problem.

## Consequences

- `logTransaction` stays one-call-per-transaction; `relayer.go`'s ABI (`txIdHash, amount, buyerHash, sellerHash`) is the stable contract — no batching/proof plumbing needed on the Go side.
- Verifiability is immediate and requires no backend-served proof: a transaction's audit-log event exists on-chain as soon as the relayer's call is mined.
- Revisit this decision once real Fuji/Mainnet gas spend at production volume is measured. If cumulative relayer gas cost becomes material, evaluate the simple batch array first (lower complexity) before Merkle-tree batching (higher complexity, adds a backend-trust dependency for proofs).