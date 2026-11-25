# Product Requirements Document (PRD) — Trading Board (backend v1)

**Status:** Finalized (unattended run; decisions from the stakeholder document)
**Version:** v1.0
**Owner:** Trading Board owner
**Stakeholders:** Owner, agent team, frontend builder, coordinator

---

## 1. Background & Goal (Why & Goal)

- **Problem Statement:** The agent team's evening advice (what to buy and sell next session)
  and the owner's actual fills live nowhere in particular. The owner cannot see in one place
  what was advised, what really filled, what is held, how much of this month's NT$10,000 is
  left, and whether the strategy is actually making money.
- **Expected Outcome:** Every evening's Daily Trade Card is recorded by the agent team; every
  fill the owner reports is reflected immediately in order progress, holdings, the monthly
  budget and monthly performance — all from one source of record, with exact money figures.
- **Out of Scope:** live prices and unrealized profit; syncing with the go-trading trade
  journal; notifications; automatic order placement; more than one user; editing or deleting
  fills (a wrong fill stays — known limitation of v1); any automatic handling of expired
  orders.

---

## 2. User Personas

- **Owner** — the single human user. Reads the card on phone or desktop in the evening and
  before the session, places orders at the broker by hand, and reports fills afterwards.
- **Agent team** — a machine client. Publishes the Daily Trade Card each evening after market
  close; may read everything the owner can read (e.g. holdings, before advising sells).
- **Operator** — whoever runs the board; sets the owner's username and password, the agent
  key, the allocation, fee discount, minimum fee and month-end start day.

---

## 3. User Stories & Acceptance Criteria

### US-01 — Sign in and keep the board private [P0]
**As the** owner, **I want** to sign in with my username and password, **so that** only I
(and my agent team) can see my trading.

```gherkin
Scenario: owner signs in with correct credentials
  Given the owner's username and password are set by the operator
  When the owner signs in with that username and password
  Then the owner receives a pass together with the moment it expires

Scenario: owner signs in with a wrong password
  When the owner signs in with a wrong password
  Then the sign-in is refused as not authenticated
  And no pass is given

Scenario: reading without a pass or key is refused
  Given no pass and no agent key are presented
  When anyone asks for the latest card
  Then the request is refused as not authenticated

Scenario: an expired pass is refused
  Given the owner's pass has expired
  When the owner asks for the latest card
  Then the request is refused as not authenticated

Scenario: the agent team may read with its key
  Given the agent key is presented
  When the agent team asks for the latest card
  Then the latest card is returned

Scenario: the owner cannot publish a card
  Given the owner's valid pass and no agent key
  When the owner tries to publish a card
  Then the publish is refused as not authenticated
```

### US-02 — Publish the Daily Trade Card [P0]
**As the** agent team, **I want** to publish tomorrow's card, **so that** the owner sees
exactly what to place.

```gherkin
Scenario: a well-formed buy order is accepted
  Given a buy order with stop-loss 48, limit 52.5, highest chase 54, take-profit 60 and 90 shares
  When the agent team publishes the card for 2026-11-26
  Then the card for 2026-11-26 is stored with that buy order waiting and 0 shares filled

Scenario: limit equal to the highest chase price is accepted
  Given a buy order with limit 54 and highest chase 54
  When the agent team publishes the card
  Then the card is accepted

Scenario: stop-loss equal to the limit is refused
  Given a buy order with stop-loss 52.5 and limit 52.5
  When the agent team publishes the card
  Then the publish is refused as invalid, saying the stop-loss must be below the limit

Scenario: take-profit equal to the limit is refused
  Given a buy order with limit 52.5 and take-profit 52.5
  When the agent team publishes the card
  Then the publish is refused as invalid, saying the take-profit must be above the limit

Scenario: limit above the highest chase price is refused
  Given a buy order with limit 55 and highest chase 54
  When the agent team publishes the card
  Then the publish is refused as invalid, saying the limit may not exceed the highest chase price

Scenario: a single odd-lot share is accepted
  Given a buy order for 1 share
  When the agent team publishes the card
  Then the card is accepted

Scenario: zero shares or a zero price is refused
  Given a buy order for 0 shares, or with a price of 0
  When the agent team publishes the card
  Then the publish is refused as invalid

Scenario: selling what is held is accepted
  Given the owner holds 50 shares of 3006
  When the agent team publishes a card selling 50 shares of 3006
  Then the card is accepted with that sell order waiting

Scenario: selling one share more than held is refused
  Given the owner holds 50 shares of 3006
  When the agent team publishes a card selling 51 shares of 3006
  Then the publish is refused as invalid, saying more shares are sold than held

Scenario: selling a stock not held is refused
  Given the owner holds no 2330
  When the agent team publishes a card selling 10 shares of 2330
  Then the publish is refused as invalid, saying the stock is not held

Scenario: two sell orders of one stock together exceeding the holding are refused
  Given the owner holds 50 shares of 3006
  When the agent team publishes one card with two sell orders of 30 shares of 3006 each
  Then the publish is refused as invalid, saying more shares are sold than held

Scenario: a no-trade card is accepted
  Given a no-trade card with the note "大盤轉弱" and no orders
  When the agent team publishes it
  Then the card is stored as no-trade with that note

Scenario: a no-trade card carrying an order is refused
  Given a no-trade card with one buy order
  When the agent team publishes it
  Then the publish is refused as invalid

Scenario: a trade card with no orders is refused
  Given a trade card with no buy and no sell orders
  When the agent team publishes it
  Then the publish is refused as invalid

Scenario: re-publishing replaces a card with no fills
  Given the card for 2026-11-26 has no fills
  When the agent team publishes 2026-11-26 again with different orders
  Then the card for 2026-11-26 shows only the new orders

Scenario: re-publishing replaces a card whose only change is a skipped order
  Given the card for 2026-11-26 has one skipped order and no fills
  When the agent team publishes 2026-11-26 again
  Then the card for 2026-11-26 is replaced

Scenario: re-publishing a card with a fill is refused
  Given an order on the card for 2026-11-26 has a fill of 10 shares
  When the agent team publishes 2026-11-26 again
  Then the publish is refused as a conflict
  And the card for 2026-11-26 still shows its original orders and the fill
```

