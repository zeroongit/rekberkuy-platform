# Operations Runbook — RekberKuy Platform

## 1. Overview
Runbook ini memuat panduan operasional harian, prosedur deployment, migrasi database, dan pemeliharaan sistem untuk tim engineering RekberKuy.

## 2. Common Operations & Commands

### A. Backend (Go core-service)
```bash
cd apps/core-service
go mod tidy                        # Install dependencies
go run ./cmd/migrate/main.go up    # Apply DB migrations (golang-migrate)
go run ./cmd/server/main.go        # Run backend server (port 8080)
go run ./cmd/seed/main.go          # Seed categories & mockup data
go test ./...                      # Run all backend unit tests
```

### B. AI Service (Python FastAPI)
```bash
cd backend-ai
python3.11 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
python main.py                     # Runs on port 8081
```

### C. Frontend (Next.js v16)
```bash
cd apps/dashboard-web
npm install
npm run dev                        # Dev server (localhost:3000)
npm run build                      # Production build
npm run test                       # Unit tests (Jest)
```

### D. Docker Compose Deployment
```bash
docker compose run --rm migrate    # Apply database migrations
docker compose up -d --build       # Build and start all services
```
