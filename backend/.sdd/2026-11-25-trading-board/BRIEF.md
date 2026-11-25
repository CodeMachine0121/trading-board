# Trading Board (backend v1) — Requirements Brief

## Goal
Give the owner a single system of record for the evening advisory loop: the agent team
publishes a Daily Trade Card (what to buy and sell in the next session), the owner reads it,
places the orders at the broker by hand and reports back what actually filled (or that an
order was skipped). From those reported fills the board works out the owner's holdings, how
much of the monthly NT$10,000 budget is left, and the realized profit and win rate of each
month. Only the owner and the agent team ever use it.

## Requirements
- **Who may do what.** The owner signs in with a username and password and receives a
  time-limited pass. The agent team identifies itself with a fixed key. Only the agent team
  may publish cards; the agent team may also read everything the owner can read. Owner
  credentials and the agent key are set by whoever operates the board; there is no sign-up.
- **Daily Trade Card.** At most one card per session date. A card is either a *trade* card
  (it carries orders) or a *no-trade* card ("今日不買", no orders, optional note). A card has
  any number of buy orders and sell orders.
- **Buy order** carries: stock code, stock name, limit price, highest price worth chasing,
  share count, estimated amount, stop-loss price, take-profit price, reason, and whether it
  was placed under the relaxed month-end rules.
- **Sell order** carries: stock code, stock name, limit price, share count, exit reason
  (take profit / stop loss / trailing stop / time stop) and reason.
- **Card validation.** Prices are positive; share counts are whole numbers of at least 1
  (odd lots are normal); for a buy, stop-loss < limit < take-profit and limit ≤ highest chase
  price. A sell order may only name a stock the owner currently holds, with at least that many
  shares.
- **Re-publishing.** Publishing again for the same session date replaces the earlier card,
  but only while none of its orders has received a fill; otherwise it is refused as a conflict.
- **Order progress.** An order is *waiting* until shares fill, *partly filled* while some
  have, *filled* when all have, or *skipped* when the owner says it was not placed / did not
  fill. A skipped order accepts no more fills. Waiting orders of past sessions are shown as
  expired by the screens; the board itself does nothing about them.
- **Fills.** The owner reports a fill against an order with price, share count, optional fee,
  and when it filled. An order may receive several fills; together they may never exceed the
  order's share count. Fills cannot be edited or removed in v1.
- **Fee estimate when the owner leaves the fee out.** Buy: broker fee = price × shares ×
  0.1425% × discount, rounded to a whole dollar, never below the minimum fee. Sell: the same
  broker fee plus securities transaction tax of 0.3% of the amount. Defaults: discount 0.6,
  minimum fee NT$1; both adjustable by the operator. All money is exact (no rounding drift),
  in New Taiwan dollars.
- **Holdings.** Buy fills open or grow a holding per stock at average cost (buy fees included
  in the cost). Sell fills shrink it; each sell earns realized profit = sale amount − sale fee −
  average cost per share × shares sold. A holding closes when its shares reach zero; a closed
  holding is a *win* if its total realized profit is above zero, otherwise a *loss*. An open
  holding shows stock, name, shares, average cost, total cost, current stop-loss and
  take-profit, the date it opened and how many calendar days it has been held.
