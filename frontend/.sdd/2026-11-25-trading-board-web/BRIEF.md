# Trading Board Web — Requirements Brief

## Goal
Give the owner a phone-first, Traditional Chinese viewing surface for the agent team's
Daily Trade Card loop: sign in, read tonight's card for the next session, report the
fills that actually happened at the broker (or skip an order), and see what the fills
add up to — open and closed positions, how much of this month's budget is left, and the
month's realized result. Taiwan colour convention applies everywhere money goes up or
down: **red means profit / up, green means loss / down**.

## Requirements
- R1 Sign-in: the owner signs in with username and password. Every other screen is only
  reachable while signed in; when the sign-in is no longer accepted the owner is sent
  back to the sign-in screen. The owner can sign out.
- R2 Today (home): shows the most recent card. Each buy order shows symbol, name, limit
  price, maximum chase price, shares, estimated amount, stop loss, take profit, reason,
  whether it is a month-end allocation, its status, filled shares and every reported
  fill. Each sell order shows symbol, name, limit price, shares, exit reason, reason,
  status, filled shares and fills.
- R3 A "no trade" card shows a clear 「今日不買」 banner with its note instead of orders.
- R4 Reporting a fill (「回報成交」): price starts at the order's limit price, shares start
  at the order's remaining shares, fee is optional (left empty → the board estimates
  it), fill time starts at "now". After a successful report the order shows its new
  status, filled shares and the new fill; the budget strip refreshes.
- R5 Skipping (「略過」) marks an order as not placed; a skipped order offers no further
  actions.
- R6 Order status shown as: 待成交 / 部分成交 / 已成交 / 已略過 / 已過期. An order is
  "expired" when its card's trading day is before today (Taipei time) and it is still
  waiting; expired orders are read-only.
- R7 Month budget strip on top of Today: remaining vs. monthly allocation, a progress
  bar of how much of this month's available money is still unspent, and the mode badge
  (「嚴格期」 before the month-end window, 「月底配置期」 inside it).
- R8 History: pick a month → list of that month's cards, newest first, each showing
  date, trade / no-trade, number of buys and sells, number of completed orders →
  open a card to see the same card view as Today.
- R9 Positions: open positions (symbol, name, shares, average cost, total cost, stop
  loss, take profit, opened date, days held) and, for a chosen month, closed positions
  with realized profit/loss, coloured win (red) / loss (green).
- R10 Monthly summary: pick a month → budget breakdown (allocation, carried over,
  recycled proceeds, spent, remaining, mode) and performance (realized profit/loss,
  closed positions, wins, losses, win rate, realized return).
- R11 Money and prices are shown exactly as the board reports them (no rounding drift),
  amounts as NT$ with thousands separators.
- R12 Before the real board service is available the whole web can be previewed with
  sample data; moving to the real service is a setting change, not a code change.

## Examples (Specification by Example)
Each example lists **only** the data that affects the behavior — nothing more.

### Rule: which screen a visitor sees (R1)
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | not signed in | owner enters correct username and password | lands on Today |
| 2 (exception) | not signed in | wrong password | stays on sign-in, sees 「帳號或密碼錯誤」 |
| 3 (exception) | not signed in | opens History directly | is taken to sign-in |
| 4 (boundary) | signed in, sign-in has since expired | opens any screen / any action is refused as not signed in | is taken to sign-in |

### Rule: what Today shows (R2, R3)
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | latest card 2026-11-26 is a trade card with 1 buy (聯電 2303) and 1 sell (晶豪科 3006) | owner opens Today | sees a 買進 section with the 聯電 order card and a 賣出 section with the 晶豪科 order card |
| 2 (boundary) | latest card is a trade card with buys but no sells | opens Today | 賣出 section says 「無賣出委託」 |
| 3 (exception) | latest card is "no trade", note 「大盤轉弱」 | opens Today | sees 「今日不買」 banner with the note, no order sections |
| 4 (exception) | no card has ever been published | opens Today | sees 「尚未發布任何交易卡」 |

