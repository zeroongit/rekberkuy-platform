# ADR-0005: 7-Layer Defense-in-Depth Security Standard

## Status
ACCEPTED

## Context
As the RekberKuy platform handles sensitive escrow funds, multi-tier vendor allocations, digital wallets, and blockchain audit logs, relying solely on perimeter security is insufficient. A compromised service or component must not lead to a total system failure. Therefore, we mandate a strict **7-Layer Defense-in-Depth Security Model** spanning infrastructure, network, transport, session, presentation, and application layers.

## Decision
All microservices (`apps/core-service`, `backend-ai`, `apps/dashboard-web`, and proxy tiers) must adhere to the following 7-Layer security controls:

### Layer 7: Application (Business Logic & Validation)
- **Zod Validation:** All incoming data payloads on both frontend (Next.js Server Actions) and backend must be validated using strict, shared Zod schemas before processing.
- **Business Logic Guardrails:** State machines (defined in `docs/application-flow-and-module-guide.md`) must be enforced strictly in the backend usecase layer to prevent illegal status transitions or fund-release bypasses.

### Layer 6: Presentation (Data Sanitization & Encoding)
- **XSS Sanitization:** All user-generated text fields must be sanitized using robust libraries (e.g., DOMPurify) before rendering in UI components.
- **Safe Output Encoding:** API responses must return structured JSON with strict `Content-Type` headers.

### Layer 5: Session & State Management
- **Secure Cookies:** All authentication and session cookies must enforce strict flags: `HttpOnly: true`, `Secure: true`, and `SameSite=Strict`.
- **Redis Session Store:** Active sessions, token blacklists, and idempotency keys are managed securely via distributed Redis instances with short-lived access token rotation.

### Layer 4: Transport (Encryption in Transit)
- **HTTPS/TLS 1.3:** All external and internal service-to-service communications must be encrypted using modern TLS standards configured via Nginx reverse proxy.
- **Port Minimization:** Only necessary public ports (80/443) are exposed to the public internet.

### Layer 3: Network (Segmentation & Routing)
- **Network Boundaries:** Isolation of public-facing gateways (Nginx) from private backend services (`core-service`, `backend-ai`, PostgreSQL, Redis) which reside within an internal network/localhost zone.
- **Rate Limiting:** IP-based and route-based rate limiting implemented at the proxy layer to prevent DDoS and brute-force attacks.

### Layer 2: Data Link (Virtual Private Cloud / Isolation)
- **Container Isolation:** Docker containers communicate over internal Docker bridge networks (`docker-compose`), ensuring isolated inter-container traffic.

### Layer 1: Physical (Infrastructure Security)
- **Cloud Provider Security:** Physical infrastructure relies on certified data center security compliance (ISO 27001) provided by the chosen VPS/Cloud hosting partner.

## Consequences
- **Enhanced Resilience:** A vulnerability in a single layer (e.g., client-side XSS) is neutralized by downstream layers (e.g., `HttpOnly` cookies and strict backend validation).
- **Development Overhead:** Developers must ensure Zod schemas, secure cookie flags, and proper sanitization are included in all new feature pull requests.