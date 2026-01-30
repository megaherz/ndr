# Quickstart Guide: Nitro Drag Royale MVP

**Feature**: 001-nitro-drag-mvp  
**Date**: 2026-01-28  
**Purpose**: Local development setup and workflow

---

## Prerequisites

### Required Software

- **Go 1.25+** — Backend language
- **Node.js 20+** — Frontend build tooling
- **Docker Desktop** — Local infrastructure (PostgreSQL, Redis, Centrifugo)
- **Make** — Build automation (comes with macOS/Linux, or via Chocolatey on Windows)
- **protoc** — Protobuf compiler (for RPC contract generation)

### Install Prerequisites (macOS)

```bash
# Install Homebrew (if not installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install Go
brew install go

# Install Node.js
brew install node@20

# Install Docker Desktop
brew install --cask docker

# Install protoc
brew install protobuf

# Install protoc-gen-go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Add Go bin to PATH (add to ~/.zshrc or ~/.bashrc)
export PATH="$PATH:$(go env GOPATH)/bin"
```

---

## Project Setup

### 1. Clone Repository

```bash
cd ~/workspace
git clone <repo_url>
cd ndr
git checkout 001-nitro-drag-mvp
```

### 2. Install Backend Dependencies

```bash
cd backend
go mod download
```

### 3. Install Frontend Dependencies

```bash
cd frontend
npm install
```

---

## Local Development Environment

### Start Full Environment

```bash
# From repo root
make dev
```

This command starts:
1. **PostgreSQL** (Docker) on `localhost:5432`
2. **Redis** (Docker) on `localhost:6379`
3. **Centrifugo** (Docker) on `localhost:8000` (WebSocket), `localhost:8001` (gRPC)
4. **Backend Go server** (live reload) on `localhost:8080`
5. **Frontend Vite dev server** (HMR) on `localhost:5173`

### Environment Variables

Local development uses `.env` files (gitignored):

**backend/.env**:
```bash
DATABASE_URL=postgres://ndr:ndr@localhost:5432/ndr?sslmode=disable
REDIS_URL=redis://localhost:6379/0
JWT_SECRET=local-dev-secret-change-in-production
CENTRIFUGO_API_KEY=local-centrifugo-key
CENTRIFUGO_SECRET=local-centrifugo-secret
CENTRIFUGO_GRPC_ADDR=localhost:8001
TONCENTER_API_KEY=your-toncenter-api-key-here
METRICS_ADDR=:9090
LOG_LEVEL=debug
```

**frontend/.env**:
```bash
VITE_API_URL=http://localhost:8080
VITE_CENTRIFUGO_URL=ws://localhost:8000/connection/websocket
VITE_AMPLITUDE_API_KEY=local-amplitude-key
```

### Docker Compose Services

**deployments/docker-compose.yml**:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:17-alpine
    environment:
      POSTGRES_USER: ndr
      POSTGRES_PASSWORD: ndr
      POSTGRES_DB: ndr
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data

  centrifugo:
    image: centrifugo/centrifugo:v4
    ports:
      - "8000:8000"  # WebSocket
      - "8001:8001"  # gRPC
    volumes:
      - ./centrifugo/config.json:/centrifugo/config.json
    command: centrifugo -c /centrifugo/config.json

volumes:
  postgres_data:
  redis_data:
