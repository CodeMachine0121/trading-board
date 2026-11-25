# 📔 Ubiquitous Language Map

**Project:** Trading Board (frontend)
**Bounded Context:** Daily Trade Card loop — publish → read → report fills → positions / budget / performance
**Maintainer:** Trading Board owner
**Last Updated:** 2026-11-25

Source of truth for the vocabulary: `docs/requirements/trading-board.md` (stakeholder input, v1).
Field names follow the shared API contract §4 so frontend and backend speak the same words.

---

## 1. Nouns & Concepts

| Domain Term | Technical Name | User-Facing Label | Definition & Business Rules | Status |
| :--- | :--- | :--- | :--- | :--- |
| Owner | — | 我 / 登入 | The single human user; signs in with username + password | Confirmed |
| Daily Trade Card | `TradeCard` | 交易卡 | At most one per trade date; status trade or noTrade | Confirmed |
| Trade Date | `tradeDate` | 交易日 | Asia/Taipei calendar date the orders are intended for | Confirmed |
| Card Summary | `TradeCardSummary` | — | One row of the month's card list: trade date, status, buy count, sell count, filled order count | Confirmed |
| Buy Order | `BuyOrder` | 買進委託 | symbol, name, limitPrice, maxChasePrice, shares, estimatedAmount, stopLossPrice, takeProfitPrice, reason, isMonthEndAllocation | Confirmed |
| Sell Order | `SellOrder` | 賣出委託 | symbol, name, limitPrice, shares, exitReason, reason | Confirmed |
| Limit Price | `limitPrice` | 限價 | > 0 | Confirmed |
| Max Chase Price | `maxChasePrice` | 最高追價 | Do not buy above it; ≥ limit price | Confirmed |
| Stop Loss / Take Profit | `stopLossPrice` / `takeProfitPrice` | 停損 / 停利 | stopLoss < limit < takeProfit | Confirmed |
| Month-End Allocation | `isMonthEndAllocation` | 月底配置 | Order placed under relaxed month-end rules | Confirmed |
| Exit Reason | `exitReason` | 出場原因 | takeProfit / stopLoss / trailingStop / timeStop | Confirmed |
| Order Status | `status` | 狀態 | pending / partiallyFilled / filled / skipped; expired is display-only | Confirmed |
| Remaining Shares | `remainingShares` | 剩餘股數 | shares − filledShares | Confirmed |
| Fill | `Fill` | 成交 | price, shares, fee, amount, filledAt; append-only | Confirmed |
| Position | `Position` | 持股 | Per symbol, average cost method; open or closed | Confirmed |
| Days Held | `daysHeld` | 持有天數 | Calendar days | Confirmed |
| Realized P&L | `realizedPnl` | 已實現損益 | Only closed / sold shares; red = profit, green = loss | Confirmed |
| Win / Loss | `win` / `loss` | 獲利 / 虧損 | Closed position with realized P&L > 0 is a win, otherwise loss | Confirmed |
| Month | `month` | 月份 | YYYY-MM in Asia/Taipei | Confirmed |
| Monthly Budget | `MonthlyBudget` | 本月預算 | allocation, carriedOver, recycledProceeds, spent, remaining, mode | Confirmed |
| Allocation | `allocation` | 月配額 | Default NT$10,000 | Confirmed |
| Carried Over | `carriedOver` | 上月結轉 | Previous month's remaining | Confirmed |
| Recycled Proceeds | `recycledProceeds` | 賣出回收 | Net proceeds of this month's sell fills | Confirmed |
| Spent | `spent` | 已投入 | Buy fill amounts + fees this month | Confirmed |
| Remaining | `remaining` | 剩餘 | Never negative | Confirmed |
| Budget Mode | `mode` | 嚴格期 / 月底配置期 | strict before day 22, monthEnd from day 22 | Confirmed |
| Monthly Performance | `MonthlyPerformance` | 本月績效 | realizedPnl, closedPositionCount, winCount, lossCount, winRate, realizedReturnPct | Confirmed |

---

## 2. Actions & Processes

| Business Action | Technical Method | Trigger | Business Impact | Notes |
| :--- | :--- | :--- | :--- | :--- |
| Sign in | `signIn` | Owner submits sign-in form | Owner can read cards | |
| Sign out | `signOut` | Owner taps 登出 | Back to sign-in | |
| Read latest card | `fetchLatestCard` | Opening Today | — | |
| Report a fill | `reportFill` | 回報成交 | Order filled shares / status change; budget / positions change | Rejected if over remaining shares or budget |
| Skip an order | `skipOrder` | 略過 | Order becomes skipped | |
| Browse a month's cards | `listCards` | History month change | — | |

---

## 3. Ambiguities & Conflicts

| Ambiguous Term | Meaning in Context A | Meaning in Context B | Resolution |
| :--- | :--- | :--- | :--- |
| expired | Display state of a past-date pending order | Not a stored status | Display-only; never sent to the board |

---

## 4. External & Enum Mapping

| Category | Code Value / Key | Domain Label | Description |
| :--- | :--- | :--- | :--- |
| Card status | `trade` / `noTrade` | 交易 / 今日不買 | |
| Order status | `pending` | 待成交 | |
| Order status | `partiallyFilled` | 部分成交 | |
| Order status | `filled` | 已成交 | |
| Order status | `skipped` | 已略過 | |
| Order status (display) | expired | 已過期 | Past trade date and still pending |
| Exit reason | `takeProfit` / `stopLoss` / `trailingStop` / `timeStop` | 停利 / 停損 / 移動停利 / 時間停損 | |
| Budget mode | `strict` / `monthEnd` | 嚴格期 / 月底配置期 | |
