# Research Document: Nitro Drag Royale MVP

**Feature**: 001-nitro-drag-mvp  
**Date**: 2026-01-28  
**Status**: Complete

## Overview

This document captures the research and technical decisions made for the Nitro Drag Royale MVP implementation. All critical architectural decisions are documented here with rationale and alternatives considered.

---

## 1. Backend Language & Framework

### Decision: Go 1.25+ with Chi Router

**Rationale**:
- **Performance**: Go provides excellent concurrency primitives (goroutines) required for managing 1,000 concurrent players and 100 concurrent matches
- **Type safety**: Static typing catches errors at compile time, critical for economy safety
- **Simplicity**: Go's standard library and ecosystem provide all needed primitives without framework bloat
- **Chi router**: Superior route organization, built-in middleware ecosystem (RequestID, Logger, Recoverer), better performance than standard `net/http` ServeMux

**Alternatives Considered**:
- **Node.js + Express**: Rejected due to lack of type safety (even with TypeScript), weaker concurrency model, floating-point number handling issues for economy
- **Rust + Actix**: Rejected due to steep learning curve, longer development time for MVP, overkill for modular monolith
- **Python + FastAPI**: Rejected due to performance concerns (GIL), lack of true concurrency, deployment complexity

**References**:
- Uber Go Style Guide: https://github.com/uber-go/guide/blob/master/style.md
- Chi Router: https://github.com/go-chi/chi

---

## 2. Real-Time Communication Architecture

### Decision: Centrifugo v4 + Redis + gRPC Proxy

**Rationale**:
- **Separation of concerns**: Centrifugo handles connection management (WebSocket, SSE), backend handles game logic
- **Horizontal scalability**: Redis broker enables multiple Centrifugo nodes (post-MVP)
- **Hot-path performance**: RPC proxy pattern (Centrifugo RPC → gRPC → Go) provides <200ms latency for gameplay commands
- **Production-ready**: Built-in presence tracking, channel subscriptions, message broadcasting, connection recovery
- **Security**: Subscription proxy enables fine-grained channel authorization without embedding logic in JWT claims

**Alternatives Considered**:
- **WebSocket in Go backend directly**: Rejected because:
  - Requires custom connection management, presence tracking, broadcasting logic
  - Harder to horizontally scale (sticky sessions or Redis pub/sub required)
  - Mixes concerns (connection management + game logic in same service)
  
- **Socket.IO**: Rejected because:
  - Node.js-centric, requires separate Node service for WebSocket layer
  - More complex than needed (fallback transports not needed for Telegram Mini App)
  
- **Ably / Pusher (SaaS)**: Rejected because:
  - Cost at scale (1,000 concurrent connections = $$$)
  - Vendor lock-in, less control over infrastructure
  - Latency concerns (external service)

**RPC Proxy Contract**: https://github.com/centrifugal/centrifugo/blob/v4/internal/proxyproto/proxy.proto

---

## 3. Database & Storage Strategy

### Decision: PostgreSQL 17+ (persistent) + Redis 7+ (volatile)

**Rationale**:
- **PostgreSQL** for system of record:
  - ACID guarantees essential for ledger/economy operations
  - JSON support for flexible data (match history, ghost replay data)
  - `DECIMAL(16,2)` type for fixed-point money math (no floating point errors)
  - Mature migration tooling (`golang-migrate`)
  - Excellent performance for transactional data
  
- **Redis** for volatile state:
  - Matchmaking queues (sorted sets for priority)
  - Idempotency keys (TTL-based expiry)
  - Session data (short-lived locks)
  - Centrifugo broker/engine (pub/sub)

**Alternatives Considered**:
- **MongoDB**: Rejected due to lack of ACID guarantees across documents, inappropriate for financial ledger
- **MySQL**: Rejected in favor of PostgreSQL's superior JSON support and standards compliance
- **In-memory only (no Redis)**: Rejected because matchmaking queues and idempotency keys require TTL and atomic operations
- **SQLite**: Rejected due to concurrency limitations (single writer)

**Fixed-Point Decimal Math**:
- Backend uses `github.com/shopspring/decimal` for all monetary calculations
- All rounding uses **rounding down (floor)** semantics (`Truncate(2)`)
- Floating point math is **forbidden** in economy code paths

**References**:
- PostgreSQL JSON: https://www.postgresql.org/docs/current/datatype-json.html
- shopspring/decimal: https://github.com/shopspring/decimal

---

## 4. Frontend Framework & Tech Stack

### Decision: React 18+ with TypeScript 5+, Vite 5+, Pixi.js 7+

**Rationale**:
- **React 18**: Mature component model, excellent ecosystem, concurrent rendering features
- **TypeScript 5** (mandatory): Type safety for contracts, catch errors at compile time, better IDE support
- **Vite 5**: Fast dev server, optimized production builds, superior DX compared to Webpack
- **Pixi.js 7**: High-performance 2D rendering via WebGL, required for Race HUD (30-45 FPS)
- **Zustand**: Lightweight global state (no Redux boilerplate)
- **React Query**: Server state caching, automatic refetching, optimistic updates

