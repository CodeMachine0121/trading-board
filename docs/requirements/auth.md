# Trading Board — Authentication & API Keys (v2) — Stakeholder Requirements

Stakeholder input for the `feat/auth` branch. All business decisions below are final.
When `/clarify` would ask a business question, answer it from this document (and from
`trading-board.md` for everything not changed here); if something is genuinely not
covered, choose the simplest secure option and record it as an assumption.

Trading Board is a standalone system. It has **its own** login and must not depend on
go-trading or any external identity provider.

## 1. What changes

| Today (v1) | After this change |
| --- | --- |
| `POST /auth/login` with username/password from config → 24 h JWT | Own OAuth 2.0 authorization server: login page + Authorization Code with PKCE → short-lived access token + rotating refresh token; logout revokes |
| One static `AGENT_API_KEY` from config | The owner creates, names, lists and revokes API keys in the web UI, one per agent-team member |
| Username `owner` | The owner signs in with an **email** address |

## 2. Owner account

- Exactly one account. Its email and a **password hash** come from configuration
  (Kubernetes Secret): `OWNER_EMAIL`, `OWNER_PASSWORD_HASH` (bcrypt or argon2id; pick one
  and provide a small command to generate the hash). The plaintext password never
  appears in code, commits, docs, logs or config committed to git.
- No sign-up, no password change, no password reset in this version.
- Failed sign-ins are throttled: after 5 failures within 15 minutes for the account,
  further attempts are refused for 15 minutes (still answering generically).
- Sign-in errors never reveal whether the email or the password was wrong.

## 3. OAuth 2.0 flow (RFC 6749 + RFC 7636 PKCE, public client)

- One registered public client, configured (not dynamic): `client_id = trading-board-web`,
  allowed redirect URIs from configuration (production
  `https://trading-board.coding-afternoon.com/auth/callback`, dev
  `http://localhost:3000/auth/callback`). PKCE with `S256` is mandatory; `plain` is refused.
- Flow:
  1. The web app creates `code_verifier`, `code_challenge`, `state`, then shows **its own
     login page** (Traditional Chinese) with the authorization parameters.
  2. The login form submits email + password + authorization parameters to the backend's
     authorization endpoint. On success the backend issues a **one-time authorization
     code** bound to client, redirect URI and code challenge, valid 60 seconds, and
     answers with the redirect target (`redirect_uri?code=…&state=…`).
  3. The web app's `/auth/callback` verifies `state` and exchanges the code at the token
     endpoint with `code_verifier` (`grant_type=authorization_code`).
  4. Token response: `access_token` (JWT, 15 minutes), `token_type=Bearer`, `expires_in`,
     and a refresh token.
- Refresh: `grant_type=refresh_token` returns a new access token **and a new refresh
  token**; the old refresh token stops working (rotation). Reusing an already-rotated
  refresh token revokes the whole token family (theft detection). Refresh tokens live
  30 days, are opaque, and are stored **hashed**.
- Logout: a revocation endpoint (RFC 7009) revokes the refresh token family; the web app
  forgets the access token.
- Token storage in the browser: access token **in memory only**; refresh token in an
  `HttpOnly`, `Secure`, `SameSite=Strict` cookie scoped to the token endpoints' path on the
  API host (the web and API hosts are same-site under `coding-afternoon.com`; local dev
  runs on localhost). Requests that use the cookie must be protected against CSRF
  (e.g. require a custom header or an `Origin` allowlist check).
- Silent refresh: on start-up and shortly before expiry the web app refreshes; if refresh
  fails it shows the login page. The user should stay signed in for up to 30 days of
  inactivity without retyping the password.
- Discovery document (`/.well-known/oauth-authorization-server`) is nice to have, not required.
- Error responses on OAuth endpoints follow OAuth's `{error, error_description}`; all
  other endpoints keep the v1 error shape.

## 4. API keys for agent-team members

- New web page **「API 金鑰」**: the owner creates a key by entering a **member name**
  (e.g. `stock-scanner`, `risk-budget-officer`, `desk-chief`; 1–40 chars, unique among
  active keys) and an optional note.
- The full key is shown **once** at creation, with a copy button, and is never
  retrievable again. Format: a recognizable prefix plus ≥ 32 bytes of randomness
  (e.g. `tbk_…`). Store only a hash (SHA-256 is fine for high-entropy keys) and a short
  display prefix.
- The list shows: member name, note, display prefix, created at, last used at, status.
  The owner can **revoke** a key (immediate, irreversible). No editing of keys.
- API keys authenticate with `X-Api-Key` exactly as before and keep v1 permissions
  (publish cards, read everything, report fills/skips). API keys can **not** manage API
  keys; only a signed-in owner can.
- Cards published with an API key record `publishedBy` = the member name; fills/skips
  reported with an API key record `reportedBy` = member name (owner actions record
  `owner`). Show `publishedBy` on the card in the UI.
- `last used at` is updated on use (at most once per minute per key is fine).
- The static `AGENT_API_KEY` configuration is removed.

## 5. Endpoints (proposed; adjust names if a convention in the codebase fits better)

| Method & path | Auth | Purpose |
| --- | --- | --- |
| `POST /api/v1/oauth/authorize` | public | Credentials + authorization params → `{redirectTo}` with code and state |
| `POST /api/v1/oauth/token` | public (+ refresh cookie) | `authorization_code` and `refresh_token` grants |
| `POST /api/v1/oauth/revoke` | public (+ refresh cookie) | Logout |
| `GET /api/v1/me` | owner | `{email}` for the header |
| `GET /api/v1/api-keys` | owner | List keys |
| `POST /api/v1/api-keys` | owner | Create → includes the full key once |
| `POST /api/v1/api-keys/{id}/revoke` | owner | Revoke |
| `POST /auth/login` | — | **Removed** |

All existing v1 business endpoints keep their behaviour; they now accept the new access
token (`Authorization: Bearer`) or an active API key.

## 6. Frontend

- Login page: email + password, Traditional Chinese, mobile-first, generic error message,
  throttling message when locked.
- `/auth/callback` route, silent refresh, logout button in the layout, signed-in email shown.
- 「API 金鑰」page as in §4, reachable from the main navigation.
- The dev mock (`server/mock`) is updated to cover the new flow so `bun run dev` still works
  without a backend.

## 7. Platform

- Same stack and rules as v1 (`backend/.claude/rules`, `frontend/.claude/rules`).
- Database: new tables for authorization codes, refresh tokens and API keys in
  `trading_board`, code-first.
- Config/Secret changes: add `OWNER_EMAIL`, `OWNER_PASSWORD_HASH`, OAuth client id,
  allowed redirect URIs, token lifetimes, cookie domain/secure flags; remove
  `OWNER_USERNAME`, `OWNER_PASSWORD`, `AGENT_API_KEY`. Update `.env.example`,
  `secret.example.yaml` and `configmap.yaml` in go-trading-deploy accordingly.
- CORS must allow credentials (cookies) from the web origins.

## 8. Out of scope

Multiple users, sign-up, password reset/change, MFA, third-party login, API key scopes or
expiry dates, editing keys.
