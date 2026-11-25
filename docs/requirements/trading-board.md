# Trading Board — Stakeholder Requirements (v1)

This document is the **stakeholder input** for both the backend and frontend agents.
All business decisions below are final for v1. When `/clarify` would ask the
stakeholder a question, answer it from this document. If something is genuinely not
covered, pick the simplest reasonable option, record it as an explicit assumption in
your BRIEF/PRD, and keep going — do not stop to ask.

This file is read-only for both agents. If you need the shared API contract changed,
write the proposal to `docs/contract-change-requests/<backend|frontend>.md` in your own
branch and mention it in your final report; the coordinator will reconcile.

## 1. Context

The owner runs a Taiwan-stock "agent team" (Claude Code subagents) that acts as a
personal swing-trading advisory squad. Every evening after market close the team
produces a **Daily Trade Card**: which stocks to buy tomorrow (limit price, share
count, amount, stop loss, take profit) and which holdings to sell. The owner places the
orders manually at the broker and reports the actual fills back.

Trading Board is the system of record and the viewing surface for that loop:

- the agent team **publishes** cards via API (machine client, API key);
- the owner **reads** the card on phone/desktop and **reports fills** (web, login);
- the board derives **positions, monthly budget and realized P&L** from the fills.

Investment mandate the board must reflect:

- Cash equities only (no leverage, margin, futures, day trading). Long only.
- Swing trading, buying confirmed breakouts after a bottoming base. Not buy-and-hold.
- Capital: **NT$10,000 is allocated every month**. The owner requires the monthly
  budget to be fully deployed within the month. Odd-lot (零股) trading is normal, so
  share counts are integers ≥ 1, not multiples of 1,000.
- "Profit" means **realized** profit only (closed / sold shares). There is no live
  price feed in v1, so no unrealized P&L.

## 2. Users

| User | How they authenticate | What they do |
| --- | --- | --- |
| Owner (single human user) | Username + password → access token (JWT) | Reads cards, reports fills, skips orders, views positions/budget/performance |
| Agent team (machine) | Static API key header `X-Api-Key` | Publishes daily cards; may also read everything |

Single-tenant. No sign-up, no multi-user. Owner credentials and the API key come from
configuration (Kubernetes Secret).

## 3. Domain rules

### 3.1 Trading calendar & time
- All business dates use **Asia/Taipei**. `tradeDate` is a calendar date (YYYY-MM-DD)
  meaning "the session the orders are intended for".
- A month is `YYYY-MM` in Asia/Taipei.

### 3.2 Daily Trade Card
- At most **one card per tradeDate**. Status is `trade` (has orders) or `noTrade`
  ("今日不買" — no orders, optional note).
- A card has 0..n **buy orders** and 0..n **sell orders**.
- Buy order fields: symbol, name, limitPrice, maxChasePrice (do not buy above this),
  shares, estimatedAmount, stopLossPrice, takeProfitPrice, reason, and
  `isMonthEndAllocation` (true when the team placed it under relaxed month-end rules).
- Sell order fields: symbol, name, limitPrice, shares, exitReason
  (`takeProfit` | `stopLoss` | `trailingStop` | `timeStop`), reason.
- Re-publishing a card for the same tradeDate **replaces** it, but only while none of
  its orders has a fill; otherwise reject with a conflict.
- Validation: prices > 0, shares integer ≥ 1, stopLossPrice < limitPrice <
  takeProfitPrice for buys, limitPrice ≤ maxChasePrice. A sell order must reference a
  symbol the owner currently holds with at least that many shares.

### 3.3 Order status
- `pending` → `partiallyFilled` → `filled`, or `skipped` (owner chose not to place /
  it did not fill). Status is derived from filled shares vs ordered shares, except
  `skipped`, which the owner sets explicitly. A skipped order cannot receive fills.
- Orders of past tradeDates that are still `pending` are shown as "expired" in the UI
  (display concern only; no background job needed).

### 3.4 Fills
- The owner reports a fill against an order: price, shares, fee (optional), filledAt.
- Multiple fills per order are allowed; total filled shares may not exceed the order's
  shares.
- If fee is omitted the backend estimates it:
  - buy fee = round(price × shares × 0.1425% × feeDiscount), minimum `minFee`
  - sell fee = same broker fee + securities transaction tax 0.3% of the amount
  - defaults: feeDiscount = 0.6, minFee = NT$1 (odd-lot). All configurable.
- Money is handled as exact decimals (never float). Amounts in TWD.

### 3.5 Positions
- Buy fills open/increase a position per symbol; cost basis includes buy fees
  (average cost method).
- Sell fills reduce it; realized P&L per sell fill =
  sell amount − sell fee − (average cost per share × shares sold).
- A position closes when its shares reach 0. A closed position is a **win** if its
  total realized P&L > 0, otherwise a **loss**.
- Each open position shows: symbol, name, shares, average cost, total cost, stop loss
  and take profit (from the latest buy order for that symbol), opened date, days held
  (calendar days).

### 3.6 Monthly budget
- `allocation` = monthly allocation (default NT$10,000, configurable).
- `carriedOver` = previous month's `remaining` (may be 0). Unspent money rolls over.
- `recycledProceeds` = net proceeds of sell fills in this month (amount − fee).
- `spent` = sum of buy fill amounts + buy fees in this month.
- `remaining` = allocation + carriedOver + recycledProceeds − spent. It may not go
  negative: a buy fill that would make it negative is rejected.
