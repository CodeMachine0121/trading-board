# Contract Traceability Matrix — Trading Board Web

Contract: PRD.md (§3 acceptance criteria + §4 business rules and edge cases)
Design map: ARCH.md
Implementation: `frontend/` (app/, server/mock/ for US-09)
Oracle: Acceptance Criteria (53 clauses: 41 AC + 12 BR). NFRs (§6: load time, browser
support, analytics N/A) are not statically auditable and are excluded; session-is-local
security is covered by AC-5.

Ceiling: static conformance audit. Test assertions and code paths are judged against
the spec's expected outcome; no invented scenarios were executed. Single mapped tests
were re-run only as corroboration (BR-4, AC-11).

`Test audit` / `Code audit` are Claude's shadow verdict (recorded before Jev ran).
`Status` / `Route` are as returned by `jev_audit.py`.

## Clauses

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Jev test | Jev code | Status | Route |
|----|--------|------------------------|------|------|------------|------------|----------|----------|--------|-------|
| AC-1 | US-01 Correct credentials open Today | Owner signed in with correct credentials ends up on Today | `app/domain/service/user-session-service.ts:13-16` | `tests/application/user-session-application.spec.ts:34-43` | shallow | produces-oracle | shallow 0.53 | produces-oracle 0.67 | 🟠 mis-asserted | escalate (low confidence) |
| AC-2 | US-01 Wrong password is refused | Owner stays on sign-in and sees 帳號或密碼錯誤 | `app/infrastructure/proxy/board-api-proxy.ts:84-96` | `tests/infrastructure/proxy/auth-proxy.spec.ts:52-60` | asserts-oracle | produces-oracle | asserts-oracle 0.91 | produces-oracle 0.43 | ✅ conforms | escalate (low confidence) |
| AC-3 | US-01 A protected screen without signing in leads to sign-in | Opening History unsigned lands on sign-in | `app/middleware/auth.global.ts:7-18` | `tests/application/user-session-application.spec.ts:53-59` | shallow | produces-oracle | no-test 0.17 | produces-oracle 0.80 | 🟡 partial | escalate (disagree) |
| AC-4 | US-01 An expired sign-in leads back to sign-in | Any screen/action with expired sign-in lands on sign-in | `app/domain/models/domains/session-domain.ts:1-11` | `tests/application/user-session-application.spec.ts:53-59` | shallow | produces-oracle | asserts-oracle 0.71 | unclear 0.48 | ❔ unclear | escalate (disagree) |
| AC-5 | US-01 Signing out | After 登出 owner is on sign-in and Today leads to sign-in | `app/layouts/default.vue:16-19` | `tests/application/user-session-application.spec.ts:61-67` | shallow | produces-oracle | shallow 0.79 | produces-oracle 0.51 | 🟠 mis-asserted | escalate (low confidence) |
| AC-6 | US-02 A trade card shows buy and sell sections | 買進 shows 聯電 with limit, max chase, shares, est. amount, stop loss, take profit, reason, status; 賣出 shows 晶豪科 with limit, shares, exit reason, reason, status | `app/domain/models/entities/buy-order.ts:34-59` | `tests/application/trade-card-application.spec.ts:45-71` | asserts-oracle | produces-oracle | asserts-oracle 0.89 | produces-oracle 0.65 | ✅ conforms | recheck |
| AC-7 | US-02 A trade card without sells | 賣出 section shows 無賣出委託 | `app/domain/models/domains/trade-card-domain.ts:20-40` | `tests/application/trade-card-application.spec.ts:73-81` | asserts-oracle | produces-oracle | asserts-oracle 0.99 | produces-oracle 0.51 | ✅ conforms | escalate (low confidence) |
| AC-8 | US-02 A month-end allocation buy is marked | Order marked 月底配置 | `app/domain/models/entities/buy-order.ts:34-45` | `tests/application/trade-card-application.spec.ts:83-93` | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.76 | ✅ conforms | recheck |
| AC-9 | US-02 A no-trade card | 今日不買 banner with 大盤轉弱, no order sections | `app/domain/models/domains/trade-card-domain.ts:20-40` | `tests/application/trade-card-application.spec.ts:95-106` | asserts-oracle | produces-oracle | asserts-oracle 0.99 | produces-oracle 0.79 | ✅ conforms | recheck |
| AC-10 | US-02 No card has been published yet | Today shows 尚未發布任何交易卡 | `app/infrastructure/proxy/board-api-proxy.ts:58-70` | `tests/application/trade-card-application.spec.ts:108-113` | shallow | produces-oracle | asserts-oracle 0.63 | produces-oracle 0.96 | ✅ conforms | escalate (disagree) |
| AC-11 | US-03 Form starts at the order's limit price and all shares | Price 52.5, shares 90, fee empty | `app/domain/models/entities/buy-order.ts:52-58` | `tests/components/organisms/OrderCard.spec.ts:34-42` | asserts-oracle | produces-oracle | asserts-oracle 1.00 | unclear 0.25 | ❔ unclear | escalate (disagree) |
| AC-12 | US-03 Form starts at the remaining shares | Shares show 30 when 60 of 90 filled | `app/domain/models/domains/order-progress-domain.ts:23-41` | `tests/application/trade-card-application.spec.ts:115-122` | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.79 | ✅ conforms | recheck |
| AC-13 | US-03 Filling exactly the remaining shares is accepted | Order shows 已成交 with the new fill listed | `app/domain/models/domains/fill-draft-domain.ts:28-58` | `tests/application/trade-card-application.spec.ts:148-160` | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.76 | ✅ conforms | recheck |
| AC-14 | US-03 More than the remaining shares is refused by the form | Form refuses with 股數不可超過剩餘 30 股, nothing reported | `app/domain/models/domains/fill-draft-domain.ts:28-45` | `tests/application/trade-card-application.spec.ts:162-171` | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.85 | ✅ conforms | accept |
| AC-15 | US-03 Zero shares or zero price is refused by the form | Refuses with 股數須為 1 以上的整數 or 價格須大於 0 | `app/domain/models/domains/fill-draft-domain.ts:28-45` | `tests/application/trade-card-application.spec.ts:162-171` | asserts-oracle | produces-oracle | asserts-oracle 0.93 | produces-oracle 0.85 | ✅ conforms | accept |
| AC-16 | US-03 The board refuses a fill that exceeds the month's budget | Owner sees board reason; order keeps previous status and filled shares | `app/infrastructure/proxy/board-api-proxy.ts:84-103` | `tests/application/trade-card-application.spec.ts:173-181` | asserts-oracle | produces-oracle | asserts-oracle 0.95 | produces-oracle 0.41 | ✅ conforms | escalate (low confidence) |
| AC-17 | US-03 Leaving the fee empty lets the board estimate it | New fill listed with board-estimated fee | `app/infrastructure/proxy/order-proxy.ts:11-22` | `tests/infrastructure/proxy/order-proxy.spec.ts:26-35` | asserts-oracle | produces-oracle | asserts-oracle 0.63 | produces-oracle 0.77 | ✅ conforms | recheck |
| AC-18 | US-03 Reporting refreshes the budget strip | Budget strip shows the new remaining amount | `app/pages/index.vue:14-26` | `tests/composables/use-trade-card-actions.spec.ts:44-53` | shallow | produces-oracle | shallow 0.27 | produces-oracle 0.72 | 🟠 mis-asserted | escalate (low confidence) |
| AC-19 | US-04 Waiting order of today's session | 待成交 with 回報成交 and 略過 | `app/domain/models/domains/order-progress-domain.ts:23-41` | `tests/domain/models/domains/order-progress-domain.spec.ts:5-22` | asserts-oracle | produces-oracle | asserts-oracle 0.63 | produces-oracle 0.57 | ✅ conforms | escalate (low confidence) |
| AC-20 | US-04 Partly bought order | 部分成交 with 回報成交 and 略過 | `app/domain/models/domains/order-progress-domain.ts:23-41` | `tests/domain/models/domains/order-progress-domain.spec.ts:5-22` | asserts-oracle | produces-oracle | asserts-oracle 0.76 | produces-oracle 0.45 | ✅ conforms | escalate (low confidence) |
| AC-21 | US-04 Fully bought order | 已成交, no actions | `app/domain/models/domains/order-progress-domain.ts:23-41` | `tests/domain/models/domains/order-progress-domain.spec.ts:5-22` | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.99 | ✅ conforms | accept |
| AC-22 | US-04 Skipping an order | After 略過, 已略過 and no actions | `app/domain/service/trade-card-service.ts:57-61` | `tests/application/trade-card-application.spec.ts:183-193` | asserts-oracle | produces-oracle | asserts-oracle 0.64 | produces-oracle 0.77 | ✅ conforms | recheck |
| AC-23 | US-04 Waiting order of a past session is expired | 已過期, no actions | `app/domain/models/domains/order-progress-domain.ts:23-41` | `tests/domain/models/domains/order-progress-domain.spec.ts:5-22` | asserts-oracle | produces-oracle | asserts-oracle 1.00 | produces-oracle 0.37 | ✅ conforms | escalate (low confidence) |
| AC-24 | US-04 Partly bought order of a past session stays reportable | 部分成交 and still offers 回報成交 | `app/domain/models/domains/order-progress-domain.ts:23-41` | `tests/domain/models/domains/order-progress-domain.spec.ts:5-22` | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.76 | ✅ conforms | recheck |
| AC-25 | US-05 Ordinary day in the strict window | 剩餘 NT$5,275 / 月配額 NT$10,000, bar 52.75%, 嚴格期 | `app/domain/models/domains/month-summary-domain.ts:28-43` | `tests/domain/models/domains/month-summary-domain.spec.ts:5-18` | asserts-oracle | produces-oracle | asserts-oracle 0.96 | unclear 0.43 | ❔ unclear | escalate (disagree) |
| AC-26 | US-05 Month-end window | Badge 月底配置期 | `app/domain/models/domains/month-summary-domain.ts:11-43` | `tests/domain/models/domains/month-summary-domain.spec.ts:5-18` | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.97 | ✅ conforms | accept |
| AC-27 | US-05 Everything spent | 剩餘 NT$0 and bar empty | `app/domain/models/domains/month-summary-domain.ts:28-43` | `tests/domain/models/domains/month-summary-domain.spec.ts:5-18` | asserts-oracle | produces-oracle | asserts-oracle 0.17 | produces-oracle 0.64 | ✅ conforms | escalate (low confidence) |
| AC-28 | US-05 Carried over and recycled money | 剩餘 NT$15,000 / 月配額 NT$10,000, bar full | `app/domain/models/domains/month-summary-domain.ts:28-43` | `tests/domain/models/domains/month-summary-domain.spec.ts:5-18` | asserts-oracle | produces-oracle | asserts-oracle 0.92 | unclear 0.20 | ❔ unclear | escalate (disagree) |
| AC-29 | US-06 Current month is shown first | November 2026 selected; cards newest first with date, 交易/今日不買, buy/sell counts, completed count | `app/domain/service/trade-card-service.ts:31-40` | `tests/application/trade-card-application.spec.ts:196-212` | asserts-oracle | produces-oracle | asserts-oracle 0.87 | produces-oracle 0.61 | ✅ conforms | recheck |
| AC-30 | US-06 Month without cards | 這個月沒有交易卡 | `app/components/organisms/CardSummaryList.vue:15-22` | `tests/components/organisms/CardSummaryList.spec.ts:9-13` | asserts-oracle | produces-oracle | asserts-oracle 1.00 | produces-oracle 1.00 | ✅ conforms | accept |
| AC-31 | US-06 Opening a past card | Same layout as Today; waiting orders 已過期 without actions | `app/pages/history/[tradeDate].vue:1-20` | `tests/application/trade-card-application.spec.ts:224-236` | asserts-oracle | produces-oracle | asserts-oracle 0.97 | unclear 0.73 | ❔ unclear | escalate (disagree) |
| AC-32 | US-07 Open positions | 聯電 with shares 90, avg cost 52.58, total cost, stop loss, take profit, opened date, days held | `app/domain/models/domains/position-domain.ts:15-32` | `tests/domain/models/domains/position-domain.spec.ts:5-18` | asserts-oracle | produces-oracle | asserts-oracle 0.51 | produces-oracle 0.17 | ✅ conforms | escalate (low confidence) |
| AC-33 | US-07 A profitable closed position | Red, labelled 獲利 | `app/domain/models/domains/realized-pnl-domain.ts:14-28` | `tests/domain/models/domains/position-domain.spec.ts:28-40` | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.75 | ✅ conforms | recheck |
| AC-34 | US-07 A losing closed position | Green, labelled 虧損 | `app/domain/models/domains/realized-pnl-domain.ts:14-28` | `tests/domain/models/domains/position-domain.spec.ts:28-40` | asserts-oracle | produces-oracle | asserts-oracle 0.83 | produces-oracle 0.51 | ✅ conforms | escalate (low confidence) |
| AC-35 | US-07 A break-even closed position | Neutral colour, labelled 虧損 | `app/domain/models/domains/realized-pnl-domain.ts:14-28` | `tests/domain/models/domains/position-domain.spec.ts:28-40` | asserts-oracle | produces-oracle | asserts-oracle 0.91 | produces-oracle 0.52 | ✅ conforms | escalate (low confidence) |
| AC-36 | US-07 Nothing held | 目前沒有持股 | `app/components/organisms/OpenPositionList.vue:10-17` | `tests/components/organisms/OpenPositionList.spec.ts:8-11` | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.99 | ✅ conforms | accept |
| AC-37 | US-08 Budget breakdown and performance | Allocation, carried over, recycled, spent, remaining, window; realized P&L, 3 closed, 2 wins, 1 loss, win rate 66.67% | `app/domain/models/domains/month-summary-domain.ts:45-80` | `tests/domain/models/domains/month-summary-domain.spec.ts:32-60` | asserts-oracle | produces-oracle | asserts-oracle 0.88 | unclear 0.31 | ❔ unclear | escalate (disagree) |
| AC-38 | US-08 No closed positions | Win rate and realized return show — | `app/domain/models/domains/month-summary-domain.ts:45-80` | `tests/domain/models/domains/month-summary-domain.spec.ts:62-70` | asserts-oracle | produces-oracle | asserts-oracle 0.80 | produces-oracle 0.67 | ✅ conforms | recheck |
| AC-39 | US-08 Month before the board's first month | 這個月沒有資料 | `app/infrastructure/proxy/month-summary-proxy.ts:34-51` | `tests/application/month-summary-application.spec.ts:50-58` | asserts-oracle | produces-oracle | asserts-oracle 0.99 | produces-oracle 0.63 | ✅ conforms | recheck |
| AC-40 | US-09 Sample data preview | Today shows sample card with 聯電 buy and 晶豪科 sell; fill on 聯電 -> 部分成交/已成交 and remaining budget drops | `server/mock/mock-board.ts:96-160` | `tests/server/mock/mock-board.spec.ts:37-55` | asserts-oracle | produces-oracle | asserts-oracle 0.65 | unclear 0.40 | ❔ unclear | escalate (disagree) |
| AC-41 | US-09 Switching to the live board | All data comes from the live board when set | `app/infrastructure/proxy/board-api-proxy.ts:72-82` | `tests/infrastructure/proxy/board-api-proxy.spec.ts:39-51` | asserts-oracle | produces-oracle | asserts-oracle 0.80 | produces-oracle 0.79 | ✅ conforms | recheck |
| BR-1 | Displayed order state rule | skipped→已略過; all filled→已成交; >0 filled→部分成交; none filled and past→已過期; else 待成交 | `app/domain/models/domains/order-progress-domain.ts:1-41` | `tests/domain/models/domains/order-progress-domain.spec.ts:5-28` | asserts-oracle | produces-oracle | asserts-oracle 0.88 | produces-oracle 0.95 | ✅ conforms | accept |
| BR-2 | Actions only for 待成交 and 部分成交 | Report/skip offered only for those two states | `app/domain/models/domains/order-progress-domain.ts:23-41` | `tests/domain/models/domains/order-progress-domain.spec.ts:5-22` | asserts-oracle | produces-oracle | asserts-oracle 0.41 | produces-oracle 0.43 | ✅ conforms | escalate (low confidence) |
| BR-3 | Fill form validation | price>0; whole shares 1..remaining; fee empty or >=0; fill time required (defaults now) | `app/domain/models/domains/fill-draft-domain.ts:1-59` | `tests/domain/models/domains/fill-draft-domain.spec.ts:1-48` | asserts-oracle | produces-oracle | asserts-oracle 0.79 | produces-oracle 0.63 | ✅ conforms | recheck |
| BR-4 | Budget bar formula | remaining ÷ (allocation+carried+recycled), clamped 0–100, empty when denominator 0 | `app/domain/models/domains/month-summary-domain.ts:28-43` | `tests/domain/models/domains/month-summary-domain.spec.ts:5-29` | asserts-oracle | produces-oracle | asserts-oracle 0.55 | diverges 0.67 | 🔴 violation | escalate (violation) |
| BR-5 | Profit colouring | >0 red 獲利, <0 green 虧損, =0 neutral counts as 虧損 | `app/domain/models/domains/realized-pnl-domain.ts:1-29` | `tests/domain/models/domains/position-domain.spec.ts:28-40` | asserts-oracle | produces-oracle | asserts-oracle 0.95 | produces-oracle 0.84 | ✅ conforms | recheck |
| BR-6 | Win rate / return two decimals; — when none closed | 66.67%; — with 0 closed | `app/domain/models/domains/month-summary-domain.ts:45-80` | `tests/domain/models/domains/month-summary-domain.spec.ts:47-70` | asserts-oracle | produces-oracle | asserts-oracle 0.80 | produces-oracle 0.17 | ✅ conforms | escalate (low confidence) |
| BR-7 | Money exact with NT$ and thousands separators | Amounts shown as reported, NT$ with thousands separators, no rounding drift | `app/utilities/decimal-formatter.ts:1-40` | `tests/utilities/decimal-formatter.spec.ts:1-45` | asserts-oracle | produces-oracle | asserts-oracle 0.88 | produces-oracle 0.76 | ✅ conforms | recheck |
| BR-8 | Today and this month are Asia/Taipei | Date boundaries follow Taipei time | `app/domain/models/domains/taipei-calendar-domain.ts:1-63` | `tests/domain/models/domains/taipei-calendar-domain.spec.ts:1-45` | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.97 | ✅ conforms | accept |
| BR-9 | Edge: board unreachable | 無法連線，請稍後再試 message with a retry | `app/infrastructure/proxy/board-api-proxy.ts:84-103` | `tests/infrastructure/proxy/board-api-proxy.spec.ts:92-98` | shallow | produces-oracle | asserts-oracle 0.29 | produces-oracle 0.60 | ✅ conforms | escalate (disagree) |
| BR-10 | Edge: board refuses an action | Its message shown next to the action; order left as it was | `app/composables/use-trade-card-actions.ts:19-50` | `tests/composables/use-trade-card-actions.spec.ts:64-75` | asserts-oracle | produces-oracle | asserts-oracle 1.00 | produces-oracle 0.77 | ✅ conforms | recheck |
| BR-11 | Edge: sign-in stops being accepted mid-session | Back to sign-in | `app/infrastructure/proxy/board-api-proxy.ts:84-96` | `tests/infrastructure/proxy/board-api-proxy.spec.ts:63-76` | shallow | produces-oracle | shallow 0.45 | unclear 0.24 | ❔ unclear | escalate (disagree) |
| BR-12 | Edge: double-tap on submit | Only one report sent (button disabled while sending) | `app/components/molecules/FillReportForm.vue:85-100` | `tests/components/molecules/FillReportForm.spec.ts:40-42` | asserts-oracle | produces-oracle | asserts-oracle 0.80 | produces-oracle 0.55 | ✅ conforms | escalate (low confidence) |

