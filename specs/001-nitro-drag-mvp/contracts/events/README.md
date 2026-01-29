# Event Contracts (Centrifugo Publications)

## Overview

These JSON schemas define the **events** that the backend publishes to Centrifugo channels. Clients subscribe to channels and receive events pushed by the server.

**Flow**:
```
Go Backend → Centrifugo gRPC Publish API → Centrifugo → Client (WebSocket)
```

## Channel Types

### 1. Personal Channel: `user:{user_id}`

User-specific events (balances, payments).

**Subscription**: Always allowed for authenticated user.

**Events**:
- `balance_updated` — FUEL/BURN balance changed
- `payment_status_changed` — Deposit/withdrawal status updated

---

### 2. Match Channel: `match:{match_id}`

Match lifecycle and heat state events.

**Subscription**: Allowed only if user is participant of the match (checked via subscription proxy).

**Events**:
- `match_found` — Match formed, countdown starting
- `heat_starting` — Heat countdown (3…2…1)
- `heat_started` — Heat in progress (Speed growing)
- `player_locked_score` — Another player locked their score
- `heat_ended` — Heat completed (all players finished or timer expired)
- `intermission` — Standings display between heats
- `match_settled` — Final settlement results

---

## Event Schemas

### balance_updated

**Channel**: `user:{user_id}`

```json
{
  "type": "balance_updated",
  "payload": {
    "fuel_balance": "1234.56",
    "burn_balance": "42.00",
    "fuel_delta": "+50.00",
    "burn_delta": "+8.00",
    "reason": "MATCH_PRIZE"
  }
}
```

**Fields**:
- `fuel_balance` — New FUEL balance (decimal string)
- `burn_balance` — New BURN balance (decimal string)
- `fuel_delta` — Change in FUEL (with sign, e.g., `"+50.00"` or `"-10.00"`)
- `burn_delta` — Change in BURN
- `reason` — Operation type (e.g., `MATCH_PRIZE`, `MATCH_BUYIN`, `DEPOSIT`, etc.)

---

### payment_status_changed

**Channel**: `user:{user_id}`

```json
{
  "type": "payment_status_changed",
  "payload": {
    "payment_id": "uuid",
    "payment_type": "DEPOSIT",
    "status": "CONFIRMED",
    "ton_amount": "10.00",
    "fuel_amount": "10.00"
  }
}
```

**Fields**:
- `payment_id` — Payment identifier
- `payment_type` — `DEPOSIT` or `WITHDRAWAL`
- `status` — `PENDING`, `CONFIRMED`, or `FAILED`
- `ton_amount` — TON amount (decimal string)
- `fuel_amount` — FUEL amount (decimal string)

---

### match_found

**Channel**: `match:{match_id}`

```json
{
  "type": "match_found",
  "payload": {
    "match_id": "uuid",
    "league": "STREET",
    "live_player_count": 7,
    "ghost_player_count": 3,
    "countdown_seconds": 3
  }
}
```

**Fields**:
- `match_id` — Match identifier
- `league` — League tier (`ROOKIE`, `STREET`, `PRO`, `TOP_FUEL`)
- `live_player_count` — Number of live players
- `ghost_player_count` — Number of Ghost players
- `countdown_seconds` — Countdown duration before Heat 1

---

### heat_starting

**Channel**: `match:{match_id}`

```json
{
  "type": "heat_starting",
  "payload": {
    "heat_number": 1,
    "countdown": 3
  }
}
```

**Fields**:
- `heat_number` — Heat number (1, 2, or 3)
- `countdown` — Countdown seconds (3…2…1)

---

### heat_started

**Channel**: `match:{match_id}`

```json
{
  "type": "heat_started",
  "payload": {
    "heat_number": 1,
    "target_speed": null,
    "started_at": "2026-01-28T12:34:56Z"
  }
}
```

**Fields**:
- `heat_number` — Heat number (1, 2, or 3)
- `target_speed` — Target Line value (decimal string, `null` for Heat 1)
- `started_at` — Heat start timestamp (ISO 8601)

---

### player_locked_score

**Channel**: `match:{match_id}`

```json
{
  "type": "player_locked_score",
  "payload": {
    "heat_number": 2,
    "player_id": "uuid",
    "player_name": "Alice",
    "heat_score": "245.50",
    "remaining_alive": 6
  }
}
```

**Fields**:
- `heat_number` — Current heat
- `player_id` — Player who locked score (UUID or `null` for Ghost)
- `player_name` — Player display name
- `heat_score` — Score locked (decimal string)
- `remaining_alive` — Players still racing (not finished)