**Alternatives Considered**:
- **Vue 3 / Svelte**: Rejected to align with Constitution Principle III (React + TypeScript mandate)
- **Create React App**: Rejected in favor of Vite (faster build times, better DX)
- **Next.js**: Rejected because SSR/SSG not needed for Telegram Mini App (client-side only)
- **Three.js**: Rejected for Pixi.js (2D game, don't need 3D overhead)
- **Canvas API directly**: Rejected due to performance (no WebGL acceleration) and complexity

**Telegram Mini App Constraints**:
- FPS capped at 30-45 (battery/performance)
- Device pixel ratio clamped to 2x max (rendering performance)
- Lifecycle handling: pause Pixi ticker on `document.visibilitychange`
- No heavy runtime filters (blur, glow, displacement) — use prebaked assets instead
- BitmapFont for frequently updated text (speed, position, countdown)

**References**:
- Telegram Mini Apps SDK: https://core.telegram.org/bots/webapps
- Pixi.js Performance: https://pixijs.io/guides/basics/render-loop.html

---

## 5. Authentication & Security Strategy

### Decision: Telegram initData Validation + Dual JWT (App + Centrifugo)

**Rationale**:
- **Telegram initData**: Built-in Telegram Mini App auth mechanism, validates user identity via HMAC signature
- **Dual JWT**:
  - **App JWT** (HTTP): Used for REST/cold-path endpoints
  - **Centrifugo JWT** (WSS/RPC): Used for Centrifugo connection, subscriptions, RPC
- **Subscription proxy**: Backend authorizes `match:{match_id}` subscriptions via gRPC proxy (checks match participation)
- **No match-scoped JWT in MVP**: Subscription authorization is simpler than generating per-match tokens

**Security Rules**:
- Backend **never trusts** `user_id` from payloads — identity derived exclusively from JWT claims (`sub`)
- Channel authorization: `user:{user_id}` always allowed for authenticated user; `match:{match_id}` requires participation check
- JWT claims: `sub` (user_id), `exp` (expiration), `iat` (issued at), `jti` (token ID)

**Alternatives Considered**:
- **Match-scoped JWT**: Rejected for MVP (adds complexity, subscription proxy is simpler)
- **Session cookies**: Rejected because Centrifugo requires JWT for WebSocket auth
- **API keys**: Rejected due to security concerns (no expiration, harder to rotate)

**References**:
- Telegram initData validation: https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
- Centrifugo JWT: https://centrifugal.dev/docs/server/authentication

---

## 6. TON Blockchain Integration

### Decision: TonCenter API (REST) + Async Reconciliation

**Rationale**:
- **TonCenter API**: Official TON HTTP API, simplifies integration (no running TON node)
- **Async reconciliation**: Deposits/withdrawals are reconciled via polling worker (eventual consistency)
- **Idempotency**: Deposit credits keyed by `tx_hash` (PostgreSQL unique constraint), withdrawals keyed by `client_req_id`
- **Ledger-based**: All TON movements recorded in ledger with explicit entries

**Flow**:
1. **Deposit**: Client initiates via TON Connect → Backend records intent → TonCenter polled by worker → On confirmation, ledger credit applied
2. **Withdrawal**: Withdrawal intent created → Funds locked in ledger → Transaction broadcasted via TonCenter → Status reconciled via polling

**Alternatives Considered**:
- **Run TON node**: Rejected due to infrastructure complexity, overkill for MVP
- **TON SDK (Go)**: Rejected in favor of simpler REST API (fewer dependencies)
- **Synchronous confirmation**: Rejected due to blockchain latency (would block user for 5-30s)

**References**:
- TonCenter API: https://toncenter.com/api/v2/
- TON Connect: https://docs.ton.org/develop/dapps/ton-connect/overview

---

## 7. Deployment & Infrastructure

### Decision: DigitalOcean App Platform + Managed Services

**Rationale**:
- **App Platform**: Simplifies deployment (Git push → auto-deploy), no Kubernetes complexity for MVP
- **Managed PostgreSQL**: Automated backups, monitoring, no ops overhead
- **Managed Redis**: High availability, automatic failover
- **Single replica for MVP**: Horizontal scaling deferred until match ownership/routing implemented

**12-Factor Config**:
- All configuration via environment variables (no config files in repo)
- `DATABASE_URL`, `REDIS_URL`, `JWT_SECRET`, `CENTRIFUGO_SECRET`, `TONCENTER_API_KEY`, `METRICS_ADDR`

**Graceful Shutdown**:
- On `SIGTERM`: Service enters DRAINING state
- New matches rejected, active matches continue up to 120 seconds
- Abort before end of Heat 1 → refund buy-ins
- Abort after Heat 1 → current heat = crash (0), settlement applied

**Alternatives Considered**:
- **Kubernetes**: Rejected for MVP (unnecessary complexity, slower iteration)
- **AWS ECS/EKS**: Rejected in favor of DigitalOcean (simpler, cost-effective for MVP)
- **Vercel/Netlify**: Rejected because backend is long-running Go service (not serverless-friendly)

**References**:
- 12-Factor App: https://12factor.net/config
- DigitalOcean App Platform: https://docs.digitalocean.com/products/app-platform/

---

## 8. Testing Strategy

### Decision: Table-Driven Tests (Go) + Dockertest (Integration) + Vitest (Frontend)

**Rationale**:
- **Go unit tests**: Table-driven test pattern (Uber Go Style Guide), `testify` for assertions
- **Go integration tests**: `dockertest` for ephemeral PostgreSQL/Redis instances
- **Frontend unit tests**: Vitest (fast, Vite-native) + React Testing Library
- **Contract tests**: Verify RPC and HTTP payloads match Protobuf/JSON schemas

**Test Principles**:
- Database tests use transactions that rollback after completion
- WebSocket tests mock Centrifugo interactions
- Financial/economy tests verify ledger invariants (sum of deltas = 0)

**Alternatives Considered**:
- **Jest**: Rejected in favor of Vitest (faster, better Vite integration)
- **Testcontainers**: Rejected in favor of `dockertest` (lighter, Go-native)

**References**:
- Dockertest: https://github.com/ory/dockertest
- Vitest: https://vitest.dev/

---

## 9. Observability & Metrics

### Decision: Prometheus + Structured Logging (Logrus) + Sentry

**Rationale**:
- **Prometheus**: Industry-standard metrics, `/metrics` endpoint on dedicated port
- **Structured logging**: `github.com/sirupsen/logrus` with mandatory fields (`match_id`, `user_id`, `command`, `client_req_id`)
- **Sentry**: Panic recovery, error tracking
- **Key metrics**:
  - RPC latency & rejects
  - Matchmaking wait time
  - Match duration & aborts
  - Settlement duration & errors
  - Rake & payouts
  - House balance gauge (`house_fuel_balance`)
  - TonCenter latency & failures

**High-cardinality labels forbidden**: No `user_id` labels (cardinality explosion)

**Alternatives Considered**:
- **Zap**: Rejected in favor of Logrus (simpler API, adequate performance)
- **OpenTelemetry**: Rejected for MVP (overkill, adds complexity)
- **Datadog / New Relic**: Rejected in favor of Prometheus (open-source, self-hosted)

**References**:
- Prometheus best practices: https://prometheus.io/docs/practices/naming/
- Logrus: https://github.com/sirupsen/logrus

---

## 10. Localization (i18n)

### Decision: Frontend-only i18n with 4 MVP languages

**Languages**: `en` (English, default), `ru` (Russian), `es-PE` (Spanish/Peru), `pt-BR` (Portuguese/Brazil)

**Rationale**:
- Localization handled **entirely on frontend** (backend returns language-agnostic data only)
- Language determined by: Telegram user language (if supported) → user-selected language (future) → fallback to `en`
- Implementation: `react-i18next` or equivalent
- Translation keys must be **stable and semantic** (e.g., `garage.balance`, `race.match_found`)

**Rules**:
- All user-facing strings **must be localized** (no hardcoded text)
- Dynamic values injected via placeholders (not string concatenation)
- Text in Pixi HUD uses localized strings from React layer

**Alternatives Considered**:
- **Backend i18n**: Rejected to keep backend stateless and simplify API contracts
- **FormatJS**: Rejected in favor of react-i18next (more popular, better ecosystem)

**References**:
- react-i18next: https://react.i18next.com/

---

## 11. Product Analytics

### Decision: Amplitude (Client-Side SDK)

**Rationale**:
- **Client-side only**: Analytics implemented entirely on frontend (non-authoritative telemetry)
- **Amplitude**: Industry-standard, supports funnel analysis, A/B testing (post-MVP), cohort analysis
- **Best-effort**: Events may be dropped, backend **never** depends on analytics events

**Core Events (MVP)**:
- Session: `app_opened`, `auth_success`, `garage_viewed`, `first_match_started`
- Match: `matchmaking_joined`, `match_found`, `heat_started`, `heat_crashed`, `match_finished`
- Economy: `balance_viewed`, `deposit_initiated`, `withdraw_initiated`

**User Properties**: `user_id` (hashed), `platform` (telegram), `client_version`, `locale`, `country`

**Privacy**:
- No private keys, wallet secrets, or raw TON addresses in analytics
- Monetary values logged as **rounded display values** only (not ledger deltas)

**Alternatives Considered**:
- **Mixpanel**: Rejected due to cost at scale
- **PostHog**: Rejected due to self-hosting complexity (not needed for MVP)
- **Google Analytics**: Rejected due to poor event schema flexibility

**References**:
- Amplitude SDK: https://www.docs.developers.amplitude.com/

---

## Phase 0 Summary

All critical technical decisions are documented above. No unresolved `NEEDS CLARIFICATION` items remain. The architecture is **constitution-compliant** and ready for Phase 1 (data model, contracts, quickstart).

**Next Steps**: Generate `data-model.md`, `contracts/`, and `quickstart.md`.
