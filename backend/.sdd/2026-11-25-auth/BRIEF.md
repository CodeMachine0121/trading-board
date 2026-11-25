# Owner Sign-in & Member Keys (backend v2) — Requirements Brief

## Goal
Replace the board's v1 way of telling who is calling. The owner now signs in with an
**email address** and password on the board's own sign-in page and then stays signed in for
up to 30 days of inactivity without retyping the password, while each short-lived pass the
browser holds can be renewed, and signing out ends the whole sign-in for good. The single
shared key the agent team used is replaced by **one key per agent-team member**, which the
owner creates, names, lists and revokes from the web screens. Everything the board records on
a member's behalf (published cards, reported fills, skipped orders) now says which member did
it. The board stays standalone: it has its own sign-in and depends on no outside identity
service. Every business rule of v1 that is not mentioned here stays exactly as it was.

## Requirements
- **One owner account.** Exactly one account. Its email and a scrambled form of its
  password that cannot be turned back into it are set by whoever operates the board; the password itself is never written down
  anywhere the operator commits or logs. There is no sign-up, no password change and no
  password reset. If the operator has not set the email or the password form, nobody can
  sign in.
- **Signing in.** The owner types email and password on the board's own sign-in page. The
  email is matched ignoring surrounding spaces and letter case; the password must match
  exactly. A refusal never says whether the email or the password was wrong.
- **Sign-in throttling.** After 5 failed sign-ins within 15 minutes, every further sign-in is
  refused for 15 minutes from the fifth failure — even with the right password — with a
  message that says to try again later, still without saying what was wrong. A successful
  sign-in clears the count. Every failed sign-in counts, whatever email was typed, because
  there is only one account.
- **Only the board's own web app may ask for a sign-in.** The web app is the one registered
  application. It names where the owner should be sent back after signing in; only the
  return addresses the operator allows are accepted. It must prove, when it collects the
  pass, that it is the same party that started the sign-in (a secret it made up beforehand
  and only revealed at the end, using the strong proof method; the weak method is refused).
- **One-time sign-in code.** A successful sign-in hands the web app a one-time code to send
  the owner back with, together with the web app's own round-trip marker unchanged. The code
  is good for 60 seconds, once, only for the same application, the same return address and
  the same proof. Any second use, late use or mismatch is refused.
- **Passes.** Trading the code in gives a short-lived pass (15 minutes) for calling the board,
  and a long-lived renewal pass (30 days) that the browser keeps where page scripts cannot
  read it.
- **Renewal.** Presenting the renewal pass gives a fresh short-lived pass and a fresh renewal
  pass; the old renewal pass stops working. Presenting a renewal pass that was already
  replaced is treated as theft: every renewal pass of that sign-in stops working and the
  owner must sign in again. Each renewal pass lives 30 days from when it was handed out, so
  using the board at least once every 30 days keeps the owner signed in. Renewal passes are
  stored only in a scrambled form.
- **Signing out** ends the whole sign-in: every renewal pass of it stops working. A
  short-lived pass already handed out keeps working until its 15 minutes are up; the web app
  simply forgets it.
- **Renewal and sign-out may not be triggered by another website.** Because the browser
  sends the renewal pass by itself, the board only honours renewal and sign-out requests
  coming from the board's own web app.
- **Who am I.** The signed-in owner can ask the board for the owner's email, to show it in the
  header.
- **Member keys.** The signed-in owner creates a key by giving a member name (1 to 40
  characters after trimming spaces; not shared with another active key, ignoring letter
  case) and an optional note (up to 200 characters). The full key is shown once, at creation,
  and can never be read again; the board keeps only a scrambled form and a short recognisable
  beginning of the key for display. A key looks like a fixed recognisable lead-in followed
  by a long random part.
- **Listing keys.** The owner sees every key, newest first, with member name, note, the
  recognisable beginning, when it was created, when it was last used (blank if never) and
  whether it is active or revoked (and when it was revoked).
