# RPC Contracts (Centrifugo → gRPC)

## Overview

These Protobuf contracts define the **hot-path gameplay commands** that are proxied from Centrifugo to the Go backend via gRPC.

**Correct Flow**:
```
1. Client (WebSocket): centrifuge.rpc('matchmaking.join', {league: 'STREET', ...})
2. Centrifugo receives via WebSocket
3. Centrifugo → Backend (gRPC): CentrifugoProxy.RPC(method='matchmaking.join', data=<JSON bytes>, user='user_id')
4. Backend: Unmarshal data → JoinMatchmakingRequest, process, marshal response → JSON bytes
5. Backend → Centrifugo (gRPC): RPCResponse(result={data: <JSON bytes>})
6. Centrifugo → Client (WebSocket): RPC result
```

**Key Point**: Backend implements **Centrifugo's `CentrifugoProxy` service**, NOT custom services. The RPC method name is passed as a string in the `method` field, and request/response data is JSON-encoded in the `data` field.

## Contract Files

- `centrifugo_proxy.proto` — **Centrifugo v4 proxy protocol** (backend MUST implement this)
- `matchmaking.proto` — Data structures for join/cancel matchmaking (JSON payload)
- `match.proto` — Data structures for in-match actions (JSON payload)

## Key Principles

1. **All requests include `client_req_id`** for idempotency
2. **All responses are JSON-serializable** (no complex nested objects)
3. **Error responses include stable machine-readable `code`** (frontend maps to localized messages)
4. **Decimal values are strings** (e.g., `"245.50"` not float) to avoid precision loss

## Error Response Format

All RPC handlers return errors in this format:

```json
{
  "error": {
    "code": "INSUFFICIENT_BALANCE",
    "message": "User balance insufficient for league buy-in"
  }
}
```

**Frontend must NOT display `message` directly** — use `code` to look up localized message via i18n.

## Authentication

- All RPC calls are authenticated via **Centrifugo JWT** (user identity in JWT `sub` claim)
- Backend never trusts user_id from RPC payload — always derives from JWT

## Rate Limiting

- RPC calls are rate-limited per user: **10 requests/second** (configurable)
- Exceeded rate limit returns error code `RATE_LIMIT_EXCEEDED`

## Idempotency

- `client_req_id` is stored in Redis with TTL (30 seconds)
- Duplicate `client_req_id` within TTL returns cached response
- This prevents double-processing due to network retries

## Backend Implementation Pattern

### 1. Implement Centrifugo Proxy Service

```go
type RPCHandler struct {
    // dependencies...
}

// Implement CentrifugoProxy.RPC method
func (h *RPCHandler) RPC(ctx context.Context, req *centrifugo.RPCRequest) (*centrifugo.RPCResponse, error) {
    // Extract user_id from req.User (authenticated by Centrifugo JWT)
    userID := req.User
    
    // Route by method name
    switch req.Method {
    case "matchmaking.join":
        return h.handleMatchmakingJoin(ctx, userID, req.Data)
    case "matchmaking.cancel":
        return h.handleMatchmakingCancel(ctx, userID, req.Data)
    case "match.earn_points":
        return h.handleMatchEarnPoints(ctx, userID, req.Data)
    case "match.give_up":
        return h.handleMatchGiveUp(ctx, userID, req.Data)
    default:
        return &centrifugo.RPCResponse{
            Error: &centrifugo.Error{Code: 404, Message: "method not found"},
        }, nil
    }
}

func (h *RPCHandler) handleMatchmakingJoin(ctx context.Context, userID string, data []byte) (*centrifugo.RPCResponse, error) {
    // Unmarshal request
    var req matchmaking.JoinMatchmakingRequest
    if err := json.Unmarshal(data, &req); err != nil {
        return &centrifugo.RPCResponse{
            Error: &centrifugo.Error{Code: 400, Message: "invalid request"},
        }, nil
    }
    
    // Process (call matchmaker service)
    resp, err := h.matchmaker.Join(ctx, userID, req.League, req.ClientReqId)
    if err != nil {
        return &centrifugo.RPCResponse{
            Error: &centrifugo.Error{Code: 500, Message: err.Error()},
        }, nil
    }
    
    // Marshal response
    respData, _ := json.Marshal(resp)
    return &centrifugo.RPCResponse{
        Result: &centrifugo.RPCResult{Data: respData},
    }, nil
}
```

### 2. Register gRPC Server

```go
grpcServer := grpc.NewServer()
centrifugo.RegisterCentrifugoProxyServer(grpcServer, rpcHandler)
```

## Protobuf Generation

Generate Go code:
```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       contracts/rpc/centrifugo_proxy.proto \
       contracts/rpc/matchmaking.proto \
       contracts/rpc/match.proto
```

## Client-Side Usage

```typescript
import { Centrifuge } from 'centrifuge';

const centrifuge = new Centrifuge('ws://localhost:8000/connection/websocket', {
  token: centrifugoJWT, // From auth endpoint
});

// Join matchmaking
const result = await centrifuge.rpc('matchmaking.join', {
  league: 'STREET',
  client_req_id: crypto.randomUUID(),
});
console.log(result.data); // JoinMatchmakingResponse JSON

// Earn points
const earnResult = await centrifuge.rpc('match.earn_points', {
  match_id: matchId,
  heat_number: 2,
  client_req_id: crypto.randomUUID(),
});
console.log(earnResult.data); // EarnPointsResponse JSON
```

## Centrifugo Configuration

Configure RPC proxy in `centrifugo/config.json`:

```json
{
  "rpc_proxy": {
    "enabled": true,
    "endpoint": "grpc://backend:8080"
  }
}
```

Reference: https://github.com/centrifugal/centrifugo/blob/v4/internal/proxyproto/proxy.proto
