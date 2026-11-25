# Trading Board backend — report

Branch `feat/backend` (app repo) and `feat/trading-board-backend` (go-trading-deploy). Nothing pushed, nothing
merged, no cluster or shared database touched.

## What was built

**Service (`backend/`)** — Go 1.26 + Gin + GORM/pgx, Clean/Onion layers mirroring go-trading
(`cmd/{server,migrate}`, `internal/{config,controller,application,domain,infrastructure}`).

- Implements shared contract §4 exactly: `POST /auth/login`, `PUT /cards/{tradeDate}` (API key only),
  `GET /cards?month=`, `GET /cards/latest`, `GET /cards/{tradeDate}`, `POST /orders/{id}/fills`,
  `POST /orders/{id}/skip`, `GET /positions?status=&month=`, `GET /months/{month}/summary`, plus
  `GET /health` and `GET /api/v1/health`. Owner endpoints accept `Authorization: Bearer` or `X-Api-Key`.
- Storage holds only cards (with orders) and the append-only fill log. Holdings, the monthly budget and
  performance are replayed from the fill log on every read by one deep module (`TradeLedgerDomain`), so no
  derived figure can drift. All money is `shopspring/decimal` / `numeric(18,4)`, rendered as strings.
- Every write (publish, fill, skip) runs inside one transaction holding a PostgreSQL advisory lock, so the
  cross-row rules (no oversell, no overspend, no replacing a filled card) cannot race.
- Code-first schema (GORM AutoMigrate) via the `migrate` binary, run as an init container before `server`
  (same image, same commit). Config entirely from env (`backend/.env.example`). Dockerfile, Makefile.
- SDD artifacts: `backend/.sdd/UL-MAP.md`, `backend/.sdd/2026-11-25-trading-board/{BRIEF,PRD,ARCH,CONTRACT}.md`.

**CI** — `.github/workflows/backend-ci.yml`, path-filtered to `backend/**`: lint (gofmt, build, vet, stale
mocks) · unit (race) · storage (real Postgres; fails if the storage tests skip) → on `main`: publish
`ghcr.io/codemachine0121/trading-board-backend:{sha7,latest}` · prune (keep 5) · bump `newTag` in
`go-trading-deploy/trading-board-backend/kustomization.yaml` via yq.

**Deploy (go-trading-deploy, commit `31c3a1c`)** — `trading-board-backend/` with `configmap.yaml`,
`deployment.yaml` (migrate init container + server, probes on `/health`), `service.yaml`, `ingress.yaml`
(`trading-board-api.coding-afternoon.com`), `secret.example.yaml`, `kustomization.yaml` (placeholder tag
`"0000000"`); Argo CD Application appended to `argocd/applications.yaml`; `database/trading-board.md`
(manual DB creation); README folder table updated. `POSTGRES_PASSWORD` is read from the existing `go-trading`
Secret via `secretKeyRef`, so the DB password still lives in one place. A local, git-ignored
`trading-board-backend/secret.yaml` with generated values was created and **not** committed.

## Workflow

clarify → prd → architecture → implement (code-first) → improve-codebase → contract, all unattended;
business questions answered from `docs/requirements/trading-board.md`, the rest recorded as assumptions.

## Tests and coverage

- `go test ./... -race` with `TEST_POSTGRES_DSN` against a local `postgres:17-alpine` container:
  **106 test functions / 234 including subtests, all pass, 0 skipped.**
- Coverage of `internal/...` (mocks excluded): **99.9 %** of statements. The only uncovered branch is the
  signing-error path in `JwtAccessTokenProxy.Issue` (HMAC signing with a byte key cannot fail in practice).
  `cmd/` (composition root / `main`) is not unit-tested; it was exercised by an end-to-end smoke run and a
  successful `docker build`.
- Test layers per the rules: domain models directly; applications with real services + domain models and
  gomock only for repositories / clock / token proxy; HTTP via the real router; storage against real
  PostgreSQL (including commit/rollback, write serialization and lock-timeout failures).
- Falsification: 55 deliberate mutations across domain, services, middleware, presenter and repositories —
  **55 killed, 0 survived** (one re-run with a compilable variant to rule out a compile-error kill).

## /contract verdict

80 clauses (60 acceptance scenarios, 15 business rules, 5 NFRs), two Jev runs:

- ✅ **70 conform (87.5 %)** · 🟠 0 mis-asserted · 🟡 0 partial · ❌ 0 gaps · ❔ 8 unclear · 🔴 2 violations
- Run 1 found three shallow tests (AC-5, AC-34, AC-49) and a message that did not say "not held" (AC-16);
  all fixed (`f69b623`) before run 2.