### US-03 — Read cards [P0]
**As the** owner, **I want** to read the latest card and past cards, **so that** I know what
to place and can review history.

```gherkin
Scenario: latest card is the one with the greatest session date
  Given cards exist for 2026-11-25 and 2026-11-26
  When the owner asks for the latest card
  Then the card for 2026-11-26 is returned with every order, its progress and its fills

Scenario: no card yet
  Given no card has ever been published
  When the owner asks for the latest card
  Then the answer is not found

Scenario: a card of a date without one
  Given no card exists for 2026-11-20
  When the owner asks for the card of 2026-11-20
  Then the answer is not found

Scenario: listing a month newest first
  Given cards for 2026-11-25 (1 buy, fully filled) and 2026-11-26 (2 buys, 1 sell, none filled)
  When the owner lists the cards of 2026-11
  Then 2026-11-26 comes first with 2 buys, 1 sell and 0 filled orders
  And 2026-11-25 comes second with 1 buy, 0 sells and 1 filled order
```

### US-04 — Report fills and skip orders [P0]
**As the** owner, **I want** to report what actually filled, or that I skipped an order,
**so that** the board knows my real position.

```gherkin
Scenario: a full fill completes the order
  Given a buy order of 90 shares with nothing filled
  When the owner reports a fill of 90 shares
  Then the order is filled with 90 shares filled

Scenario: a partial fill
  Given a buy order of 90 shares with nothing filled
  When the owner reports a fill of 40 shares
  Then the order is partly filled with 40 shares filled

Scenario: the last shares complete the order
  Given a buy order of 90 shares with 40 filled
  When the owner reports a fill of 50 shares
  Then the order is filled with 90 shares filled and two fills listed

Scenario: overfilling is refused
  Given a buy order of 90 shares with 40 filled
  When the owner reports a fill of 51 shares
  Then the fill is refused as invalid, saying it exceeds the order
  And the order still shows 40 shares filled

Scenario: a skipped order takes no fills
  Given a skipped order
  When the owner reports a fill
  Then the fill is refused as a conflict

Scenario: skipping a waiting order
  Given a waiting order
  When the owner skips it
  Then the order is skipped

Scenario: skipping a partly filled order keeps what filled
  Given a buy order of 90 shares with 40 filled
  When the owner skips it
  Then the order is skipped with 40 shares filled
  And the owner still holds those 40 shares

Scenario: a fully filled order cannot be skipped
  Given a fully filled order
  When the owner skips it
  Then the skip is refused as a conflict

Scenario: an unknown order
  Given no order with the given identity exists
  When the owner reports a fill or skips it
  Then the answer is not found

Scenario: buy fee estimated when left out
  Given a buy order
  When the owner reports a fill of 90 shares at 52.5 without a fee
  Then the fill's amount is 4,725 and its fee is 4

Scenario: minimum fee applies to a tiny buy
  When the owner reports a buy fill of 1 share at 10 without a fee
  Then the fill's fee is 1

Scenario: sell fee includes transaction tax
  Given the owner holds 50 shares of 3006
  When the owner reports a sell fill of 50 shares at 68 without a fee
  Then the fill's amount is 3,400 and its fee is 13

Scenario: a given fee is kept
  When the owner reports a buy fill of 90 shares at 52.5 with fee 20
  Then the fill's fee is 20
```

