# Trading Board v2 — Owner Sign-in & Member API Keys — Delivery Report

Branch `feat/auth` (this repo) and `feat/trading-board-auth` (go-trading-deploy). Nothing pushed or merged;
no `kubectl`, no shared database touched. Requirements: `docs/requirements/auth.md` (final) +
`docs/requirements/trading-board.md` for everything unchanged.

Workflow run twice (backend, then frontend), each from its subproject with its own rules:
`/clarify → /prd → /architecture → /implement (code-first) → /improve-codebase → /contract`.
Artifacts: `backend/.sdd/2026-11-25-auth/` and `frontend/.sdd/2026-11-25-auth/`
(`BRIEF.md`, `PRD.md`, `ARCH.md`, `CONTRACT.md`; both `UL-MAP.md` updated).

---

## 1. What changed

### Backend (Go)
- **Own OAuth 2.0 authorization server** for one public client (`trading-board-web`), Authorization Code +
  PKCE (S256 only; `plain` refused), configured redirect-URI allow-list (exact match).
- **Owner account** from config: `OWNER_EMAIL` + `OWNER_PASSWORD_HASH` (**bcrypt**). Email matched
  case/space-insensitively; the bcrypt check runs even for an unknown email; refusals never say which part
  was wrong. Empty email or hash admits nobody.
- **Sign-in lock**: 5 failures within 15 min → refused for 15 min from the 5th (`429 temporarily_unavailable`),
  refusals during the lock are not counted, a success clears the count (stored in `SignInFailures`).
- **One-time authorization code** (60 s, SHA-256 digest stored, bound to client/redirect/challenge, used up on
  any trade-in attempt — even with a wrong verifier).
- **Tokens**: access token = HS256 JWT, 15 min. Refresh token = 256-bit opaque, stored as SHA-256 digest,
  30 days sliding, **rotated** on every use; reusing a rotated one **revokes the whole family**. Delivered
  only as an `HttpOnly; SameSite=Strict; Path=/api/v1/oauth` cookie (`Secure` and `Domain` from config).
- **CSRF**: token(refresh)/revoke require an `Origin` in `CORS_ALLOWED_ORIGINS` (missing Origin refused);
  CORS now answers `Access-Control-Allow-Credentials: true`.
- **Member API keys** (`tbk_` + 43 random chars), SHA-256 digest + 12-char display prefix stored, full key
  shown once; unique active member name (case-insensitive, partial unique index); revoke is final and
  idempotent; `last used` refreshed at most once a minute. Keys keep v1 permissions (publish, read,
  fills/skips) and can never manage keys (`403 forbidden`).
- **Attribution**: `publishedBy` on cards (and card summaries), `reportedBy` on fills, `skippedBy` on orders —
  member name, or `owner`; older rows are empty. A second skip keeps the first skipper.
- **Removed**: `POST /api/v1/auth/login`, `OWNER_USERNAME`, `OWNER_PASSWORD`, `AGENT_API_KEY`.
- **New command** `cmd/hash-password` (reads the password from stdin, prints the bcrypt hash).
- Schema (code-first, `cmd/migrate`): new tables `AuthorizationCodes`, `RefreshTokens`, `ApiKeys`,
  `SignInFailures`; new columns `TradeCards.published_by`, `TradeOrders.skipped_by`, `OrderFills.reported_by`
  (`NOT NULL DEFAULT ''`, so existing rows migrate cleanly).

### Frontend (Nuxt)
- Sign-in page: **email** + password; generic error, lock message, unreachable message.
- PKCE + state made with Web Crypto; verifier/state kept in `sessionStorage` for the one redirect and removed
  on use; `/auth/callback` checks state, trades the code, opens the home screen (or shows
  「登入流程已失效，請重新登入」 with a link back).