- **Unclear** (AC-37, 38, 39, 48, 50, 52, 58, AC-5): Jev cannot evaluate decimal arithmetic or mock wiring by
  reading; the mapped tests assert the exact PRD figures, pass, and were falsified.
- **Violations, both low confidence**: NFR-1 (0.24) — recheck found no divergence (credentials only from env,
  passes require expiry, wrong password answers only 401 "帳號或密碼錯誤"); NFR-5 (0.37) — there is no
  performance test; reads replay the whole fill log (documented known debt, fine for one owner).
- 29 clauses routed `escalate`, almost all for low Jev confidence; full list in `CONTRACT.md` → "Needs James".

## Assumptions (from BRIEF A1–A15, carried into the PRD)

1. Holdings are replayed in the order fills were **reported**; opening date = first buy fill's date, closing
   date = date of the sell fill that emptied it. Selling to zero then buying again opens a new holding.
2. An open holding's stop-loss / take-profit come from the most recent (by session date) buy order of that
   stock with a fill in the holding.
3. Cost of shares sold = total cost × sold ÷ held, rounded to the cent; the closing sale takes all remaining
   cost (so a holding's total realized P&L is exact). Average cost shown to 4 decimals.
4. Broker fee and transaction tax are each rounded to whole NT$ (half up).
5. A fill belongs to the month of its `filledAt` (Taipei); a card to the month of its `tradeDate`.
6. A buy fill is refused if it would make its own month **or any later month** negative (money carries forward).
7. With nothing recorded yet, the first month is the current month.
8. Mode: past month `monthEnd`, future month `strict`, current month by today's day.
9. `winRate` and `realizedReturnPct` are percentages, 2 dp, `"0"` when the divisor is 0.
10. `filledOrderCount` counts fully filled orders only.
11. Partly filled orders may be skipped (filled shares stay held); filled orders may not; re-skipping is a no-op.
12. Several sell orders of one stock on a card are summed against the holding.
13. `GET /cards` without `month` = current month; `GET /positions` without `status` = open.
14. Pass lifetime 24 h (configurable).
15. `filledAt` and `estimatedAmount` are taken as given (no check against the card date or limit × shares);
    `estimatedAmount` must be > 0 and a given fee ≥ 0.
16. Owner password is compared (constant-time) against the configured plain value from the Secret; there is
    no hashing because there is no user store.

## Contract change requests

Written to `docs/contract-change-requests/backend.md` (commit `7dcb6d9`) for the coordinator:
position object shape (open vs closed semantics), flat monthly-summary object (percent strings), additive
`side` field on orders, the full `error.code` list, bare-array list responses, body/path `tradeDate` rule,
`fee: null` = estimate, timestamps rendered `+08:00`, both health paths.

## Manual steps for the owner

1. **Create the database** on the shared Postgres (doc: `go-trading-deploy/database/trading-board.md`):
   `kubectl -n go-trading exec -it statefulset/postgres -- psql -U postgres -c 'CREATE DATABASE trading_board;'`
2. **Create the Secret** `trading-board-backend` (owner username/password, agent API key, JWT signing key):
   apply the local, git-ignored `go-trading-deploy.worktrees/trading-board-backend/trading-board-backend/secret.yaml`
   (generated values; owner username is `owner` — change it if you like), or create it as shown in
   `secret.example.yaml`. Give the agent team the `AGENT_API_KEY`.
3. **Repository setup for CI**: the app repo needs the `DEPLOY_REPO_SSH_KEY` secret (deploy key with write
   access to go-trading-deploy), the same as go-trading; GHCR push uses `GITHUB_TOKEN` (`packages: write`).
   If the GHCR package is private, the cluster needs pull access as for go-trading.
4. **Cloudflare**: add the public hostname `trading-board-api.coding-afternoon.com` → Traefik.
5. Push/merge both branches; after the first CI publish has replaced the placeholder tag, apply the Argo
   Application (`kubectl apply -f argocd/applications.yaml`).
6. Coordinator: reconcile `docs/contract-change-requests/backend.md` with the frontend.

## Known limitations

- Fills cannot be edited or deleted (v1 scope); a mistyped fill distorts holdings and budget.
- Every read replays the full fill log — fine for one owner; revisit with a snapshot if reads exceed ~100 ms.
- `backend/README.md` is still the generic coding-rules README shipped with the repo bootstrap; service
  facts (env vars, routes) live in `.env.example`, `docs/contract-change-requests/backend.md` and this report.