- **Revoking keys.** The owner can revoke a key; from that moment it admits nobody, and it can
  never be turned back on. Revoking a key that is already revoked changes nothing. A revoked
  key's member name may be used again for a new key. Keys cannot be edited.
- **What a key may do.** An active key may do exactly what the v1 agent key could: publish
  cards, read everything, report fills and skip orders. A key can never create, list or
  revoke keys; only the signed-in owner can. Publishing a card stays something only a key
  may do.
- **Last used.** Each time an active key is used, its "last used" time is brought up to date,
  at most once a minute per key.
- **Attribution.** A card published with a key records that key's member name as its
  publisher, and the card shows it. A fill reported or an order skipped with a key records
  the member name as who reported it; the same done by the signed-in owner records "the owner".
  Records made before this change have no attribution and show none.
- **The single shared agent key and the username sign-in are gone.**

## Examples (Specification by Example)
Each example lists **only** the data that affects the behavior — nothing more.

### Rule: Signing in
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy) | The owner's email is set, no recent failures | Sign in with the owner's email in different letter case with extra spaces, and the right password | Signed in; a one-time code is handed back with the web app's round-trip marker unchanged |
| 2 (exception) | — | Sign in with the right email and a wrong password | Refused: "email or password is wrong" |
| 3 (exception) | — | Sign in with an unknown email and any password | Refused with exactly the same words as #2 |
| 4 (exception) | The operator set no owner email | Sign in with anything | Refused with the same words as #2 |

### Rule: Sign-in throttling
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (boundary) | 4 failures in the last 15 minutes | 5th wrong attempt at 20:00 | Refused as wrong; the account is now locked until 20:15 |
| 2 (exception) | Locked until 20:15 | Right password at 20:10 | Refused: "too many attempts, try again later" |
| 3 (boundary) | Locked until 20:15 | Right password at 20:15 | Signed in; failure count cleared |
| 4 (boundary) | 5 failures, but the first and fifth are 16 minutes apart | Right password right after | Signed in (not 5 within 15 minutes) |
| 5 (happy) | 3 failures, then a success | 2 more failures | Not locked (the success cleared the count) |

### Rule: Asking for a sign-in
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy) | The board's own web app, return address on the allowed list, strong proof method | Sign in | Code handed back for that return address |
| 2 (exception) | Unknown application | Sign in | Refused; the owner is not sent anywhere |
| 3 (exception) | Return address not on the allowed list | Sign in | Refused; the owner is not sent anywhere |
| 4 (exception) | Weak proof method, or no proof at all | Sign in | Refused |

### Rule: Trading the one-time code in
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy) | Code issued 20 seconds ago, unused | Traded in with the matching secret, same application and return address | Short-lived pass for 15 minutes plus a renewal pass |
| 2 (boundary) | Code issued 61 seconds ago | Traded in correctly | Refused as expired |
| 3 (exception) | Code already traded in once | Traded in again | Refused |
| 4 (exception) | Unused code | Traded in with a secret that does not match | Refused, and the code is used up |
| 5 (exception) | Unused code issued for return address A | Traded in naming return address B | Refused |

### Rule: Renewal and theft detection
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy) | Renewal pass R1, current | Renew | New short-lived pass and renewal pass R2; R1 no longer works |
| 2 (exception) | R1 was already replaced by R2 | Someone renews with R1 | Refused; R2 also stops working; owner must sign in again |
| 3 (boundary) | R2 handed out 29 days 23 hours ago | Renew | Works (R3 handed out, good for another 30 days) |
| 4 (boundary) | R2 handed out 30 days ago | Renew | Refused as expired |
| 5 (exception) | Renewal request comes from another website | Renew | Refused |
| 6 (exception) | No renewal pass presented | Renew | Refused |

### Rule: Signing out
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy) | Signed in with renewal pass R2 | Sign out | R2 and every earlier pass of that sign-in stop working |
| 2 (boundary) | No renewal pass or an unknown one | Sign out | Answered as done (nothing to end) |
| 3 (exception) | Sign-out request comes from another website | Sign out | Refused; nothing ends |

