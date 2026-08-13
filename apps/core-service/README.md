# ⚙️ RekberKuy - Core Service (Backend)

This module is the main backend for the **RekberKuy** platform, built with **Go 1.25** applying a pure **Clean Architecture** pattern. This backend manages all business logic, escrow transactions (Goods, Services, Events), user management (CRM Loyalty), and the *relayer* function for recording *gasless transactions* to the blockchain.

## 🏗️ Architecture (Clean Architecture)
This system is very strict in separating concerns through the following hierarchical layers:

1. **`internal/delivery/handlers/`**: REST API handlers (Gin). HTTP request parsing, payload validation, and JSON responses — no business logic lives here.
2. **`internal/usecase/`**: Pure business logic (escrow flows, `finance_calculator`, dispute resolution).
3. **`internal/repository/`**: Database access (PostgreSQL via Supabase, Redis) — implements the `domain` interfaces.
4. **`internal/domain/`**: Independent entity *structs* and *interface* contracts. Imports nothing else within the project.

> **Golden Rule:** An outer *layer* may only call a deeper *layer*. `domain` must not import packages from any *layer*.

## 💸 Transaction Context
- **Purely IDR**: This backend manages finances in Rupiah currency. There is no storage of users' crypto addresses.
- **Gasless Relayer**: The backend is responsible for firing *smart contracts* to Avalanche automatically in the background (*background job*). End users are not burdened with *gas fees*.
- **ACID Compliance**: All wallet/financial mutations must use *database transactions* and *pessimistic locking* (`SELECT ... FOR UPDATE`).

## 🚀 How to Run Locally

1. **Environment Preparation**: copy `.env.example` to `.env` and fill in real values (`DATABASE_URL` is required).
2. **Install Dependencies**:
   ```bash
   go mod tidy
   ```
3. **Apply Database Migrations** (required before the first run — the server no longer auto-creates tables):
   ```bash
   go run ./cmd/migrate/main.go up
   ```
   The schema is managed by [golang-migrate](https://github.com/golang-migrate/migrate) under `db/migrations/`.
   Run `go run ./cmd/migrate/main.go help` for `down` / `fresh` / `steps` / `force` / `version`.
4. **Run the Server**:
   ```bash
   go run ./cmd/server/main.go
   ```
5. **Build Binary**:
   ```bash
   go build -o bin/server ./cmd/server/main.go
   ```
6. **Seed Database** (optional, for development/testing):
   ```bash
   go run ./cmd/seed/main.go
   ```
   Injects a 3-tier category structure (Category → SubCategory → SubSubCategory) for the
   Goods, Services, Events & Vendor domains, along with mockup data (user profile, RekberPay wallet, and
   dummy transactions) ready to use for testing the `auto_release_worker`.

## 📌 Implementation Status

- ✅ Complete domain: `auth`, `category` (3-tier), `dispute`, `finance`, `profile`, `transaction`, `wallet`, `worker`
- ✅ Repository, usecase, and handlers for Goods/Services/Events transactions, user, vendor, wallet, KYC
- ✅ Credential-based auth: `POST /api/v1/auth/register` + `POST /api/v1/auth/login` (bcrypt + JWT)
- ✅ Middleware: JWT auth (RBAC), config-driven CORS, idempotency key (anti double-spending)
- ✅ Worker: `auto_release_worker` (auto-release escrow) & `crm_worker` (Loyalty Tiering evaluation)
- ✅ Database seeder (`cmd/seed/main.go`)
- ✅ Versioned SQL migrations via golang-migrate (`db/migrations/` + `cmd/migrate`); GORM AutoMigrate retired
- ✅ Adapters: `internal/midtrans` (Snap + signature), `internal/relayer` (go-ethereum gasless), `internal/fraud` (HTTP → backend-ai). Each falls back to a stub when its external service is not configured.
- ✅ Unit tests (`*_test.go`) across usecases, repositories (sqlmock), handlers, adapters, and config
- 🚧 `backend-ai` (Python KYC/fraud) and the `TransactionLogger` contract are out of tree; the Go adapters call stubs until those services exist

## 🧪 Testing
Every new *usecase* addition must include unit tests.
```bash
# Run all tests
go test ./...
```
