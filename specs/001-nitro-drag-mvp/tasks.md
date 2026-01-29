# Tasks: Nitro Drag Royale MVP

**Input**: Design documents from `/specs/001-nitro-drag-mvp/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/, research.md, quickstart.md

**Tests**: Tests are OPTIONAL - only included if explicitly requested in the feature specification. This task list focuses on implementation tasks.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

**Priority Changes**: TON wallet integration (US5) moved to Phase 4 (right after US1) to enable early deposit/withdrawal testing.

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

**Frontend**: `frontend/` at repository root (Telegram Mini App)
- `frontend/src/components/` — React components
- `frontend/src/pages/` — Top-level page components
- `frontend/src/services/` — Business logic layer
- `frontend/src/stores/` — Zustand stores

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create project directory structure per plan.md (backend/, frontend/, deployments/)
- [ ] T002 [P] Initialize Go module in backend/ with go.mod (Go 1.25+, include logrus dependency)
- [ ] T003 [P] Initialize React project in frontend/ with Vite 5+ and TypeScript 5+ as Telegram Mini App
- [ ] T004 [P] Create docker-compose.yml in deployments/ for PostgreSQL 17+, Redis 7+, Centrifugo v4
- [ ] T005 [P] Create Makefile in backend/ with targets: dev, test, migrate-up, migrate-down, proto-gen
- [ ] T006 [P] Setup golang-migrate for database migrations in backend/internal/storage/postgres/migrations/
- [ ] T007 [P] Create backend/internal/config/ package for 12-factor env var configuration
- [ ] T008 [P] Setup Prometheus metrics endpoint in backend/internal/metrics/
- [ ] T009 [P] Create .env.example files for backend and frontend with required environment variables
- [ ] T010 [P] Setup Telegram Mini Apps SDK in frontend/src/services/telegram/sdk.ts

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T011 Create single database migration in backend/internal/storage/postgres/migrations/000001_initial_schema.up.sql with all 9 tables (users, wallets, system_wallets, ledger_entries, matches, match_participants, ghost_replays, match_settlements, payments) and seed data for system wallets (HOUSE_FUEL, RAKE_FUEL)
- [ ] T012 [P] Create PostgreSQL models in backend/internal/storage/postgres/models/ for all 9 entities
- [ ] T013 [P] Create Redis client wrapper in backend/internal/storage/redis/client.go
- [ ] T014 [P] Generate Protobuf code from backend/proto/ using protoc (matchmaking.proto, match.proto)
- [ ] T015 [P] Create Chi router setup in backend/cmd/server/main.go with logrus middleware (RequestID, Logger, Recoverer)
- [ ] T016 [P] Create JWT utility package in backend/internal/auth/jwt.go for token generation/validation
- [ ] T017 [P] Create Centrifugo gRPC client wrapper in backend/internal/centrifugo/client.go
- [ ] T018 [P] Create decimal utility package in backend/internal/decimal/ for fixed-point arithmetic (shopspring/decimal)
- [ ] T019 [P] Setup frontend routing with React Router in frontend/src/App.tsx
- [ ] T020 [P] Create Zustand stores structure in frontend/src/stores/ (auth, wallet, match)
- [ ] T021 [P] Create React Query setup in frontend/src/services/api/client.ts
- [ ] T022 [P] Create Centrifuge.js client wrapper in frontend/src/services/centrifugo/client.ts
- [ ] T023 [P] Create TON Connect UI setup in frontend/src/services/wallet/tonconnect.ts
- [ ] T024 [P] Initialize Telegram WebApp API in frontend/src/services/telegram/webapp.ts

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Enter and Complete First Race (Priority: P1) 🎯 MVP

**Goal**: New player can join Rookie league, play 3 heats with speed acceleration mechanic, and receive winnings/BURN based on final position

**Independent Test**: Onboard new user, provide initial FUEL, observe complete race cycle (Garage → Matchmaking → Race → Settlement)

### Implementation for User Story 1

**Backend: Auth Module**

- [ ] T025 [P] [US1] Create auth module structure in backend/internal/modules/auth/
- [ ] T026 [P] [US1] Implement Telegram initData validation in backend/internal/modules/auth/telegram.go
- [ ] T027 [US1] Implement AuthService with Authenticate() method in backend/internal/modules/auth/service.go
- [ ] T028 [US1] Create user repository interface in backend/internal/storage/postgres/repository/users.go
- [ ] T029 [US1] Implement HTTP handler POST /api/v1/auth/telegram in backend/internal/modules/gateway/http/auth_handler.go

**Backend: Account Module**

- [ ] T030 [P] [US1] Create account module structure in backend/internal/modules/account/
- [ ] T031 [P] [US1] Create wallet repository interface in backend/internal/storage/postgres/repository/wallets.go
- [ ] T032 [P] [US1] Create ledger repository interface in backend/internal/storage/postgres/repository/ledger.go
- [ ] T033 [US1] Implement AccountService with GetWallet() method in backend/internal/modules/account/service.go
- [ ] T034 [US1] Implement ledger operations (DebitFuel, CreditFuel, RecordEntry) in backend/internal/modules/account/ledger.go
- [ ] T035 [US1] Implement HTTP handler GET /api/v1/garage in backend/internal/modules/gateway/http/garage_handler.go

**Backend: Matchmaker Module**

- [ ] T036 [P] [US1] Create matchmaker module structure in backend/internal/modules/matchmaker/
- [ ] T037 [P] [US1] Implement Redis queue operations in backend/internal/modules/matchmaker/queue.go
- [ ] T038 [US1] Implement MatchmakerService with JoinQueue() method in backend/internal/modules/matchmaker/service.go
- [ ] T039 [US1] Implement CancelQueue() method in backend/internal/modules/matchmaker/service.go
- [ ] T040 [US1] Implement lobby formation logic (10 players max, 20s timeout) in backend/internal/modules/matchmaker/lobby.go
- [ ] T041 [US1] Implement RPC handler matchmaking.join in backend/internal/modules/gateway/rpc/matchmaking_handler.go
- [ ] T042 [US1] Implement RPC handler matchmaking.cancel in backend/internal/modules/gateway/rpc/matchmaking_handler.go

**Backend: Game Engine Module**

- [ ] T043 [P] [US1] Create gameengine module structure in backend/internal/modules/gameengine/
- [ ] T044 [P] [US1] Create match repository interface in backend/internal/storage/postgres/repository/matches.go
- [ ] T045 [P] [US1] Create match participants repository interface in backend/internal/storage/postgres/repository/match_participants.go
- [ ] T046 [P] [US1] Implement speed formula calculation Speed = 500 * ((e^(0.08·t) - 1) / (e^(0.08·25) - 1)) in backend/internal/modules/gameengine/physics.go
- [ ] T047 [P] [US1] Implement crash seed generation (cryptographic hash) in backend/internal/modules/gameengine/provable_fairness.go
- [ ] T048 [US1] Implement GameEngineService with CreateMatch() method in backend/internal/modules/gameengine/service.go
- [ ] T049 [US1] Implement in-memory match state machine (FORMING → IN_PROGRESS → COMPLETED) in backend/internal/modules/gameengine/state.go
- [ ] T050 [US1] Implement heat lifecycle (countdown → active → intermission) in backend/internal/modules/gameengine/heat.go
- [ ] T051 [US1] Implement EarnPoints (lock score) logic in backend/internal/modules/gameengine/earn_points.go
- [ ] T052 [US1] Implement RPC handler match.earn_points in backend/internal/modules/gateway/rpc/match_handler.go
- [ ] T053 [US1] Implement early heat end optimization (all players finished) in backend/internal/modules/gameengine/heat.go
- [ ] T054 [US1] Implement settlement calculation (positions, prizes, BURN rewards) in backend/internal/modules/gameengine/settlement.go
- [ ] T055 [US1] Integrate settlement with ledger (apply prize/BURN entries) in backend/internal/modules/gameengine/settlement.go

**Backend: Gateway Module (Events)**

- [ ] T056 [P] [US1] Create gateway module structure in backend/internal/modules/gateway/
- [ ] T057 [P] [US1] Implement Centrifugo publisher methods (PublishToUser, PublishToMatch) in backend/internal/modules/gateway/centrifugo.go
- [ ] T058 [P] [US1] Implement event schemas (match_found, heat_started, heat_ended, match_settled) in backend/internal/modules/gateway/events/
- [ ] T059 [US1] Publish match_found event to user:{user_id} channel in backend/internal/modules/matchmaker/service.go
- [ ] T060 [US1] Publish heat_started event to match:{match_id} channel in backend/internal/modules/gameengine/heat.go
- [ ] T061 [US1] Publish heat_ended event to match:{match_id} channel in backend/internal/modules/gameengine/heat.go
- [ ] T062 [US1] Publish match_settled event to match:{match_id} channel in backend/internal/modules/gameengine/settlement.go
- [ ] T063 [US1] Publish balance_updated event to user:{user_id} channel after settlement in backend/internal/modules/account/service.go

**Frontend: Auth & Garage**

- [ ] T064 [P] [US1] Create auth store in frontend/src/stores/authStore.ts (JWT tokens, user state)
- [ ] T065 [P] [US1] Implement Telegram initData extraction in frontend/src/services/auth/telegram.ts
- [ ] T066 [US1] Create login API call in frontend/src/services/api/auth.ts (POST /api/v1/auth/telegram)
- [ ] T067 [US1] Implement automatic authentication on app launch in frontend/src/App.tsx (extract initData, call auth API, store tokens)
- [ ] T068 [P] [US1] Create wallet store in frontend/src/stores/walletStore.ts (balances, league access)
- [ ] T069 [US1] Create garage API call in frontend/src/services/api/garage.ts (GET /api/v1/garage)
- [ ] T070 [US1] Create Garage page component in frontend/src/pages/Garage.tsx (league selector, balances, "RACE NOW")
- [ ] T071 [US1] Create league card components in frontend/src/components/garage/LeagueCard.tsx

**Frontend: Matchmaking**

- [ ] T072 [P] [US1] Create match store in frontend/src/stores/matchStore.ts (match state, participants, heat data)
- [ ] T073 [US1] Implement RPC call matchmaking.join in frontend/src/services/centrifugo/rpc.ts
- [ ] T074 [US1] Implement RPC call matchmaking.cancel in frontend/src/services/centrifugo/rpc.ts
- [ ] T075 [US1] Subscribe to user:{user_id} channel on login in frontend/src/services/centrifugo/subscriptions.ts
- [ ] T076 [US1] Handle match_found event and navigate to Race HUD in frontend/src/services/centrifugo/eventHandlers.ts
- [ ] T077 [US1] Create Matchmaking page component in frontend/src/pages/Matchmaking.tsx (status, timer, cancel button)

**Frontend: Race HUD (Pixi.js)**

- [ ] T078 [P] [US1] Create Pixi.js app initialization in frontend/src/pixi/app.ts
- [ ] T079 [P] [US1] Create Race scene setup in frontend/src/pixi/scenes/RaceScene.ts
- [ ] T080 [P] [US1] Create Speed display component (BitmapFont) in frontend/src/pixi/sprites/SpeedDisplay.ts
- [ ] T081 [P] [US1] Create countdown overlay in frontend/src/pixi/sprites/CountdownOverlay.ts
- [ ] T082 [US1] Implement RPC call match.earn_points in frontend/src/services/centrifugo/rpc.ts
- [ ] T083 [US1] Create Race page component in frontend/src/pages/Race.tsx (Pixi container, "EARN POINTS" button)
- [ ] T084 [US1] Subscribe to match:{match_id} channel when match starts in frontend/src/services/centrifugo/subscriptions.ts
- [ ] T085 [US1] Handle heat_started event and start Pixi animation in frontend/src/services/centrifugo/eventHandlers.ts
- [ ] T086 [US1] Handle heat_ended event and transition to intermission in frontend/src/services/centrifugo/eventHandlers.ts
- [ ] T087 [US1] Update Speed display in Pixi ticker loop in frontend/src/pixi/scenes/RaceScene.ts
- [ ] T088 [US1] Implement spectator state (controls disabled, score frozen) in frontend/src/pages/Race.tsx

**Frontend: Intermission**

- [ ] T089 [P] [US1] Create Intermission component in frontend/src/components/race/Intermission.tsx (standings table, 5s timer)
- [ ] T090 [US1] Calculate position deltas (↑ ↓ =) in frontend/src/components/race/Intermission.tsx
- [ ] T091 [US1] Highlight player's own row in standings table in frontend/src/components/race/Intermission.tsx
- [ ] T092 [US1] Auto-transition to next heat after 5 seconds in frontend/src/components/race/Intermission.tsx

**Frontend: Settlement**

- [ ] T093 [P] [US1] Create Settlement page component in frontend/src/pages/Settlement.tsx
- [ ] T094 [US1] Display top 3 podium in frontend/src/components/settlement/Podium.tsx
- [ ] T095 [US1] Display remaining players (4th-10th) in frontend/src/components/settlement/RankedList.tsx
- [ ] T096 [US1] Display FUEL win/loss and BURN reward in frontend/src/components/settlement/PrizeDisplay.tsx
- [ ] T097 [US1] Handle match_settled event and navigate to Settlement in frontend/src/services/centrifugo/eventHandlers.ts
- [ ] T098 [US1] Implement "RACE AGAIN" button (check balance, return to matchmaking) in frontend/src/pages/Settlement.tsx
- [ ] T099 [US1] Implement "BACK TO GARAGE" button in frontend/src/pages/Settlement.tsx

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 5 - TON Wallet Integration & Gas Station (Priority: P2) 🎯 EARLY

**Goal**: Players can connect TON wallet, deposit TON for FUEL, withdraw FUEL to TON (BURN non-convertible)

**Independent Test**: Connect TON wallet, deposit 10 TON, verify FUEL credit, attempt withdrawal, verify TON transfer, verify BURN cannot be converted

**Why Early**: TON deposits/withdrawals enable real monetary testing early. Independent module can be developed in parallel with User Story 2.

### Implementation for User Story 5

**Backend: Payments Module**

- [ ] T100 [P] [US5] Create payments module structure in backend/internal/modules/payments/
- [ ] T101 [P] [US5] Create TonCenter API client in backend/internal/toncenter/client.go
- [ ] T102 [P] [US5] Create payments repository interface in backend/internal/storage/postgres/repository/payments.go
- [ ] T103 [US5] Implement PaymentsService with CreateDeposit() method in backend/internal/modules/payments/service.go
- [ ] T104 [US5] Implement CreateWithdrawal() method in backend/internal/modules/payments/service.go
- [ ] T105 [US5] Implement deposit reconciliation worker (poll TonCenter, credit on confirm) in backend/internal/modules/payments/deposit_worker.go
- [ ] T106 [US5] Implement withdrawal broadcasting (lock FUEL, broadcast tx via TonCenter) in backend/internal/modules/payments/withdrawal_worker.go
- [ ] T107 [US5] Apply deposit ledger entries (DEPOSIT, credit FUEL) in backend/internal/modules/payments/service.go
- [ ] T108 [US5] Apply withdrawal ledger entries (WITHDRAWAL, debit FUEL) in backend/internal/modules/payments/service.go
- [ ] T109 [US5] Implement idempotency checks (client_request_id, ton_tx_hash) in backend/internal/modules/payments/service.go
- [ ] T110 [US5] Implement HTTP handler POST /api/v1/payments/deposit in backend/internal/modules/gateway/http/payments_handler.go
- [ ] T111 [US5] Implement HTTP handler POST /api/v1/payments/withdraw in backend/internal/modules/gateway/http/payments_handler.go
- [ ] T112 [US5] Implement HTTP handler GET /api/v1/payments/history in backend/internal/modules/gateway/http/payments_handler.go

**Frontend: Gas Station**

- [ ] T113 [P] [US5] Create deposit API call in frontend/src/services/api/payments.ts (POST /api/v1/payments/deposit)
- [ ] T114 [P] [US5] Create withdrawal API call in frontend/src/services/api/payments.ts (POST /api/v1/payments/withdraw)
- [ ] T115 [P] [US5] Create payment history API call in frontend/src/services/api/payments.ts (GET /api/v1/payments/history)
- [ ] T116 [US5] Create Gas Station page component in frontend/src/pages/GasStation.tsx
- [ ] T117 [US5] Create wallet connection button (TON Connect) in frontend/src/components/gasstation/ConnectWallet.tsx
- [ ] T118 [US5] Create deposit form in frontend/src/components/gasstation/DepositForm.tsx
- [ ] T119 [US5] Create withdrawal form in frontend/src/components/gasstation/WithdrawForm.tsx
- [ ] T120 [US5] Display payment history in frontend/src/components/gasstation/PaymentHistory.tsx
- [ ] T121 [US5] Show error message if user tries to convert BURN in frontend/src/components/gasstation/WithdrawForm.tsx

**Checkpoint**: TON wallet integration complete, deposits/withdrawals functional, BURN non-convertible

---

## Phase 5: User Story 2 - Ghost Player System & Matchmaking Speed (Priority: P1) 🎯 MVP

**Goal**: Fill remaining lobby slots with Ghost players (historical replays) when <10 live players available, ensuring matches start within 20 seconds

**Independent Test**: Simulate low player availability, verify lobby fills with Ghosts within 20s, Ghosts behave identically to live players, can win prizes

### Implementation for User Story 2

**Backend: Ghosts Module**

- [ ] T122 [P] [US2] Create ghosts module structure in backend/internal/modules/ghosts/
- [ ] T123 [P] [US2] Create ghost replays repository interface in backend/internal/storage/postgres/repository/ghost_replays.go
- [ ] T124 [P] [US2] Implement stratified sampling algorithm (select by score percentiles) in backend/internal/modules/ghosts/selection.go
- [ ] T125 [US2] Implement GhostsService with SelectGhosts() method in backend/internal/modules/ghosts/service.go
- [ ] T126 [US2] Implement CreateReplay() method to persist player performance in backend/internal/modules/ghosts/service.go
- [ ] T127 [US2] Integrate Ghost selection in lobby formation (fill remaining slots) in backend/internal/modules/matchmaker/lobby.go
- [ ] T128 [US2] Implement Ghost buy-in deduction from HOUSE_FUEL in backend/internal/modules/gameengine/settlement.go
- [ ] T129 [US2] Implement Ghost prize credit to HOUSE_FUEL in backend/internal/modules/gameengine/settlement.go
- [ ] T130 [US2] Create match participants with is_ghost=TRUE for Ghosts in backend/internal/modules/gameengine/service.go

**Backend: Ghost Replay Behavior**

- [ ] T131 [P] [US2] Load Ghost behavioral_data (lock timings) in backend/internal/modules/gameengine/state.go
- [ ] T132 [US2] Auto-lock Ghost scores at behavioral_data.heat{N}_lock_time in backend/internal/modules/gameengine/heat.go
- [ ] T133 [US2] Exclude Ghosts from leaderboard queries (WHERE is_ghost = FALSE) in backend/internal/storage/postgres/repository/matches.go

**Frontend: Ghost Display**

- [ ] T134 [P] [US2] Display Ghosts identically to live players in race HUD in frontend/src/pages/Race.tsx
- [ ] T135 [US2] Display Ghosts in intermission standings (no visual distinction) in frontend/src/components/race/Intermission.tsx
- [ ] T136 [US2] Display Ghosts in settlement (no visual distinction) in frontend/src/pages/Settlement.tsx

**Post-Match: Ghost Replay Creation**

- [ ] T137 [US2] Call CreateReplay() for all live players after settlement in backend/internal/modules/gameengine/settlement.go

**Checkpoint**: Ghost system complete - matches fill instantly, Ghosts indistinguishable from live players during race

---

## Phase 6: User Story 3 - Multi-League Progression & Economy (Priority: P2)

**Goal**: Players can progress through Rookie (10 FUEL) → Street (50 FUEL) → Pro (300 FUEL) → Top Fuel (3000 FUEL) with increasing stakes and BURN rewards

**Independent Test**: Provide test account with FUEL, play across leagues, verify buy-in deductions, prize calculations (8% rake), BURN rewards scale with league

### Implementation for User Story 3

**Backend: League Logic**

- [ ] T138 [P] [US3] Define league constants (buy-ins, BURN tables) in backend/internal/modules/gameengine/leagues.go
- [ ] T139 [US3] Implement Rookie race count validation (max 3) in backend/internal/modules/matchmaker/service.go
- [ ] T140 [US3] Increment rookie_races_completed after Rookie settlement in backend/internal/modules/gameengine/settlement.go
- [ ] T141 [US3] Implement BURN reward calculation per league/position in backend/internal/modules/gameengine/burn_rewards.go
- [ ] T142 [US3] Apply BURN ledger entries (operation_type=MATCH_BURN_REWARD) in backend/internal/modules/gameengine/settlement.go
- [ ] T143 [US3] Exclude Rookie from BURN rewards (FR-038) in backend/internal/modules/gameengine/burn_rewards.go

**Backend: Balance Validation**

- [ ] T144 [US3] Validate balance >= league buy-in before joining queue in backend/internal/modules/matchmaker/service.go
- [ ] T145 [US3] Return error INSUFFICIENT_BALANCE if validation fails in backend/internal/modules/matchmaker/service.go

**Frontend: League Restrictions**

- [ ] T146 [P] [US3] Disable Rookie after 3 races with tooltip in frontend/src/components/garage/LeagueCard.tsx
- [ ] T147 [US3] Disable leagues with insufficient balance with "GO TO GAS STATION" CTA in frontend/src/components/garage/LeagueCard.tsx
- [ ] T148 [US3] Display BURN balance in Garage in frontend/src/pages/Garage.tsx
- [ ] T149 [US3] Display BURN reward in settlement (if applicable) in frontend/src/components/settlement/PrizeDisplay.tsx

**Checkpoint**: All 4 leagues functional with correct buy-ins, rake, prizes, and BURN rewards

---

## Phase 7: User Story 4 - Disconnect/Reconnect Handling (Priority: P2)

**Goal**: Players who disconnect mid-race have 3 seconds to reconnect; if successful, gameplay resumes; if failed, heat score = 0 (crashed)

**Independent Test**: Simulate network interruptions, verify "Reconnecting…" overlay, reconnect within 3s resumes gameplay, timeout >3s auto-crashes

### Implementation for User Story 4

**Backend: Connection Tracking**

- [ ] T150 [P] [US4] Track active connections per user in Redis (key: conn:{user_id}) in backend/internal/modules/gateway/connections.go
- [ ] T151 [US4] Implement disconnect detection via Centrifugo webhooks in backend/internal/modules/gateway/webhooks.go
- [ ] T152 [US4] Start 3-second grace period timer on disconnect in backend/internal/modules/gameengine/reconnect.go
- [ ] T153 [US4] Auto-crash current heat if timer expires (score = 0) in backend/internal/modules/gameengine/reconnect.go
- [ ] T154 [US4] Cancel timer if reconnection succeeds within 3s in backend/internal/modules/gameengine/reconnect.go

**Backend: Pre-Race Disconnect**

- [ ] T155 [US4] Detect disconnect during countdown (before heat 1) in backend/internal/modules/gameengine/state.go
- [ ] T156 [US4] Remove player from lobby and refund buy-in in backend/internal/modules/gameengine/state.go

**Frontend: Reconnect Overlay**

- [ ] T157 [P] [US4] Create "Reconnecting…" overlay component in frontend/src/components/race/ReconnectOverlay.tsx
- [ ] T158 [US4] Show overlay on Centrifugo disconnect event in frontend/src/services/centrifugo/client.ts
- [ ] T159 [US4] Hide overlay on successful reconnect in frontend/src/services/centrifugo/client.ts
- [ ] T160 [US4] Sync match state after reconnect (request current heat data) in frontend/src/services/centrifugo/reconnect.ts

**Checkpoint**: Disconnect/reconnect handling complete, fair auto-crash for timeout >3s

---

## Phase 8: User Story 6 - Target Line Mechanic for Chase Tension (Priority: P3)

**Goal**: In Heat 2 and Heat 3, display Target Line at score to beat (Heat 1 winner or current leader total), with visual/audio feedback when crossed

**Independent Test**: Observe Heat 2 and Heat 3, verify Target Line at correct position, fixed at heat start, includes Ghost results, triggers feedback when crossed

### Implementation for User Story 6

**Backend: Target Line Calculation**

- [ ] T161 [P] [US6] Calculate Target Line for Heat 2 (Heat 1 winner score) in backend/internal/modules/gameengine/heat.go
- [ ] T162 [P] [US6] Calculate Target Line for Heat 3 (current leader total) in backend/internal/modules/gameengine/heat.go
- [ ] T163 [US6] Include Target Line in heat_started event payload in backend/internal/modules/gateway/events/match_events.go

**Frontend: Target Line Display**

- [ ] T164 [P] [US6] Create Target Line sprite in frontend/src/pixi/sprites/TargetLine.ts
- [ ] T165 [US6] Position Target Line at correct Speed value in frontend/src/pixi/scenes/RaceScene.ts
- [ ] T166 [US6] Show one-time tooltip "Reach this line to beat the current leader" in frontend/src/components/race/TargetLineTooltip.tsx
- [ ] T167 [US6] Trigger visual pulse and audio cue when Speed crosses Target Line in frontend/src/pixi/scenes/RaceScene.ts
- [ ] T168 [US6] Hide Target Line in Heat 1 in frontend/src/pixi/scenes/RaceScene.ts

**Checkpoint**: Target Line mechanic complete, increases PvP tension in Heats 2 and 3

---

## Phase 9: User Story 7 - Intermission Standings & Position Tracking (Priority: P3)

**Goal**: After each heat, display 5-second intermission with standings table (positions, names, scores, deltas), player row highlighted

**Independent Test**: Complete Heat 1 and Heat 2, verify 5-second timer, standings show all 10 players, position deltas appear, player row highlighted

### Implementation for User Story 7

**Backend: Intermission Data**

- [ ] T169 [P] [US7] Calculate standings after each heat in backend/internal/modules/gameengine/heat.go
- [ ] T170 [US7] Include standings in heat_ended event payload in backend/internal/modules/gateway/events/match_events.go

**Frontend: Enhanced Intermission**

- [ ] T171 [P] [US7] Display player names in standings table in frontend/src/components/race/Intermission.tsx
- [ ] T172 [P] [US7] Display total scores in standings table in frontend/src/components/race/Intermission.tsx
- [ ] T173 [US7] Calculate and display position deltas (↑ ↓ =) in frontend/src/components/race/Intermission.tsx
- [ ] T174 [US7] Highlight player's own row with distinct background in frontend/src/components/race/Intermission.tsx
- [ ] T175 [US7] Auto-scroll to ensure player row is visible in frontend/src/components/race/Intermission.tsx

**Checkpoint**: Intermission standings complete, provides PvP feedback and builds tension

---

## Phase 10: User Story 8 - Early Heat End Optimization (Priority: P3)

**Goal**: If all alive players lock scores before crash/timer expiry, heat ends immediately (saves waiting time)

**Independent Test**: Have all 10 players lock scores early (e.g., t=8s), verify heat ends immediately, displays final Speed at t=8s, transitions to intermission

### Implementation for User Story 8

**Backend: Early Heat End**

- [ ] T176 [P] [US8] Track alive player count per heat in backend/internal/modules/gameengine/heat.go
- [ ] T177 [US8] Check if all alive players locked after each EarnPoints call in backend/internal/modules/gameengine/earn_points.go
- [ ] T178 [US8] End heat immediately if condition met (publish heat_ended early) in backend/internal/modules/gameengine/heat.go
- [ ] T179 [US8] Record actual heat duration in match_participants table in backend/internal/modules/gameengine/heat.go

**Frontend: Early End Handling**

- [ ] T180 [US8] Stop Pixi ticker loop on early heat_ended event in frontend/src/pixi/scenes/RaceScene.ts
- [ ] T181 [US8] Display final Speed value from event payload in frontend/src/pages/Race.tsx

**Checkpoint**: Early heat end optimization complete, reduces unnecessary waiting

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T182 [P] Add error handling for all HTTP endpoints in backend/internal/modules/gateway/http/
- [ ] T183 [P] Add error handling for all RPC handlers in backend/internal/modules/gateway/rpc/
- [ ] T184 [P] Implement unified error overlay component in frontend/src/components/common/ErrorOverlay.tsx
- [ ] T185 [P] Add contextual loading states in frontend/src/components/common/LoadingState.tsx
- [ ] T186 [P] Implement localization (i18n) with 4 languages in frontend/src/locales/ (en, ru, es-PE, pt-BR)
- [ ] T187 [P] Integrate Amplitude analytics SDK in frontend/src/services/analytics/amplitude.ts
- [ ] T188 [P] Add core analytics events (app_opened, match_started, match_finished) in frontend/src/services/analytics/events.ts
- [ ] T189 [P] Add Prometheus metrics for RPC latency in backend/internal/metrics/rpc.go
- [ ] T190 [P] Add Prometheus metrics for matchmaking wait time in backend/internal/metrics/matchmaking.go
- [ ] T191 [P] Add Prometheus metrics for house balance gauge in backend/internal/metrics/economy.go
- [ ] T192 [P] Implement graceful shutdown (SIGTERM, DRAINING state) in backend/cmd/server/main.go
- [ ] T193 [P] Add structured logging with logrus for all module operations in backend/internal/modules/
- [ ] T194 [P] Integrate Sentry for error tracking in backend/cmd/server/main.go
- [ ] T195 [P] Integrate Sentry for error tracking in frontend/src/main.tsx
- [ ] T196 Run validation against quickstart.md scenarios

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-10)**: All depend on Foundational phase completion
  - User Story 1 (Phase 3): Can start after Foundational - Core gameplay loop
  - User Story 5 (Phase 4): Can start after Foundational - Independent TON module (parallel with US1)
  - User Story 2 (Phase 5): Depends on User Story 1 completion (extends matchmaker/gameengine)
  - User Story 3 (Phase 6): Depends on User Story 1 completion (extends economy)
  - User Story 4 (Phase 7): Depends on User Story 1 completion (extends gateway/gameengine)
  - User Story 6 (Phase 8): Depends on User Story 1 completion (extends race HUD)
  - User Story 7 (Phase 9): Depends on User Story 1 completion (extends intermission)
  - User Story 8 (Phase 10): Depends on User Story 1 completion (extends heat lifecycle)
- **Polish (Phase 11)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - Core gameplay loop
- **User Story 5 (P2)**: Can start after Foundational (Phase 2) - Independent TON integration (can run parallel with US1)
- **User Story 2 (P1)**: Depends on User Story 1 - Extends matchmaker/gameengine with Ghost system
- **User Story 3 (P2)**: Depends on User Story 1 - Extends economy with leagues and BURN
- **User Story 4 (P2)**: Depends on User Story 1 - Adds reconnect handling to gameplay
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

## Parallel Example: Foundational Phase

```bash
# Launch all database migrations sequentially (dependencies):
Task: "Create database migrations for Users table"
Task: "Create database migrations for Wallets table"
...

