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
│   ├── core-service/     # Backend Go — clean architecture (owns every decision & all money movement)
│   └── dashboard-web/     # Frontend Next.js v16
├── backend-ai/           # KYC + fraud-risk SCORING — Python/FastAPI/Groq (verification only, never decides)
├── blockchain/           # Smart contract audit log — Hardhat v3 (TransactionLogger.sol)
├── e2e-qa/               # QA Automation & E2E testing
├── docs/
│   ├── general-user-guide.md              # Task-focused user guide
│   ├── application-flow-and-module-guide.md # Business flows, state machines & backend module map — READ THIS FIRST
│   └── adr/                               # Architecture Decision Records — READ BEFORE redesigning any cross-service integration
├── CONTEXT.md            # Domain glossary (Transaction, Release, Disburse, etc.)
├── .github/workflows/ci.yml        # CI/CD pipeline
└── CLAUDE.md             # ← you are reading this
```

---

## 🔗 Service Boundary Contracts (read this before touching an interface)

RekberKuy has **three independently-developed backend services** (`core-service` Go, `backend-ai`
Python, `blockchain` Solidity). If you are working inside only ONE of these folders, it is easy to
break the system by changing an interface without checking its counterpart in a different
language/folder — that counterpart will NOT show up in a search scoped to your current folder.

| Boundary | Canonical contract | Must stay in sync with |
|---|---|---|
| core-service ↔ backend-ai (fraud) | `apps/core-service/internal/fraud/client.go` — `fraudAnalyzeRequest`/`fraudAnalyzeResponse` structs, calls `POST /api/v1/fraud/score` | `backend-ai/main.py` — `FraudRequest`/`FraudResponse` Pydantic models, same route |
| core-service ↔ backend-ai (KYC) | `apps/core-service/internal/domain/profile.go` (`KYCSubmission`) | `backend-ai/main.py` — `POST /api/v1/kyc/verify` exists but **has no Go caller yet** (see gap note below) |
| core-service ↔ blockchain | `apps/core-service/internal/relayer/relayer.go` — `loggerABI` constant (function + event signature) | `blockchain/contracts/TransactionLogger.sol` — `logTransaction` function + `TransactionLogged` event. Function selector and event topic0 **must stay byte-identical**; changing either breaks the Go relayer's calldata packing/log decoding. |
| Any pair of services | `docs/adr/` — decisions that constrain how services are allowed to talk to each other | Check for a relevant ADR before redesigning an integration point; don't reverse a documented decision without discussing it first |

**Rule of thumb:** if a task changes a function signature, a request/response shape, or an event
schema in one of the files above, open and check the counterpart file in the *same* change —
even though it lives in a different folder and a different language. The Go/Python/Solidity
boundary is not a reason to treat the other side as "someone else's problem."

**Known gap — do not silently "fix" it:** `backend-ai`'s KYC endpoint works, but `core-service`
never calls it, and there is no admin approve/reject flow for KYC at all yet (`KYCApproved` /
`KYCRejected` are currently unreachable states). If your task touches KYC, read
[`docs/application-flow-and-module-guide.md`](./docs/application-flow-and-module-guide.md#ai-assisted-admin-verification--current-status)
first — this is a documented, scoped-for-later gap, not something to wire up as a side effect of
an unrelated task.

---

## ⚙️ Important Commands

### Backend (Go)
```bash
cd apps/core-service
go mod tidy                        # Install dependencies
go run ./cmd/migrate/main.go up    # Apply DB migrations (golang-migrate) — run before first start
go run ./cmd/server/main.go        # Run server
go run ./cmd/seed/main.go          # Seed categories (3-tier) + mockup data
go test ./...                      # Run all tests
go build -o bin/server ./cmd/server/main.go  # Build binary
```

### AI Service (Python — verification only, never decides)
```bash
cd backend-ai
python3.11 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env               # then set GROQ_API_KEY (optional — service boots without it)
python main.py                     # serves on :8081 (matches core-service AI_SERVICE_URL default)
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
npx hardhat ignition deploy ./ignition/modules/TransactionLogger.ts --network localhost
# Production: pass DISTINCT owner (cold wallet) & relayer (hot key = DEPLOYER_PRIVATE_KEY)
# via --parameters params.json — see ignition/modules/TransactionLogger.ts for the shape.
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
- `dispute.go` — `Dispute` entity + `DisputeStatus`/`Outcome` — implemented, admin-mediated (see [ADR-0003](./docs/adr/0003-admin-mediated-dispute-resolution-ai-deferred.md))
- `review.go` — buyer review of counterparty after `RELEASED`
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

> **IMPORTANT:** The smart contract in the `blockchain/` folder (`TransactionLogger.sol`)
> functions **ONLY as a transaction audit log**. Not for holding funds, not on-chain escrow,
> not business logic.