```

**deployments/centrifugo/config.json**:
```json
{
  "engine": "redis",
  "redis_address": "redis:6379",
  "token_hmac_secret_key": "local-centrifugo-secret",
  "api_key": "local-centrifugo-key",
  "grpc_api": true,
  "grpc_api_address": "0.0.0.0:8001",
  "allowed_origins": ["http://localhost:5173"],
  "namespaces": [
    {
      "name": "user",
      "proxy_subscribe": {
        "enabled": true,
        "endpoint": "grpc://backend:8080"
      }
    },
    {
      "name": "match",
      "proxy_subscribe": {
        "enabled": true,
        "endpoint": "grpc://backend:8080"
      }
    }
  ]
}
```

---

## Database Migrations

### Run Migrations

```bash
# From backend directory
make migrate-up
```

Migrations are applied sequentially from `backend/internal/storage/postgres/migrations/`:
```
000001_create_users.up.sql
000002_create_wallets.up.sql
000003_create_system_wallets.up.sql
000004_create_ledger_entries.up.sql
000005_create_matches.up.sql
000006_create_match_participants.up.sql
000007_create_ghost_replays.up.sql
000008_create_match_settlements.up.sql
000009_create_payments.up.sql
```

### Create New Migration

```bash
make migrate-create NAME=add_new_field
```

This creates:
- `000010_add_new_field.up.sql`
- `000010_add_new_field.down.sql`

### Rollback Migration

```bash
make migrate-down
```

---

## Generate Code from Contracts

### Generate Protobuf (RPC Contracts)

```bash
# From repo root
make proto-gen
```

Generates Go code from `proto/**/*.proto` → `backend/internal/proto/`

### Generate OpenAPI Types (HTTP Contracts)

```bash
# From frontend directory
npm run generate-types
```

Generates TypeScript types from `specs/001-nitro-drag-mvp/contracts/http/*.json` → `frontend/src/types/api/`

---

## Development Workflow

### Backend Development

**Live Reload**: Backend uses `air` for live reload on file changes.

```bash
# From backend directory
air
```

**Run Tests**:
```bash
# Unit tests
make test

# Integration tests (requires Docker)
make test-integration

# Test coverage
make test-coverage
```

**Lint Code**:
```bash
make lint
```

### Frontend Development

**Vite Dev Server**: Runs with Hot Module Replacement (HMR).

```bash
# From frontend directory
npm run dev
```

**Run Tests**:
```bash
# Unit tests (Vitest)
npm run test

# Component tests
npm run test:ui

# Test coverage
npm run test:coverage
```

**Lint Code**:
```bash
npm run lint
```

**Type Check**:
```bash
npm run type-check
```

---

## Common Tasks

### Seed Test Data

```bash
# From backend directory
make seed-test-data
```

Creates:
- 10 test users
- Initial FUEL balances (1000 FUEL each)
- 20 Ghost replays (5 per league)

### Clear Database

```bash
make migrate-down
make migrate-up
make seed-test-data
```

### View Logs

```bash
# Backend logs (structured JSON)
tail -f backend/logs/ndr.log | jq

# Centrifugo logs
docker logs -f ndr-centrifugo

# PostgreSQL logs
docker logs -f ndr-postgres
```

### Access Database

```bash
# psql
docker exec -it ndr-postgres psql -U ndr -d ndr

# Run query
docker exec -it ndr-postgres psql -U ndr -d ndr -c "SELECT COUNT(*) FROM users;"
```

### Access Redis

```bash
# redis-cli
docker exec -it ndr-redis redis-cli

# View matchmaking queue
ZRANGE matchmaking:STREET 0 -1 WITHSCORES
```

---

## Testing Real-Time Features

### Test Centrifugo Connection

```bash
# From frontend directory
npm run dev

# Open browser console, should see:
# "Connected to Centrifugo"
# "Subscribed to user:<user_id>"
```

### Test RPC Commands

```javascript
// In browser console
const response = await centrifuge.rpc('matchmaking.join', {
  league: 'STREET',
  client_req_id: crypto.randomUUID()
});
console.log(response);
```

### Publish Test Event

```bash
# From backend directory
make publish-test-event USER_ID=<user_id>
```

Publishes test `balance_updated` event to `user:<user_id>` channel.

---

## Troubleshooting

### Backend won't start

**Error**: `dial tcp 127.0.0.1:5432: connect: connection refused`

**Fix**: Start Docker services first:
```bash
docker-compose -f deployments/docker-compose.yml up -d
```

### Migrations fail

**Error**: `dirty database version X`

**Fix**: Force version and retry:
```bash
make migrate-force VERSION=X
make migrate-up
```

### Frontend can't connect to Centrifugo

**Error**: `WebSocket connection failed`

**Fix**: Check Centrifugo is running and CORS is configured:
```bash
docker logs ndr-centrifugo
# Should see: "WebSocket server started on :8000"
```

### Protobuf generation fails

**Error**: `protoc-gen-go: program not found or is not executable`

**Fix**: Install protoc plugins:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

---

## Makefile Targets

### Backend (`backend/Makefile`)

| Target | Description |
|--------|-------------|
| `make dev` | Start backend with live reload (air) |
| `make build` | Build production binary |
| `make test` | Run unit tests |
| `make test-integration` | Run integration tests (dockertest) |
| `make test-coverage` | Generate test coverage report |
| `make lint` | Run golangci-lint |
| `make migrate-up` | Apply all migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-create NAME=<name>` | Create new migration |
| `make proto-gen` | Generate Go code from Protobuf |
| `make seed-test-data` | Seed database with test data |

### Frontend (`frontend/Makefile`)

| Target | Description |
|--------|-------------|
| `make dev` | Start Vite dev server (alias for `npm run dev`) |
| `make build` | Build production bundle |
| `make test` | Run unit tests (Vitest) |
| `make lint` | Run ESLint |
| `make type-check` | Run TypeScript type checking |
| `make generate-types` | Generate TypeScript types from OpenAPI |

### Root (`Makefile`)

| Target | Description |
|--------|-------------|
| `make dev` | Start full local environment (Docker + backend + frontend) |
| `make stop` | Stop all Docker services |
| `make clean` | Clean build artifacts and Docker volumes |
| `make proto-gen` | Generate Protobuf code (backend) |
| `make test` | Run all tests (backend + frontend) |

---

## Next Steps

1. **Explore codebase structure**: See [plan.md](./plan.md) Project Structure section
2. **Review data model**: See [data-model.md](./data-model.md)
3. **Review API contracts**: See `contracts/http/` and `contracts/rpc/`
4. **Start implementing tasks**: See [tasks.md](./tasks.md) (generated by `/speckit.tasks`)

---

## Resources

- **Uber Go Style Guide**: https://github.com/uber-go/guide/blob/master/style.md
- **Chi Router Docs**: https://go-chi.io/
- **Centrifugo v4 Docs**: https://centrifugal.dev/docs/getting-started/introduction
- **Pixi.js Docs**: https://pixijs.io/guides/
- **TON Connect Docs**: https://docs.ton.org/develop/dapps/ton-connect/overview
- **Vite Docs**: https://vitejs.dev/guide/
- **React Query Docs**: https://tanstack.com/query/latest

---

**Happy Coding! 🚀**