# Launch all infrastructure setup in parallel:
Task: "Create PostgreSQL models in backend/internal/storage/postgres/models/"
Task: "Create Redis client wrapper in backend/internal/storage/redis/client.go"
Task: "Generate Protobuf code from backend/proto/"
Task: "Create Chi router setup in backend/cmd/server/main.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 + User Story 5 + User Story 2)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. **Parallel Development**:
   - Developer A: Complete Phase 3: User Story 1 (core gameplay)
   - Developer B: Complete Phase 4: User Story 5 (TON integration)
4. Complete Phase 5: User Story 2 (Ghost system)
5. **STOP and VALIDATE**: Test MVP independently with Rookie league + TON deposits/withdrawals
6. Deploy/demo if ready

**MVP delivers**: Complete gameplay loop + Ghost players + Real money deposits/withdrawals

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 + 5 (parallel) → Test independently → Deploy/Demo (Core + Money!)
3. Add User Story 2 → Test Ghost system → Deploy/Demo (MVP Complete!)
4. Add User Story 3 → Test multi-league economy → Deploy/Demo
5. Add User Story 4 → Test reconnect handling → Deploy/Demo
6. Add User Stories 6-8 → Polish UX → Deploy/Demo
7. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - **Developer A**: User Story 1 (core gameplay)
   - **Developer B**: User Story 5 (TON integration - independent)
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
- **Logrus used directly** - No internal logger package wrapper
- **Telegram Mini App** - Frontend initialized as Telegram Mini App with SDK integration in foundational phase
- Commit after each task or logical group (one commit, one task per Constitution II)
- Stop at any checkpoint to validate story independently
- **Avoid**: vague tasks, same file conflicts, cross-story dependencies that break independence
- **MVP scope**: User Story 1 + User Story 5 + User Story 2 (P1 stories + early TON) deliver complete testable product
- **Full MVP**: All P1 + P2 stories (US1-US5) required for launch
- **P3 stories**: Nice-to-have UX enhancements, can defer post-launch

---

## Task Summary

- **Total Tasks**: 196
- **Phase 1 (Setup)**: 10 tasks
- **Phase 2 (Foundational)**: 14 tasks (includes single migration file + Telegram Mini App SDK)
- **Phase 3 (US1 - Core Gameplay)**: 75 tasks
- **Phase 4 (US5 - TON Integration)**: 22 tasks ⚡ MOVED EARLY
- **Phase 5 (US2 - Ghost System)**: 16 tasks
- **Phase 6 (US3 - Leagues)**: 12 tasks
- **Phase 7 (US4 - Reconnect)**: 11 tasks
- **Phase 8 (US6 - Target Line)**: 8 tasks
- **Phase 9 (US7 - Intermission)**: 7 tasks
- **Phase 10 (US8 - Early End)**: 6 tasks
- **Phase 11 (Polish)**: 15 tasks

**Early MVP** (US1 + US5 + US2): **121 tasks** ⚡ TON testing enabled early
**Full Launch** (US1-US5): **160 tasks**
**Complete Feature** (All stories): **196 tasks**
