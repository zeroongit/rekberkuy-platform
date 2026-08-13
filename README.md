# RekberKuy Platform
Tailored Escrow & Payment Settlement System for Goods, Services, and Events

## Overview
RekberKuy is a comprehensive modern escrow platform designed to secure transactions across physical/digital goods, professional services, and event management. It provides end-to-end management of buyer-seller funds locking, milestone-based disbursements, multi-tier vendor allocations, and gasless blockchain audit logging.

## Documentation

- [General user guide](docs/general-user-guide.md) — task-focused instructions for everyday platform users.
- [Application flow and module guide](docs/application-flow-and-module-guide.md) — detailed business flows, domain state machines, and technical map.

## Core Features

### Goods Escrow Management
- Automated funds locking and release lifecycle for physical and digital products
- Shipping courier tracking and delivery status confirmation
- Auto-confirmation timeout system powered by background workers
- Dispute resolution and mediation workflow for buyer-seller conflicts
- Transaction history with immutable audit logs

### Services & Milestone Payment
- Proposal and project posting for clients and freelancers
- Milestone-based payment release structure (phased disbursements)
- Project deadline monitoring and deliverable verification
- Escrow protection for both service providers and clients
- Performance review and rating integration

### Event & Vendor Management
- End-to-end event management system for Event Organizers (EO)
- Integrated Multi-Tier Vendor Marketplace (Sound System, Catering, Venue, Decor)
- Vendor allocation tracking and contract escrow management
- Phased vendor payout requests with invoice upload and review workflow
- Live event budget monitoring and disbursement dashboard

### Wallet & Financial Ledger
- Internal RekberPay digital wallet with real-time balance tracking
- Concurrency protection via Serializable Isolation and SELECT FOR UPDATE
- Automated platform fee, payment gateway fee, and net payout calculation
- Multi-channel top-up integration via Midtrans payment gateway
- Monthly CRM Loyalty Tiering evaluation with automated point distribution

### Gasless Blockchain Audit Logging
- Immutable transaction logging on Avalanche network
- Backend relayer system subsidizing gas fees for seamless UX
- Publicly verifiable transaction proofs without requiring crypto wallets
- Complete transparent audit trail for high-value event escrow

### Identity & Fraud Protection (AI-Powered)
- Automated KYC (Know Your Customer) document verification
- AI-driven ID card and selfie verification scoring via Groq
- Automated fraud scoring prior to releasing locked funds
- Idempotency key middleware preventing double-spending and duplicate requests

### Administrative Features
- Role-based access control (User, Verified Merchant, Vendor, Event Organizer, Admin)
- Real-time transaction state monitoring and dispute override controls
- Multi-tier taxonomy management for Goods, Services, Events, and Vendors
- Background worker monitoring and job scheduling

## Tech Stack

Built with:
- Go 1.25 (Core Service microservice with Clean Architecture)
- Next.js v16 (Modern web application frontend with App Router)
- Python 3.11 (FastAPI AI microservice for KYC & Fraud Detection)
- PostgreSQL & Supabase (Scalable relational database with BaaS features)

### Additional Technologies
- **Frontend**: Shadcn/UI, Tailwind CSS v4, TypeScript
- **Database & Cache**: PostgreSQL, Redis
- **Blockchain**: Avalanche Network, Hardhat v3, Solidity
- **AI & ML**: Groq AI API, FastAPI
- **Payment Gateway**: Midtrans Sandbox/Production
- **CI/CD & DevOps**: Docker, GitHub Actions

## Getting Started

```sh
# Start Core Service Backend (Go)
cd apps/core-service
go run ./cmd/migrate/main.go up   # apply database migrations (golang-migrate)
go run ./cmd/server/main.go

# Seed Master Taxonomy & Mockup Data
go run ./cmd/seed/main.go

# Start Web Dashboard (Next.js)
cd apps/dashboard-web
npm install
npm run dev