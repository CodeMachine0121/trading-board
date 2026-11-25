# Code-first oracle (pinned before BUILD; frozen)

Expected outcomes derived from PRD §3/§4 only. Owner component per ARCH.md.

| Case | PRD scenario | Expected outcome | Owner |
|---|---|---|---|
| C1 | US-01 Correct credentials open Today | `signIn(owner/pw)` stores session (token, expiry) and `isSignedIn()` is true; login page emits navigation to `/` | UserSessionApplication, LoginForm |
| C2 | US-01 Wrong password is refused | login answered 401 → `CredentialsRejectedError`; nothing stored; LoginForm shows 「帳號或密碼錯誤」 | AuthProxy, LoginForm |
| C3 | US-01 Protected screen without signing in | no stored session → `isSignedIn()` false (guard redirects to `/login`) | UserSessionApplication |
| C4 | US-01 Expired sign-in | stored session with expiresAt ≤ now → `isSignedIn()` false; any board call answered 401 → `SignedOutError` and the signed-out hook is called once | SessionDomain, BoardApiProxy |
| C5 | US-01 Signing out | `signOut()` clears the stored session; `isSignedIn()` false | UserSessionApplication |
| C6 | US-02 Trade card shows buy and sell sections | card DTO: buy section 1 order titled 「聯電 2303」 with rows 限價 52.5, 最高追價 54, 股數 90, 預估金額 NT$4,725, 停損 48, 停利 60; sell section 1 order 「晶豪科 3006」 with rows 限價 68, 股數 50, 出場原因 停利 | TradeCardDomain, BuyOrder/SellOrder |
| C7 | US-02 Trade card without sells | sell section empty text 「無賣出委託」 | TradeCardDomain |
| C8 | US-02 Month-end allocation | buy order with isMonthEndAllocation true carries tag 「月底配置」; false → no tag | BuyOrder |
| C9 | US-02 No-trade card | DTO isNoTrade true, banner 「今日不買」, note 「大盤轉弱」, no orders rendered | TradeCardDomain, TradeCardView |
| C10 | US-02 No card yet | latest answered 404 → application returns null → Today shows 「尚未發布任何交易卡」 | TradeCardProxy, Today |
| C11 | US-03 Form defaults | order 90 sh @52.5, 0 filled → form price "52.5", shares "90", fee "" | OrderProgressDomain, FillReportForm |
| C12 | US-03 Defaults to remaining | 90 sh, 60 filled → shares "30" | OrderProgressDomain |
| C13 | US-03 Exactly remaining accepted | 30 remaining, report 30 @ 52 → proxy called with shares 30, price "52"; returned card shows 「已成交」 | FillDraftDomain, TradeCardService |
| C14 | US-03 Over remaining refused | 31 with 30 remaining → `FillDraftInvalidError` 「股數不可超過剩餘 30 股」; proxy not called | FillDraftDomain |
| C15 | US-03 Zero refused | shares 0 → 「股數須為 1 以上的整數」; price 0 → 「價格須大於 0」; proxy not called | FillDraftDomain |
| C16 | US-03 Board refuses over budget | proxy answered 400/409 with message M → `BoardRejectedError` with message M; OrderCard shows M; card not re-read | BoardApiProxy, OrderCard |
| C17 | US-03 Fee empty → estimated | fee "" → request body has no `fee`; fill row shows returned fee | FillDraftDomain, OrderProxy, Fill |
| C18 | US-03 Refresh budget strip | Today reloads budget strip after a successful report | Today page (wiring) |
| C19 | US-04 Waiting today | pending, today's date → 「待成交」, actionable true | OrderProgressDomain |
| C20 | US-04 Partly bought | 30/90 partiallyFilled → 「部分成交」, actionable true | OrderProgressDomain |
| C21 | US-04 Fully bought | 90/90 filled → 「已成交」, actionable false | OrderProgressDomain |
| C22 | US-04 Skip | skip → proxy skip called with order id, card re-read; skipped → 「已略過」, actionable false | TradeCardService, OrderProgressDomain |
| C23 | US-04 Past waiting expired | pending, trade date yesterday → 「已過期」, actionable false | OrderProgressDomain |
| C24 | US-04 Past partly bought | partiallyFilled 30/90, yesterday → 「部分成交」, actionable true | OrderProgressDomain |
| C25 | US-05 Strict window | alloc 10000, carried 0, recycled 0, spent 4725, remaining 5275, strict → 「剩餘 NT$5,275」, 「月配額 NT$10,000」, percent 52.75, 「嚴格期」 | MonthSummaryDomain |
| C26 | US-05 Month-end | mode monthEnd → 「月底配置期」 | MonthSummaryDomain |
| C27 | US-05 Everything spent | remaining 0 → 「剩餘 NT$0」, percent 0 | MonthSummaryDomain |
| C28 | US-05 Carried + recycled | alloc 10000, carried 2000, recycled 3000, spent 0, remaining 15000 → 「剩餘 NT$15,000」, percent 100 | MonthSummaryDomain |
| C29 | US-06 Current month first | now 2026-11-25 Taipei → month "2026-11", label 「2026 年 11 月」, next null; rows newest first with 交易/今日不買, buy/sell counts, filled count | TaipeiCalendarDomain, TradeCardService |
| C30 | US-06 Month without cards | empty list → 「這個月沒有交易卡」 | CardSummaryList |
| C31 | US-06 Opening past card | fetchCard("2026-11-20") on 2026-11-25 → pending orders 「已過期」 | TradeCardService |
| C32 | US-07 Open positions | 聯電 90 sh avg 52.58 → rows shares 90, 平均成本 52.58, total cost, stop loss, take profit, opened date, days held | PositionDomain |
| C33 | US-07 Profitable closed | realizedPnl 1200 → tone profit, label 獲利, 「+NT$1,200」 | PositionDomain |
| C34 | US-07 Losing closed | −800 → tone loss, label 虧損, 「-NT$800」 | PositionDomain |
| C35 | US-07 Break-even | 0 → tone neutral, label 虧損 | PositionDomain |
| C36 | US-07 Nothing held | empty → 「目前沒有持股」 | OpenPositionList |
| C37 | US-08 Breakdown + performance | 3 closed, 2 wins, 1 loss → win rate 「66.67%」; breakdown shows all six budget items | MonthSummaryDomain |
| C38 | US-08 No closed | closedPositionCount 0 → win rate 「—」, realized return 「—」 | MonthSummaryDomain |
| C39 | US-08 Month before first | summary 404 → page DTO summary null → 「這個月沒有資料」 | MonthSummaryProxy, MonthSummaryPanel |
| C40 | US-09 Sample data | mock: login owner/owner → token; latest card has 聯電 buy + 晶豪科 sell; fill 90 → filled and budget remaining drops | server mock |
| C41 | US-09 Switch to live | proxies request `${apiBaseUrl}/cards/latest` with the configured base URL | BoardApiProxy |
| C42 | §4 Unreachable | network failure / 5xx → `BoardUnreachableError` | BoardApiProxy |
| C43 | §4 rule: budget bar denominator (added in FALSIFY) | alloc 10000, carried 2000, recycled 0, remaining 6000 → percent 50 (not 60) | MonthSummaryDomain |
