# Contract Traceability Matrix — Owner Sign-in & Member Keys (web v2)

Contract: PRD.md
Design map: ARCH.md
Implementation: `frontend/app`, `frontend/server` (practice board)
Oracle: Acceptance Criteria (26 scenarios) + Core Business Rules (5) + Non-Functional Requirements (2) = 33 clauses

> Static conformance audit: it judges test assertions and code paths against each clause's spec-derived
> oracle. It does not execute invented scenarios.

**Two passes.** Pass 1 (commit 0750120) found that the "which page shows next" part of sign-in, renewal failure
and sign-out lived in pages/layout/middleware that no test reached (AC-1/5/9/14/15/16, BR-4 mis-asserted).
Those rules were moved into `useSignInFlow` with injected navigation and tested (commit 273aa57, behaviour
unchanged; 6 mutations of the new rules all killed), and every non-accepted clause was re-audited.
Rows marked ¹ show their pass-1 status; Status/Route are the pass-2 result as returned by `jev_audit.py`.

## Clauses

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Jev test | Jev code | Status | Route |
|----|--------|------------------------|------|------|------------|------------|----------|----------|--------|-------|
| AC-1 | US-01 the right email and password open the home screen | Home screen opens and the header shows the owner's email | app/composables/use-sign-in-flow.ts:1-73<br>app/domain/service/user-session-service.ts:1-107 | tests/composables/use-sign-in-flow.spec.ts:65-79<br>tests/composables/use-sign-in-flow.spec.ts:92-106 | asserts-oracle | produces-oracle | shallow 0.19 | produces-oracle 0.88 | 🟠 mis-asserted ¹pass 1: mis-asserted | escalate (disagree) |
| AC-2 | US-01 a wrong password or unknown email keeps the owner on the sign-in page | Sign-in page stays, showing 「電子郵件或密碼錯誤」 | app/composables/use-sign-in-flow.ts:1-73<br>app/infrastructure/proxy/board-api-proxy.ts:1-132 | tests/composables/use-sign-in-flow.spec.ts:73-95<br>tests/infrastructure/proxy/auth-proxy.spec.ts:62-75 | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.97 | ✅ conforms ¹pass 1: conforms | accept |
| AC-3 | US-01 a locked sign-in says to try later | Sign-in page stays, showing 「嘗試次數過多，請 15 分鐘後再試」 | app/composables/use-sign-in-flow.ts:1-73<br>app/infrastructure/proxy/board-api-proxy.ts:1-132 | tests/composables/use-sign-in-flow.spec.ts:73-95<br>tests/infrastructure/proxy/auth-proxy.spec.ts:63-76 | asserts-oracle | produces-oracle | asserts-oracle 0.99 | produces-oracle 0.96 | ✅ conforms ¹pass 1: conforms | accept |
| AC-4 | US-01 an unreachable board says to try later | Sign-in page stays, showing 「無法連線，請稍後再試」 | app/composables/use-sign-in-flow.ts:1-73<br>app/infrastructure/proxy/board-api-proxy.ts:1-132 | tests/composables/use-sign-in-flow.spec.ts:73-95 | asserts-oracle | produces-oracle | asserts-oracle 1.00 | produces-oracle 0.97 | ✅ conforms ¹pass 1: conforms | accept |
| AC-5 | US-02 a matching marker and accepted code sign the owner in | Home screen opens, signed in | app/composables/use-sign-in-flow.ts:1-73<br>app/domain/service/user-session-service.ts:1-107 | tests/composables/use-sign-in-flow.spec.ts:92-106 | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.95 | ✅ conforms ¹pass 1: mis-asserted | accept |
| AC-6 | US-02 a marker that differs, or none made, stops the sign-in | Shows 「登入流程已失效，請重新登入」 with a link to the sign-in page; the code is not traded in | app/composables/use-sign-in-flow.ts:1-73<br>app/domain/service/user-session-service.ts:1-107 | tests/composables/use-sign-in-flow.spec.ts:101-119<br>tests/application/user-session-application.spec.ts:109-125 | asserts-oracle | produces-oracle | asserts-oracle 0.87 | produces-oracle 0.91 | ✅ conforms ¹pass 1: conforms | accept |
| AC-7 | US-02 a refused code stops the sign-in | Shows 「登入流程已失效，請重新登入」 with a link to the sign-in page | app/domain/service/user-session-service.ts:1-107<br>app/pages/auth/callback.vue:1-62 | tests/application/user-session-application.spec.ts:120-132 | asserts-oracle | produces-oracle | asserts-oracle 0.89 | produces-oracle 0.84 | ✅ conforms | recheck (mid confidence) |
| AC-8 | US-03 a valid renewal pass opens screens without signing in | Screen shows without the sign-in page | app/composables/use-sign-in-flow.ts:1-73<br>app/application/user-session-application.ts:1-40 | tests/composables/use-sign-in-flow.spec.ts:126-147<br>tests/application/user-session-application.spec.ts:137-149 | asserts-oracle | produces-oracle | asserts-oracle 0.75 | produces-oracle 0.79 | ✅ conforms ¹pass 1: conforms | recheck (mid confidence) |
| AC-9 | US-03 no or a refused renewal pass shows the sign-in page | Sign-in page shows | app/composables/use-sign-in-flow.ts:1-73<br>app/application/user-session-application.ts:1-40 | tests/composables/use-sign-in-flow.spec.ts:127-149 | asserts-oracle | produces-oracle | asserts-oracle 0.76 | produces-oracle 0.92 | ✅ conforms ¹pass 1: mis-asserted | recheck (mid confidence) |
| AC-10 | US-03 a pass about to run out is renewed before asking for data | A new short-lived pass is obtained before the board is asked; data shows | app/domain/service/user-session-service.ts:1-107<br>app/domain/models/domains/session-domain.ts:1-16 | tests/application/user-session-application.spec.ts:153-165<br>tests/domain/models/domains/session-domain.spec.ts:2-17 | asserts-oracle | produces-oracle | asserts-oracle 0.93 | produces-oracle 0.91 | ✅ conforms | accept |
| AC-11 | US-03 a pass with time left is used as it is | Board asked directly with no renewal | app/domain/service/user-session-service.ts:1-107<br>app/domain/models/domains/session-domain.ts:1-16 | tests/application/user-session-application.spec.ts:161-173 | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.39 | ✅ conforms ¹pass 1: conforms | escalate (low confidence) |
| AC-12 | US-03 simultaneous needs share one renewal | Only one renewal happens; both get their data | app/domain/service/user-session-service.ts:1-107 | tests/application/user-session-application.spec.ts:169-181<br>tests/application/user-session-application.spec.ts:179-193 | asserts-oracle | produces-oracle | asserts-oracle 0.99 | produces-oracle 0.92 | ✅ conforms | accept |
| AC-13 | US-03 a pass the board no longer accepts sends the owner to sign in | Sign-in page shows | app/infrastructure/proxy/board-api-proxy.ts:1-132<br>app/plugins/dependencies.ts:1-84 | tests/infrastructure/proxy/board-api-proxy.spec.ts:52-64 | asserts-oracle | produces-oracle | asserts-oracle 0.49 | produces-oracle 0.84 | ✅ conforms ¹pass 1: conforms | escalate (low confidence) |
| AC-14 | US-04 signing out ends the sign-in | Board ends the sign-in and the sign-in page shows | app/composables/use-sign-in-flow.ts:1-73<br>app/domain/service/user-session-service.ts:1-107 | tests/composables/use-sign-in-flow.spec.ts:151-171<br>tests/application/user-session-application.spec.ts:192-206 | asserts-oracle | produces-oracle | asserts-oracle 0.92 | produces-oracle 0.93 | ✅ conforms ¹pass 1: mis-asserted | accept |
| AC-15 | US-04 signing out works even when the board is unreachable | Sign-in page shows and the pass is forgotten on this page | app/composables/use-sign-in-flow.ts:1-73<br>app/domain/service/user-session-service.ts:1-107 | tests/composables/use-sign-in-flow.spec.ts:151-171 | asserts-oracle | produces-oracle | asserts-oracle 0.61 | produces-oracle 0.41 | ✅ conforms ¹pass 1: mis-asserted | escalate (low confidence) |
| AC-16 | US-04 signing out a second time is harmless | Sign-in page shows with no error | app/composables/use-sign-in-flow.ts:1-73<br>app/domain/service/user-session-service.ts:1-107 | tests/composables/use-sign-in-flow.spec.ts:166-176 | asserts-oracle | produces-oracle | asserts-oracle 0.29 | produces-oracle 0.47 | ✅ conforms ¹pass 1: mis-asserted | escalate (low confidence) |
| AC-17 | US-05 a new key is shown once with a copy button | Full key shown with copy button and warning 「這把金鑰只會顯示這一次」; list shows stock-scanner 啟用中 | app/domain/models/entities/created-api-key.ts:1-15<br>app/components/molecules/CreatedApiKeyPanel.vue:1-74 | tests/application/api-key-application.spec.ts:31-47<br>tests/components/molecules/CreatedApiKeyPanel.spec.ts:16-26 | asserts-oracle | produces-oracle | asserts-oracle 0.92 | produces-oracle 0.89 | ✅ conforms | accept |
| AC-18 | US-05 an empty member name is not sent | Nothing sent to the board; member name field asks to be filled in | app/domain/service/api-key-service.ts:1-29<br>app/components/molecules/ApiKeyCreateForm.vue:1-68 | tests/application/api-key-application.spec.ts:44-54<br>tests/components/molecules/ApiKeyCreateForm.spec.ts:16-26 | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.97 | ✅ conforms | accept |
| AC-19 | US-05 the board's refusal is shown in its own words | 「這個成員名稱已有啟用中的金鑰」 is shown | app/infrastructure/proxy/board-api-proxy.ts:1-132<br>app/composables/use-api-key-screen.ts:1-87 | tests/application/api-key-application.spec.ts:52-62<br>tests/composables/use-api-key-screen.spec.ts:41-59 | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.60 | ✅ conforms ¹pass 1: conforms | recheck (mid confidence) |
| AC-20 | US-05 copying the key confirms it | Key is copied and the button shows 「已複製」 | app/composables/use-api-key-screen.ts:1-87<br>app/components/molecules/CreatedApiKeyPanel.vue:1-74 | tests/components/molecules/CreatedApiKeyPanel.spec.ts:24-42<br>tests/composables/use-api-key-screen.spec.ts:56-68 | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.75 | ✅ conforms | recheck (mid confidence) |
| AC-21 | US-05 the full key disappears for good | After creating another or leaving, the earlier key is no longer shown anywhere | app/composables/use-api-key-screen.ts:1-87<br>app/pages/api-keys.vue:1-71 | tests/composables/use-api-key-screen.spec.ts:41-59 | asserts-oracle | produces-oracle | asserts-oracle 0.95 | produces-oracle 0.89 | ✅ conforms | accept |
| AC-22 | US-05 revoking after confirming | Key shows 已撤銷 with no actions | app/components/organisms/ApiKeyList.vue:1-167<br>app/composables/use-api-key-screen.ts:1-87 | tests/components/organisms/ApiKeyList.spec.ts:29-41<br>tests/composables/use-api-key-screen.spec.ts:67-81 | asserts-oracle | produces-oracle | asserts-oracle 0.93 | produces-oracle 0.63 | ✅ conforms | recheck (mid confidence) |
| AC-23 | US-05 cancelling a revoke changes nothing | Key stays 啟用中 | app/components/organisms/ApiKeyList.vue:1-167 | tests/components/organisms/ApiKeyList.spec.ts:39-51 | asserts-oracle | produces-oracle | asserts-oracle 0.61 | produces-oracle 0.87 | ✅ conforms | recheck (mid confidence) |
| AC-24 | US-05 a never-used key says so | Last-used time shows 「從未使用」 | app/domain/models/domains/api-key-domain.ts:1-25 | tests/application/api-key-application.spec.ts:63-85 | asserts-oracle | produces-oracle | asserts-oracle 1.00 | produces-oracle 1.00 | ✅ conforms | accept |
| AC-25 | US-06 a card names its publisher | Shows 「發布者：desk-chief」 | app/domain/models/domains/trade-card-domain.ts:1-43<br>app/components/organisms/TradeCardView.vue:1-175 | tests/application/trade-card-application.spec.ts:44-59<br>tests/components/organisms/TradeCardView.spec.ts:23-39 | asserts-oracle | produces-oracle | asserts-oracle 1.00 | produces-oracle 0.91 | ✅ conforms | accept |
| AC-26 | US-06 an older card shows no publisher | No publisher line is shown | app/domain/models/domains/trade-card-domain.ts:1-43<br>app/components/organisms/TradeCardView.vue:1-175 | tests/application/trade-card-application.spec.ts:45-60<br>tests/components/organisms/TradeCardView.spec.ts:23-40 | asserts-oracle | produces-oracle | asserts-oracle 0.92 | produces-oracle 0.93 | ✅ conforms | accept |
| BR-1 | The short-lived pass is kept only while the page is open; the renewal pass is never visible to page scripts | Access token only in memory; refresh token never handled by scripts | app/infrastructure/proxy/memory-session-storage-proxy.ts:1-25<br>app/infrastructure/proxy/auth-proxy.ts:1-102 | tests/infrastructure/proxy/memory-session-storage-proxy.spec.ts:1-18<br>tests/infrastructure/proxy/auth-proxy.spec.ts:71-97 | asserts-oracle | produces-oracle | asserts-oracle 0.71 | produces-oracle 0.93 | ✅ conforms | recheck (mid confidence) |
| BR-2 | The one-off secret and marker live for one sign-in and are discarded once used | Pending sign-in taken once and removed | app/infrastructure/proxy/browser-pending-sign-in-storage-proxy.ts:1-51<br>app/domain/service/user-session-service.ts:1-107 | tests/infrastructure/proxy/browser-pending-sign-in-storage-proxy.spec.ts:11-23 | asserts-oracle | produces-oracle | asserts-oracle 0.99 | produces-oracle 0.99 | ✅ conforms | accept |
| BR-3 | A refused renewal or a pass the board no longer accepts always ends at the sign-in page | Sign-in page | app/composables/use-sign-in-flow.ts:1-73<br>app/domain/service/user-session-service.ts:1-107 | tests/composables/use-sign-in-flow.spec.ts:127-149<br>tests/infrastructure/proxy/board-api-proxy.spec.ts:52-64 | asserts-oracle | produces-oracle | asserts-oracle 0.63 | produces-oracle 0.76 | ✅ conforms ¹pass 1: conforms | recheck (mid confidence) |
| BR-4 | Signing out always succeeds on the screen | Sign-in page shows regardless of board | app/composables/use-sign-in-flow.ts:1-73<br>app/domain/service/user-session-service.ts:1-107 | tests/composables/use-sign-in-flow.spec.ts:151-171 | asserts-oracle | produces-oracle | asserts-oracle 0.79 | produces-oracle 0.48 | ✅ conforms ¹pass 1: mis-asserted | escalate (low confidence) |
| BR-5 | The full member key is visible only right after its creation | Key only in creation response and page state | app/domain/models/dto/api-key-dto.ts:1-17<br>app/composables/use-api-key-screen.ts:1-87 | tests/application/api-key-application.spec.ts:31-47<br>tests/composables/use-api-key-screen.spec.ts:41-59 | asserts-oracle | produces-oracle | asserts-oracle 0.77 | produces-oracle 0.80 | ✅ conforms | recheck (mid confidence) |
| NFR-1 | Security: nothing proving the sign-in is saved on the device; renewal pass never readable by scripts; one-off secret only sent to the board when trading the code | As stated | app/infrastructure/proxy/memory-session-storage-proxy.ts:1-25<br>app/infrastructure/proxy/auth-proxy.ts:1-102 | tests/infrastructure/proxy/memory-session-storage-proxy.spec.ts:1-18<br>tests/infrastructure/proxy/web-crypto-sign-in-challenge-proxy.spec.ts:1-19 | asserts-oracle | produces-oracle | asserts-oracle 0.79 | produces-oracle 0.83 | ✅ conforms | recheck (mid confidence) |
| NFR-2 | Compatibility: current mobile Safari and Chrome; works against the practice board | Practice board answers the v2 flow | server/mock/mock-board.ts:1-703<br>server/routes/mock/api/v1/[...path].ts:1-109 | tests/server/mock/mock-board.spec.ts:22-84<br>tests/server/mock/mock-board.spec.ts:55-73 | asserts-oracle | produces-oracle | asserts-oracle 0.57 | produces-oracle 0.28 | ✅ conforms ¹pass 1: unclear | escalate (low confidence) |