- **Access token in page memory only** (nothing in `localStorage`); the HttpOnly cookie renews it on start-up
  and whenever < 1 min is left, through **one shared in-flight renewal** (parallel renewals would trip the
  board's theft detection). A refused token sends the owner to the sign-in page.
- Header shows the signed-in email; sign-out revokes on the board and always succeeds on screen.
- New **「API 金鑰」** screen in the main nav: create (member name + note), full key shown once with copy button
  and warning, list (status, prefix, created, last used / 「從未使用」), revoke with on-screen confirmation.
- Cards show 「發布者：<member>」.
- Practice board (`server/mock`, `bun run dev`): speaks the whole v2 flow incl. the refresh cookie, lock,
  rotation/theft detection, `me`, keys. Sign in with `owner@example.com` / `owner`.

### Deploy repo (`feat/trading-board-auth`, commit `609b473`)
- `configmap.yaml`: `OAUTH_CLIENT_ID`, `OAUTH_REDIRECT_URIS` (prod + localhost), `AUTH_REFRESH_COOKIE_DOMAIN`
  (empty = API host only; the web host is same-site), `AUTH_REFRESH_COOKIE_SECURE=true`, CORS-with-credentials note.
- `secret.example.yaml`: `OWNER_EMAIL`, `OWNER_PASSWORD_HASH`, `AUTH_ACCESS_TOKEN_SIGNING_KEY`; agent key removed.
- Git-ignored `trading-board-backend/secret.yaml` rewritten locally: `OWNER_EMAIL`, `OWNER_PASSWORD_HASH`
  (bcrypt of the owner password), existing `AUTH_ACCESS_TOKEN_SIGNING_KEY` kept; `OWNER_USERNAME`,
  `OWNER_PASSWORD`, `AGENT_API_KEY` dropped. Git-ignored `backend/.env` written for local runs (hash only).

### tw-stock agent team
- `~/workspace/TradingNotes/tw-stock-team/.claude/skills/trading-board/SKILL.md` (not a git repo): each member
  now uses its own key, `TRADING_BOARD_API_KEY_<MEMBER>` (e.g. `TRADING_BOARD_API_KEY_DESK_CHIEF`), never
  borrows another's; a 401 means that member's key was revoked or mistyped.

---

## 2. Endpoints

| Method & path | Auth | Request → Response |
| --- | --- | --- |
| `POST /api/v1/oauth/authorize` | public | JSON `{email, password, response_type:"code", client_id, redirect_uri, code_challenge, code_challenge_method:"S256", state}` → `200 {redirectTo}` |
| `POST /api/v1/oauth/token` | public (+ cookie, Origin) | form `grant_type=authorization_code&code&code_verifier&client_id&redirect_uri` or `grant_type=refresh_token` → `200 {access_token, token_type:"Bearer", expires_in:900}` + `Set-Cookie trading_board_refresh_token` |
| `POST /api/v1/oauth/revoke` | public (+ cookie, Origin) | → `200`, cookie cleared (unknown/missing token also `200`) |
| `GET /api/v1/me` | owner | → `{email}` |
| `GET /api/v1/api-keys` | owner | → `[{id, memberName, note, displayPrefix, status, createdAt, lastUsedAt, revokedAt}]` newest first |
| `POST /api/v1/api-keys` | owner | `{memberName, note}` → `201 {…, key}` (only time the key is returned) |
| `POST /api/v1/api-keys/{id}/revoke` | owner | → the key, `status:"revoked"` |
| `POST /api/v1/auth/login` | — | **removed** (404) |

OAuth errors: `{error, error_description}` — `access_denied` 401, `temporarily_unavailable` 429,
`invalid_client` / `invalid_request` / `unsupported_response_type` / `invalid_grant` / `unsupported_grant_type` 400,
foreign/missing Origin on cookie requests 403 `invalid_request`. All other endpoints keep
`{error:{code,message}}`; a member key on an owner-only route → 403 `forbidden`.
v1 business endpoints are unchanged and accept `Authorization: Bearer` or an active `X-Api-Key`;
publishing still requires a member key.

---

## 3. Tests, coverage, verification

| | Result |
| --- | --- |
| Backend `go test ./... -race` (with `TEST_POSTGRES_DSN` → local Docker Postgres 17) | all green: 168 top-level tests / 377 incl. subtests, 0 failures; `go vet` clean |
| Backend coverage (`-coverpkg=./internal/...`, mocks excluded) | 99.9 % statements; 100 % of the new/changed code. Only gap: `JwtAccessTokenProxy.Issue` 83.3 % (HMAC signing-error branch, pre-existing) |
| Backend falsification | 24 + 3 hand mutations of the critical rules (lock window/end, code expiry, PKCE, redirect binding, reuse→family revoke, last-used throttle, name length/case, origin check, HttpOnly, owner-only, revoked keys, attribution, rollback of refusals…) — all killed |
| Frontend `bun run verify` (eslint, stylelint, token check, `nuxt typecheck`, vitest) | all green: 46 files, 295 tests |
| Frontend coverage (V8, project scope) | 100 % statements / branches / functions / lines |
| Frontend falsification | 24 + 6 mutations (renewal margin, single-flight, state check, sign-out, 429, OAuth error text, cookies, form encoding, show-once, confirm step, navigation rules…) — all killed |
| `nuxt generate` | builds; `/auth/callback` and `/api-keys` generated |
| End-to-end (headless Chromium, not part of the suite) | practice board: unsigned → login, wrong password message, sign-in, header email, HttpOnly cookie only, reload keeps sign-in, create/copy/duplicate/cancel/revoke key, sign-out, forged callback refused. **Real Go backend + web app**: sign-in, reload via refresh cookie, key creation, key authenticates against the API, sign-out. Server logs contained no password or key |

## 4. `/contract` verdicts

| Side | Clauses | Result | Notes |
| --- | --- | --- | --- |
| Backend | 56 (47 AC + 6 BR + 3 NFR) | **55 ✅ conforms, 1 ❔ unclear (AC-14)**, 0 violations/gaps | Pass 1 flagged unasserted refusal wording and a missing "key + token both valid → member" test; fixed in tests (`3cb38c0`), re-audited. AC-14 (fresh code → 15-min pass + 30-day HttpOnly renewal cookie): Jev's code confidence stayed low; both mapped tests pass and assert exactly that |
| Frontend | 33 (26 AC + 5 BR + 2 NFR) | **32 ✅ conforms, 1 🟠 mis-asserted (AC-1)**, 0 violations/gaps | Pass 1 flagged untested navigation; rules moved into `useSignInFlow` and tested (`273aa57`). AC-1: the header-email binding is in the layout, which no unit test mounts (verified only in the headless run) |

Low-confidence or disagreeing Jev routes are listed under "Needs James" in each `CONTRACT.md`
(backend 15, frontend 7); none is a violation. Matrices: `backend/.sdd/2026-11-25-auth/CONTRACT.md`,
`frontend/.sdd/2026-11-25-auth/CONTRACT.md`.

---

## 5. Assumptions (not covered by the requirements; simplest secure option)

1. Every failed sign-in counts toward the lock whatever email was typed; refusals during the lock don't extend it.
2. Email matched ignoring case and surrounding spaces; password exact.
3. A code traded in with a wrong verifier is used up.
4. The refresh token is returned **only** in the HttpOnly cookie, never in the JSON body.
5. Refresh tokens live 30 days from issue (sliding), no absolute cap on a sign-in.
6. Sign-out with no/unknown refresh token answers 200. Access tokens stay valid until their 15 min end.
7. Member names unique among active keys ignoring case; note ≤ 200 chars; revoking twice is a no-op.
8. Skips record `skippedBy` like fills record `reportedBy`; older rows carry no attribution.
9. Publishing remains member-key-only (as in v1); a request carrying both a valid key and a valid owner token is the member.
10. Cookie-carrying requests (refresh, revoke) must have an allowed `Origin`; a missing Origin is refused.
11. bcrypt was chosen (cost 10) for `OWNER_PASSWORD_HASH`.
12. Web: renewal on demand when < 1 min is left (no background timer); verifier/state in `sessionStorage` for one
    redirect; sign-out always succeeds on screen; revoke needs an on-screen confirmation; fill/skip reporters not shown yet.
13. All auth writes reuse the existing ledger-wide advisory-lock transaction (single user, simplest race-free option).

---

## 6. Manual steps for the owner

1. **Review & merge** `feat/auth` (this repo) and `feat/trading-board-auth` (go-trading-deploy). Backend and web
   must ship together: the old `/auth/login` and the shared agent key stop working with this release.
2. **Apply the Secret** (already filled locally, git-ignored):
   `kubectl apply -f trading-board-backend/secret.yaml` in the go-trading-deploy worktree. It holds `OWNER_EMAIL`,
   `OWNER_PASSWORD_HASH` and the existing signing key; the removed keys are gone from it. To change the password
   later: in `backend/`, `read -rs PASSWORD && printf '%s' "$PASSWORD" | go run ./cmd/hash-password; unset PASSWORD`,
   put the hash in the Secret (single-quoted in YAML), restart the deployment.
3. **Apply the ConfigMap** (via Argo after merge) and restart `trading-board-backend`; the migrate init container
   adds the new tables/columns. Make sure the production redirect URI
   `https://trading-board.coding-afternoon.com/auth/callback` is the web app's real address.
4. **Sign in** at `https://trading-board.coding-afternoon.com` with your email and password.
5. **Create the first API keys**: open 「API 金鑰」, enter a member name (e.g. `desk-chief`, `stock-scanner`,
   `risk-budget-officer`, `position-keeper`, …) and an optional note, press 建立金鑰, then 複製金鑰 — the full key is
   shown **only once**. Repeat per member.
6. **Give each member its key**: set `TRADING_BOARD_URL=https://trading-board-api.coding-afternoon.com` and
   `TRADING_BOARD_API_KEY_<MEMBER>` (e.g. `TRADING_BOARD_API_KEY_DESK_CHIEF`) in the environment the tw-stock
   team runs in (the updated `trading-board` skill reads these). Never commit them.
7. Revoke a member's key any time on the same screen; it stops working immediately.
8. Local development against a local backend: add `http://localhost:3000/auth/callback` to `OAUTH_REDIRECT_URIS`,
   `http://localhost:3000` to `CORS_ALLOWED_ORIGINS`, and set `AUTH_REFRESH_COOKIE_SECURE=false` (see the two
   `.env.example` files). A laptop on localhost against the **production** API can sign in but cannot renew
   silently (the SameSite=Strict cookie is cross-site there).
