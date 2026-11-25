# Trading Board frontend — delivery report

Branch `feat/frontend` (app repo) · branch `feat/trading-board-frontend` (go-trading-deploy).
Nothing was pushed or merged. Workflow run unattended:
`/clarify → /prd → /architecture → /implement (code-first) → /improve-codebase → /contract`.
Specs live in `frontend/.sdd/` (`UL-MAP.md`, `2026-11-25-trading-board-web/{BRIEF,PRD,ARCH,ORACLE,CONTRACT}.md`).

## What was built

**App (`frontend/`)** — Nuxt 3 SPA (`ssr: false`, `nuxt generate` → nginx), Bun, TypeScript
strict, Traditional Chinese, mobile-first, red = profit / green = loss. Mirrors
go-trading-frontend's layout, lint (ESLint layer rules, stylelint token/BEM rules, design
token name check), Vitest setup, Dockerfile and nginx.conf.

| Screen | Route | What it does |
| --- | --- | --- |
| 登入 | `/login` | Username + password → JWT kept in localStorage; wrong password shows 「帳號或密碼錯誤」 |
| 今日 | `/` | Budget strip (remaining / allocation, bar, 嚴格期 / 月底配置期) + latest card; 今日不買 banner; buy / sell order cards with every field, status badge, fills; 回報成交 (price = limit, shares = remaining, fee optional, time = now Taipei) and 略過 |
| 歷史 | `/history?month=` → `/history/:tradeDate` | Month stepper → card list (newest first) → same card view; past waiting orders show 已過期 read-only |
| 持股 | `/positions?month=` | Open positions; closed positions for a month coloured by realized result |
| 月報 | `/summary?month=` | Budget breakdown (§3.6) and performance (§3.7) |

Architecture (per `frontend/.claude/rules`): `.vue` → Application → Domain Service →
`I*Proxy` interfaces; one Proxy per resource (`AuthProxy`, `TradeCardProxy`, `OrderProxy`,
`PositionProxy`, `MonthSummaryProxy`, `SessionStorageProxy`) over a shared `BoardApiProxy`
that translates 401 / 4xx / 5xx-or-network into sentinel errors. Business display rules live
in Domain Models (`OrderProgressDomain`, `FillDraftDomain`, `MonthSummaryDomain`,
`PositionDomain`, `RealizedPnlDomain`, `TaipeiCalendarDomain`, `TradeCardDomain`); money is
`decimal.js` end to end. Components only see DTOs.

**Mock layer** — `frontend/server/routes/mock/api/v1/[...path].ts` + `server/mock/mock-board.ts`:
a Nitro server route implementing contract §4 in memory (seeded: closed 力積電 win, open 晶豪科,
a 今日不買 card, an expired order, tomorrow's card with 聯電 / 南亞科(月底配置) buys and a 晶豪科
sell; fee estimate, budget, average cost, carry-over). It exists only under `nuxt dev`.
Switching is configuration only: `NUXT_PUBLIC_API_BASE_URL` (default `/mock/api/v1`,
credentials `owner` / `owner`; production build arg
`https://trading-board-api.coding-afternoon.com/api/v1`).

**CI** — `.github/workflows/frontend-ci.yml` (repo root): path-filtered to `frontend/**`;
eslint, stylelint, token check, typecheck, tests with coverage, `nuxt generate`; on `main`:
build + push `ghcr.io/codemachine0121/trading-board-frontend:{sha7,latest}`, prune to 5
versions, bump `trading-board-frontend/kustomization.yaml` in go-trading-deploy (rebase-retry).

**Deploy (go-trading-deploy, commit `e0f265a`)** — `trading-board-frontend/{deployment,service,ingress,kustomization}.yaml`
(2 replicas, rolling update, `/health` probes, non-root, host
`trading-board.coding-afternoon.com`, tag placeholder `0000000`), Argo CD Application appended
to `argocd/applications.yaml`, README folder table + tunnel route row. No ConfigMap and no
Secret: the reference bakes the API URL at build time and the frontend holds no secrets.

## Verification

