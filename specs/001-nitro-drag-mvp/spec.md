# Feature Specification: Nitro Drag Royale MVP

**Feature Branch**: `001-nitro-drag-mvp`  
**Created**: 2026-01-28  
**Status**: Draft  
**Input**: NITRO DRAG ROYALE — Complete MVP Game Design Document for a Telegram Mini App PvP crash tournament racing game

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Enter and Complete First Race (Priority: P1)

A new player opens Nitro Drag Royale in Telegram, joins their first Rookie league race, experiences all 3 heats with the speed acceleration mechanic, and receives their winnings or BURN rewards based on final position.

**Why this priority**: This is the absolute core experience — the complete gameplay loop from entry to exit. Without this, there is no product. It validates the entire crash tournament mechanic, PvP tension, and reward distribution.

**Independent Test**: Can be fully tested by onboarding a new user, providing them with initial FUEL, and observing them complete one full race cycle. Delivers the core value proposition: "high-stakes drag racing where timing beats greed."

**Acceptance Scenarios**:

1. **Given** player has 10+ FUEL and is in Garage, **When** they tap "RACE NOW" on Rookie league, **Then** they enter matchmaking with status "FINDING RACERS" and a countdown timer
2. **Given** matchmaking succeeds within 20 seconds, **When** match is formed, **Then** buy-in (10 FUEL) is deducted and player sees Race HUD with countdown "3…2…1"
3. **Given** Heat 1 starts, **When** time progresses, **Then** Speed increases from 0 following deterministic formula: `Speed = 500 * ((e^(0.08·t) - 1) / (e^(0.08·25) - 1))`
4. **Given** player sees Speed = 250, **When** they tap "EARN POINTS", **Then** their Heat 1 score is locked at 250 and they enter spectator state
5. **Given** all players finish Heat 1, **When** intermission displays, **Then** player sees standings table with positions, names, scores, and position delta indicators for 5 seconds
6. **Given** Heat 2 starts, **When** race begins, **Then** Target Line appears at position equal to Heat 1 winner's score with one-time tooltip "Reach this line to beat the current leader"
7. **Given** player completes all 3 heats with total score of 780, **When** settlement displays, **Then** player sees their final position (e.g., 6th), prize pool breakdown, FUEL change, and BURN reward if applicable
8. **Given** settlement is displayed, **When** player taps "RACE AGAIN", **Then** they return to matchmaking for same league (if sufficient balance) or Garage if insufficient
9. **Given** two players finish with identical total scores (e.g., both 780), **When** settlement calculates positions, **Then** player with higher Heat 3 score ranks higher (if Heat 3 tied, use Heat 2, then Heat 1)
10. **Given** player finishes 6th place in Rookie league, **When** settlement displays, **Then** BURN reward shown = 0 (Rookie league does not grant BURN per FR-038)

---

### User Story 2 - Ghost Player System & Matchmaking Speed (Priority: P1)

A player enters matchmaking at an off-peak time when fewer than 10 live players are available. The system fills remaining slots with Ghost players (historical replays), ensuring match starts within 20 seconds maximum.

**Why this priority**: Fast matchmaking is Design Pillar #1 and critical to user retention. Without Ghost players, the product cannot deliver "no waiting, no dead time" promise, especially during MVP with limited player base.

**Independent Test**: Can be tested by simulating low player availability and verifying: (1) lobby fills with Ghosts within 20s, (2) Ghosts behave identically to live players during race, (3) Ghosts can win prizes, (4) buy-ins from Ghosts come from house balance, (5) Ghosts excluded from leaderboards.

**Acceptance Scenarios**:

1. **Given** only 3 live players available, **When** matchmaking runs, **Then** system fills 7 slots with Ghost players selected from historical replays
2. **Given** lobby has 3 live + 7 Ghost players, **When** race starts, **Then** all 10 players are visually indistinguishable and participate fully
3. **Given** Ghost player finishes 1st place, **When** settlement displays, **Then** Ghost receives 50% of prize pool from house balance
4. **Given** Ghost player participates in match, **When** leaderboards update, **Then** Ghost's results are excluded from all leaderboard statistics
5. **Given** matchmaking timer reaches 20 seconds with only 1 live player found, **When** match is formed, **Then** system fills remaining 9 slots with Ghost players and match starts normally

