# Contract change requests — frontend

Raised by the frontend agent while building against the shared API contract v1
(`docs/requirements/trading-board.md` §4). None of these changes existing endpoints or
fields; they pin down shapes §4 leaves open. The frontend already implements the shapes
below (each is isolated in one proxy file under `frontend/app/infrastructure/proxy/`, so
adopting a different shape is a one-file change on the frontend side).

## CCR-F1 — Response shape of `GET /positions`

§4 says only "Positions". The frontend assumes a **bare JSON array** of:

```json
{
  "symbol": "2303",
  "name": "聯電",
  "shares": 90,
  "averageCost": "52.58",
  "totalCost": "4732",
  "stopLossPrice": "48",
  "takeProfitPrice": "60",
  "openedDate": "2026-11-20",
  "closedDate": null,
  "daysHeld": 5,
  "realizedPnl": "0"
}
```

- Decimals are strings (as everywhere in §4). `shares` is 0 for closed positions.
- `stopLossPrice` / `takeProfitPrice` may be `null` (no buy order known); `closedDate` is
  `null` while open; `realizedPnl` is the position's total realized P&L (§3.5), `"0"`
  while nothing was sold.
- `daysHeld`: calendar days from `openedDate` to today (open) or to `closedDate` (closed).
- Closed positions with `month=YYYY-MM` filter by `closedDate` (as §4 already says).

## CCR-F2 — Response shape of `GET /months/{month}/summary`

§4 says "fields from §3.6, §3.7". The frontend assumes one **flat** object:

```json
{
  "month": "2026-11",
  "allocation": "10000",
  "carriedOver": "0",
  "recycledProceeds": "3300",
  "spent": "9800",
  "remaining": "3500",
  "mode": "strict",
  "realizedPnl": "1200",
  "closedPositionCount": 3,
  "winCount": 2,
  "lossCount": 1,
  "winRate": "0.6667",
  "realizedReturnPct": "8.5"
}
```

- `realizedReturnPct` is a **percentage** (`"8.5"` = 8.5 %) and may be `null` when nothing
  was sold in the month.
- The frontend does **not** read `winRate` (it computes `winCount ÷ closedPositionCount`
  itself), so its unit (ratio vs percent) does not matter to the web; please document it.
- A month before the first month (§3.6) answers **404** — the web shows 「這個月沒有資料」.
  The current month with no activity yet should answer 200 with zeros (the Today budget
  strip depends on it).

## CCR-F3 — Response shape of `GET /cards?month=`

The frontend assumes a **bare JSON array** of
`{ tradeDate, status, buyCount, sellCount, filledOrderCount }`, and an empty array (not
404) for a month without cards. `filledOrderCount` = orders whose status is `filled`.

## CCR-F4 — Error responses on `POST /orders/{id}/fills` and `/skip`

The web shows `error.message` to the owner verbatim (Traditional Chinese preferred), e.g.
the budget rejection (§3.6) and "more than the remaining shares" (§3.4). Please use 409 for
business-rule conflicts (budget exceeded, skipped order, fills on a filled order) and 400
for malformed input, as §4 lists. `POST /auth/login` must answer **401** for wrong
credentials (the web distinguishes it from an expired session only by being the login
call).

## CCR-F5 — CORS (not a contract field, but required for the web to work)

The web runs on `https://trading-board.coding-afternoon.com` (and `http://localhost:3000`
in development) and calls `https://trading-board-api.coding-afternoon.com/api/v1` from the
browser. The backend must allow those origins with the `Authorization` and
`Content-Type` request headers (preflight on `POST`/`PUT`).