### Notes

- **AC-1** — Navigation to Today lives in pages/login.vue (pages are wiring-only and untested by design); tests pin session storage, not the landing screen.
- **AC-3** — Redirect lives in middleware/auth.global.ts, which has no test; only isSignedIn()=false is asserted.
- **AC-4** — Redirect after expiry is wired in plugins/dependencies.ts (untested); hook invocation and isSignedIn()=false are asserted.
- **AC-5** — Layout navigation to /login untested; clearing the stored session is asserted.
- **AC-10** — Empty-state text is in pages/index.vue (untested); null result asserted.
- **AC-11** — Re-audited after adding OrderCard.spec.ts:44 and trade-card-application.spec.ts:115 (first run: mis-asserted — existing tests only covered a 30-share remainder). Second run shown.
- **AC-18** — Budget reload is page wiring (pages/index.vue:18); the composable test only asserts the callback fires.
- **BR-4** — Recheck: formula, clamp and zero-denominator path match the oracle; mapped test re-run green; denominator mutant was killed in /implement FALSIFY. Claude believes Jev is wrong here; status left as returned.
- **BR-9** — Retry button lives in pages/index.vue (untested).
- **BR-11** — Navigation part is plugin wiring (untested).

## Orphans (code with no clause)

| Code | Description | Verdict |
|------|-------------|---------|
| `app/pages/history/[tradeDate].vue:40-45` | 「找不到這一天的交易卡」 when a history URL names a date without a card | undocumented (sensible fallback) |
| `app/domain/models/domains/trade-card-domain.ts:37` | 「發布於 …」 published-at line on the card | undocumented (contract field `publishedAt`) |
| `app/domain/models/domains/taipei-calendar-domain.ts:44-62` | Future / malformed `?month=` falls back to the current month; month kept in the URL | undocumented |
| `app/pages/history/[tradeDate].vue:24-32` | 「‹ 回到歷史」 back link | undocumented |
| `server/mock/mock-board.ts` | Dev-only network mock also enforces sell-quantity, skip and budget rules | supports US-09; not shipped |