---

### User Story 3 - Multi-League Progression & Economy (Priority: P2)

A player progresses from Rookie (10 FUEL buy-in, first 3 races only, no BURN) through Street (50 FUEL, BURN rewards start) to Pro (300 FUEL) and Top Fuel (3000 FUEL), experiencing increasing stakes and BURN reward tiers.

**Why this priority**: League progression creates retention and monetization structure. While single-league gameplay is testable independently, the full economy (FUEL + BURN) and progression incentives require multi-league implementation.

**Independent Test**: Can be tested by providing a test account with sufficient FUEL to play across leagues and verifying: (1) Rookie restrictions (3 races max, no BURN), (2) correct buy-in deductions per league, (3) prize pool calculations (8% rake), (4) BURN rewards scale with league and position.

**Acceptance Scenarios**:

1. **Given** new player completes 3 Rookie races, **When** they return to Garage, **Then** Rookie league is visually disabled with tooltip "Rookie league limited to first 3 races"
2. **Given** player has 200 FUEL, **When** they select Street league (50 FUEL buy-in), **Then** they can enter matchmaking and play normally
3. **Given** player finishes 6th in Street league, **When** settlement displays, **Then** they receive 15 BURN (per BURN rewards table)
4. **Given** 10 players in Top Fuel league (3000 FUEL each), **When** match completes, **Then** total pool = 30,000 FUEL, rake = 2,400 FUEL, remaining = 27,600 FUEL distributed 50%/30%/20% to top 3
5. **Given** player has 40 FUEL balance, **When** they tap Pro league (300 FUEL buy-in), **Then** league is disabled with tooltip "Insufficient FUEL" and CTA "GO TO GAS STATION"

---

### User Story 4 - Disconnect/Reconnect Handling (Priority: P2)

During an active heat, a player's connection drops. They have up to 3 seconds to reconnect. If reconnection succeeds, gameplay resumes. If it fails, their heat score = 0 (crashed).

