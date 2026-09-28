# ADR-001: Backend Clean Architecture & Local JWT Authentication

## Status
ACCEPTED

## Context
RekberKuy platform requires a robust, testable, and maintainable backend architecture in Go to handle complex escrow logic, wallet ledger mutations, and cross-domain transactions without tight coupling to external frameworks or database drivers. Furthermore, while Supabase configuration keys are present, relying on Supabase Auth introduced an external dependency with zero initial implementation.

## Decision
1. **Clean Architecture:** Enforce strict 4-layer separation in `apps/core-service/`: `domain/` (entities & interfaces), `repository/` (DB persistence), `usecase/` (business logic), and `delivery/handlers/` (HTTP transport). Dependencies flow strictly inward.
2. **Local JWT Authentication:** Implement local email/password authentication using bcrypt hashing and issue self-signed HS256 JWT tokens stored in `HttpOnly`, `Secure`, `SameSite=Strict` cookies.

## Consequences
- Business logic is completely isolated from HTTP and database concerns, making unit testing straightforward.
- The team has full ownership of credential management and token issuance without third-party auth service lock-in.
