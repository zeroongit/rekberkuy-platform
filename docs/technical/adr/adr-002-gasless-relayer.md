# ADR-002: Gasless Blockchain Relayer for Audit Logging

## Status
ACCEPTED

## Context
Cryptocurrency payments and direct wallet interactions are not viable for legal and user-experience reasons in Indonesia. However, users and regulators require an immutable, verifiable audit trail for completed escrow transactions. Requiring users to pay gas or hold crypto on Avalanche is a severe UX barrier.

## Decision
Implement a **Gasless Transaction** model via a backend Relayer (`apps/core-service/internal/relayer/relayer.go`). The backend automatically pays the gas fee and signs `logTransaction` calls to the deployed Solidity smart contract (`TransactionLogger.sol`) on Avalanche Fuji Testnet / Mainnet upon successful transaction completion.

## Consequences
- Users enjoy zero-crypto friction while benefiting from cryptographic auditability.
- The platform operator bears the gas cost as an operational expense.
- The smart contract function selector and event signatures (`TransactionLogged`) must remain strictly byte-identical to the relayer ABI.