Given that crypto payments are not legal in Indonesia, this platform implements a **Gasless
Transaction** approach. This means users do not interact directly with the blockchain at all.
The backend acts as a *Relayer* that executes and pays the *gas fee* automatically in the
background for every completed transaction recording — see the exact call sequence in
[`docs/application-flow-and-module-guide.md`](./docs/application-flow-and-module-guide.md#ai-assisted-admin-verification--current-status).

**What the smart contract may do:**
- Record `keccak256` hashes (`txIdHash`/`buyerHash`/`sellerHash`) + `amount`, emitted as the
  `TransactionLogged` event — no on-chain storage of the log entries themselves (event-only, for
  minimal relayer gas cost — see [ADR-0004](./docs/adr/0004-per-transaction-logging-instead-of-merkle-batching.md))
- Rotate the authorized `relayer` wallet via owner-only `setRelayer()`, without redeploying —
  so a compromised relayer key doesn't fracture the audit trail across two contracts

**What must NOT be in the smart contract:**
- Escrow logic or fund holding
- Any business logic (fee calculations, validation, etc.)
- Sensitive user data (raw IDs — only hashes ever reach the chain)

**ABI stability:** `logTransaction`'s signature and the `TransactionLogged` event schema must
stay byte-identical to `loggerABI` in `apps/core-service/internal/relayer/relayer.go` — see
[Service Boundary Contracts](#-service-boundary-contracts-read-this-before-touching-an-interface) above.

**Target network:** Avalanche C-Chain (`avalancheFuji`, chainId 43113, for development/staging;
Mainnet, chainId 43114, for production — only the Fuji network is configured in
`hardhat.config.ts` today).

---

## 🧪 Testing

| Test Type | Tool | Location |
|------------|------|----------|
| Frontend unit test | Jest | `apps/dashboard-web/__tests__/` |
| E2E & QA Automation | Playwright | `e2e-qa/` |
| API test | Postman | `docs/api/` (collection) |
| Smart contract test | Hardhat | `blockchain/test/` |
| Backend unit test | Go test | `apps/core-service/**/*_test.go` |
| AI service test | *(none yet — known gap)* | `backend-ai/` has no `*_test.go`; don't assume coverage exists when modifying it |

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
- Functions are only for **write** (record transaction) and **read** (read log) — no escrow, no fee math
- Emit an event for every new recording; prefer events over storage (`SSTORE`) wherever possible — the backend relayer pays gas on every call, so gas cost is a direct operating cost, not a one-time deployment concern
- Do not use `mapping` that stores sensitive data
- Access control via hand-written custom errors + modifiers (`onlyRelayer`, `onlyOwner`
  pattern in `TransactionLogger.sol`) — avoid pulling in OpenZeppelin for a single-purpose
  contract; every import adds deployment bytecode cost
- Any change to a public function or event signature must be reflected in `relayer.go`'s
  `loggerABI` in the same change — see [Service Boundary Contracts](#-service-boundary-contracts-read-this-before-touching-an-interface)

### IPFS Proof Storage Rules
- All milestone deliverables, vendor invoices, and event completion proofs MUST be uploaded to IPFS.
- The PostgreSQL database only stores the IPFS CID (`ipfs_cid` VARCHAR).
- Frontend renders IPFS files via Gateway URL: `https://gateway.pinata.cloud/ipfs/{ipfs_cid}`.

---

## 🔑 Environment Variables

Never hardcode secrets. Each service has its own `.env.example` — copy it, never commit the
real `.env`: `apps/core-service/.env.example`, `backend-ai/.env.example`,
`apps/dashboard-web/.env.example`.

Crucial cross-service variables (must be set consistently across services that share them):
- `DATABASE_URL` — PostgreSQL connection via Supabase (`core-service` only)
- `SUPABASE_SERVICE_ROLE_KEY` — do not expose to the frontend
- `DEPLOYER_PRIVATE_KEY` — the relayer's hot wallet key. Used by `core-service` to sign
  `logTransaction` calls AND by `blockchain/hardhat.config.ts` (`avalancheFuji` network) to
  deploy/sign — keep the same key in both `.env` files in development
- `AVALANCHE_RPC_URL` — Avalanche RPC endpoint (same value needed in both `core-service` and
  `blockchain`)
- `CONTRACT_ADDRESS` — deployed `TransactionLogger` address; `core-service` needs this to call
  the contract, so update it there immediately after every redeploy
- `GROQ_API_KEY` — optional; `backend-ai` runs in low-confidence fallback mode without it, so
  omitting it is safe for local dev but never for production fraud/KYC scoring
- `AI_SERVICE_URL` (`core-service`) — must point at wherever `backend-ai` is actually running
  (default `http://localhost:8081`, matching `backend-ai`'s own `PORT` default)
- `FRAUD_UNSAFE_THRESHOLD` / `FRAUD_FAIL_OPEN` — the fraud DECISION lives entirely in
  `core-service`; `backend-ai` has no equivalent config, it only returns a score

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
- Domain glossary (shared vocabulary across all three services): [`CONTEXT.md`](./CONTEXT.md)
- Architecture decisions that constrain cross-service integration: [`docs/adr/`](./docs/adr/)
- Domain model: [`apps/core-service/internal/domain/`](./apps/core-service/internal/domain/)
- AI service: [`backend-ai/README.md`](./backend-ai/README.md)
- Smart contract: [`blockchain/README.md`](./blockchain/README.md)
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