### Rule: Creating a member key
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy) | Signed in as owner | Create for stock-scanner, note "daily scan" | Key shown once in full; listed as active with its beginning |
| 2 (boundary) | — | Member name of 40 characters | Created |
| 3 (boundary) | — | Member name of 41 characters, or only spaces | Refused |
| 4 (exception) | Active key for stock-scanner | Create for Stock-Scanner | Refused: name already in use |
| 5 (boundary) | Only a revoked key for stock-scanner | Create for stock-scanner | Created |
| 6 (exception) | Caller is a member key | Create a key | Refused: only the signed-in owner |

### Rule: Listing and revoking keys
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy) | Keys A (older) and B (newer, revoked) | Owner lists | B then A, each with status, last used, created |
| 2 (happy) | Active key A | Owner revokes A | A is revoked; a call made with A is now refused |
| 3 (boundary) | Key A already revoked | Owner revokes A again | Nothing changes; A stays revoked |
| 4 (exception) | No such key | Owner revokes it | Refused: not found |
| 5 (exception) | Caller is a member key | List or revoke | Refused: only the signed-in owner |

### Rule: Using a member key
| # | Given (only relevant data) | When | Then |
|---|---|---|---|
| 1 (happy) | Active key of desk-chief | Publish a card | Published; the card shows publisher desk-chief |
| 2 (happy) | Active key of risk-budget-officer | Report a fill | Fill recorded as reported by risk-budget-officer |
| 3 (happy) | Signed-in owner | Skip an order | Recorded as skipped by the owner |
| 4 (exception) | Signed-in owner, no key | Publish a card | Refused: only a member key may publish |
| 5 (exception) | Revoked or unknown key | Read the latest card | Refused: not signed in |
| 6 (boundary) | Key last used at 20:00:10 | Used again at 20:00:50 | Last used stays 20:00:10 |
| 7 (boundary) | Key last used at 20:00:10 | Used again at 20:01:15 | Last used becomes 20:01:15 |

## Out of Scope
- More than one user, sign-up, password change or reset, second-factor checks, sign-in with
  outside accounts.
- Key permissions narrower than the v1 agent key, key expiry dates, editing keys.
- Ending a short-lived pass before its 15 minutes are up.

## Open Decisions
Items the PRD author should resolve:
- Exact wording of refusals and the shape of answers for the sign-in steps (they follow the
  common sign-in standard's shape; everything else keeps the v1 shape).
- How the board recognises that a renewal or sign-out request comes from its own web app.

## Context / Background
- Source: the v2 stakeholder requirements for sign-in and keys (final decisions) and
  the v1 stakeholder requirements for everything unchanged. Run unattended; no
  clarification questions were asked — every question was answered from those documents.
- Stages chosen by the stakeholder: clarify → prd → architecture → implement (code-first) →
  improve-codebase → contract.
- **Assumptions** (not covered by the documents; simplest secure option chosen):
  1. Every failed sign-in counts towards the lock, whatever email was typed; attempts refused
     during the lock do not extend it; a success clears the count.
  2. The email is matched ignoring case and surrounding spaces.
  3. A code traded in with a wrong secret is used up (cannot be retried).
  4. The renewal pass is handed to the browser only in the protected place it keeps it, not
     in a form page scripts can read.
  5. Each renewal pass lives 30 days from when it was handed out (sliding), with no overall
     cap on how long one sign-in can be renewed.
  6. Signing out with no or an unknown renewal pass is answered as done.
  7. Member names are unique among active keys ignoring letter case; the note is at most 200
     characters.
  8. Revoking an already revoked key is answered with the key unchanged.
  9. Skipping an order records who skipped it, the same way a fill records who reported it.
  10. Cards, fills and skips recorded before this change carry no attribution.
  11. Publishing a card remains something only a member key may do (as in v1).
