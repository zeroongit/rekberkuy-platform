# 🧠 Gemini Code Assist - Global Documentation

The main guide (Global Rules) for developing the **RekberKuy Platform**. Read this file to understand the big picture of the system before writing any code.

**IMPORTANT**: There is also a module-specific `gemini.md` file in each folder (`apps/core-service/`, `apps/dashboard-web/`, and `blockchain/`) with technical details per module.

---

## 🚀 Project Overview
**RekberKuy** is a trusted joint-account (escrow) platform for **Goods**, **Services**, and **Event** transactions (plus a Vendor Marketplace).
- **Core Principle**: All transactions and money logic (purely IDR) run on the **Backend**.
- **Transparency**: Every completed transaction is recorded as a permanent *Audit Log* on the **Avalanche Blockchain**.
- **Legal Compliance (Gasless)**: Because crypto payments are not yet legal in Indonesia, the system runs purely on Rupiah (IDR). Blockchain interaction is implemented via **Gasless Transactions** executed by the Backend acting as a relayer.

---

## 🏗️ Main System Structure & Architecture

### 1. Backend (`apps/core-service/`)
- **Stack**: Golang 1.25, PostgreSQL (via Supabase), Redis.
- **Architecture**: Pure Clean Architecture.
  - **Dependency Flow**: `delivery/http` ➔ `usecase` ➔ `repository` ➔ `domain`.
  - **Strict Rules**: Do not skip layers (e.g. `delivery` calling `repository` directly) or import in reverse (e.g. `domain` importing `usecase`).
- **Financial Context**: This system is purely **Rupiah (IDR)**. Do not use or request crypto addresses from users. The Backend automatically broadcasts transactions to the blockchain in the background.

### 2. Frontend (`apps/dashboard-web/`)
- **Stack**: Next.js v16, Tailwind CSS v4, Shadcn/UI, TypeScript.
- **Architecture**: Strictly **App Router** (`src/app/`).
- **Strict Rules**:
  - Use *Server Components* by default.
  - Prefer existing Shadcn/UI components before building your own.
  - Do not use CSS modules or styled-components (Tailwind only).
  - Avoid the `any` type unless absolutely necessary.

### 3. Blockchain (`blockchain/`)
- **Stack**: Solidity, Hardhat v3, Avalanche C-Chain.
- **Sole Purpose**: **AUDIT LOG ONLY**.
- **Strict Rules**: Do NOT put escrow logic, hold funds, perform fee calculations, or store PII/sensitive data (including in `mapping`) on-chain.

---

## 🧪 Testing Conventions
- **Backend**: Unit tests in Go (`go test`).
- **Frontend**: Unit tests with Jest (`npm run test`).
- **Smart Contract**: Hardhat Test (`npx hardhat test`).
- **E2E & QA**: Playwright & automation scenarios live in the `e2e-qa/` folder.

Every new function/usecase in the backend and smart contract **MUST** have a unit test covering both success and failure scenarios.

---

## 🚫 Absolute Restrictions (Do NOT)
1. **DO NOT** hardcode *secrets*, database URLs, API keys, or RPC URLs in code. Always use *Environment Variables* (see `.env.example`).
2. **DO NOT** ignore `error` handling in Golang (always `return error`, never `panic`).
3. **DO NOT** ever mutate a user's wallet balance (`wallet_repository.go`) without using a Database Transaction (`BeginTx`) and *Anti Race-Condition* protection (`SELECT ... FOR UPDATE`).
4. **DO NOT** turn the *smart contract* (`Counter.sol` / `TransactionLogger.sol`) into a *hold funds* function. Payment processing logic (Midtrans/RekberPay) is strictly handled by the backend.

---

## 🤝 Git Workflow & Contribution
- Use the **Conventional Commits** format:
  - `feat:` (new feature)
  - `fix:` (bug fix)
  - `refactor:` (refactoring without changing functionality)
  - `test:` (adding tests)
- Always branch from `develop` (`feature/feature-name`).
