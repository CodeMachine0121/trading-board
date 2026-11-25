# Contract change requests — backend

Section 4 of `docs/requirements/trading-board.md` is implemented exactly. The items below are places where
the contract leaves a shape or a behaviour open; the backend made the simplest choice and documents it here
so the coordinator can confirm it or reconcile it with the frontend. Nothing here removes or renames a
contract field.

## 1. Position object (`GET /positions`) — shape not specified

Bare JSON array of:

```json
{
  "symbol": "2303",
  "name": "聯電",
  "status": "open",
  "shares": 100,
  "averageCost": "50.04",
  "totalCost": "5004",
  "stopLossPrice": "48",
  "takeProfitPrice": "60",
  "openedDate": "2026-11-20",
  "closedDate": null,
  "daysHeld": 5,
  "realizedPnl": "0",
  "result": null
}
```

- **Open** (`status: "open"`): `shares` / `totalCost` / `averageCost` are what is held now (average cost to
  4 decimals, buy fees included); `realizedPnl` is profit already realized by partial sales of this holding;
  `stopLossPrice` / `takeProfitPrice` come from the most recent (by `tradeDate`) buy order of that symbol that
  has a fill in this holding; `closedDate` and `result` are `null`; `daysHeld` = calendar days from
  `openedDate` to today (Asia/Taipei), never negative.
- **Closed** (`status: "closed"`): `shares` = total shares bought into the holding, `totalCost` = total buy
  cost (incl. fees), `averageCost` = average buy cost; `realizedPnl` = total realized P&L; `result` =
  `"win"` (> 0) or `"loss"`; `closedDate` = date of the sell fill that emptied it; `daysHeld` =
  `closedDate − openedDate`; `stopLossPrice` / `takeProfitPrice` are `null`.
- Defaults: `status` omitted = `open`; `month` filters closed holdings by `closedDate` only and is ignored
  for open ones. Open holdings are listed in opening order, closed ones newest close first.
- A holding that is sold to zero and bought again is a **new** holding with its own `openedDate`.

## 2. Monthly summary (`GET /months/{month}/summary`) — flat object

```json
{
  "month": "2026-11",
  "mode": "strict",
  "allocation": "10000",
  "carriedOver": "0",
  "recycledProceeds": "0",
  "spent": "4729",
  "remaining": "5271",
  "realizedPnl": "0",
  "closedPositionCount": 0,
  "winCount": 0,
  "lossCount": 0,
  "winRate": "0",
  "realizedReturnPct": "0"
}
```

- `winRate` and `realizedReturnPct` are **percentages** as decimal strings rounded to 2 places
  (`"50"` = 50 %, `"6.01"` = 6.01 %); both are `"0"` when nothing closed / nothing was sold.
- `mode`: current month follows today's day (Asia/Taipei, month-end from the 22nd); a past month reports
  `monthEnd`, a future month `strict`.
- Months before the first month → 404 `monthNotFound`. When nothing has ever been recorded, the first month
  is the current month (so the budget strip has something to show from day one).

## 3. Order objects carry `side`

Every order object (inside cards and in the fill / skip responses) has an extra `"side": "buy" | "sell"`, so
the fill / skip responses are self-describing. Additive; clients may ignore it. Buy-only fields are omitted
on sell orders and `exitReason` is omitted on buy orders.

## 4. Error codes (`error.code`)

| Status | Codes |
| --- | --- |
| 400 | `invalidRequest` (unreadable body / query, bad date or month, body `tradeDate` ≠ path), `invalidCard`, `sellExceedsHolding` (published sells exceed — or name a stock not in — the holding), `invalidFill`, `fillExceedsOrder` |
| 401 | `unauthenticated` |
| 404 | `cardNotFound`, `orderNotFound`, `monthNotFound` |
| 409 | `cardHasFills` (re-publish after a fill), `orderSkipped` (fill on a skipped order), `orderAlreadyFilled` (skip a filled order), `insufficientShares` (sell fill > held), `insufficientBudget` (buy fill would make a month's remaining negative) |
| 500 | `internalError` (details only in the server log) |

`message` is Traditional Chinese, meant for display.

## 5. Smaller behaviours the contract leaves open

- List endpoints (`GET /cards`, `GET /positions`) return bare JSON arrays; `GET /cards` without `month`
  means the current month (Asia/Taipei).
- `PUT /cards/{tradeDate}`: the body's `tradeDate` is optional; if present it must equal the path. The
  owner's bearer token alone is refused with 401 (publishing needs `X-Api-Key`).
- `POST /orders/{id}/fills`: `fee` omitted or `null` → estimated. `filledAt` is required (RFC 3339).
  Timestamps in responses are rendered in Asia/Taipei (`+08:00`).
- `POST /orders/{id}/skip`: allowed for `pending` and `partiallyFilled` (already-filled shares stay held);
  skipping an already skipped order returns it unchanged; a `filled` order → 409.
- Health: `GET /health` and `GET /api/v1/health` → `{"status":"ok"}`, no credentials.
- Login: `expiresAt` defaults to 24 h after sign-in.
