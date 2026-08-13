# CLAUDE.md — RekberKuy Platform

This guide helps Claude AI understand the context, architecture, and code conventions of the RekberKuy project. Read this entire file before making any changes.

---

## 🧠 Project Overview

**RekberKuy** is a joint account (escrow) platform for **Goods**, **Services**, and **Event** transactions. Every transaction is recorded as an immutable audit log on the Avalanche blockchain. The platform also provides a vendor marketplace for Event Organizers (EO).

**Three core business domains:**
- Buying and selling goods (physical & digital)
- Services (freelance / professional)
- Event procurement + vendor marketplace

---

## 📁 Project Structure

```
rekberkuy-platform/
├── apps/
│   ├── core-service/     # Backend Go — clean architecture
│   └── dashboard-web/     # Frontend Next.js v16
├── blockchain/           # Smart contract audit log — Hardhat v3
├── e2e-qa/               # QA Automation & E2E testing
├── docs/
│   ├── general-user-guide.md              # Task-focused user guide
│   └── application-flow-and-module-guide.md # Business flows, state machines & backend module map — READ THIS FIRST
├── .github/workflows/ci.yml        # CI/CD pipeline
└── CLAUDE.md             # ← you are reading this
```

---

## ⚙️ Important Commands

### Backend (Go)
```bash
cd apps/core-service
go mod tidy                        # Install dependencies
go run ./cmd/server/main.go        # Run server
go test ./...                      # Run all tests
go build -o bin/server ./cmd/server/main.go  # Build binary
```

### Frontend (Next.js)
```bash
cd apps/dashboard-web
npm install                        # Install dependencies
npm run dev                        # Dev server (localhost:3000)
npm run build                      # Production build
npm run lint                       # Lint check
npm run test                       # Unit test (Jest)
```

### Blockchain (Hardhat)
```bash
cd blockchain
npm install
npx hardhat compile                # Compile smart contract
npx hardhat test                   # Run contract tests
npx hardhat node                   # Run local blockchain (port 8545)
npx hardhat ignition deploy ./ignition/modules/Counter.ts --network localhost
```

### QA / E2E
```bash
cd e2e-qa
npm install
npm run test                       # Run all QA scenarios
```

---

## 🏛️ Backend Architecture (Clean Architecture)

Layers are sequential — **do not skip layers, do not import in reverse**:

```
delivery/handlers/  →  usecase/  →  repository/  →  domain/
```

| Layer | Location | Responsibility |
|-------|----------|----------------|
| `domain/` | `internal/domain/` | Entity structs & interface contracts |
| `repository/` | `internal/repository/` | Database access — implementation of domain interfaces |
| `usecase/` | `internal/usecase/` | Pure business logic — must not know about HTTP |
| `delivery/` | `internal/delivery/handlers/` | HTTP handlers — only request parsing & response |

### Existing domain files (`internal/domain/`)
- `auth.go` — JWT custom claims & auth contract
- `profile.go` — `UserProfile`, `VendorProfile`, `CRMLoyalty` (tiering), `KYCSubmission`
- `transaction.go` — escrow `Transaction` entity, universal status state machine, milestones, event vendor payouts/allocations
- `finance.go` — `PlatformFinance`, `EventAuditResult`, fee constants
- `wallet.go` — `RekberPayWallet`, ledger mutations, idempotency contract
- `category.go` — 3-tier taxonomy (Goods / Services / Events / Vendors)
- `dispute.go` — dispute entity (reserved for the dispute-resolution module)
- `services.go` — external service ports: `FraudClient`, `Relayer`, `MidtransClient`
- `worker.go` — background worker contract (`CRMWorker`)
- `unit_of_work.go` — transactional boundary (`UnitOfWork` + `TxStores`)

### Dependency rules
- `domain/` must not import any other package within the project
- `usecase/` may only import `domain/`
- `repository/` may only import `domain/`
- `delivery/` may import `usecase/` and `domain/`

---

## 🌐 Frontend Architecture (Next.js v16)

- Use **App Router** (`src/app/`) — not Pages Router
- UI components from **Shadcn/UI** — do not build UI components from scratch if they already exist in Shadcn
- Styling only with **Tailwind CSS** — no CSS modules or styled-components
- All components must be **TypeScript** — no `.js` or `.jsx` files
- Use **Server Components** by default; add `"use client"` only when interactivity is truly required

---

## ⛓️ Blockchain — Role & Limitations

> **IMPORTANT:** The smart contract in the `blockchain/` folder functions **ONLY as a transaction audit log**. Not for holding funds, not on-chain escrow, not business logic.

Given that crypto payments are not legal in Indonesia, this platform implements a **Gasless Transaction** approach. This means users do not interact directly with the blockchain at all. The Backend acts as a *Relayer* that executes and pays the *gas fee* automatically in the background for every completed transaction recording.

