# ADR-004: Per-Transaction On-Chain Logging Instead of Merkle Batching

## Status
ACCEPTED

## Context
When recording completed escrow transactions on the Avalanche blockchain, we evaluated whether to batch multiple transactions into a Merkle tree root commit or log each transaction individually.

## Decision
Record each completed escrow transaction as an individual `logTransaction` call (`TransactionLogger.sol`), rather than Merkle-tree batching.

## Consequences
- Immediate public verifiability: anyone can inspect a transaction's audit log event on-chain the moment it completes without needing backend proof-generation infrastructure.
- Gas cost per transaction is minimized by keeping the smart contract event-only (`bytes32` hashes + amount) with no dynamic on-chain storage.
- If future high-volume scaling demands gas optimization, Merkle batching can be reconsidered.
