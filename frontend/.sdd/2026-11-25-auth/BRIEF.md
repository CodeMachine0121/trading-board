# Owner Sign-in & Member Keys (web v2) — Requirements Brief

## Goal
Bring the board's new way of signing in and its member keys to the owner's phone-first,
Traditional Chinese web screens. The owner signs in with an email address on the board's own
sign-in page, stays signed in for up to 30 days of inactivity without retyping the password,
sees who is signed in and can sign out from every screen. A new 「API 金鑰」 screen lets the
owner create a key for each agent-team member (seeing it in full exactly once, with a copy
button), see all keys and revoke one. Cards say which member published them. Everything else on
the screens stays as in web v1. The development stand-in for the board answers the new sign-in
and key screens too, so the screens still work without the real board.

## Requirements
- **Sign-in page.** Email and password, Traditional Chinese, comfortable on a phone. A refusal
  shows one general message that never says which of the two was wrong. While sign-in is locked
  after repeated failures, it shows the "too many attempts" message instead.
- **Sign-in round trip.** Before showing the sign-in page the web app makes up a one-off secret
  and a round-trip marker. After a successful sign-in the board sends the owner back to the web
  app's return page with a one-time code; the return page checks the marker is the one it made,
  trades the code (with the secret) for passes and opens the home screen. A wrong or missing
  marker, or a refused code, shows a message and a way back to the sign-in page.
- **Where the passes live.** The short-lived pass is kept only while the page is open, never
  saved on the device. The long-lived renewal pass is kept by the browser where page scripts
  cannot read it.
- **Staying signed in.** When the web app opens, and shortly before the short-lived pass runs
  out, it renews silently. If renewal is refused, the owner sees the sign-in page. Two renewals
  are never sent at the same time (a second one would look like theft and end the sign-in).
- **Who is signed in.** Every signed-in screen shows the owner's email next to the sign-out
  button.
- **Signing out** ends the sign-in on the board and returns to the sign-in page, even if the
  board cannot be reached at that moment.
- **「API 金鑰」 screen**, reachable from the main navigation:
  - lists every key, newest first: member name, note, the first few characters of the key, created
    at, last used at (「從未使用」 when never), and status 啟用中 / 已撤銷;
  - a form with member name (required, 1–40 characters) and an optional note creates a key;
  - right after creating, the full key is shown once with a copy button and a warning that it
    will not be shown again; leaving or creating another hides it for good;
  - an active key can be revoked after a confirmation step on the screen; a revoked key offers
    no actions;
  - the board's refusals (name already in use, name too long) are shown as the board words them.
- **Publisher on the card.** A card shows which member published it; older cards without one
  show nothing.
- **Practice board** (the stand-in used while building the screens) answers sign-in, the return trip, renewal, sign-out, who am I and
  the key screens.

## Examples (Specification by Example)
Each example lists **only** the data that affects the behavior — nothing more.

### Rule: Signing in
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy) | Right email and password | Owner signs in | Home screen opens; header shows the owner's email |
| 2 (exception) | Wrong password or unknown email | Owner signs in | Stays on sign-in page with 「電子郵件或密碼錯誤」 |
| 3 (exception) | Sign-in locked by the board | Owner signs in | Stays on sign-in page with 「嘗試次數過多，請 15 分鐘後再試」 |
| 4 (exception) | Board unreachable | Owner signs in | Stays on sign-in page with the general "try again later" message |

### Rule: Return page
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy) | Marker matches; code accepted | Return page opens | Home screen opens, signed in |
| 2 (exception) | Marker differs from the one the web app made, or none was made | Return page opens | Message 「登入流程已失效，請重新登入」 and a link to sign in; code not traded |
| 3 (exception) | Board refuses the code | Return page opens | Same message and link |

### Rule: Staying signed in
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy) | Valid renewal pass, page just opened | Owner opens a screen | Screen shows without the sign-in page |
| 2 (exception) | No or refused renewal pass | Owner opens a screen | Sign-in page |
| 3 (boundary) | Short-lived pass runs out in 30 seconds | Screen asks the board for data | Renewed first, then the data shows |
| 4 (boundary) | Short-lived pass runs out in 10 minutes | Screen asks the board for data | Asked directly, no renewal |
| 5 (exception) | Two screens ask for data while renewal is needed | Both ask | Only one renewal happens; both get their data |
| 6 (exception) | Board says the pass is no longer accepted | Screen asks for data | Sign-in page |

### Rule: Signing out
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy) | Signed in | Owner taps 登出 | Board ends the sign-in; sign-in page shows |
| 2 (exception) | Board unreachable | Owner taps 登出 | Sign-in page shows anyway; the pass is forgotten on this page |
| 3 (boundary) | Already signed out in another tab | Owner taps 登出 | Sign-in page shows; no error |

### Rule: Member keys
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy) | — | Create for stock-scanner, note "daily scan" | Full key shown once with copy button and warning; list shows stock-scanner 啟用中 |
| 2 (exception) | Member name empty | Create | Not sent; field asks for a member name |
| 3 (exception) | Board says the name is in use | Create | Board's words shown: 「這個成員名稱已有啟用中的金鑰」 |
| 4 (happy) | Key just shown | Tap copy | Key copied; button confirms 「已複製」 |
| 5 (boundary) | Key just shown | Create another or leave the screen | The earlier key is no longer shown anywhere |
| 6 (happy) | Active key | Tap 撤銷, then confirm | Shows 已撤銷 with no actions |
| 7 (boundary) | Active key | Tap 撤銷, then cancel | Nothing changes |
| 8 (boundary) | Key never used | List | Last used shows 「從未使用」 |

### Rule: Publisher on the card
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy) | Card published by desk-chief | Owner reads it | Shows 「發布者：desk-chief」 |
| 2 (boundary) | Card from before attribution | Owner reads it | No publisher line |

## Out of Scope
- More than one user, sign-up, password change or reset, second-factor checks, outside accounts.
- Editing keys, key permissions or expiry; showing a key again after creation.
- Showing who reported fills or skips (the board records it; the screens do not show it yet).

## Open Decisions
- How long before the short-lived pass runs out to renew (assumed: within the last minute).

## Context / Background
- Source: the v2 stakeholder requirements for sign-in and keys, the v1 stakeholder requirements,
  and the board's own v2 brief and PRD (built first). No clarification questions were asked.
- Stages: clarify → prd → architecture → implement (code-first) → improve-codebase → contract.
- **Assumptions** (simplest secure option):
  1. Renew when the short-lived pass has less than one minute left, on the next request, instead
     of a background timer; a renewal already under way is shared, never repeated.
  2. The one-off secret and marker are kept only for the duration of one sign-in in the browser
     tab, and are discarded once the return page has used them.
  3. A failed return trip asks the owner to sign in again rather than retrying.
  4. Signing out always succeeds on the screen, even when the board cannot be reached.
  5. Revoking needs an on-screen confirmation (no browser pop-up).
  6. Who reported fills or skipped orders is not shown in this version.
