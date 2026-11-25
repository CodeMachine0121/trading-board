# Product Requirements Document (PRD) — Owner Sign-in & Member Keys (backend v2)

**Status:** Finalized (unattended run; decisions from the v2 stakeholder requirements)
**Version:** v2.0
**Owner:** Trading Board owner
**Stakeholders:** Owner, agent-team members, frontend builder, operator

---

## 1. Background & Goal (Why & Goal)

- **Problem Statement:** In v1 the owner signs in with a username and gets a 24-hour pass
  that cannot be ended early, and the whole agent team shares one fixed key that only the
  operator can change. A leaked pass or key cannot be stopped, nobody can tell which member
  published a card, and the owner must retype the password every day.
- **Expected Outcome:** The owner signs in with an email on the board's own sign-in page,
  stays signed in for up to 30 days of inactivity, and can end the sign-in at any time. A
  stolen renewal pass is detected the moment it is reused. Each agent-team member has its own
  member key that the owner creates and revokes in the web screens without the operator, and
  every card, fill and skip says who made it.
- **Out of Scope:** more than one user, sign-up, password change or reset, second-factor
  checks, sign-in with outside accounts, narrower member-key permissions, key expiry dates,
  editing keys, ending a short-lived pass before it runs out.

---

## 2. User Personas

- **Owner** — the single human user. Signs in on phone or desktop with email and password;
  manages member keys; reads cards, reports fills and skips orders.
- **Agent-team member** — one machine client per member (for example stock-scanner,
  risk-budget-officer, desk-chief). Calls the board with its own member key to publish cards,
  read everything and report fills or skips.
- **Operator** — whoever runs the board; sets the owner's email and the scrambled form of the
  password, the web app's allowed return addresses and pass lifetimes.

---

## 3. User Stories & Acceptance Criteria

### US-01 — Sign in with email and password [P0]
**As the** owner, **I want** to sign in with my email and password on the board's own page,
**so that** only I can open the board.

```gherkin
Scenario: owner signs in with the right email and password
  Given the operator has set the owner's email and password
  And there were no recent failed sign-ins
  When the owner signs in with the email in different letter case and with extra spaces, and the right password
  Then the sign-in succeeds
  And the web app receives a one-time sign-in code together with its round-trip marker unchanged

Scenario: wrong password is refused without saying which part was wrong
  When the owner signs in with the right email and a wrong password
  Then the sign-in is refused with "電子郵件或密碼錯誤"

Scenario: unknown email is refused with the same words
  When someone signs in with an unknown email
  Then the sign-in is refused with exactly the same words "電子郵件或密碼錯誤"

Scenario: nobody can sign in while the owner account is not set
  Given the operator has not set the owner's email
  When someone signs in with any email and password
  Then the sign-in is refused with "電子郵件或密碼錯誤"
```

### US-02 — Lock sign-in after repeated failures [P0]
**As the** owner, **I want** repeated wrong guesses to lock sign-in for a while, **so that**
nobody can guess my password.

```gherkin
Scenario: the fifth failure within 15 minutes locks sign-in
  Given 4 failed sign-ins in the last 15 minutes
  When a 5th wrong sign-in is made at 20:00
  Then it is refused as "電子郵件或密碼錯誤"
  And sign-in is locked until 20:15

Scenario: the right password is refused while locked
  Given sign-in is locked until 20:15
  When the owner signs in with the right password at 20:10
  Then the sign-in is refused with "嘗試次數過多，請 15 分鐘後再試"

Scenario: the lock ends after 15 minutes
  Given sign-in is locked until 20:15
  When the owner signs in with the right password at 20:15
  Then the sign-in succeeds
  And the failure count starts again from zero

Scenario: five failures spread over more than 15 minutes do not lock
  Given 5 failed sign-ins whose first and fifth are 16 minutes apart
  When the owner signs in with the right password right after
  Then the sign-in succeeds

Scenario: a success clears earlier failures
  Given 3 failed sign-ins followed by a successful sign-in
  When 2 more wrong sign-ins are made
  Then sign-in is not locked
```

### US-03 — Only the board's web app may ask for a sign-in [P0]
**As the** owner, **I want** the board to hand sign-in codes only to its own web app at an
allowed return address, **so that** no other site can collect my sign-in.

```gherkin
Scenario: the registered web app asks correctly
  Given the board's own web app, a return address on the allowed list and the strong proof method
  When the owner signs in
  Then the code is handed back for that return address

Scenario: an unknown application is refused
  Given an application the board does not know
  When the owner signs in
  Then the sign-in is refused as "不明的應用程式"
  And the owner is not sent anywhere

Scenario: a return address not on the allowed list is refused
  Given a return address not on the allowed list
  When the owner signs in
  Then the sign-in is refused as "不允許的返回網址"
  And the owner is not sent anywhere

Scenario: the weak proof method or a missing proof is refused
  Given the web app uses the weak proof method, or sends no proof
  When the owner signs in
  Then the sign-in is refused as "必須使用 S256 驗證"
```