---

### heat_ended

**Channel**: `match:{match_id}`

```json
{
  "type": "heat_ended",
  "payload": {
    "heat_number": 2,
    "crash_speed": "312.75",
    "reason": "CRASH"
  }
}
```

**Fields**:
- `heat_number` — Heat number
- `crash_speed` — Final Speed when engine crashed (decimal string)
- `reason` — `CRASH` (engine crashed), `ALL_FINISHED` (early heat end), or `TIMER_EXPIRED`

---

### intermission

**Channel**: `match:{match_id}`

```json
{
  "type": "intermission",
  "payload": {
    "heat_number": 2,
    "standings": [
      {
        "position": 1,
        "player_id": "uuid",
        "player_name": "Alice",
        "total_score": "780.25",
        "delta": "="
      },
      {
        "position": 2,
        "player_id": "uuid",
        "player_name": "Bob",
        "total_score": "760.00",
        "delta": "↑"
      }
    ],
    "duration_seconds": 5
  }
}
```

**Fields**:
- `heat_number` — Heat just completed
- `standings` — Array of player standings (all 10 players)
  - `position` — Current position (1-10)
  - `player_id` — Player UUID (or `null` for Ghost)
  - `player_name` — Display name
  - `total_score` — Sum of all heats so far (decimal string)
  - `delta` — Position change indicator (`"↑"`, `"↓"`, `"="`)
- `duration_seconds` — Intermission duration (5 seconds)

---

### match_settled

**Channel**: `match:{match_id}`

```json
{
  "type": "match_settled",
  "payload": {
    "match_id": "uuid",
    "final_standings": [
      {
        "position": 1,
        "player_id": "uuid",
        "player_name": "Alice",
        "total_score": "1350.75",
        "prize_fuel": "450.00",
        "burn_reward": "0.00"
      },
      {
        "position": 4,
        "player_id": "uuid",
        "player_name": "Charlie",
        "total_score": "980.50",
        "prize_fuel": "0.00",
        "burn_reward": "8.00"
      }
    ]
  }
}
```

**Fields**:
- `match_id` — Match identifier
- `final_standings` — Array of all 10 players with final results
  - `position` — Final position (1-10)
  - `player_id` — Player UUID (or `null` for Ghost)
  - `player_name` — Display name
  - `total_score` — Total score across all 3 heats (decimal string)
  - `prize_fuel` — FUEL prize won (decimal string, `"0.00"` for 4th-10th)
  - `burn_reward` — BURN reward received (decimal string, `"0.00"` for 1st-3rd)

---

## Key Principles

1. **All events are JSON objects** with `type` and `payload` fields
2. **All monetary values are decimal strings** (not floats)
3. **Player IDs are UUIDs** (or `null` for Ghosts)
4. **Timestamps are ISO 8601 format** (UTC)
5. **Frontend must NOT use events as source of truth** — events are notifications only
   - Canonical state (e.g., balances) is always read via HTTP APIs
6. **Event delivery is best-effort** — clients must handle missed events gracefully
7. **RPC commands have synchronous responses** — no need for acknowledgment events
   - RPC calls like `centrifuge.rpc('matchmaking.join', {...})` return results immediately
   - Events are only for asynchronous notifications (balance changes, match updates, etc.)

## Centrifugo Publish API

Backend publishes events via gRPC:

```go
// Publish to personal channel
centrifugoClient.Publish(ctx, &gocent.PublishRequest{
  Channel: fmt.Sprintf("user:%s", userID),
  Data:    eventJSON,
})

// Publish to match channel
centrifugoClient.Publish(ctx, &gocent.PublishRequest{
  Channel: fmt.Sprintf("match:%s", matchID),
  Data:    eventJSON,
})
```

## TypeScript Type Generation

Generate TypeScript types from JSON schemas:
```bash
# (Manual types for now, consider json-schema-to-typescript for automation)
```

Example TypeScript types:
```typescript
type BalanceUpdatedEvent = {
  type: 'balance_updated';
  payload: {
    fuel_balance: string;
    burn_balance: string;
    fuel_delta: string;
    burn_delta: string;
    reason: string;
  };
};

type MatchFoundEvent = {
  type: 'match_found';
  payload: {
    match_id: string;
    league: 'ROOKIE' | 'STREET' | 'PRO' | 'TOP_FUEL';
    live_player_count: number;
    ghost_player_count: number;
    countdown_seconds: number;
  };
};

// ... (all event types)
```
