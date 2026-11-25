# Product Requirements Document (PRD) — Owner Sign-in & Member Keys (web v2)

**Status:** Finalized (unattended run; decisions from the v2 stakeholder requirements)
**Version:** v2.0
**Owner:** Trading Board owner
**Stakeholders:** Owner, agent-team members, board builder

---

## 1. Background & Goal (Why & Goal)

- **Problem Statement:** The web screens sign the owner in with a username and keep a 24-hour pass
  saved on the device; the owner retypes the password every day and the agent team shares one key
  only the operator can change.
- **Expected Outcome:** The owner signs in with an email on the board's own sign-in page, stays
  signed in for up to 30 days of inactivity, sees who is signed in and can sign out; nothing that
  proves the sign-in is saved where page scripts can read it. The owner creates, sees and revokes
  each agent-team member's key on a 「API 金鑰」 screen, and every card names its publisher.
- **Out of Scope:** more than one user, sign-up, password change or reset, second-factor checks,
  outside accounts; editing keys, key permissions or expiry, showing a key again; showing who
  reported fills or skipped orders.

---

## 2. User Personas

- **Owner** — the single human user, mostly on a phone, sometimes on a desktop, in the evening and
  before the session.

---

## 3. User Stories & Acceptance Criteria

### US-01 — Sign in with email and password [P0]
**As the** owner, **I want** to sign in with my email and password, **so that** I can open the board.

```gherkin
Scenario: the right email and password open the home screen
  Given the owner's right email and password
  When the owner signs in
  Then the home screen opens
  And the header shows the owner's email

Scenario: a wrong password or unknown email keeps the owner on the sign-in page
  Given a wrong password or an unknown email
  When the owner signs in
  Then the sign-in page stays, showing 「電子郵件或密碼錯誤」

Scenario: a locked sign-in says to try later
  Given the board has locked sign-in after repeated failures
  When the owner signs in
  Then the sign-in page stays, showing 「嘗試次數過多，請 15 分鐘後再試」

Scenario: an unreachable board says to try later
  Given the board cannot be reached
  When the owner signs in
  Then the sign-in page stays, showing 「無法連線，請稍後再試」
```

### US-02 — Finish the sign-in on the return page [P0]
**As the** owner, **I want** the return page to finish my sign-in safely, **so that** nobody can
slip me someone else's sign-in.

```gherkin
Scenario: a matching marker and accepted code sign the owner in
  Given the round-trip marker matches the one the web app made
  And the board accepts the one-time code
  When the return page opens
  Then the home screen opens, signed in

Scenario: a marker that differs, or none made, stops the sign-in
  Given the round-trip marker differs from the one the web app made, or none was made
  When the return page opens
  Then it shows 「登入流程已失效，請重新登入」 with a link to the sign-in page
  And the code is not traded in

Scenario: a refused code stops the sign-in
  Given the board refuses the one-time code
  When the return page opens
  Then it shows 「登入流程已失效，請重新登入」 with a link to the sign-in page
```

### US-03 — Stay signed in [P0]
**As the** owner, **I want** to stay signed in without retyping my password, **so that** checking
the board in the evening is one tap.

```gherkin
Scenario: a valid renewal pass opens screens without signing in
  Given the browser keeps a valid renewal pass
  And the page was just opened
  When the owner opens a screen
  Then the screen shows without the sign-in page

Scenario: no or a refused renewal pass shows the sign-in page
  Given the browser keeps no renewal pass, or the board refuses it
  When the owner opens a screen
  Then the sign-in page shows

Scenario: a pass about to run out is renewed before asking for data
  Given the short-lived pass runs out in 30 seconds
  When a screen asks the board for data
  Then a new short-lived pass is obtained before the board is asked
  And the screen shows the data with no sign-in page

Scenario: a pass with time left is used as it is
  Given the short-lived pass runs out in 10 minutes
  When a screen asks the board for data
  Then the board is asked directly with no renewal

Scenario: simultaneous needs share one renewal
  Given renewal is needed
  When two screens ask the board for data at the same time
  Then only one renewal happens
  And both get their data

Scenario: a pass the board no longer accepts sends the owner to sign in
  Given the board says the pass is no longer accepted
  When a screen asks for data
  Then the sign-in page shows
```

### US-04 — Sign out [P0]
**As the** owner, **I want** to sign out, **so that** this device is no longer signed in.