### Rule: fill form defaults and limits (R4)
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | buy 90 shares @ limit 52.5, 0 filled | opens 回報成交 | price 52.5, shares 90, fee empty |
| 2 (boundary) | buy 90 shares, 60 filled | opens 回報成交 | shares defaults to 30 |
| 3 (boundary) | 30 remaining | enters 30 shares | accepted |
| 4 (exception) | 30 remaining | enters 31 shares | form refuses: 「股數不可超過剩餘 30 股」 |
| 5 (exception) | — | enters 0 shares or price 0 | form refuses |
| 6 (exception) | budget remaining NT$1,000 | reports a buy fill costing NT$4,725 | board refuses; owner sees the board's reason and the order is unchanged |
| 7 (happy) | fee left empty | submits valid fill | board estimates the fee; the new fill shows it |

### Rule: how an order's progress is labelled and what the owner can do with it (R5, R6)
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | order is for today's session, nothing bought yet | owner looks at it | 「待成交」, owner can report a fill or skip it |
| 2 (happy) | 30 of 90 shares bought | owner looks at it | 「部分成交」, owner can report a fill or skip it |
| 3 (boundary) | all 90 of 90 shares bought | owner looks at it | 「已成交」, nothing more to do |
| 4 (exception) | owner chose to skip it | owner looks at it | 「已略過」, nothing more to do |
| 5 (boundary) | order was for yesterday's session, nothing bought | owner looks at it | 「已過期」, nothing more to do |
| 6 (boundary) | order was for yesterday's session, 30 of 90 bought | owner looks at it | 「部分成交」, owner can still report the rest |

### Rule: budget strip (R7)
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | allocation 10,000, carried 0, recycled 0, spent 4,725, day 10 | view Today | 「剩餘 NT$5,275 / 月配額 NT$10,000」, bar at 52.75 %, badge 「嚴格期」 |
| 2 (boundary) | day 22 (mode reported as month-end) | view | badge 「月底配置期」 |
| 3 (boundary) | spent everything | view | remaining NT$0, bar empty |
| 4 (boundary) | carried 2,000, recycled 3,000, spent 0 | view | remaining NT$15,000 / allocation NT$10,000, bar full |

### Rule: profit/loss colouring (R9, R10)
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | closed position, realized +1,200 | view | shown red, labelled 「獲利」 |
| 2 (exception) | closed position, realized −800 | view | shown green, labelled 「虧損」 |
| 3 (boundary) | closed position, realized 0 | view | shown neutral, labelled 「虧損」 (break-even counts as a loss) |

### Rule: monthly summary (R10)
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | 2026-11: 3 closed, 2 wins | view | win rate 66.67 % |
| 2 (boundary) | month with no closed positions | view | win rate and realized return shown as 「—」 |
| 3 (exception) | month before the board's first month | view | 「這個月沒有資料」 |

### Rule: which month is shown first when choosing a month (R8, R9, R10)
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | today is 25 November 2026 in Taiwan | owner opens History, closed positions or the monthly summary | November 2026 is shown first |
| 2 (boundary) | no card was published in November 2026 | owner chooses November 2026 in History | sees 「這個月沒有交易卡」 |

## Out of Scope
- Live prices and unrealized profit/loss.
- Editing or deleting a reported fill (a wrong fill is a known limitation).
- Publishing cards from the web (only the agent team publishes).
- Notifications, multiple users, sign-up, password change.
- Other languages than Traditional Chinese.

## Open Decisions
Resolved here as assumptions (see Context); the PRD records them:
- A1 The shapes of the positions list and the monthly summary are not spelled out in the
  shared agreement with the board service; the web assumes the natural field names from
  the domain rules and raises a change request so both sides agree.
- A2 Progress bar = remaining ÷ (allocation + carried over + recycled proceeds).
- A3 Break-even (0) closed positions are labelled loss but coloured neutral.
- A4 A month before the board's first month is shown as "no data".
- A5 Fill time defaults to the current time; the owner may change it.
- A6 Partially filled orders from past trading days stay reportable (only fully-waiting
  ones are "expired").

## Context / Background
Stakeholder input: `docs/requirements/trading-board.md` §1–§5, §7 (final for v1). No
clarification round was needed — every business question is answered there; remaining
gaps are recorded as assumptions A1–A6. Run unattended; stages chosen by the owner:
clarify → prd → architecture → implement (code-first) → improve-codebase → contract.