**Why this priority**: Network reliability is essential for fairness and trust (Design Pillar #3). While not strictly required for basic gameplay, handling disconnects gracefully prevents user frustration and ensures economy safety.

**Independent Test**: Can be tested by simulating network interruptions at various race stages and verifying: (1) inline "Reconnecting…" overlay appears, (2) successful reconnect within 3s resumes gameplay, (3) disconnect >3s auto-crashes with score = 0, (4) no buy-in refunds for mid-race disconnects.

**Acceptance Scenarios**:

1. **Given** player is in Heat 2 with Speed = 180, **When** network disconnects, **Then** inline overlay displays "Reconnecting…" and game state freezes
2. **Given** "Reconnecting…" overlay is active for 2 seconds, **When** connection restores, **Then** overlay disappears and gameplay resumes from last server-synced state
3. **Given** "Reconnecting…" overlay is active for 4 seconds, **When** 3-second window expires, **Then** player's heat score = 0 (treated as crash) and they enter spectator state
4. **Given** player disconnects before race starts (during countdown), **When** disconnect persists, **Then** they are removed from lobby and buy-in is refunded
5. **Given** player crashes due to disconnect timeout, **When** settlement displays, **Then** their crashed heats show score = 0 and final position reflects this

---

### User Story 5 - TON Wallet Integration & Gas Station (Priority: P2)

A player runs low on FUEL. They navigate to "Gas Station," connect their TON wallet via TON Connect, deposit TON, and receive FUEL at the configured exchange rate. They can also withdraw FUEL back to TON (excluding BURN, which is non-convertible).

**Why this priority**: Without deposits/withdrawals, the economy is closed-loop. This unlocks real monetary stakes and player acquisition. It's P2 (not P1) because core gameplay can be tested with pre-funded accounts, but it's essential for MVP launch.

**Independent Test**: Can be tested by: (1) connecting TON wallet via TON Connect UI, (2) depositing TON and verifying FUEL credit, (3) attempting withdrawal and verifying TON transfer, (4) attempting to convert BURN to TON (should fail with error "BURN is non-convertible").

**Acceptance Scenarios**:

1. **Given** player taps "GO TO GAS STATION", **When** Gas Station screen opens, **Then** they see current FUEL balance, TON wallet connection status, and "DEPOSIT" / "WITHDRAW" options
2. **Given** player has no connected wallet, **When** they tap "DEPOSIT", **Then** TON Connect modal appears requesting wallet connection
3. **Given** player connects wallet successfully, **When** they initiate deposit of 10 TON, **Then** transaction is processed and FUEL balance increases by equivalent amount
4. **Given** player has 500 FUEL and 50 BURN, **When** they attempt withdrawal, **Then** only FUEL is convertible to TON (BURN remains in account)
5. **Given** transaction fails (network error, insufficient gas, user cancels), **When** error occurs, **Then** unified error overlay displays contextual message and no balances change

---

### User Story 6 - Target Line Mechanic for Chase Tension (Priority: P3)

In Heat 2 and Heat 3, players see a visual Target Line on their HUD representing the score they need to beat (Heat 1 winner for Heat 2, current total leader for Heat 3). Crossing the line triggers brief visual/audio feedback.

**Why this priority**: Target Line reduces perceived randomness and increases PvP tension (Design Pillars #2 and #3), but core gameplay functions without it. It's P3 because it's a UX enhancement rather than functional requirement.

**Independent Test**: Can be tested by observing Heat 2 and Heat 3 races and verifying: (1) Target Line appears at correct position, (2) is fixed at heat start, (3) includes Ghost results, (4) triggers feedback when crossed, (5) tooltip appears only once on first exposure.

**Acceptance Scenarios**:

1. **Given** Heat 1 winner scored 320, **When** Heat 2 starts, **Then** Target Line appears at Speed = 320 with one-time tooltip "Reach this line to beat the current leader"
2. **Given** player is in Heat 2 with Speed = 300, **When** Speed crosses 320 (Target Line), **Then** brief visual pulse and audio cue play
3. **Given** after Heat 2, leader has total score 650, **When** Heat 3 starts, **Then** Target Line updates to represent 650 as the target
4. **Given** Target Line is set at heat start, **When** other players finish during heat, **Then** Target Line does NOT update mid-heat (remains fixed)
5. **Given** Ghost player has highest Heat 1 score, **When** Heat 2 starts, **Then** Target Line uses Ghost's score (not excluded)

---

### User Story 7 - Intermission Standings & Position Tracking (Priority: P3)

After each heat completes, all players see a 5-second intermission overlay showing current standings table with positions, player names, total scores, and position delta indicators (↑ ↓ =). Player's own row is highlighted.

**Why this priority**: Intermission provides critical PvP feedback and builds tension before next heat. It's P3 because races can function without it, but it significantly improves perceived fairness and competitive engagement.

**Independent Test**: Can be tested by completing Heat 1 and Heat 2, verifying: (1) 5-second timer displays, (2) standings show all 10 players ranked by total score, (3) position deltas appear (↑ if improved, ↓ if dropped, = if unchanged), (4) player's row is visually highlighted, (5) auto-transitions to next heat countdown.

**Acceptance Scenarios**:

1. **Given** Heat 1 completes, **When** intermission displays, **Then** standings table shows all 10 players ranked by Heat 1 scores with no position deltas (first heat baseline)
2. **Given** player was 8th after Heat 1, scored 400 in Heat 2, and moved to 5th, **When** Heat 2 intermission displays, **Then** player's row shows "5th" with "↑" indicator and is highlighted
3. **Given** intermission timer shows 5 seconds, **When** countdown completes, **Then** screen auto-transitions to next heat countdown "3…2…1"
4. **Given** intermission is displaying, **When** player taps screen, **Then** no action occurs (input is blocked during intermission)
5. **Given** all 10 players visible in standings, **When** standings table renders, **Then** player's own name and row are always visible and highlighted (auto-scroll if needed)

**See also**: Functional requirements FR-056 through FR-060 for detailed intermission specifications.

---

### User Story 8 - Early Heat End Optimization (Priority: P3)

If all alive players lock their scores before the engine crashes or timer expires, the heat ends immediately rather than waiting for full 25-second duration. This shows final Speed value and transitions to intermission.

**Why this priority**: Reduces unnecessary waiting when all decisions are made, supporting Design Pillar #1 (Fast Sessions). It's P3 because it's an optimization, not a functional requirement — races work fine without it.

**Independent Test**: Can be tested by having all 10 players lock scores early (e.g., at t=8s) and verifying: (1) heat ends immediately, (2) final Speed at t=8s is displayed, (3) transition to intermission occurs without waiting for t=25s, (4) all scores are correctly recorded.

**Acceptance Scenarios**:

1. **Given** all 10 players lock scores by t=8 seconds, **When** last player locks, **Then** heat ends immediately and displays final Speed = [calculated value at t=8]
2. **Given** 9 players locked, 1 player waiting, **When** waiting player's engine crashes at t=12, **Then** heat ends immediately after crash
3. **Given** heat ends early via all-finished rule, **When** transition occurs, **Then** intermission displays with all 10 scores correctly recorded
4. **Given** heat timer at t=8s with all finished, **When** early end triggers, **Then** players do NOT wait additional 17 seconds (optimization saves time)
5. **Given** heat ends at t=25s (timer expiry), **When** not all players finished, **Then** remaining players who didn't lock are treated as crashed (score = 0)

---

### Edge Cases

- **What happens when Ghost player wins 1st place?**  
  Ghost receives 50% of prize pool from house balance. Live players receive remaining prizes. Short-term platform PnL variance is accepted by design.

- **What happens when matchmaking times out after 20 seconds with only 1 live player?**  
  Match starts normally with 1 live player + 9 Ghosts. Buy-in is deducted and match proceeds as usual.

- **What happens when player disconnects before race starts (during countdown)?**  
  Player is removed from lobby and buy-in is refunded. Match proceeds with remaining players + Ghosts to fill slots.

- **What happens when player disconnects mid-race and reconnects after 3-second window?**  
  Player's current heat score = 0 (treated as crash). They can spectate remaining heats but cannot earn points. Previous heat scores remain valid.

- **What happens when transaction fails (deposit/withdrawal) due to network error or insufficient gas?**  
  Unified error overlay displays contextual message. No balances change. Buy-in is never deducted if error occurs before race start.

- **What happens when player taps "EARN POINTS" at Speed = 499?**  
  Heat score is locked at 499. Player enters spectator state and observes remaining racers.

- **What happens when engine crashes at Speed = 0 (immediate crash)?**  
  Heat score = 0. Player enters spectator state immediately.

- **What happens when player completes 3 Rookie races?**  
  Rookie league becomes visually disabled. Tooltip explains "Rookie league limited to first 3 races. Try Street league!" No BURN rewards were granted during Rookie races.

- **What happens when player has insufficient FUEL for selected league?**  
  League is disabled/tappable. Tap shows tooltip "Insufficient FUEL. Current balance: X FUEL. Required: Y FUEL" with CTA "GO TO GAS STATION."

- **What happens when all 10 players in lobby are Ghosts (0 live players)?**  
  This should not occur as matchmaking requires at least 1 live player to initiate. A match cannot start with 0 live players.

- **What happens when player's total score across 3 heats = 0 (crashed all heats)?**  
  Player finishes in last place (10th). Settlement shows position 10th, prize = 0 FUEL, BURN reward = [maximum for league, e.g., 25 BURN for Street].

- **What happens when two players have identical total scores?**  
  Tiebreaker rule: Higher Heat 3 score wins. If Heat 3 scores also tied, use Heat 2 score. If all heats tied (extremely rare), use Heat 1 score. This rewards strong final-heat performance and creates dramatic tension.

- **What happens when Target Line position equals exact Speed player locked at?**  
  Player successfully "reached the line to beat the current leader" — visual/audio feedback triggers. Player is tied or ahead depending on other heats.

- **What happens when intermission displays but player force-closes app?**  
  Player is reconnected to race state when they reopen. If more than 3 seconds passed during next heat, they auto-crash for that heat.

- **What happens when settlement displays and player has 0 FUEL remaining?**  
  "RACE AGAIN" button is disabled with tooltip "Insufficient FUEL for this league. Go to Gas Station or try a lower league."

## Requirements *(mandatory)*

### Technical Constraints

- **TC-001**: Application MUST be built using Telegram Mini Apps SDK (official) for native integration and TON Connect support
- **TC-002**: Frontend MUST leverage Telegram WebApp API for UI components and platform features (BackButton for navigation, MainButton for primary CTAs, HapticFeedback for interactions, ClosingConfirmation for active match warnings)
- **TC-003**: Backend MUST use centralized Golang server for core game logic and economy safety
- **TC-004**: Real-time race state synchronization MUST use Centrifugo for WebSocket/SSE-based pub-sub messaging
- **TC-005**: Persistent storage MUST use PostgreSQL for transactional data (player accounts, match history, economy transactions)
- **TC-006**: Volatile storage MUST use Redis for real-time state (active races, matchmaking queues, session data)

### Functional Requirements

**Core Gameplay**

- **FR-001**: System MUST support matches of exactly 10 players (live + Ghost combined)
- **FR-002**: System MUST allow matches with at least 1 live player; match starts even with single live player plus 9 Ghosts
- **FR-003**: System MUST fill remaining slots (when live players <10) with Ghost players selected from historical real player replays
- **FR-004**: Each match MUST consist of exactly 3 Heats (Heat 1 Baseline, Heat 2 Chase, Heat 3 Final)
- **FR-005**: Speed MUST increase deterministically following formula: `Speed = 500 * ((e^(0.08·t) - 1) / (e^(0.08·25) - 1))` where t ∈ [0, 25] seconds, rounded to 2 decimal places
- **FR-006**: Speed MUST be clamped to maximum 500.00 and MUST NOT grow beyond t=25 seconds
- **FR-007**: Player MUST be able to press "EARN POINTS" at any moment during active heat to lock current Speed as heat score
- **FR-008**: If player does not press "EARN POINTS" before engine crashes, heat score MUST be 0
- **FR-009**: Player total score MUST equal sum of all 3 heat scores
- **FR-010**: System MUST rank players 1-10 by total score at match end
- **FR-010a**: When two or more players have identical total scores, system MUST use Heat 3 score as first tiebreaker (higher wins), then Heat 2 score if needed, then Heat 1 score if needed (calculated at runtime in settlement logic, no database storage required)
- **FR-011**: Crash timing and outcomes MUST be consistent and verifiable across all players (no client manipulation possible)
- **FR-011a**: Server MUST generate crash point for each heat using cryptographic seed (pre-committed hash) before heat starts
- **FR-011b**: Crash point MUST be verifiable by players after heat completion (seed + hash reveal for provable fairness)
- **FR-011c**: Crash timing MUST follow uniform random distribution where t ∈ [0, 25] seconds has equal probability

**Target Line Mechanic**

- **FR-012**: Target Line MUST be disabled in Heat 1
- **FR-013**: In Heat 2, Target Line MUST be set to Heat 1 winner's score (including Ghosts)
- **FR-014**: In Heat 3, Target Line MUST be set to current leader's total score (Heat 1 + Heat 2, including Ghosts)
- **FR-015**: Target Line MUST be fixed at heat start and MUST NOT update mid-heat
- **FR-016**: When player crosses Target Line, system MUST display brief visual and audio feedback
- **FR-017**: On first Target Line appearance for a player, system MUST display one-time tooltip: "Reach this line to beat the current leader" (stored in frontend session storage, resets on logout)

**Economy — Currencies**

- **FR-018**: System MUST support three currencies: TON (external), FUEL (hard currency), BURN (meta currency)
- **FR-019**: FUEL MUST be used for match buy-ins and prize payouts
- **FR-020**: BURN MUST be granted based on final match position and league tier (non-convertible to TON)
- **FR-021**: TON deposits MUST credit player's FUEL balance at configured exchange rate (default: 1 TON = 100 FUEL, configurable via TON_FUEL_EXCHANGE_RATE environment variable)
- **FR-022**: TON withdrawals MUST debit player's FUEL balance at same exchange rate (BURN excluded, non-convertible)

**Economy — Leagues & Buy-ins**

- **FR-023**: System MUST support four leagues: Rookie (10 FUEL), Street (50 FUEL), Pro (300 FUEL), Top Fuel (3000 FUEL)
- **FR-024**: Rookie league MUST be available only for first 3 races per player
- **FR-025**: Rookie league MUST NOT grant BURN rewards
- **FR-026**: System MUST enforce buy-in deduction only when match countdown starts (match successfully formed and confirmed, not during matchmaking phase)
- **FR-027**: If player balance < league buy-in, league MUST be visually disabled with explanatory tooltip and CTA "GO TO GAS STATION"

**Economy — Prize Pool & Rake**

- **FR-028**: All players (live + Ghost) MUST pay buy-in into prize pool
- **FR-029**: System MUST apply 8% rake to total prize pool
- **FR-030**: Remaining prize pool MUST be distributed: 1st place = 50%, 2nd place = 30%, 3rd place = 20%
- **FR-031**: Players finishing 4th-10th MUST receive 0 FUEL prizes
- **FR-032**: Ghost buy-ins MUST be paid from house balance; Ghost prizes MUST be paid to house balance

**Economy — BURN Rewards**

- **FR-033**: Players finishing 1st-3rd MUST receive 0 BURN
- **FR-034**: BURN rewards MUST be granted once per match based on final position and league tier
- **FR-035**: Street league BURN: 4th-5th = 8 BURN, 6th-7th = 15 BURN, 8th-10th = 25 BURN (max 25)
- **FR-036**: Pro league BURN: 4th-5th = 20 BURN, 6th-7th = 40 BURN, 8th-10th = 70 BURN (max 70)
- **FR-037**: Top Fuel league BURN: 4th-5th = 40 BURN, 6th-7th = 70 BURN, 8th-10th = 100 BURN (hard cap 100)
- **FR-038**: Rookie league MUST NOT grant BURN rewards regardless of position

**Ghost Players**

- **FR-039**: Ghost players MUST be replays of real historical player races, selected using stratified sampling by score percentiles to ensure diverse performance distribution
- **FR-040**: Ghost players MUST be visually indistinguishable from live players during race
- **FR-041**: Ghost players MUST participate fully in prize pool (pay buy-in, can win prizes)
- **FR-042**: Ghost players CAN finish in any position including 1st place
- **FR-043**: Ghost players MUST be excluded from leaderboards and social statistics
- **FR-044**: Ghost buy-ins and prizes MUST transact with house balance, not player balances

**Matchmaking**

- **FR-045**: Matchmaking MUST display dedicated screen with status "FINDING RACERS" and progress indicator/timer
- **FR-046**: Matchmaking MUST have maximum duration of 20 seconds from when backend receives join request (server-side timer)
- **FR-047**: Player MUST be able to press "CANCEL" to abort matchmaking and return to Garage with no buy-in deduction
- **FR-048**: If matchmaking times out (20s) or is cancelled, player MUST return to Garage with no buy-in deduction
- **FR-049**: When match is successfully formed, player MUST transition directly to Race HUD with countdown "3…2…1"

**Race HUD States**

- **FR-051**: Pre-heat state MUST display countdown overlay, Speed = 0, controls disabled, opponents visible but idle
- **FR-052**: Active heat state MUST show Speed growing continuously, "EARN POINTS" button available
- **FR-053**: Post-finish (spectator) state MUST show "YOUR HEAT SCORE: X", "CURRENT POSITION: N / Y" (updates live), Speed frozen, controls disabled
- **FR-054**: If all alive players lock score before engine crashes, heat MUST end immediately (early heat end rule). "Alive players" = players who have not yet locked their score (excluding spectators who already locked or crashed)
- **FR-055**: When early heat end triggers, system MUST display final Speed value and transition to intermission without waiting for full 25-second duration

**Intermission**

- **FR-056**: Intermission MUST last exactly 5 seconds
- **FR-057**: Intermission MUST display standings table with: position, player name, total score, position delta indicators (↑ ↓ =)
- **FR-058**: Player's own row MUST be visually highlighted and always visible
- **FR-059**: Intermission MUST block all player input (tap/gesture has no effect)
- **FR-060**: After 5 seconds, system MUST auto-transition to next heat countdown (or settlement if Heat 3 completed)

**Settlement**

- **FR-061**: Settlement MUST display top 3 players in separate podium block
- **FR-062**: Settlement MUST display remaining players (4th-10th) in ranked list below podium
- **FR-063**: Player's own row MUST be visible and highlighted regardless of position
- **FR-064**: Settlement MUST show FUEL win/loss clearly (e.g., "+450 FUEL" or "-300 FUEL")
- **FR-065**: Settlement MUST show BURN reward inline when applicable (e.g., "+15 BURN")
- **FR-066**: Primary CTA MUST be "RACE AGAIN" (same league); secondary CTA MUST be "BACK TO GARAGE"
- **FR-067**: If player balance < league buy-in, "RACE AGAIN" MUST be disabled with explanatory message

**Connection & Reconnect**

- **FR-068**: On disconnect during active race, system MUST display inline overlay "Reconnecting…"
- **FR-069**: Player MUST have up to 3 seconds to reconnect (grace period starts when Centrifugo disconnect event is received by backend)
- **FR-070**: If reconnection succeeds within 3 seconds, gameplay MUST resume from last server-synced state
- **FR-071**: If disconnect persists >3 seconds, player's current heat score MUST be 0 (auto-crash) and player enters spectator state
- **FR-072**: If disconnect occurs before race starts (during countdown), player MUST be removed from lobby and buy-in MUST be refunded

**Error Handling**

- **FR-073**: System MUST display unified error overlay with contextual copy for all error scenarios
- **FR-074**: Error scenarios MUST include: matchmaking failure, network issues, server errors, transaction failures
- **FR-075**: If error occurs before race start, buy-in MUST never be deducted (economy safety guarantee)

**Loading States**

- **FR-076**: System MUST use contextual loaders instead of generic spinners (e.g., "Finding racers…", "Preparing race…", "Calculating results…")
- **FR-077**: Loaders MUST always block user input to prevent race conditions

**Garage**

- **FR-078**: Garage MUST display: league selector, current FUEL balance, "RACE NOW" button
- **FR-079**: Unavailable leagues MUST be visually disabled but tappable
- **FR-080**: Tapping unavailable league MUST show tooltip/modal explaining restriction with CTA "GO TO GAS STATION"

**TON Wallet Integration**

- **FR-081**: System MUST enable secure wallet connection with user approval flow for TON deposits and withdrawals
- **FR-082**: Player MUST be able to deposit TON and receive equivalent FUEL
- **FR-083**: Player MUST be able to withdraw FUEL and receive equivalent TON
- **FR-084**: BURN currency MUST NOT be convertible to TON or FUEL
- **FR-085**: If transaction fails (network error, insufficient gas, user cancels), system MUST display error message and preserve all balances unchanged

### Key Entities

- **Player**: Represents a live user account with FUEL balance, BURN balance, TON wallet connection status, race history (count, Rookie races completed), current league access
- **Ghost**: Represents a historical replay of a real player's race performance (timing of score locks per heat, final scores, behavioral data) used to fill lobby slots
- **Match**: Represents a single 10-player race session with match ID, league tier, lobby composition (live vs Ghost count), prize pool size, rake amount, timestamp
- **Heat**: Represents one of 3 rounds within a match with heat number (1-3), Target Line value (if applicable), per-player heat scores, crash events, duration
- **League**: Represents a tier of competition with name (Rookie/Street/Pro/Top Fuel), buy-in amount, BURN reward table, access restrictions (Rookie = first 3 races only)
- **Prize Pool**: Represents the economic distribution of a match with total buy-ins, rake (8%), 1st place prize (50% of remaining), 2nd place prize (30%), 3rd place prize (20%)
- **Transaction**: Represents a FUEL or TON currency movement with transaction type (buy-in, prize, deposit, withdrawal), amount, timestamp, status (pending/confirmed/failed)
- **Standings**: Represents the intermediate ranking state after each heat with player position (1-10), total score (sum of completed heats), position delta vs previous heat (↑ ↓ =)

## Clarifications

### Session 2026-01-28

- Q: Tech stack for Telegram Mini App implementation? → A: Telegram Mini Apps SDK (official)
- Q: Backend architecture for race state synchronization and matchmaking? → A: Centralized Golang server with Centrifugo
- Q: Database and storage solution for player state, match history, and Ghost replays? → A: PostgreSQL + Redis
- Q: How is crash timing determined to ensure fairness and verifiability? → A: Server generates crash point using cryptographic seed (pre-committed hash)
- Q: What probability distribution should determine when crashes occur within the 25-second window? → A: Uniform random distribution

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: New players can complete their first full race (Garage → Matchmaking → 3 Heats → Settlement) in under 3 minutes
- **SC-002**: Matchmaking forms matches within 20 seconds for 95% of attempts (accounting for Ghost player fill)
- **SC-003**: Players reconnecting within 3 seconds of disconnect resume gameplay without losing current heat progress
- **SC-004**: Ghost players successfully fill lobbies when live player count <10, maintaining full 10-player matches 100% of the time (when ≥2 live players available)
- **SC-005**: Prize pool mathematics are correct 100% of the time: 8% rake applied, remaining distributed 50%/30%/20% to top 3, Ghost prizes transact with house balance
- **SC-006**: BURN rewards are granted correctly 100% of the time per league tier and final position (positions 1-3 = 0 BURN, positions 4-10 = per table)
- **SC-007**: Economy safety guarantee holds 100% of the time: buy-ins never deducted if error occurs before race start or matchmaking cancelled
- **SC-008**: Speed acceleration follows deterministic formula with <0.1% variance from expected value at any given timestamp
- **SC-009**: Target Line appears at correct position (Heat 1 winner score or current leader total) 100% of the time in Heats 2 and 3
- **SC-010**: Early heat end optimization triggers when all players finish, reducing average heat duration by 20% when applicable
- **SC-011**: Players successfully deposit TON and receive FUEL credit within 30 seconds of transaction confirmation
- **SC-012**: Players successfully withdraw FUEL and receive TON within 30 seconds of transaction confirmation
- **SC-013**: Rookie league restrictions enforce correctly: available for first 3 races only, no BURN rewards, disabled after 3 completions
- **SC-014**: 90% of players understand crash tournament risk/reward mechanic after first race (measured via retention to race #2)
- **SC-015**: Players perceive Ghost players as indistinguishable from live players during races (measured via post-race surveys showing <5% detection rate)
- **SC-016**: System handles 100 concurrent matches (1,000 concurrent live players) without performance degradation (defined as: RPC latency <200ms at 95th percentile, matchmaking <20s, heat state broadcast <100ms)
- **SC-017**: 95% of disconnects <3 seconds result in successful reconnection and gameplay resumption
- **SC-018**: Zero instances of buy-in deduction when matchmaking times out or is cancelled by player
- **SC-019**: Settlement screen accurately reflects all prize and BURN rewards with zero discrepancies between displayed and credited amounts
- **SC-020**: Inter-heat intermission displays for exactly 5 seconds and shows correct standings with position deltas 100% of the time
