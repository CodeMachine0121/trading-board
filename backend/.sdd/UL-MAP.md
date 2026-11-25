# 📔 Ubiquitous Language Map

**Project:** Trading Board (backend)
**Bounded Context:** Swing-trade advisory loop — daily trade cards, reported fills, holdings, monthly budget
**Maintainer:** Trading Board owner
**Last Updated:** 2026-11-25 (v2: sign-in & member keys)

---

## 1. Nouns & Concepts

| Domain Term | Technical Name | User-Facing Label | Definition & Business Rules | Status |
| :--- | :--- | :--- | :--- | :--- |
| Owner | owner | 我 | The single human user. v2: signs in with **email** + password; manages member keys; reads cards, reports fills, skips orders. | Confirmed |
| Agent team | agent / api key | 顧問團隊 | v1: one machine client with a fixed key. v2: a set of members, each calling with its own member key. | Confirmed |
| Pass | access token | — | Time-limited proof the owner signed in (v2: 15 minutes). | Confirmed |
| Renewal pass | refresh token | — | Long-lived (30 days, sliding) pass the browser keeps out of reach of page scripts; each renewal replaces it; reuse of a replaced one ends the sign-in. Stored scrambled. | Confirmed |
| Sign-in (session) | refresh token family | — | One sign-in and all the renewal passes descending from it; signing out or theft detection ends it as a whole. | Confirmed |
| One-time sign-in code | authorization code | — | Handed back after a successful sign-in; 60 s, single use, bound to application, return address and proof. | Confirmed |
| Web app (registered application) | OAuth client `trading-board-web` | — | The only application allowed to ask for a sign-in; return addresses are allowed by the operator. | Confirmed |
| Sign-in lock | sign-in throttle | 嘗試次數過多 | 5 failures within 15 min lock sign-in for 15 min. | Confirmed |
| Agent-team member | member name | 成員名稱 | One member of the agent team (e.g. stock-scanner); 1–40 characters, unique among active keys ignoring case. | Confirmed |
| Member key | ApiKey | API 金鑰 | A member's own key (`tbk_…`), shown once at creation, stored scrambled with a display prefix; active or revoked. | Confirmed |
| Publisher | publishedBy | 發布者 | Member name that published a card. | Confirmed |
| Reported by / skipped by | reportedBy / skippedBy | 回報者 | Member name or `owner` that reported a fill or skipped an order. | Confirmed |
| Daily Trade Card (card) | TradeCard | 每日交易卡 | At most one per session date; status trade or no-trade; carries buy and sell orders. | Confirmed |
| Session date | tradeDate | 交易日 | Calendar date (Taipei) the orders are intended for. | Confirmed |
| No-trade card | status `noTrade` | 今日不買 | A card with no orders and an optional note. | Confirmed |
| Buy order | TradeOrder (side buy) | 買進 | Stock, limit price, highest chase price, shares, estimated amount, stop-loss, take-profit, reason, month-end flag. | Confirmed |
| Sell order | TradeOrder (side sell) | 賣出 | Stock, limit price, shares, exit reason, reason. May only sell what is held. | Confirmed |
| Highest chase price | maxChasePrice | 追價上限 | Do not buy above this; limit price ≤ it. | Confirmed |
| Month-end allocation | isMonthEndAllocation | 月底配置 | Order placed under relaxed month-end rules. | Confirmed |
| Exit reason | exitReason | 出場原因 | take profit / stop loss / trailing stop / time stop. | Confirmed |
| Fill | OrderFill | 成交 | Price, shares, fee, amount (= price × shares), fill time. Append-only. | Confirmed |
| Fee | fee | 手續費 | Broker fee (and, on sells, transaction tax); estimated when omitted. | Confirmed |
| Holding (position) | Position / PositionDomain | 持股 | Per-stock lot opened by buy fills, closed when shares reach zero; average cost method. | Confirmed |
| Realized profit | realizedPnl | 已實現損益 | Sale amount − sale fee − average cost × shares sold. | Confirmed |
| Win / loss | result | 勝 / 敗 | Closed holding with total realized profit > 0 is a win, else a loss. | Confirmed |
| Monthly budget | MonthlyBudget | 月預算 | allocation + carried over + recycled proceeds − spent = remaining (never negative). | Confirmed |
| Allocation | allocation | 月配額 | NT$10,000 per month by default. | Confirmed |
| Carried over | carriedOver | 上月結轉 | Previous month's remaining. | Confirmed |
| Recycled proceeds | recycledProceeds | 回收資金 | This month's sale amounts minus sale fees. | Confirmed |
| Spent | spent | 已投入 | This month's purchase amounts plus purchase fees. | Confirmed |
| Budget mode | mode | 嚴格 / 月底 | strict before the 22nd, month-end from the 22nd (Taipei). Reported only. | Confirmed |
| Monthly performance | MonthlyPerformance | 月績效 | realized profit, closed count, wins, losses, win rate, realized return. | Confirmed |

---

## 2. Actions & Processes

| Business Action | Technical Method | Trigger | Business Impact | Notes |
| :--- | :--- | :--- | :--- | :--- |
| Sign in | login / authorize | Owner submits email + password | v2: web app receives a one-time sign-in code | Lock after 5 failures |
| Trade in code | token (authorization_code) | Web app after the return | Pass + renewal pass | |
| Renew | token (refresh_token) | Web app on start-up / before expiry | New pass + new renewal pass | Reuse ends the sign-in |
| Sign out | revoke | Owner | Sign-in ended | |
| Create member key | create api key | Owner | Key shown once | |
| Revoke member key | revoke api key | Owner | Key admits nobody, final | |
| Publish a card | publish card | Agent team, each evening | Card for the session date created or replaced | Replace only while no fill exists |
| Report a fill | report fill | Owner after the broker fills | Order progress, holding, budget change | Refused if it would overfill, oversell or overspend |
| Skip an order | skip order | Owner chose not to place / it did not fill | Order becomes skipped; accepts no more fills | |
| Read holdings | list positions | Owner opens holdings | — | |
| Read monthly summary | month summary | Owner opens the month | — | |

---

## 3. Ambiguities & Conflicts

| Ambiguous Term | Meaning in Context A | Meaning in Context B | Resolution |
| :--- | :--- | :--- | :--- |
| Expired order | Waiting order of a past session (screens) | — | Display concern only; the board stores no such state. |

---

## 4. External & Enum Mapping

| Category | Code Value / Key | Domain Label | Description |
| :--- | :--- | :--- | :--- |
| Card status | `trade` | 有交易 | Has at least one order |
| Card status | `noTrade` | 今日不買 | No orders |
| Order status | `pending` | 等待中 | No shares filled |
| Order status | `partiallyFilled` | 部分成交 | Some shares filled |
| Order status | `filled` | 全部成交 | All shares filled |
| Order status | `skipped` | 略過 | Set by the owner |
| Exit reason | `takeProfit` / `stopLoss` / `trailingStop` / `timeStop` | 停利 / 停損 / 移動停利 / 時間停損 | |
| Budget mode | `strict` / `monthEnd` | 嚴格 / 月底 | |
| Member key status | `active` / `revoked` | 啟用中 / 已撤銷 | |
| Position status | `open` / `closed` | 持有中 / 已出場 | |