### US-05 — Holdings [P0]
**As the** owner, **I want** to see what I hold and how closed holdings ended, **so that** I
know my exposure and my results.

```gherkin
Scenario: first buy opens a holding
  Given no holding of 2303
  When a buy fill of 100 shares at 50 with fee 4 is reported
  Then the open holding of 2303 has 100 shares, total cost 5,004 and average cost 50.04

Scenario: a second buy averages the cost
  Given an open holding of 100 shares with total cost 5,004
  When a buy fill of 100 shares at 60 with fee 5 is reported
  Then the holding has 200 shares, total cost 11,009 and average cost 55.045

Scenario: a partial sale realizes profit at average cost
  Given an open holding of 200 shares with total cost 11,009
  When a sell fill of 100 shares at 70 with fee 17 is reported
  Then that sale's realized profit is 1,478.5
  And 100 shares remain with total cost 5,504.5

Scenario: selling the rest closes the holding as a win
  Given the holding above with 100 shares and total cost 5,504.5
  When a sell fill of 100 shares at 50 with fee 12 is reported
  Then that sale's realized profit is −516.5
  And the holding is closed with total realized profit 962 and counted as a win

Scenario: a closed holding with no profit is a loss
  Given a holding whose total realized profit is 0 when it closes
  Then it is counted as a loss

Scenario: selling more than held is refused
  Given an open holding of 50 shares of 3006
  When a sell fill of 60 shares is reported
  Then the fill is refused as a conflict, saying more shares are sold than held

Scenario: buying again after closing opens a new holding
  Given 2303 was held and fully sold
  When a new buy fill of 10 shares of 2303 is reported
  Then a new open holding of 10 shares appears with its own opening date
  And the earlier closed holding is unchanged

Scenario: open holding details
  Given an open holding of 2303 with opening date 2026-11-20, filled from a buy order with stop-loss 48 and take-profit 60
  When the owner lists open holdings on 2026-11-25
  Then 2303 shows stop-loss 48, take-profit 60, opening date 2026-11-20 and 5 days held

Scenario: closed holdings of a month
  Given one holding closed on 2026-11-24 and another closed on 2026-12-02
  When the owner lists closed holdings of 2026-11
  Then only the holding closed on 2026-11-24 is listed, with its realized profit and win or loss
```

### US-06 — Monthly budget and performance [P0]
**As the** owner, **I want** to see how much of this month's money is left and how the month
performed, **so that** I deploy the whole budget and judge the strategy.

```gherkin
Scenario: first month budget
  Given the first activity is in 2026-11 and November's buy fills cost 4,725 plus fee 4
  When the owner reads the summary of 2026-11
  Then allocation is 10,000, carried over 0, spent 4,729 and remaining 5,271

Scenario: unspent money rolls over
  Given November's remaining is 5,271 and nothing happens in December
  When the owner reads the summary of 2026-12
  Then carried over is 5,271 and remaining is 15,271

Scenario: sale proceeds are recycled
  Given December has one sell fill of amount 3,400 with fee 13
  When the owner reads the summary of 2026-12
  Then recycled proceeds are 3,387

Scenario: a buy that uses exactly the remaining money is accepted
  Given the month's remaining is 4,729
  When a buy fill costing 4,729 including its fee is reported
  Then the fill is accepted and the remaining becomes 0

Scenario: a buy that would overspend is refused
  Given the month's remaining is 4,728
  When a buy fill costing 4,729 including its fee is reported
  Then the fill is refused as a conflict, saying the budget is insufficient
  And the remaining stays 4,728

Scenario: a month before the first one does not exist
  Given the first activity is in 2026-11
  When the owner reads the summary of 2026-10
  Then the answer is not found

Scenario: strict on the 21st
  Given today is the 21st of the month in Taipei
  When the owner reads the current month's summary
  Then the mode is strict

Scenario: month-end from the 22nd
  Given today is the 22nd of the month in Taipei
  When the owner reads the current month's summary
  Then the mode is month-end

Scenario: monthly performance
  Given in 2026-11 holding A closed with +962 (win) and holding B closed with −300 (loss)
  And the cost of the shares sold in 2026-11 is 11,009
  When the owner reads the summary of 2026-11
  Then realized profit is 662, 2 holdings closed, 1 win, 1 loss, win rate 50% and realized return 6.01%

Scenario: a month with nothing sold
  Given nothing was sold in the month
  When the owner reads the summary
  Then realized profit is 0, 0 closed, win rate 0 and realized return 0
```