| Check | Result |
| --- | --- |
| `bun run verify` (eslint, stylelint, tokens, typecheck, tests) | green |
| Tests | 36 files, 213 tests, all passing |
| Coverage (v8, `app/**` minus wiring-only plugins/pages/layouts/middleware/app.vue) | 100 % statements / branches / functions / lines |
| Mutation (manual, 31 mutants on key rules) | 31/31 killed (one survivor found and fixed by adding a case) |
| `nuxt generate` with the production URL | builds; URL baked into HTML, no mock URL in output |
| `docker build` + run | image builds; `/health` → `ok`; `/history/2026-11-20` → SPA shell (200) |
| Mock API smoke (curl) | login, 401s, latest card, positions, month summary behave per contract |
| `kubectl kustomize trading-board-frontend` (offline render) | renders Service / Deployment / Ingress with the expected image and host |
| Browser visual check | **not done** — the Chrome extension was not connected |

`server/mock/**` is excluded from the coverage gate (dev-only, not shipped) and has its own
smoke spec (`tests/server/mock/mock-board.spec.ts`).

## /contract verdict

53 clauses (41 AC + 12 BR), Jev ran on all of them: **40 conforms (75.5 %)**,
3 mis-asserted (AC-1, AC-5, AC-18), 1 partial (AC-3), 8 unclear (AC-4, AC-11, AC-25, AC-28,
AC-31, AC-37, AC-40, BR-11), 1 violation flagged by Jev (BR-4). No gaps.

- The mis-asserted / partial / several unclear clauses share one cause: their last step
  (navigate to Today / to sign-in, reload the budget strip, empty-state text) lives in pages,
  layout, middleware or the composition root, which are wiring-only and untested by the
  reference project's convention. The logic underneath is tested.
- BR-4 (budget bar formula): Jev says the code diverges (0.67). I re-read it and re-ran its
  test; the formula, clamp and zero-denominator path match the PRD and the denominator mutant
  is killed. I believe this is a Jev false positive — **please confirm**.
- AC-11 was mis-asserted on the first pass (no test pinned the untouched-order defaults);
  a test was added after the audit and the clause re-audited.

Full matrix: `frontend/.sdd/2026-11-25-trading-board-web/CONTRACT.md`.

## Assumptions (not covered by the requirements)

1. Positions, month summary and card-list response shapes (see CCRs below).
2. Budget bar = remaining ÷ (allocation + carried over + recycled), 0–100 %.
3. Break-even closed positions are labelled 虧損 but coloured neutral.
4. A month before the first month (404) shows 「這個月沒有資料」; months in the future are not selectable.
5. Fill time defaults to now (Taipei); entered times are interpreted as Taipei time.
6. Only fully-waiting past orders are 已過期; partly filled past orders stay reportable.
7. After a fill or skip the card is re-read rather than patched from the "updated order" response.
8. Win rate is computed on the web from `winCount ÷ closedPositionCount`.
9. Single locale: no vue-i18n (the reference has it; the requirement is zh-TW only).
10. Session expiry: no refresh token in the contract, so an expired JWT sends the owner back to sign-in.

## Contract change requests

`docs/contract-change-requests/frontend.md` (commit `3da101d`):
CCR-F1 positions shape · CCR-F2 month summary shape (`realizedReturnPct` as percent, 404 only
before the first month) · CCR-F3 card list as bare array, empty for an empty month ·
CCR-F4 error codes/messages for fills, skip and login · CCR-F5 CORS for the web origins.

## Manual steps for the owner

1. Review and push `feat/frontend` and go-trading-deploy `feat/trading-board-frontend`; merge both.
2. In the trading-board GitHub repo add secret `DEPLOY_REPO_SSH_KEY` (deploy key with write access to go-trading-deploy).
3. After the first CI publish, make the new GHCR package `trading-board-frontend` pullable by the cluster (public, like go-trading-frontend, or add an imagePullSecret) — new packages start private.
4. Only then apply the Argo Application (`kubectl apply -f argocd/applications.yaml`); until CI bumps it the tag is the `0000000` placeholder.
5. Add the Cloudflare Tunnel route `trading-board.coding-afternoon.com → http://localhost:80`.
6. Make sure the backend allows CORS from `https://trading-board.coding-afternoon.com` (and `http://localhost:3000` for dev).
7. Local dev: `cd frontend && bun install && bun run dev` → http://localhost:3000, sign in with `owner` / `owner` (mock). Point at a real backend with `NUXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1 bun run dev`.

## Known limitations / follow-ups

- No browser visual pass was possible in this run; worth a quick look on a phone.
- `frontend/README.md` is still the generic coding-rules README from the bootstrap commit; it was left untouched.
- Pages / layout / middleware wiring has no tests (reference convention); see the contract notes.
- Wrong fills cannot be corrected (v1 out of scope).