- `mode` = `strict` before day 22 of the month, `monthEnd` from day 22 (Asia/Taipei;
  start day configurable). The board only reports the mode; the agent team applies it.
- The first month is the month of the first card or fill; earlier months do not exist.

### 3.7 Monthly performance
- realizedPnl (sum of realized P&L of sell fills in the month), closedPositionCount,
  winCount, lossCount, winRate, and realizedReturnPct = realizedPnl ÷ cost basis of
  shares sold in the month.

## 4. Shared API contract v1

Base path `/api/v1`, JSON, camelCase. Decimals are **strings** (e.g. `"52.5"`).
Dates `YYYY-MM-DD`, months `YYYY-MM`, timestamps RFC 3339 with offset.
Errors: `{ "error": { "code": "string", "message": "string" } }` with 400 (validation),
401, 404, 409 (conflict). Owner endpoints accept `Authorization: Bearer <jwt>`;
all endpoints also accept `X-Api-Key`.

| Method & path | Who | Purpose |
| --- | --- | --- |
| `POST /auth/login` | public | `{username,password}` → `{accessToken, expiresAt}` |
| `PUT /cards/{tradeDate}` | api key | Publish / replace the card (body below) → full card |
| `GET /cards?month=YYYY-MM` | owner | Card summaries for a month, newest first: `tradeDate, status, buyCount, sellCount, filledOrderCount` |
| `GET /cards/latest` | owner | Full card with the greatest tradeDate (404 if none) |
| `GET /cards/{tradeDate}` | owner | Full card |
| `POST /orders/{orderId}/fills` | owner | `{price, shares, fee?, filledAt}` → updated order |
| `POST /orders/{orderId}/skip` | owner | Mark order skipped → updated order |
| `GET /positions?status=open\|closed&month=YYYY-MM` | owner | Positions (month filter applies to closed, by close date) |
| `GET /months/{month}/summary` | owner | Budget + performance for the month (fields from §3.6, §3.7) |

Card body (`PUT`) and full card response:

```json
{
  "tradeDate": "2026-11-26",
  "status": "trade",
  "note": "string",
  "buyOrders": [{
    "id": "uuid (response only)",
    "symbol": "2303", "name": "聯電",
    "limitPrice": "52.5", "maxChasePrice": "54", "shares": 90,
    "estimatedAmount": "4725", "stopLossPrice": "48", "takeProfitPrice": "60",
    "reason": "string", "isMonthEndAllocation": false,
    "status": "pending (response only)", "filledShares": 0, "fills": []
  }],
  "sellOrders": [{
    "id": "uuid (response only)",
    "symbol": "3006", "name": "晶豪科", "limitPrice": "68", "shares": 50,
    "exitReason": "takeProfit", "reason": "string",
    "status": "pending (response only)", "filledShares": 0, "fills": []
  }],
  "publishedAt": "timestamp (response only)"
}
```

Fill object: `{ id, price, shares, fee, amount, filledAt }`.

## 5. Frontend scope (Traditional Chinese UI, mobile-first)

1. **Login** page.
2. **Today** (home): latest card. Buy section and sell section as order cards showing
   every field; status badge; "回報成交" opens a fill form (price defaults to limit
   price, shares default to remaining shares, fee optional); "略過" skips. A `noTrade`
   card shows a clear "今日不買" banner. A month-budget strip on top: remaining /
   allocation, progress bar, `strict` / `monthEnd` mode badge.
3. **History**: month picker → list of cards → card detail (same component as Today,
   read-only for expired orders).
4. **Positions**: open positions table/cards; closed positions for a month with
   realized P&L, win/loss coloring (Taiwan convention: **red = up/profit,
   green = down/loss**).
5. **Monthly summary**: budget breakdown (§3.6) and performance (§3.7).

The frontend develops against the contract in §4 with a mock layer until the backend
is available; switching to the real API must be a configuration change.

## 6. Non-functional & platform

- Backend: Go + Gin, PostgreSQL via ORM with code-first schema sync (follow
  `go-trading` conventions and `backend/.claude/rules`).
- Frontend: Nuxt 3 + Bun (follow `go-trading-frontend` conventions and
  `frontend/.claude/rules`).
- Database: the **same PostgreSQL server** as go-trading
  (`postgres.go-trading.svc.cluster.local`), but a **new database `trading_board`**.
  Creating that database is a documented manual step (the `database/` folder in
  go-trading-deploy is not managed by Argo).
- Kubernetes namespace `go-trading` (same as the platform). Hostnames:
  backend `trading-board-api.coding-afternoon.com`, frontend
  `trading-board.coding-afternoon.com`.
- CI/CD follows go-trading: GitHub Actions builds/tests, pushes the image to GHCR, then
  bumps the image tag in go-trading-deploy's kustomization; Argo CD syncs.
- Health endpoint for probes.

## 7. Out of scope for v1

Live prices / unrealized P&L, syncing with go-trading's trade journal, notifications
(Telegram etc.), automatic order placement, multi-user, editing or deleting fills
(v1: append-only; a wrong fill is a known limitation to note in the PRD).