- **Monthly budget.** allocation (NT$10,000 by default, adjustable) + carried over (last
  month's remaining) + recycled proceeds (this month's sale amounts minus sale fees) − spent
  (this month's purchase amounts plus purchase fees) = remaining. Remaining may never go below
  zero: a buy fill that would push it below zero is refused. The month is *strict* before the
  22nd and *month-end* from the 22nd (start day adjustable); the board only reports it. The
  first month is the month of the first card or fill; earlier months do not exist.
- **Monthly performance.** realized profit of the month, number of holdings closed in the
  month, wins, losses, win rate, and realized return = realized profit ÷ cost of the shares
  sold in the month.
- **Calendar.** Every date and month is reckoned in Taipei time.
- **Being able to tell the board is up**, so whoever runs it can see it is alive.
- **Agreed wording with the screens.** What the board says and accepts is already agreed with
  the people building the screens (section 4 of the stakeholder document) and must be
  honoured exactly as written.

## Examples (Specification by Example)
Each example lists **only** the data that affects the behavior — nothing more.

### Rule: only the agent team publishes; the owner must be signed in to read
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | correct username and password | owner signs in | receives a pass and when it expires |
| 2 (exception) | wrong password | owner signs in | refused as not authenticated |
| 3 (happy) | agent key | agent team publishes a card | accepted |
| 4 (exception) | owner's pass, no agent key | owner tries to publish a card | refused as not authenticated |
| 5 (happy) | agent key | agent team reads the latest card | allowed |
| 6 (exception) | no pass, no key / expired pass | anyone reads the latest card | refused as not authenticated |

### Rule: a buy order's prices must be in the right order
| # | Given (stop / limit / chase / take-profit, shares) | When | Then |
|---|---|---|---|
| 1 (happy) | 48 / 52.5 / 54 / 60, 90 shares | publish | accepted |
| 2 (boundary) | 48 / 54 / 54 / 60 (limit = chase) | publish | accepted |
| 3 (exception) | 52.5 / 52.5 / 54 / 60 (stop = limit) | publish | refused as invalid |
| 4 (exception) | 48 / 52.5 / 54 / 52.5 (take-profit = limit) | publish | refused as invalid |
| 5 (exception) | 48 / 55 / 54 / 60 (limit above chase) | publish | refused as invalid |
| 6 (boundary) | 1 share | publish | accepted (odd lot) |
| 7 (exception) | 0 shares, or a price of 0 | publish | refused as invalid |

### Rule: a sell order may only sell what is held
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | holding 3006 × 50 | publish a sell of 3006 × 50 | accepted |
| 2 (boundary) | holding 3006 × 50 | publish a sell of 3006 × 51 | refused as invalid |
| 3 (exception) | no holding of 2330 | publish a sell of 2330 × 10 | refused as invalid |
| 4 (exception) | holding 3006 × 50 | one card with two sells of 3006 × 30 | refused (together they exceed what is held) |

### Rule: a card is either trade or no-trade
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | no-trade, note "大盤轉弱" , no orders | publish | accepted, shown as "今日不買" |
| 2 (exception) | no-trade with one buy order | publish | refused as invalid |
| 3 (exception) | trade with no orders at all | publish | refused as invalid |

### Rule: re-publishing replaces the card only while nothing has filled
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | card for 11-26, no fills | publish 11-26 again | earlier card replaced by the new one |
| 2 (boundary) | card for 11-26, one order skipped, no fills | publish 11-26 again | replaced |
| 3 (exception) | card for 11-26, one order has a fill of 10 shares | publish 11-26 again | refused as a conflict; earlier card unchanged |

### Rule: order progress follows the filled shares
| # | Given (ordered / already filled) | When | Then |
|---|---|---|---|
| 1 (happy) | 90 / 0 | report fill of 90 | filled |
| 2 (boundary) | 90 / 0 | report fill of 40 | partly filled, 40 filled |
| 3 (boundary) | 90 / 40 | report fill of 50 | filled, 90 filled |
| 4 (exception) | 90 / 40 | report fill of 51 | refused as invalid (exceeds order) |
| 5 (exception) | skipped order | report a fill | refused as a conflict |
| 6 (happy) | waiting order | owner skips it | skipped |
| 7 (boundary) | partly filled 40 of 90 | owner skips it | skipped; the 40 filled shares stay held |
| 8 (exception) | fully filled order | owner skips it | refused as a conflict |

### Rule: fee is estimated when left out
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | buy 90 @ 52.5 (amount 4,725), no fee | report fill | fee 4 (4,725 × 0.1425% × 0.6 = 4.04 → 4) |
| 2 (boundary) | buy 1 @ 10, no fee | report fill | fee 1 (0.0086 → 0, raised to the minimum 1) |
| 3 (happy) | sell 50 @ 68 (amount 3,400), no fee | report fill | fee 3 + tax 10 = 13 |
| 4 (happy) | buy 90 @ 52.5, fee 20 given | report fill | fee 20 kept as given |

### Rule: holdings at average cost, realized profit on sale
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | none | buy 100 @ 50, fee 4 | holding 100 shares, total cost 5,004, average 50.04 |
| 2 (happy) | holding 100, total cost 5,004 | buy 100 @ 60, fee 5 | 200 shares, total cost 11,009, average 55.045 |
| 3 (happy) | holding 200, total cost 11,009 | sell 100 @ 70, fee 17 | realized profit 7,000 − 17 − 5,504.5 = 1,478.5; 100 left |
| 4 (boundary) | holding 100, total cost 5,504.5 | sell 100 @ 50, fee 12 | realized −516.5; holding closes; total realized 962 > 0 → win |
| 5 (exception) | holding 3006 × 50 | report a sell fill of 60 | refused (cannot sell more than held) |
| 6 (boundary) | holding closed, buy again later | buy 10 | a new holding opens with its own opening date |

### Rule: monthly budget and its rollover
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | first month 2026-11, buys 4,725 + fee 4 | read November | allocation 10,000, carried over 0, spent 4,729, remaining 5,271 |
| 2 (happy) | November remaining 5,271 | read December, nothing spent | carried over 5,271, remaining 15,271 |
| 3 (happy) | December has a sale of 3,400 with fee 13 | read December | recycled proceeds 3,387 |
| 4 (boundary) | remaining exactly 4,729 | buy fill costing 4,729 incl. fee | accepted, remaining 0 |
| 5 (exception) | remaining 4,728 | buy fill costing 4,729 incl. fee | refused as a conflict |
| 6 (exception) | first activity in 2026-11 | read 2026-10 | not found |
| 7 (boundary) | 21st / 22nd of the month (Taipei) | read current month | strict / month-end |

### Rule: monthly performance
| # | Given | When | Then |
|---|---|---|---|
| 1 (happy) | in the month: closed A (+962, win), closed B (−300, loss), cost of shares sold 11,009 | read summary | realized 662, closed 2, wins 1, losses 1, win rate 50%, return 6.01% |
| 2 (boundary) | nothing sold in the month | read summary | realized 0, closed 0, win rate 0, return 0 |

## Out of Scope
- Live prices and unrealized profit.
- Syncing with the go-trading trade journal.
- Notifications of any kind.
- Placing orders automatically.
- More than one user.
- Editing or deleting fills (a wrong fill stays; known limitation).
- Any automatic handling of expired orders.

## Open Decisions
Items the PRD author should resolve (recorded as assumptions; chosen as the simplest option):
- A1. Holdings are worked out in the order fills were reported; opening date = date of its
  first buy fill, closing date = date of the sell fill that brought it to zero.
- A2. The stop-loss / take-profit shown for an open holding come from the most recent (by
  session date) buy order of that stock that has at least one fill in the holding.
- A3. Cost of shares sold = total cost × shares sold ÷ shares held, rounded to the cent; the
  sale that closes the holding takes the whole remaining cost, so a holding's total realized
  profit is exact. Average cost is shown to 4 decimal places.
- A4. Broker fee and transaction tax are each rounded to the nearest whole dollar (half up).
- A5. A fill belongs to the month of its fill time (Taipei); a card belongs to the month of
  its session date.
- A6. The no-negative-remaining rule checks the fill's month and every later month that
  already exists, since money carries forward.
- A7. When nothing has ever been recorded, the first month is the current month.
- A8. Mode for a past month is month-end, for a future month strict; for the current month it
  follows today's day.
- A9. Win rate and realized return are percentages to 2 decimal places; both are 0 when the
  month has nothing closed / sold.
- A10. "Filled order count" on card summaries counts fully filled orders only.
- A11. Partly filled orders may be skipped (rest did not fill); fully filled orders may not;
  skipping an already-skipped order just returns it unchanged.
- A12. Several sell orders of the same stock on one card together may not exceed the holding.
- A13. When the month is left out of the card list it means the current month; when the
  holding status is left out it means open; the month filter applies only to closed holdings.
- A14. The owner's pass lasts 24 hours by default (adjustable).
- A15. Fill time and estimated amount are taken as given (no checks against the card date or
  limit × shares), beyond the estimated amount being positive and a given fee not negative.

## Context / Background
- Stakeholder input: `docs/requirements/trading-board.md` (v1, final). Run unattended — no
  live stakeholder; every question above was answered from that document or recorded as an
  assumption.
- Jev sized this as **large (1.00)**. Chosen stages: clarify → prd → architecture → implement
  (code-first) → improve-codebase → contract.
- A frontend is built in parallel against the same shared contract; anything the contract
  does not spell out (holding and monthly-summary shapes) must be proposed back to the
  coordinator rather than changed unilaterally.
- Reference service whose conventions to mirror: go-trading.
