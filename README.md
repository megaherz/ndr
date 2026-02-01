# Nitro Drag Royale MVP

A PvP crash tournament racing game implemented as a Telegram Mini App.

## 🏁 Overview

Nitro Drag Royale is a fast-paced multiplayer racing game where players compete in 10-player matches across 3 heats. The game features:

- **Real-time multiplayer racing** with speed acceleration mechanics
- **Fast matchmaking** (<20s) using Ghost players (historical replays)
- **Multi-league progression** (Rookie → Street → Pro → Top Fuel)
- **TON blockchain integration** for deposits/withdrawals
- **Provably fair gameplay** with cryptographic crash seeds

## 🏗 Architecture

- **Backend**: Go 1.25+ with Chi router (Modular Monolith)
- **Frontend**: React 18+ with TypeScript 5+ (Telegram Mini App)
- **Real-time**: Centrifugo v4 + Redis + gRPC proxy
- **Database**: PostgreSQL 17+ (persistent) + Redis 7+ (volatile)
- **Blockchain**: TON via TonCenter API

## 🚀 Quick Start

### Prerequisites

- Go 1.25+
- Node.js 20+
- Docker Desktop
- Make

### Setup

1. **Clone and navigate to the project**:
   ```bash
   git clone <repo_url>
   cd ndr
   ```

2. **Start infrastructure services**:
   ```bash
   make dev
   ```

3. **In separate terminals, start backend and frontend**:
   ```bash
   # Terminal 1: Backend
   make dev-backend
   
   # Terminal 2: Frontend  
   make dev-frontend
   ```

4. **Access the application**:
   - Frontend: http://localhost:5173
   - Backend API: http://localhost:8080
   - Metrics: http://localhost:9090/metrics

### Mobile Debugging

For debugging the Telegram Mini App on mobile devices, we use [Eruda](https://github.com/liriliri/eruda):

- **Auto-enabled** in development mode
- **Manual activation**: Add `?debug=true` to URL or use `NDR_DEBUG.show()`
- **Features**: Console, DOM inspector, network monitor, device info
- **See**: [Frontend README](frontend/README.md) for detailed debugging guide

### Environment Configuration

Copy the example environment files and configure:

```bash
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env
```

Edit the `.env` files with your configuration values.

## 📁 Project Structure

```
ndr/
├── backend/                 # Go backend (Modular Monolith)
│   ├── cmd/server/         # Main entry point
│   ├── internal/
│   │   ├── modules/        # Domain modules
│   │   ├── storage/        # Database layer
│   │   ├── config/         # Configuration
│   │   └── metrics/        # Prometheus metrics
│   └── proto/              # Protobuf definitions
├── frontend/               # React frontend (Telegram Mini App)
│   ├── src/
│   │   ├── components/     # React components
│   │   ├── services/       # Business logic
│   │   ├── stores/         # State management
│   │   └── pixi/          # Game rendering
├── deployments/            # Infrastructure
│   ├── docker-compose.yml # Local development
│   └── centrifugo/        # Centrifugo config
└── specs/                  # Documentation
    └── 001-nitro-drag-mvp/ # Feature specification
```

## 🛠 Development

### Backend Commands

```bash
cd backend

# Development with live reload
make dev

# Run tests
make test

# Database migrations
make migrate-up
make migrate-down

# Generate protobuf code
make proto-gen

# Lint code
make lint
```

### Frontend Commands

```bash
cd frontend

# Development server
npm run dev

# Run tests
npm test

# Type checking
npm run type-check

# Generate API types
npm run generate-types
```

## 🎮 Game Flow

1. **Authentication**: Telegram initData validation + JWT issuance
2. **Garage**: View balances, select league, join matchmaking
3. **Matchmaking**: Queue with other players, filled with Ghosts if needed
4. **Race**: 3 heats with speed acceleration, "EARN POINTS" to lock score
5. **Settlement**: Prize distribution, BURN rewards, provable fairness proof

## 🏆 Leagues & Economy

| League | Buy-in | Prize Pool | BURN Rewards |
|--------|--------|------------|--------------|
| Rookie | 10 FUEL | 92 FUEL (8% rake) | None |
| Street | 50 FUEL | 460 FUEL | Yes |
| Pro | 300 FUEL | 2,760 FUEL | Yes |
| Top Fuel | 3,000 FUEL | 27,600 FUEL | Yes |

## 📊 Monitoring

- **Metrics**: Prometheus endpoint at `:9090/metrics`
- **Health**: GET `/health`
- **Logs**: Structured JSON (production) or colored text (development)

## 🧪 Testing

### Backend Tests
```bash
make test              # Unit tests
make test-coverage     # Coverage report
```

### Frontend Tests
```bash
npm test              # Unit tests (Vitest)
npm run test:ui       # Component tests
npm run test:coverage # Coverage report
```

## 📚 Documentation

- [Implementation Plan](specs/001-nitro-drag-mvp/plan.md)
- [Data Model](specs/001-nitro-drag-mvp/data-model.md)
- [API Contracts](specs/001-nitro-drag-mvp/contracts/)
- [Development Guide](specs/001-nitro-drag-mvp/quickstart.md)
- [Task Breakdown](specs/001-nitro-drag-mvp/tasks.md)

## 🤝 Contributing

1. Follow the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
2. Use TypeScript only (no JavaScript files)
3. One commit per task (atomic commits)
4. Test your changes before committing