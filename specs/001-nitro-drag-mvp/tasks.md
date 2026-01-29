# Tasks: Nitro Drag Royale MVP

**Input**: Design documents from `/specs/001-nitro-drag-mvp/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/, research.md, quickstart.md

**Tests**: Tests are OPTIONAL - only included if explicitly requested in the feature specification. This task list focuses on implementation tasks.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

**Backend**: `backend/` at repository root
- `backend/cmd/server/` — Main entry point
- `backend/internal/modules/` — Domain modules (auth, account, matchmaker, gameengine, ghosts, payments, gateway)
- `backend/internal/storage/` — PostgreSQL + Redis DAL
- `backend/proto/` — Protobuf definitions

**Frontend**: `frontend/` at repository root
- `frontend/src/components/` — React components
- `frontend/src/pages/` — Top-level page components
- `frontend/src/services/` — Business logic layer
- `frontend/src/stores/` — Zustand stores

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create project directory structure per plan.md (backend/, frontend/, deployments/)
- [ ] T002 [P] Initialize Go module in backend/ with go.mod (Go 1.25+)
- [ ] T003 [P] Initialize React project in frontend/ with Vite 5+ and TypeScript 5+
- [ ] T004 [P] Create docker-compose.yml in deployments/ for PostgreSQL 17+, Redis 7+, Centrifugo v4
- [ ] T005 [P] Create Makefile in backend/ with targets: dev, test, migrate-up, migrate-down, proto-gen
- [ ] T006 [P] Setup golang-migrate for database migrations in backend/internal/storage/postgres/migrations/
- [ ] T007 [P] Create backend/internal/config/ package for 12-factor env var configuration
- [ ] T008 [P] Create backend/internal/logger/ package for structured logging (logrus)
- [ ] T009 [P] Setup Prometheus metrics endpoint in backend/internal/metrics/
- [ ] T010 [P] Create .env.example files for backend and frontend with required environment variables

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T011 Create database migrations for Users table in backend/internal/storage/postgres/migrations/000001_create_users.up.sql
- [ ] T012 Create database migrations for Wallets table in backend/internal/storage/postgres/migrations/000002_create_wallets.up.sql
- [ ] T013 Create database migrations for System Wallets table in backend/internal/storage/postgres/migrations/000003_create_system_wallets.up.sql
- [ ] T014 Create database migrations for Ledger Entries table in backend/internal/storage/postgres/migrations/000004_create_ledger_entries.up.sql
- [ ] T015 Create database migrations for Matches table in backend/internal/storage/postgres/migrations/000005_create_matches.up.sql
- [ ] T016 Create database migrations for Match Participants table in backend/internal/storage/postgres/migrations/000006_create_match_participants.up.sql
- [ ] T017 Create database migrations for Ghost Replays table in backend/internal/storage/postgres/migrations/000007_create_ghost_replays.up.sql
- [ ] T018 Create database migrations for Match Settlements table in backend/internal/storage/postgres/migrations/000008_create_match_settlements.up.sql
- [ ] T019 Create database migrations for Payments table in backend/internal/storage/postgres/migrations/000009_create_payments.up.sql
- [ ] T020 Seed System Wallets with HOUSE_FUEL and RAKE_FUEL in backend/internal/storage/postgres/migrations/000010_seed_system_wallets.up.sql
- [ ] T021 [P] Create PostgreSQL models in backend/internal/storage/postgres/models/ for all 9 entities
- [ ] T022 [P] Create Redis client wrapper in backend/internal/storage/redis/client.go
- [ ] T023 [P] Generate Protobuf code from backend/proto/ using protoc (matchmaking.proto, match.proto)
- [ ] T024 [P] Create Chi router setup in backend/cmd/server/main.go with middleware (RequestID, Logger, Recoverer)
- [ ] T025 [P] Create JWT utility package in backend/internal/auth/jwt.go for token generation/validation
- [ ] T026 [P] Create Centrifugo gRPC client wrapper in backend/internal/centrifugo/client.go
- [ ] T027 [P] Create decimal utility package in backend/internal/decimal/ for fixed-point arithmetic (shopspring/decimal)
- [ ] T028 [P] Setup frontend routing with React Router in frontend/src/App.tsx
- [ ] T029 [P] Create Zustand stores structure in frontend/src/stores/ (auth, wallet, match)
- [ ] T030 [P] Create React Query setup in frontend/src/services/api/client.ts
- [ ] T031 [P] Create Centrifuge.js client wrapper in frontend/src/services/centrifugo/client.ts
- [ ] T032 [P] Create TON Connect UI setup in frontend/src/services/wallet/tonconnect.ts

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Enter and Complete First Race (Priority: P1) 🎯 MVP

**Goal**: New player can join Rookie league, play 3 heats with speed acceleration mechanic, and receive winnings/BURN based on final position

**Independent Test**: Onboard new user, provide initial FUEL, observe complete race cycle (Garage → Matchmaking → Race → Settlement)

### Implementation for User Story 1

**Backend: Auth Module**

- [ ] T033 [P] [US1] Create auth module structure in backend/internal/modules/auth/
- [ ] T034 [P] [US1] Implement Telegram initData validation in backend/internal/modules/auth/telegram.go
- [ ] T035 [US1] Implement AuthService with Authenticate() method in backend/internal/modules/auth/service.go
- [ ] T036 [US1] Create user repository interface in backend/internal/storage/postgres/repository/users.go
- [ ] T037 [US1] Implement HTTP handler POST /api/v1/auth/telegram in backend/internal/modules/gateway/http/auth_handler.go

**Backend: Account Module**

- [ ] T038 [P] [US1] Create account module structure in backend/internal/modules/account/
- [ ] T039 [P] [US1] Create wallet repository interface in backend/internal/storage/postgres/repository/wallets.go
- [ ] T040 [P] [US1] Create ledger repository interface in backend/internal/storage/postgres/repository/ledger.go
- [ ] T041 [US1] Implement AccountService with GetWallet() method in backend/internal/modules/account/service.go
- [ ] T042 [US1] Implement ledger operations (DebitFuel, CreditFuel, RecordEntry) in backend/internal/modules/account/ledger.go
- [ ] T043 [US1] Implement HTTP handler GET /api/v1/garage in backend/internal/modules/gateway/http/garage_handler.go

**Backend: Matchmaker Module**

- [ ] T044 [P] [US1] Create matchmaker module structure in backend/internal/modules/matchmaker/
- [ ] T045 [P] [US1] Implement Redis queue operations in backend/internal/modules/matchmaker/queue.go
- [ ] T046 [US1] Implement MatchmakerService with JoinQueue() method in backend/internal/modules/matchmaker/service.go
- [ ] T047 [US1] Implement CancelQueue() method in backend/internal/modules/matchmaker/service.go
- [ ] T048 [US1] Implement lobby formation logic (10 players max, 20s timeout) in backend/internal/modules/matchmaker/lobby.go
- [ ] T049 [US1] Implement RPC handler matchmaking.join in backend/internal/modules/gateway/rpc/matchmaking_handler.go
- [ ] T050 [US1] Implement RPC handler matchmaking.cancel in backend/internal/modules/gateway/rpc/matchmaking_handler.go

**Backend: Game Engine Module**

- [ ] T051 [P] [US1] Create gameengine module structure in backend/internal/modules/gameengine/
- [ ] T052 [P] [US1] Create match repository interface in backend/internal/storage/postgres/repository/matches.go
- [ ] T053 [P] [US1] Create match participants repository interface in backend/internal/storage/postgres/repository/match_participants.go
- [ ] T054 [P] [US1] Implement speed formula calculation Speed = 500 * ((e^(0.08·t) - 1) / (e^(0.08·25) - 1)) in backend/internal/modules/gameengine/physics.go
- [ ] T055 [P] [US1] Implement crash seed generation (cryptographic hash) in backend/internal/modules/gameengine/provable_fairness.go
- [ ] T056 [US1] Implement GameEngineService with CreateMatch() method in backend/internal/modules/gameengine/service.go
- [ ] T057 [US1] Implement in-memory match state machine (FORMING → IN_PROGRESS → COMPLETED) in backend/internal/modules/gameengine/state.go
- [ ] T058 [US1] Implement heat lifecycle (countdown → active → intermission) in backend/internal/modules/gameengine/heat.go
- [ ] T059 [US1] Implement EarnPoints (lock score) logic in backend/internal/modules/gameengine/earn_points.go
- [ ] T060 [US1] Implement RPC handler match.earn_points in backend/internal/modules/gateway/rpc/match_handler.go
- [ ] T061 [US1] Implement early heat end optimization (all players finished) in backend/internal/modules/gameengine/heat.go
- [ ] T062 [US1] Implement settlement calculation (positions, prizes, BURN rewards) in backend/internal/modules/gameengine/settlement.go
- [ ] T063 [US1] Integrate settlement with ledger (apply prize/BURN entries) in backend/internal/modules/gameengine/settlement.go

**Backend: Gateway Module (Events)**

- [ ] T064 [P] [US1] Create gateway module structure in backend/internal/modules/gateway/
- [ ] T065 [P] [US1] Implement Centrifugo publisher methods (PublishToUser, PublishToMatch) in backend/internal/modules/gateway/centrifugo.go
- [ ] T066 [P] [US1] Implement event schemas (match_found, heat_started, heat_ended, match_settled) in backend/internal/modules/gateway/events/
- [ ] T067 [US1] Publish match_found event to user:{user_id} channel in backend/internal/modules/matchmaker/service.go
- [ ] T068 [US1] Publish heat_started event to match:{match_id} channel in backend/internal/modules/gameengine/heat.go
- [ ] T069 [US1] Publish heat_ended event to match:{match_id} channel in backend/internal/modules/gameengine/heat.go
- [ ] T070 [US1] Publish match_settled event to match:{match_id} channel in backend/internal/modules/gameengine/settlement.go
- [ ] T071 [US1] Publish balance_updated event to user:{user_id} channel after settlement in backend/internal/modules/account/service.go

**Frontend: Auth & Garage**

- [ ] T072 [P] [US1] Create auth store in frontend/src/stores/authStore.ts (JWT tokens, user state)
- [ ] T073 [P] [US1] Implement Telegram initData extraction in frontend/src/services/auth/telegram.ts
- [ ] T074 [US1] Create login API call in frontend/src/services/api/auth.ts (POST /api/v1/auth/telegram)
- [ ] T075 [US1] Create Auth page component in frontend/src/pages/Auth.tsx (Telegram login button)
- [ ] T076 [P] [US1] Create wallet store in frontend/src/stores/walletStore.ts (balances, league access)
- [ ] T077 [US1] Create garage API call in frontend/src/services/api/garage.ts (GET /api/v1/garage)
- [ ] T078 [US1] Create Garage page component in frontend/src/pages/Garage.tsx (league selector, balances, "RACE NOW")
- [ ] T079 [US1] Create league card components in frontend/src/components/garage/LeagueCard.tsx

**Frontend: Matchmaking**

- [ ] T080 [P] [US1] Create match store in frontend/src/stores/matchStore.ts (match state, participants, heat data)
- [ ] T081 [US1] Implement RPC call matchmaking.join in frontend/src/services/centrifugo/rpc.ts
- [ ] T082 [US1] Implement RPC call matchmaking.cancel in frontend/src/services/centrifugo/rpc.ts
- [ ] T083 [US1] Subscribe to user:{user_id} channel on login in frontend/src/services/centrifugo/subscriptions.ts
- [ ] T084 [US1] Handle match_found event and navigate to Race HUD in frontend/src/services/centrifugo/eventHandlers.ts
- [ ] T085 [US1] Create Matchmaking page component in frontend/src/pages/Matchmaking.tsx (status, timer, cancel button)

**Frontend: Race HUD (Pixi.js)**

- [ ] T086 [P] [US1] Create Pixi.js app initialization in frontend/src/pixi/app.ts
- [ ] T087 [P] [US1] Create Race scene setup in frontend/src/pixi/scenes/RaceScene.ts
- [ ] T088 [P] [US1] Create Speed display component (BitmapFont) in frontend/src/pixi/sprites/SpeedDisplay.ts
- [ ] T089 [P] [US1] Create countdown overlay in frontend/src/pixi/sprites/CountdownOverlay.ts
- [ ] T090 [US1] Implement RPC call match.earn_points in frontend/src/services/centrifugo/rpc.ts
- [ ] T091 [US1] Create Race page component in frontend/src/pages/Race.tsx (Pixi container, "EARN POINTS" button)
- [ ] T092 [US1] Subscribe to match:{match_id} channel when match starts in frontend/src/services/centrifugo/subscriptions.ts
- [ ] T093 [US1] Handle heat_started event and start Pixi animation in frontend/src/services/centrifugo/eventHandlers.ts
- [ ] T094 [US1] Handle heat_ended event and transition to intermission in frontend/src/services/centrifugo/eventHandlers.ts
- [ ] T095 [US1] Update Speed display in Pixi ticker loop in frontend/src/pixi/scenes/RaceScene.ts
- [ ] T096 [US1] Implement spectator state (controls disabled, score frozen) in frontend/src/pages/Race.tsx

**Frontend: Intermission**

- [ ] T097 [P] [US1] Create Intermission component in frontend/src/components/race/Intermission.tsx (standings table, 5s timer)
- [ ] T098 [US1] Calculate position deltas (↑ ↓ =) in frontend/src/components/race/Intermission.tsx
- [ ] T099 [US1] Highlight player's own row in standings table in frontend/src/components/race/Intermission.tsx
- [ ] T100 [US1] Auto-transition to next heat after 5 seconds in frontend/src/components/race/Intermission.tsx

**Frontend: Settlement**

- [ ] T101 [P] [US1] Create Settlement page component in frontend/src/pages/Settlement.tsx
- [ ] T102 [US1] Display top 3 podium in frontend/src/components/settlement/Podium.tsx
- [ ] T103 [US1] Display remaining players (4th-10th) in frontend/src/components/settlement/RankedList.tsx
- [ ] T104 [US1] Display FUEL win/loss and BURN reward in frontend/src/components/settlement/PrizeDisplay.tsx
- [ ] T105 [US1] Handle match_settled event and navigate to Settlement in frontend/src/services/centrifugo/eventHandlers.ts
- [ ] T106 [US1] Implement "RACE AGAIN" button (check balance, return to matchmaking) in frontend/src/pages/Settlement.tsx
- [ ] T107 [US1] Implement "BACK TO GARAGE" button in frontend/src/pages/Settlement.tsx

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Ghost Player System & Matchmaking Speed (Priority: P1) 🎯 MVP

**Goal**: Fill remaining lobby slots with Ghost players (historical replays) when <10 live players available, ensuring matches start within 20 seconds

**Independent Test**: Simulate low player availability, verify lobby fills with Ghosts within 20s, Ghosts behave identically to live players, can win prizes

### Implementation for User Story 2

**Backend: Ghosts Module**

- [ ] T108 [P] [US2] Create ghosts module structure in backend/internal/modules/ghosts/
- [ ] T109 [P] [US2] Create ghost replays repository interface in backend/internal/storage/postgres/repository/ghost_replays.go
- [ ] T110 [P] [US2] Implement stratified sampling algorithm (select by score percentiles) in backend/internal/modules/ghosts/selection.go
- [ ] T111 [US2] Implement GhostsService with SelectGhosts() method in backend/internal/modules/ghosts/service.go
- [ ] T112 [US2] Implement CreateReplay() method to persist player performance in backend/internal/modules/ghosts/service.go
- [ ] T113 [US2] Integrate Ghost selection in lobby formation (fill remaining slots) in backend/internal/modules/matchmaker/lobby.go
- [ ] T114 [US2] Implement Ghost buy-in deduction from HOUSE_FUEL in backend/internal/modules/gameengine/settlement.go
- [ ] T115 [US2] Implement Ghost prize credit to HOUSE_FUEL in backend/internal/modules/gameengine/settlement.go
- [ ] T116 [US2] Create match participants with is_ghost=TRUE for Ghosts in backend/internal/modules/gameengine/service.go

**Backend: Ghost Replay Behavior**

- [ ] T117 [P] [US2] Load Ghost behavioral_data (lock timings) in backend/internal/modules/gameengine/state.go
- [ ] T118 [US2] Auto-lock Ghost scores at behavioral_data.heat{N}_lock_time in backend/internal/modules/gameengine/heat.go
- [ ] T119 [US2] Exclude Ghosts from leaderboard queries (WHERE is_ghost = FALSE) in backend/internal/storage/postgres/repository/matches.go

**Frontend: Ghost Display**

- [ ] T120 [P] [US2] Display Ghosts identically to live players in race HUD in frontend/src/pages/Race.tsx
- [ ] T121 [US2] Display Ghosts in intermission standings (no visual distinction) in frontend/src/components/race/Intermission.tsx
- [ ] T122 [US2] Display Ghosts in settlement (no visual distinction) in frontend/src/pages/Settlement.tsx

**Post-Match: Ghost Replay Creation**

- [ ] T123 [US2] Call CreateReplay() for all live players after settlement in backend/internal/modules/gameengine/settlement.go

**Checkpoint**: Ghost system complete - matches fill instantly, Ghosts indistinguishable from live players during race

---

## Phase 5: User Story 3 - Multi-League Progression & Economy (Priority: P2)

**Goal**: Players can progress through Rookie (10 FUEL) → Street (50 FUEL) → Pro (300 FUEL) → Top Fuel (3000 FUEL) with increasing stakes and BURN rewards

**Independent Test**: Provide test account with FUEL, play across leagues, verify buy-in deductions, prize calculations (8% rake), BURN rewards scale with league

### Implementation for User Story 3

**Backend: League Logic**

- [ ] T124 [P] [US3] Define league constants (buy-ins, BURN tables) in backend/internal/modules/gameengine/leagues.go
- [ ] T125 [US3] Implement Rookie race count validation (max 3) in backend/internal/modules/matchmaker/service.go
- [ ] T126 [US3] Increment rookie_races_completed after Rookie settlement in backend/internal/modules/gameengine/settlement.go
- [ ] T127 [US3] Implement BURN reward calculation per league/position in backend/internal/modules/gameengine/burn_rewards.go
- [ ] T128 [US3] Apply BURN ledger entries (operation_type=MATCH_BURN_REWARD) in backend/internal/modules/gameengine/settlement.go
- [ ] T129 [US3] Exclude Rookie from BURN rewards (FR-038) in backend/internal/modules/gameengine/burn_rewards.go

**Backend: Balance Validation**

- [ ] T130 [US3] Validate balance >= league buy-in before joining queue in backend/internal/modules/matchmaker/service.go
- [ ] T131 [US3] Return error INSUFFICIENT_BALANCE if validation fails in backend/internal/modules/matchmaker/service.go

**Frontend: League Restrictions**

- [ ] T132 [P] [US3] Disable Rookie after 3 races with tooltip in frontend/src/components/garage/LeagueCard.tsx
- [ ] T133 [US3] Disable leagues with insufficient balance with "GO TO GAS STATION" CTA in frontend/src/components/garage/LeagueCard.tsx
- [ ] T134 [US3] Display BURN balance in Garage in frontend/src/pages/Garage.tsx
- [ ] T135 [US3] Display BURN reward in settlement (if applicable) in frontend/src/components/settlement/PrizeDisplay.tsx

**Checkpoint**: All 4 leagues functional with correct buy-ins, rake, prizes, and BURN rewards

---

## Phase 6: User Story 4 - Disconnect/Reconnect Handling (Priority: P2)

**Goal**: Players who disconnect mid-race have 3 seconds to reconnect; if successful, gameplay resumes; if failed, heat score = 0 (crashed)

**Independent Test**: Simulate network interruptions, verify "Reconnecting…" overlay, reconnect within 3s resumes gameplay, timeout >3s auto-crashes

### Implementation for User Story 4

**Backend: Connection Tracking**

- [ ] T136 [P] [US4] Track active connections per user in Redis (key: conn:{user_id}) in backend/internal/modules/gateway/connections.go
- [ ] T137 [US4] Implement disconnect detection via Centrifugo webhooks in backend/internal/modules/gateway/webhooks.go
- [ ] T138 [US4] Start 3-second grace period timer on disconnect in backend/internal/modules/gameengine/reconnect.go
- [ ] T139 [US4] Auto-crash current heat if timer expires (score = 0) in backend/internal/modules/gameengine/reconnect.go
- [ ] T140 [US4] Cancel timer if reconnection succeeds within 3s in backend/internal/modules/gameengine/reconnect.go

**Backend: Pre-Race Disconnect**

- [ ] T141 [US4] Detect disconnect during countdown (before heat 1) in backend/internal/modules/gameengine/state.go
- [ ] T142 [US4] Remove player from lobby and refund buy-in in backend/internal/modules/gameengine/state.go

**Frontend: Reconnect Overlay**

- [ ] T143 [P] [US4] Create "Reconnecting…" overlay component in frontend/src/components/race/ReconnectOverlay.tsx
- [ ] T144 [US4] Show overlay on Centrifugo disconnect event in frontend/src/services/centrifugo/client.ts
- [ ] T145 [US4] Hide overlay on successful reconnect in frontend/src/services/centrifugo/client.ts
- [ ] T146 [US4] Sync match state after reconnect (request current heat data) in frontend/src/services/centrifugo/reconnect.ts

**Checkpoint**: Disconnect/reconnect handling complete, fair auto-crash for timeout >3s

---

## Phase 7: User Story 5 - TON Wallet Integration & Gas Station (Priority: P2)

**Goal**: Players can connect TON wallet, deposit TON for FUEL, withdraw FUEL to TON (BURN non-convertible)

**Independent Test**: Connect TON wallet, deposit 10 TON, verify FUEL credit, attempt withdrawal, verify TON transfer, verify BURN cannot be converted

### Implementation for User Story 5

**Backend: Payments Module**

- [ ] T147 [P] [US5] Create payments module structure in backend/internal/modules/payments/
- [ ] T148 [P] [US5] Create TonCenter API client in backend/internal/toncenter/client.go
- [ ] T149 [P] [US5] Create payments repository interface in backend/internal/storage/postgres/repository/payments.go
- [ ] T150 [US5] Implement PaymentsService with CreateDeposit() method in backend/internal/modules/payments/service.go
- [ ] T151 [US5] Implement CreateWithdrawal() method in backend/internal/modules/payments/service.go
- [ ] T152 [US5] Implement deposit reconciliation worker (poll TonCenter, credit on confirm) in backend/internal/modules/payments/deposit_worker.go
- [ ] T153 [US5] Implement withdrawal broadcasting (lock FUEL, broadcast tx via TonCenter) in backend/internal/modules/payments/withdrawal_worker.go
- [ ] T154 [US5] Apply deposit ledger entries (DEPOSIT, credit FUEL) in backend/internal/modules/payments/service.go
- [ ] T155 [US5] Apply withdrawal ledger entries (WITHDRAWAL, debit FUEL) in backend/internal/modules/payments/service.go
- [ ] T156 [US5] Implement idempotency checks (client_request_id, ton_tx_hash) in backend/internal/modules/payments/service.go
- [ ] T157 [US5] Implement HTTP handler POST /api/v1/payments/deposit in backend/internal/modules/gateway/http/payments_handler.go
- [ ] T158 [US5] Implement HTTP handler POST /api/v1/payments/withdraw in backend/internal/modules/gateway/http/payments_handler.go
- [ ] T159 [US5] Implement HTTP handler GET /api/v1/payments/history in backend/internal/modules/gateway/http/payments_handler.go

**Frontend: Gas Station**

- [ ] T160 [P] [US5] Create TON Connect integration in frontend/src/services/wallet/tonconnect.ts
- [ ] T161 [P] [US5] Create deposit API call in frontend/src/services/api/payments.ts (POST /api/v1/payments/deposit)
- [ ] T162 [P] [US5] Create withdrawal API call in frontend/src/services/api/payments.ts (POST /api/v1/payments/withdraw)
- [ ] T163 [P] [US5] Create payment history API call in frontend/src/services/api/payments.ts (GET /api/v1/payments/history)
- [ ] T164 [US5] Create Gas Station page component in frontend/src/pages/GasStation.tsx
- [ ] T165 [US5] Create wallet connection button (TON Connect) in frontend/src/components/gasstation/ConnectWallet.tsx
- [ ] T166 [US5] Create deposit form in frontend/src/components/gasstation/DepositForm.tsx
- [ ] T167 [US5] Create withdrawal form in frontend/src/components/gasstation/WithdrawForm.tsx
- [ ] T168 [US5] Display payment history in frontend/src/components/gasstation/PaymentHistory.tsx
- [ ] T169 [US5] Show error message if user tries to convert BURN in frontend/src/components/gasstation/WithdrawForm.tsx

**Checkpoint**: TON wallet integration complete, deposits/withdrawals functional, BURN non-convertible

---

## Phase 8: User Story 6 - Target Line Mechanic for Chase Tension (Priority: P3)

**Goal**: In Heat 2 and Heat 3, display Target Line at score to beat (Heat 1 winner or current leader total), with visual/audio feedback when crossed

**Independent Test**: Observe Heat 2 and Heat 3, verify Target Line at correct position, fixed at heat start, includes Ghost results, triggers feedback when crossed

### Implementation for User Story 6

**Backend: Target Line Calculation**

- [ ] T170 [P] [US6] Calculate Target Line for Heat 2 (Heat 1 winner score) in backend/internal/modules/gameengine/heat.go
- [ ] T171 [P] [US6] Calculate Target Line for Heat 3 (current leader total) in backend/internal/modules/gameengine/heat.go
- [ ] T172 [US6] Include Target Line in heat_started event payload in backend/internal/modules/gateway/events/match_events.go

**Frontend: Target Line Display**

- [ ] T173 [P] [US6] Create Target Line sprite in frontend/src/pixi/sprites/TargetLine.ts
- [ ] T174 [US6] Position Target Line at correct Speed value in frontend/src/pixi/scenes/RaceScene.ts
- [ ] T175 [US6] Show one-time tooltip "Reach this line to beat the current leader" in frontend/src/components/race/TargetLineTooltip.tsx
- [ ] T176 [US6] Trigger visual pulse and audio cue when Speed crosses Target Line in frontend/src/pixi/scenes/RaceScene.ts
- [ ] T177 [US6] Hide Target Line in Heat 1 in frontend/src/pixi/scenes/RaceScene.ts

**Checkpoint**: Target Line mechanic complete, increases PvP tension in Heats 2 and 3

---

## Phase 9: User Story 7 - Intermission Standings & Position Tracking (Priority: P3)

**Goal**: After each heat, display 5-second intermission with standings table (positions, names, scores, deltas), player row highlighted

**Independent Test**: Complete Heat 1 and Heat 2, verify 5-second timer, standings show all 10 players, position deltas appear, player row highlighted

### Implementation for User Story 7

**Backend: Intermission Data**

- [ ] T178 [P] [US7] Calculate standings after each heat in backend/internal/modules/gameengine/heat.go
- [ ] T179 [US7] Include standings in heat_ended event payload in backend/internal/modules/gateway/events/match_events.go

**Frontend: Enhanced Intermission**

- [ ] T180 [P] [US7] Display player names in standings table in frontend/src/components/race/Intermission.tsx
- [ ] T181 [P] [US7] Display total scores in standings table in frontend/src/components/race/Intermission.tsx
- [ ] T182 [US7] Calculate and display position deltas (↑ ↓ =) in frontend/src/components/race/Intermission.tsx
- [ ] T183 [US7] Highlight player's own row with distinct background in frontend/src/components/race/Intermission.tsx
- [ ] T184 [US7] Auto-scroll to ensure player row is visible in frontend/src/components/race/Intermission.tsx

**Checkpoint**: Intermission standings complete, provides PvP feedback and builds tension

---

## Phase 10: User Story 8 - Early Heat End Optimization (Priority: P3)

**Goal**: If all alive players lock scores before crash/timer expiry, heat ends immediately (saves waiting time)

**Independent Test**: Have all 10 players lock scores early (e.g., t=8s), verify heat ends immediately, displays final Speed at t=8s, transitions to intermission

### Implementation for User Story 8

**Backend: Early Heat End**

- [ ] T185 [P] [US8] Track alive player count per heat in backend/internal/modules/gameengine/heat.go
- [ ] T186 [US8] Check if all alive players locked after each EarnPoints call in backend/internal/modules/gameengine/earn_points.go
- [ ] T187 [US8] End heat immediately if condition met (publish heat_ended early) in backend/internal/modules/gameengine/heat.go
- [ ] T188 [US8] Record actual heat duration in match_participants table in backend/internal/modules/gameengine/heat.go

**Frontend: Early End Handling**

- [ ] T189 [US8] Stop Pixi ticker loop on early heat_ended event in frontend/src/pixi/scenes/RaceScene.ts
- [ ] T190 [US8] Display final Speed value from event payload in frontend/src/pages/Race.tsx

**Checkpoint**: Early heat end optimization complete, reduces unnecessary waiting

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T191 [P] Add error handling for all HTTP endpoints in backend/internal/modules/gateway/http/
- [ ] T192 [P] Add error handling for all RPC handlers in backend/internal/modules/gateway/rpc/
- [ ] T193 [P] Implement unified error overlay component in frontend/src/components/common/ErrorOverlay.tsx
- [ ] T194 [P] Add contextual loading states in frontend/src/components/common/LoadingState.tsx
- [ ] T195 [P] Implement localization (i18n) with 4 languages in frontend/src/locales/ (en, ru, es-PE, pt-BR)
- [ ] T196 [P] Integrate Amplitude analytics SDK in frontend/src/services/analytics/amplitude.ts
- [ ] T197 [P] Add core analytics events (app_opened, match_started, match_finished) in frontend/src/services/analytics/events.ts
- [ ] T198 [P] Add Prometheus metrics for RPC latency in backend/internal/metrics/rpc.go
- [ ] T199 [P] Add Prometheus metrics for matchmaking wait time in backend/internal/metrics/matchmaking.go
- [ ] T200 [P] Add Prometheus metrics for house balance gauge in backend/internal/metrics/economy.go
- [ ] T201 [P] Implement graceful shutdown (SIGTERM, DRAINING state) in backend/cmd/server/main.go
- [ ] T202 [P] Add structured logging for all module operations in backend/internal/modules/
- [ ] T203 [P] Integrate Sentry for error tracking in backend/cmd/server/main.go
- [ ] T204 [P] Integrate Sentry for error tracking in frontend/src/main.tsx
- [ ] T205 Run validation against quickstart.md scenarios

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-10)**: All depend on Foundational phase completion
  - User Story 1 (Phase 3): Can start after Foundational - No dependencies on other stories
  - User Story 2 (Phase 4): Depends on User Story 1 completion (extends matchmaker and gameengine)
  - User Story 3 (Phase 5): Depends on User Story 1 completion (extends gameengine and account)
  - User Story 4 (Phase 6): Depends on User Story 1 completion (extends gateway and gameengine)
  - User Story 5 (Phase 7): Can start after Foundational - No dependencies on other stories (independent module)
  - User Story 6 (Phase 8): Depends on User Story 1 completion (extends gameengine and race HUD)
  - User Story 7 (Phase 9): Depends on User Story 1 completion (extends intermission)
  - User Story 8 (Phase 10): Depends on User Story 1 completion (extends heat lifecycle)
- **Polish (Phase 11)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - Core gameplay loop
- **User Story 2 (P1)**: Depends on User Story 1 - Extends matchmaker/gameengine with Ghost system
- **User Story 3 (P2)**: Depends on User Story 1 - Extends economy with leagues and BURN
- **User Story 4 (P2)**: Depends on User Story 1 - Adds reconnect handling to gameplay
- **User Story 5 (P2)**: Can start after Foundational - Independent TON integration
- **User Story 6 (P3)**: Depends on User Story 1 - UX enhancement for Heats 2 and 3
- **User Story 7 (P3)**: Depends on User Story 1 - UX enhancement for intermission
- **User Story 8 (P3)**: Depends on User Story 1 - Optimization for heat lifecycle

### Within Each User Story

- Backend models and repositories before services
- Services before HTTP/RPC handlers
- Backend modules before frontend integration
- Frontend API clients before UI components
- Core implementation before edge cases

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, User Story 1 and User Story 5 can start in parallel (independent modules)
- Within each user story, tasks marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members (if dependencies satisfied)

---

## Parallel Example: User Story 1 Backend Modules

```bash
# Launch all auth module tasks together:
Task: "Create auth module structure in backend/internal/modules/auth/"
Task: "Implement Telegram initData validation in backend/internal/modules/auth/telegram.go"