### US-04 — Trade the one-time code for passes [P0]
**As the** web app, **I want** to trade the one-time code for a short-lived pass and a renewal
pass, **so that** the owner is signed in.

```gherkin
Scenario: a fresh code is traded in
  Given a code handed out 20 seconds ago and not yet used
  When the web app trades it in with the matching secret, the same application and return address
  Then it receives a pass good for 15 minutes
  And a renewal pass good for 30 days is kept by the browser out of reach of page scripts

Scenario: a code older than 60 seconds is refused
  Given a code handed out 61 seconds ago
  When the web app trades it in correctly
  Then the exchange is refused as "授權碼無效或已過期"

Scenario: a code cannot be used twice
  Given a code that was already traded in
  When it is traded in again
  Then the exchange is refused as "授權碼無效或已過期"

Scenario: a wrong secret uses the code up
  Given an unused code
  When it is traded in with a secret that does not match
  Then the exchange is refused as "授權碼無效或已過期"
  And trading it in again with the right secret is also refused

Scenario: a different return address is refused
  Given an unused code handed out for return address A
  When it is traded in naming return address B
  Then the exchange is refused as "授權碼無效或已過期"
```

### US-05 — Stay signed in through renewal, and catch theft [P0]
**As the** owner, **I want** the web app to renew my pass silently, **so that** I stay signed
in for up to 30 days of inactivity, and a stolen renewal pass is caught.

```gherkin
Scenario: renewal replaces the renewal pass
  Given the current renewal pass R1
  When the web app renews
  Then it receives a new pass good for 15 minutes and a new renewal pass R2
  And R1 no longer works

Scenario: reusing a replaced renewal pass ends the whole sign-in
  Given R1 was already replaced by R2
  When someone renews with R1
  Then the renewal is refused as "登入已失效，請重新登入"
  And R2 no longer works either

Scenario: a renewal pass just under 30 days old still works
  Given R2 was handed out 29 days 23 hours ago
  When the web app renews
  Then it receives R3, good for another 30 days

Scenario: a renewal pass 30 days old has expired
  Given R2 was handed out 30 days ago
  When the web app renews
  Then the renewal is refused as "登入已失效，請重新登入"

Scenario: renewal from another website is refused
  Given a renewal request that does not come from the board's own web app
  When it is received
  Then it is refused as "請求來源不被允許"
  And the renewal pass keeps working

Scenario: renewal without a renewal pass is refused
  Given no renewal pass is presented
  When the web app renews
  Then the renewal is refused as "登入已失效，請重新登入"
```

### US-06 — Sign out [P0]
**As the** owner, **I want** to sign out, **so that** this sign-in cannot be renewed again.

```gherkin
Scenario: signing out ends every renewal pass of the sign-in
  Given the owner is signed in with renewal pass R2 that replaced R1
  When the owner signs out
  Then R2 no longer works
  And the browser no longer keeps a renewal pass

Scenario: signing out with no or an unknown renewal pass is answered as done
  Given no renewal pass, or one the board does not know
  When the owner signs out
  Then the sign-out is answered as done

Scenario: sign-out from another website is refused
  Given a sign-out request that does not come from the board's own web app
  When it is received
  Then it is refused as "請求來源不被允許"
  And the renewal pass keeps working
```

### US-07 — Show who is signed in [P1]
**As the** owner, **I want** to see my email in the header, **so that** I know I am signed in.

```gherkin
Scenario: the signed-in owner asks who they are
  Given the owner is signed in as owner@example.com
  When the web app asks who is signed in
  Then it is told owner@example.com

Scenario: a member key cannot ask who the owner is
  Given a caller with an active member key
  When it asks who is signed in
  Then it is refused as "只有登入的擁有者可以使用"
```

### US-08 — Create a member key [P0]
**As the** owner, **I want** to create a key for each agent-team member, **so that** each
member calls the board with its own key.

```gherkin
Scenario: a key is created and shown once
  Given the owner is signed in
  When the owner creates a key for member stock-scanner with the note "daily scan"
  Then the full key is shown once, beginning with "tbk_"
  And the key appears in the list as active, showing only its first characters

Scenario: a 40-character member name is accepted
  When the owner creates a key for a 40-character member name
  Then the key is created

Scenario: a member name too long or blank is refused
  When the owner creates a key for a 41-character member name, or one of only spaces
  Then the creation is refused as "成員名稱須為 1 到 40 個字"

Scenario: a member name in use by an active key is refused, ignoring letter case
  Given an active key for stock-scanner
  When the owner creates a key for Stock-Scanner
  Then the creation is refused as "這個成員名稱已有啟用中的金鑰"

Scenario: the name of a revoked key may be reused
  Given only a revoked key for stock-scanner
  When the owner creates a key for stock-scanner
  Then the key is created

Scenario: a member key cannot create keys
  Given a caller with an active member key
  When it creates a key
  Then it is refused as "只有登入的擁有者可以使用"
```