### US-07 — Being able to tell the board is up [P1]
**As the** operator, **I want** a way to check the board is alive, **so that** the hosting
platform can restart it when it is not.

```gherkin
Scenario: alive check
  Given the board is running
  When the operator checks whether it is alive
  Then it answers that it is alive, without needing a pass or key
```

---

## 4. Business Flow & Logic

- **Flow:** evening — agent team publishes the card (may re-publish until something fills) →
  owner reads it → next session the owner places orders at the broker → owner reports each
  fill or skips the order → holdings, budget and performance follow from the reported fills.
- **Core Business Rules:**
  - Card: at most one per session date; trade (≥1 order) or no-trade (0 orders, note).
  - Buy order: prices > 0; shares whole ≥ 1; stop-loss < limit < take-profit;
    limit ≤ highest chase price; estimated amount > 0.
  - Sell order: prices > 0; shares whole ≥ 1; all sells of one stock on a card together ≤
    shares currently held.
  - Replace only while no order of the card has a fill.
  - Order progress: waiting / partly filled / filled derived from filled vs ordered shares;
    skipped set by the owner (waiting or partly filled only; already skipped stays as is).
  - Fill: price > 0, shares whole ≥ 1, fee ≥ 0 when given; filled shares never exceed the
    order; amount = price × shares.
  - Fee estimate (whole dollars, half up): broker = max(round(amount × 0.1425% × discount),
    minimum fee); sell adds round(amount × 0.3%). Defaults discount 0.6, minimum 1.
  - Holding (in order of reporting): average cost incl. buy fees; cost of shares sold =
    total cost × sold ÷ held rounded to the cent, the closing sale takes all remaining cost;
    realized = amount − fee − cost of shares sold; a sell fill may not exceed shares held.
  - Stop-loss / take-profit of an open holding: from the most recent (by session date) buy
    order of that stock with a fill in this holding.
  - Days held: calendar days from the opening date to today (open) or to the closing date
    (closed), Taipei.
  - Budget per month M (Taipei): carried over = remaining of M−1 (0 for the first month);
    recycled = Σ(sale amount − sale fee) of fills in M; spent = Σ(purchase amount + fee) of
    fills in M; remaining = allocation + carried over + recycled − spent ≥ 0. A buy fill is
    refused if it would make the remaining of its month, or of any later existing month,
    negative.
  - First month = month of the earliest session date or fill time; if nothing exists yet,
    the current month. Earlier months: not found.
  - Mode: current month — strict before the start day (22nd), month-end from it; past
    months month-end; future months strict.
  - Performance of M: realized = Σ realized of sell fills in M; closed / wins / losses count
    holdings closed in M; win rate = wins ÷ closed × 100; realized return = realized ÷ cost
    of shares sold in M × 100; both percentages to 2 decimals, 0 when the divisor is 0.
  - Card summary's filled order count = orders fully filled.
- **Edge Cases:**
  - Two fills reported at the same time for the same order or month: only combinations that
    keep every rule (no overfill, no oversell, no overspend) may land.
  - Publishing while a fill is being reported: replacement and the fill must not both land.
  - Storage unavailable: the request fails without partial changes.

---

## 5. UI/UX Design & Interaction

- N/A for the backend. Screens are built separately against the agreed contract; expired
  display of past waiting orders is a screen concern.

---

## 6. Non-Functional Requirements

- **Performance:** single user; every read answers well under a second for years of data.
- **Security:** owner credentials and agent key supplied by the operator, never stored in
  the source; passes expire (24 h default); a wrong password reveals nothing more than
  "not authenticated".
- **Correctness of money:** all money exact, never approximate.
- **Compatibility:** honour the agreed contract (section 4 of the stakeholder document)
  exactly; shapes it leaves open are proposed back to the coordinator.
- **Operability:** settings come from the environment; storage structure is kept in step
  automatically before the board starts serving.
- **Analytics / Tracking:** N/A.

---

## 7. Dependencies & Risks

- **External Dependencies:** the shared database server used by go-trading (a new, separate
  database `trading_board`, created by hand by the operator); the hosting platform's deploy
  flow (same as go-trading).
- **Known Risks:**
  - Fills cannot be corrected in v1; a mistyped fill distorts holdings and budget until a
    later version adds correction.
  - Holding and monthly-summary shapes are not fully spelled out in the contract; the
    frontend may guess differently → contract change request to the coordinator.
  - A backdated fill (fill time in an earlier month) changes that month's and later months'
    figures; allowed by design.

---

## 8. Appendix

- Stakeholder document: `docs/requirements/trading-board.md`
- Brief: `BRIEF.md` (assumptions A1–A15)