# Launch all account module tasks together:
Task: "Create account module structure in backend/internal/modules/account/"
Task: "Create wallet repository interface in backend/internal/storage/postgres/repository/wallets.go"
Task: "Create ledger repository interface in backend/internal/storage/postgres/repository/ledger.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 + User Story 2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (core gameplay)
4. Complete Phase 4: User Story 2 (Ghost system)
5. **STOP and VALIDATE**: Test MVP independently with Rookie league only
6. Deploy/demo if ready

**MVP delivers**: Complete gameplay loop with fast matchmaking via Ghost players

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 + 2 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 3 → Test multi-league economy → Deploy/Demo
4. Add User Story 4 → Test reconnect handling → Deploy/Demo
5. Add User Story 5 → Test TON integration → Deploy/Demo
6. Add User Stories 6-8 → Polish UX → Deploy/Demo
7. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (core gameplay)
   - Developer B: User Story 5 (TON integration - independent)
3. After User Story 1:
   - Developer A: User Story 2 (Ghosts)
   - Developer B: User Story 3 (leagues)
   - Developer C: User Story 4 (reconnect)
4. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files/modules, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group (one commit, one task per Constitution II)
- Stop at any checkpoint to validate story independently
- **Avoid**: vague tasks, same file conflicts, cross-story dependencies that break independence
- **MVP scope**: User Story 1 + User Story 2 (P1 stories) deliver complete core experience
- **Full MVP**: All P1 + P2 stories (US1-US5) required for launch
- **P3 stories**: Nice-to-have UX enhancements, can defer post-launch

---

## Task Summary

- **Total Tasks**: 205
- **Phase 1 (Setup)**: 10 tasks
- **Phase 2 (Foundational)**: 22 tasks
- **Phase 3 (US1 - Core Gameplay)**: 75 tasks
- **Phase 4 (US2 - Ghost System)**: 16 tasks
- **Phase 5 (US3 - Leagues)**: 12 tasks
- **Phase 6 (US4 - Reconnect)**: 11 tasks
- **Phase 7 (US5 - TON Integration)**: 23 tasks
- **Phase 8 (US6 - Target Line)**: 8 tasks
- **Phase 9 (US7 - Intermission)**: 7 tasks
- **Phase 10 (US8 - Early End)**: 6 tasks
- **Phase 11 (Polish)**: 15 tasks

**MVP Task Count** (US1 + US2): 123 tasks
**Full Launch** (US1-US5): 169 tasks
**Complete Feature** (All stories): 205 tasks