**What the smart contract may do:**
- Record the hash/ID of completed transactions
- Store the timestamp & final status of transactions
- Emit events for indexing & transparency purposes

**What must NOT be in the smart contract:**
- Escrow logic or fund holding
- Any business logic (fee calculations, validation, etc.)
- Sensitive user data

**Target network:** Avalanche C-Chain (Fuji Testnet for development, Mainnet for production)

---

## 🧪 Testing

| Test Type | Tool | Location |
|------------|------|----------|
| Frontend unit test | Jest | `apps/dashboard-web/__tests__/` |
| E2E & QA Automation | Playwright | `e2e-qa/` |
| API test | Postman | `docs/api/` (collection) |
| Smart contract test | Hardhat | `blockchain/test/` |
| Backend unit test | Go test | `apps/core-service/**/*_test.go` |

**Testing rules:**
- Every new usecase **must** have a unit test
- Every new endpoint **must** be added to the Postman collection
- Main transaction flows (goods/services/event) **must** have E2E scenarios in `e2e-qa/`

---

## 📐 Code Conventions

### Go (Backend)
- File names: `snake_case` (example: `transaction_usecase.go`)
- Struct & interface names: `PascalCase`
- Exported function names: `PascalCase`, internal functions: `camelCase`
- Error handling: always return `error`, do not `panic` except in `main.go`
- Interfaces are defined in `domain/`, implemented in `repository/` or `usecase/`
- Use `context.Context` as the first parameter in all functions that touch I/O

```go
// ✅ Correct
func (u *transactionUsecase) CreateTransaction(ctx context.Context, req domain.Transaction) (domain.Transaction, error) {}

// ❌ Wrong — no context, no error return
func CreateTx(req domain.Transaction) domain.Transaction {}
```

### TypeScript (Frontend)
- Component file names: `PascalCase.tsx` (example: `TransactionCard.tsx`)
- Utility/hook file names: `camelCase.ts` (example: `useTransaction.ts`)
- Always define types — avoid `any`
- Use `interface` for component props, `type` for union/intersection
- Component names must be descriptive and reflect the business domain

```tsx
// ✅ Correct
interface TransactionCardProps {
  transactionId: string
  status: 'pending' | 'completed' | 'disputed'
}

// ❌ Wrong
const Card = ({ data }: { data: any }) => {}
```

### Solidity (Smart Contract)
- Functions are only for **write** (record transaction) and **read** (read log)
- Emit an event for every new recording
- Do not use `mapping` that stores sensitive data

---

## 🔑 Environment Variables

Never hardcode secrets. All config is in `apps/core-service/.env`.

Crucial variables that must exist:
- `DATABASE_URL` — PostgreSQL connection via Supabase
- `SUPABASE_SERVICE_ROLE_KEY` — do not expose to the frontend
- `DEPLOYER_PRIVATE_KEY` — blockchain deployment wallet private key, **highly sensitive**
- `AVALANCHE_RPC_URL` — Avalanche RPC endpoint

---

## 🚫 Things You Must Not Do

- ❌ Do not commit `.env` files to the repository
- ❌ Do not hardcode URLs, ports, or credentials in code
- ❌ Do not add business logic in the `delivery/handlers/` layer
- ❌ Do not import `usecase/` from `domain/` (violates clean architecture)
- ❌ Do not use `any` in TypeScript unless there is truly no other choice
- ❌ Do not add escrow logic or fund holding to the smart contract
- ❌ Do not push directly to the `main` or `develop` branch — always via Merge Request

---

## ✅ Checklist Before Making Changes

Before writing code, make sure you have:
- [ ] Read `docs/application-flow-and-module-guide.md` to understand the relevant business flow
- [ ] Understood which layer needs to change (domain / repository / usecase / delivery)
- [ ] Not violated the dependency rules between layers
- [ ] Prepared tests for the new code
- [ ] Used names consistent with the existing business domain

---

## 📚 Important References

- Business flows, state machines & module map: [`docs/application-flow-and-module-guide.md`](./docs/application-flow-and-module-guide.md)
- General user guide: [`docs/general-user-guide.md`](./docs/general-user-guide.md)
- Domain model: [`apps/core-service/internal/domain/`](./apps/core-service/internal/domain/)
- Frontend app: [`apps/dashboard-web/`](./apps/dashboard-web/)
- CI/CD pipeline: [`.github/workflows/ci.yml`](./.github/workflows/ci.yml)

---

## Agent skills

### Issue tracker

Issues are tracked as GitHub issues in this repo (via the `gh` CLI). See `docs/agents/issue-tracker.md`.

### Triage labels

Five canonical triage roles map 1:1 to GitHub labels of the same name. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context — one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.