### US-09 — List and revoke member keys [P0]
**As the** owner, **I want** to see and revoke member keys, **so that** I can stop a member or
a leaked key at once.

```gherkin
Scenario: keys are listed newest first
  Given key A created before key B, and B is revoked
  When the owner lists keys
  Then B is listed before A, each with status, last used and created time

Scenario: a revoked key stops working at once
  Given an active key A
  When the owner revokes A
  Then A is shown as revoked
  And a call made with A is refused as not signed in

Scenario: revoking an already revoked key changes nothing
  Given key A is already revoked
  When the owner revokes A again
  Then A stays revoked with its original revoke time

Scenario: revoking an unknown key is refused
  When the owner revokes a key that does not exist
  Then it is refused as "找不到這把金鑰"

Scenario: a member key cannot list or revoke keys
  Given a caller with an active member key
  When it lists or revokes keys
  Then it is refused as "只有登入的擁有者可以使用"
```

### US-10 — Member keys do the agent team's work, and say who did it [P0]
**As the** owner, **I want** every card, fill and skip to say who made it, **so that** I know
which member advised what.

```gherkin
Scenario: a card published with a member key names its publisher
  Given an active key of desk-chief
  When desk-chief publishes a card
  Then the card is published and shows publisher desk-chief

Scenario: a fill reported with a member key names its reporter
  Given an active key of risk-budget-officer
  When risk-budget-officer reports a fill
  Then the fill shows it was reported by risk-budget-officer

Scenario: an order skipped by the owner names the owner
  Given the owner is signed in
  When the owner skips an order
  Then the order shows it was skipped by owner

Scenario: the owner alone cannot publish
  Given the owner is signed in and presents no member key
  When the owner publishes a card
  Then it is refused as "只有顧問團隊的 API 金鑰可以發布交易卡"

Scenario: a revoked or unknown key is refused
  Given a revoked or unknown member key
  When it reads the latest card
  Then it is refused as not signed in

Scenario: last used is not refreshed within the same minute
  Given a key last used at 20:00:10
  When it is used again at 20:00:50
  Then its last used time stays 20:00:10

Scenario: last used is refreshed after a minute
  Given a key last used at 20:00:10
  When it is used again at 20:01:15
  Then its last used time becomes 20:01:15
```

---

## 4. Business Flow & Logic

- **Flow:** web app makes a secret and its proof and a round-trip marker → owner types email
  and password on the web app's sign-in page → board checks lock, application, return address,
  proof method and credentials → board hands back the return address with a one-time code
  and the marker → web app checks the marker and trades the code plus the secret for a
  short-lived pass and a renewal pass → before the pass runs out (and on start-up) the web app
  renews → signing out ends the sign-in.
- **Core Business Rules:**
  - Sign-in lock: 5 failures within 15 minutes lock sign-in for 15 minutes from the fifth
    failure; refusals during the lock are not counted; a success clears the count.
  - One-time code: 60 seconds, single use, bound to application, return address and proof;
    any trade-in attempt uses it up.
  - Renewal passes belong to a sign-in; each renewal replaces the one presented; reuse of a
    replaced one ends the sign-in; each lives 30 days from issue; only scrambled forms are kept.
  - Member keys: unique active member name (letter case ignored), 1–40 characters, note up to
    200 characters; only scrambled form and the first characters are kept; revoking is final.
  - Attribution: member name for member keys, "owner" for the signed-in owner; older records
    carry none.
  - A caller presenting both a member key and an owner pass is the member (the key is checked
    first, as in v1, but an invalid key with a valid owner pass is the owner).
- **Edge Cases:**
  - Two renewals racing with the same renewal pass: one wins, the other counts as reuse and
    ends the sign-in (the owner signs in again; safety over convenience).
  - Storage unavailable during sign-in or renewal: refused with the general server error;
    nothing is half-issued.

---

## 5. UI/UX Design & Interaction

N/A for the backend; refusal wording is Traditional Chinese, as in v1.

---

## 6. Non-Functional Requirements

- **Security:** password kept only as a scrambled form chosen by the operator; renewal passes
  and member keys kept only scrambled; refusal never reveals which credential was wrong;
  renewal and sign-out honoured only from the board's own web app; the full member key shown
  once; nothing secret written to the log.
- **Compatibility:** sign-in follows the common sign-in standard for public web apps with the
  strong proof method; its refusals follow that standard's shape; every other refusal keeps
  the v1 shape.
- **Performance:** "last used" refreshed at most once a minute per key.

---

## 7. Dependencies & Risks

- **External Dependencies:** none — the board is standalone and uses no outside identity
  service.
- **Known Risks:** the frontend must be released together with this change (the old sign-in
  is removed); every agent-team member needs a new key created by the owner after release,
  because the shared agent key stops working.

---

## 8. Appendix

- v2 stakeholder requirements (sign-in and keys); v1 stakeholder requirements.
- Assumptions are listed in `BRIEF.md` → Context / Background.