None implements an Out-of-Scope item (no fill edit/delete, no publishing, no live prices).

## Summary

- Conforms: 40/53 clauses ✅ (75.5%)
- Violations: BR-4  (Jev only; Claude's recheck disagrees — see Notes)
- Mis-asserted: AC-1, AC-5, AC-18  (page/layout wiring untested by design)
- Partial: AC-3
- Gaps: —
- Unclear: AC-4, AC-11, AC-25, AC-28, AC-31, AC-37, AC-40, BR-11
- Orphans: 5
- Jev: ran on 53 clauses (AC-11 twice)
- Needs James: AC-1 (low confidence), AC-2 (low confidence), AC-3 (disagree), AC-4 (disagree), AC-5 (low confidence), AC-7 (low confidence), AC-10 (disagree), AC-11 (disagree), AC-16 (low confidence), AC-18 (low confidence), AC-19 (low confidence), AC-20 (low confidence), AC-23 (low confidence), AC-25 (disagree), AC-27 (low confidence), AC-28 (disagree), AC-31 (disagree), AC-32 (low confidence), AC-34 (low confidence), AC-35 (low confidence), AC-37 (disagree), AC-40 (disagree), BR-2 (low confidence), BR-4 (violation), BR-6 (low confidence), BR-9 (disagree), BR-11 (disagree), BR-12 (low confidence)
