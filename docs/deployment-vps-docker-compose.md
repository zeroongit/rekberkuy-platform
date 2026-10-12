# Deploying to a VPS with Docker Compose

This covers a self-hosted VPS deployment of all three backend services plus
the frontend, behind a single nginx entrypoint. It does **not** cover
Kubernetes — see the note at the end for why that's deferred.

## What gets deployed

| Container | Image built from | Purpose |
|---|---|---|
| `nginx` | `nginx:1.27-alpine` (config only) | Single public entrypoint: `/` → dashboard-web, `/api/` → core-service |
| `dashboard-web` | `apps/dashboard-web/Dockerfile` | Next.js frontend (standalone build) |
| `core-service` | `apps/core-service/Dockerfile` | Go backend — owns all money movement & decisions |
| `backend-ai` | `backend-ai/Dockerfile` | Python fraud/KYC scoring — **not published to the host**; only `core-service` talks to it, over the internal Docker network |
| `postgres` | `postgres:16-alpine` | Database (or point `DATABASE_URL` at managed Supabase instead — see `docker-compose.yml`'s top comment) |
| `redis` | `redis:7-alpine` | Idempotency backend (optional — core-service falls back to Postgres if unreachable) |

The blockchain contract (`blockchain/`) is **not** part of this stack — it's
deployed once to Avalanche Fuji/Mainnet via Hardhat Ignition (see
`CLAUDE.md`'s Blockchain commands), not run as a long-lived container.

## First-time setup

```bash
# 1. Root compose env (Postgres credentials + frontend build-time vars)
cp .env.compose.example .env
# edit .env — set a real POSTGRES_PASSWORD, your real domain in NEXT_PUBLIC_API_URL

# 2. core-service runtime env
cp apps/core-service/.env.example apps/core-service/.env
# edit apps/core-service/.env — see "Gotcha" section below before filling DATABASE_URL/REDIS_URL

# 3. backend-ai runtime env
cp backend-ai/.env.example backend-ai/.env
# edit backend-ai/.env — set GROQ_API_KEY for real scoring (optional for a first boot)

# 4. Apply database migrations (one-off — NOT run automatically by `up`)
docker compose run --rm migrate

# 5. Build and start everything
docker compose up -d --build

# 6. (optional) seed category taxonomy + mockup data
docker compose run --rm core-service /app/seed
```

## ⚠️ Gotcha: container hostnames, not `localhost`

`apps/core-service/.env.example`'s defaults (`DATABASE_URL=...@localhost:5432/...`,
empty `REDIS_URL`) are written for **running the Go binary directly on your
machine**, where Postgres/Redis are reachable at `localhost`. Inside Docker
Compose, `core-service` runs in its own container — `localhost` there means
*the core-service container itself*, not the `postgres` container next to it.

When filling `apps/core-service/.env` for Compose, point these at the
Compose service names instead (Docker's internal DNS resolves them):

```
DATABASE_URL=postgresql://rekberkuy:<same password as .env POSTGRES_PASSWORD>@postgres:5432/rekberkuy
REDIS_URL=redis://redis:6379
```

`AI_SERVICE_URL` does **not** need manual editing — `docker-compose.yml`
already overrides it to `http://backend-ai:8081` for you.

## TLS / HTTPS

`nginx/nginx.conf` only listens on port 80 for now. Once you have a domain
pointed at the VPS, the standard approach is:

1. Add a `certbot` service (or run certbot on the host directly) to obtain a
   Let's Encrypt certificate for your domain.
2. Mount the certificate into the `nginx` container and add a `443 ssl`
   `server` block redirecting port 80 → 443.

This is deliberately not pre-built here since it needs your real domain name
first — ask when you're ready to wire it up.

## Why Kubernetes is deferred

K8s adds real operational cost (cluster management, ingress controllers,
secrets management, manifests to maintain) that isn't justified yet for a
solo-dev, pre-launch project on a single VPS. Docker Compose gives the same
"define the whole stack as code" benefit at a fraction of the complexity.
Revisit K8s if/when there's a concrete need Compose can't meet — e.g.
multi-node horizontal scaling under real production traffic, not before.