```gherkin
Scenario: signing out ends the sign-in
  Given the owner is signed in
  When the owner taps 登出
  Then the board ends the sign-in
  And the sign-in page shows

Scenario: signing out works even when the board is unreachable
  Given the board cannot be reached
  When the owner taps 登出
  Then the sign-in page shows
  And the pass is forgotten on this page

Scenario: signing out a second time is harmless
  Given the owner already signed out in another tab
  When the owner taps 登出
  Then the sign-in page shows with no error
```

### US-05 — Manage member keys [P0]
**As the** owner, **I want** a 「API 金鑰」 screen, **so that** each agent-team member has its own
key I can stop at any time.

```gherkin
Scenario: a new key is shown once with a copy button
  When the owner creates a key for member stock-scanner with the note "daily scan"
  Then the full key is shown with a copy button and the warning 「這把金鑰只會顯示這一次」
  And the list shows stock-scanner as 啟用中

Scenario: an empty member name is not sent
  Given the member name is empty
  When the owner creates a key
  Then nothing is sent to the board
  And the member name field asks to be filled in

Scenario: the board's refusal is shown in its own words
  Given the board says the member name is in use
  When the owner creates a key
  Then 「這個成員名稱已有啟用中的金鑰」 is shown

Scenario: copying the key confirms it
  Given a key was just shown
  When the owner taps copy
  Then the key is copied
  And the button shows 「已複製」

Scenario: the full key disappears for good
  Given a key was just shown
  When the owner creates another key or leaves the screen
  Then the earlier key is no longer shown anywhere

Scenario: revoking after confirming
  Given an active key
  When the owner taps 撤銷 and confirms
  Then the key shows 已撤銷 with no actions

Scenario: cancelling a revoke changes nothing
  Given an active key
  When the owner taps 撤銷 and cancels
  Then the key stays 啟用中

Scenario: a never-used key says so
  Given a key that was never used
  When the owner lists keys
  Then its last-used time shows 「從未使用」
```

### US-06 — See who published a card [P1]
**As the** owner, **I want** to see which member published a card, **so that** I know whose advice
it is.

```gherkin
Scenario: a card names its publisher
  Given a card published by desk-chief
  When the owner reads it
  Then it shows 「發布者：desk-chief」

Scenario: an older card shows no publisher
  Given a card from before publishers were recorded
  When the owner reads it
  Then no publisher line is shown
```

---

## 4. Business Flow & Logic

- **Flow:** sign-in page makes a one-off secret, its proof and a round-trip marker → owner signs
  in → board sends the owner to the return page with a one-time code → return page checks the
  marker, trades the code with the secret → home screen. On opening the app: renew silently.
  Before each request: renew when less than a minute is left (one renewal at a time).
- **Core Business Rules:**
  - The short-lived pass is kept only while the page is open; the renewal pass is never visible
    to page scripts.
  - The one-off secret and marker live for one sign-in and are discarded once used.
  - A refused renewal or a pass the board no longer accepts always ends at the sign-in page.
  - Signing out always succeeds on the screen.
  - The full member key is visible only right after its creation.
- **Edge Cases:** the return page opened directly (no marker made) → treated as a failed round
  trip; board unreachable on sign-out → still signed out on the screen.

---

## 5. UI/UX Design & Interaction

- Sign-in page: Traditional Chinese, mobile-first, email + password, one general error, locked
  message, busy state while signing in.
- Header: owner's email and 登出 on every signed-in screen; main navigation gains 「API 金鑰」.
- 「API 金鑰」 screen: create form on top, the one-time key panel after creation, list below with
  empty state 「還沒有任何金鑰」; revoke asks for an on-screen confirmation (no browser pop-up).

---

## 6. Non-Functional Requirements

- **Security:** nothing that proves the sign-in is saved on the device by the web app; the
  renewal pass is never readable by page scripts; the one-off secret never leaves the browser
  except to the board when trading the code.
- **Compatibility:** current mobile Safari and Chrome; works against the practice board.

---

## 7. Dependencies & Risks

- **External Dependencies:** the board's v2 sign-in, renewal, sign-out, who-am-I and key
  management (built first).
- **Known Risks:** a laptop running the screens against the production board cannot renew
  silently (the browser does not send the renewal pass across sites); development uses the
  practice board or a local board.

---

## 8. Appendix

- `BRIEF.md` (assumptions), v2 stakeholder requirements, board PRD of this feature.