Notes on rechecks:
- AC-1 (🟠): the redirect, the code exchange and the navigation to the home screen are asserted in
  `use-sign-in-flow.spec.ts`; the header showing the email is the layout binding of `fetchCurrentOwner()`
  (tested in `user-session-application.spec.ts`), and the layout itself is not mounted by any unit test.
  A headless-Chromium run against the practice board and against the real Go backend showed the header
  email after sign-in, but that run is not part of the repository's test suite.
- AC-13: the redirect to the sign-in page on a refused access token is wired in `dependencies.ts`
  (`onSignedOut` → forget session → `navigateTo('/login')`); the proxy test asserts the callback fires once.

## Orphans (code with no clause)

| Code | Description | Verdict |
|------|-------------|---------|
| `app/domain/errors/clipboard-unavailable-error.ts` | message when the clipboard cannot be written | undocumented (UX fallback) |
| `app/composables/use-api-key-screen.ts` revoke failure | revoke refusal shown on the screen | undocumented |
| `server/mock/mock-board.ts` | practice board also locks sign-in after 5 failures, rotates refresh tokens | mirrors the board's rules (backend PRD) |

No orphan implements an Out-of-Scope item (fill/skip reporters are not displayed).

## Summary

- Conforms: 32/33 clauses ✅ (96%)
- Violations: none
- Mis-asserted: AC-1
- Partial: none
- Gaps: none
- Unclear: none
- Orphans: 3 (none out of scope)
- Jev: ran on 33 clauses (pass 1) and re-ran on 17 (pass 2)
- Needs James: AC-1 (disagree), AC-11 (low confidence), AC-13 (low confidence), AC-15 (low confidence), AC-16 (low confidence), BR-4 (low confidence), NFR-2 (low confidence)
