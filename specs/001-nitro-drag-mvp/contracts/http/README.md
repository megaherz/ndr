# HTTP API Contracts (Cold Path)

## Overview

These OpenAPI 3.0 contracts define the **cold-path REST endpoints** for canonical state reads and TON payments.

**Flow**:
```
Client → HTTPS → Go Backend HTTP API
```

## Contract Files

- `auth.json` — Telegram authentication (JWT issuance)
- `garage.json` — Garage state (balances, league access)
- `payments.json` — TON deposits/withdrawals, payment history

## Key Principles

1. **All endpoints use JWT authentication** (Bearer token in `Authorization` header)
   - Exception: `/auth/telegram` (issues JWTs, no auth required)
2. **All monetary values are decimal strings** (e.g., `"10.50"` not float)
3. **Error responses include stable machine-readable `code`**
4. **Idempotency**: Payment endpoints require `client_request_id` (UUID)

## Error Response Format

All HTTP endpoints return errors in this format:

```json
{
  "error": {
    "code": "INSUFFICIENT_BALANCE",
    "message": "User balance insufficient for operation"
  }
}
```

**HTTP Status Codes**:
- `400` — Bad Request (validation error, insufficient balance)
- `401` — Unauthorized (invalid or expired JWT)
- `404` — Not Found (resource does not exist)
- `409` — Conflict (duplicate idempotency key)
- `500` — Internal Server Error (backend failure)

## Authentication

All endpoints except `/auth/telegram` require JWT in `Authorization` header:

```
Authorization: Bearer <app_jwt>
```

JWT claims:
- `sub` — user_id (UUID)
- `exp` — expiration timestamp
- `iat` — issued at timestamp
- `jti` — token identifier

## Rate Limiting

HTTP endpoints are rate-limited per user:
- **Cold path**: 100 requests/minute
- Exceeded rate limit returns `429 Too Many Requests`

## CORS

Backend must support CORS for Telegram Mini App origin:
```
Access-Control-Allow-Origin: https://web.telegram.org
```

## Pagination

List endpoints (e.g., `/payments/history`) support pagination:
- `limit` — max results per page (default 20, max 100)
- `offset` — skip N results

## Decimal String Format

All monetary values are represented as **decimal strings** to avoid floating-point precision loss:

**Good**:
```json
{"fuel_balance": "1234.56"}
```

**Bad**:
```json
{"fuel_balance": 1234.56}
```

Frontend must parse decimal strings using appropriate decimal library (e.g., `decimal.js`).

## OpenAPI Validation

Validate contract files:
```bash
npx @openapitools/openapi-generator-cli validate -i contracts/http/auth.json
npx @openapitools/openapi-generator-cli validate -i contracts/http/garage.json
npx @openapitools/openapi-generator-cli validate -i contracts/http/payments.json
```

## TypeScript Type Generation

Generate TypeScript types from OpenAPI:
```bash
npx openapi-typescript contracts/http/auth.json -o frontend/src/types/api/auth.ts
npx openapi-typescript contracts/http/garage.json -o frontend/src/types/api/garage.ts
npx openapi-typescript contracts/http/payments.json -o frontend/src/types/api/payments.ts
```
