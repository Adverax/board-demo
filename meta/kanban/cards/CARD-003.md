# CARD-003: The board page and the browser check

**Status:** done
**Priority:** P2
**Category:** feature
**Estimate:** 2c
**Complexity:** architectural
**Closure:** closed
**Revision pending:** false
**Skill:** golang-pro
**TDD:** —
**States:** enumerable
**Branch:** card/003-board-page-browser-check
**Worktree:** —
**Source:** meta/architecture/handoff.md#increment-3
**Idea:** — (no match in backlog-scored.yml)
**Wave:** 3
**Depends on:** CARD-002
**Touches:** web/**, cmd/board/**.go, internal/httpapi/**.go, go.mod, go.sum
**UI surface:** user-facing
**Readiness:** 7/7
**Review score:** 8.6
**Started:** 2026-10-24T07:34:24Z
**Closed:** 2026-10-24T08:34:28Z
**Actual:** 0.1d
**Merge commit:** 98e99b3091ac47e98b07501f197a37b877d70ff6
**Blocked by:** —

## What to implement

Increment 3 of `meta/architecture/handoff.md` — the page and the browser check.

Components: COMP-003 (embedded page), COMP-004 (serving and embedding), COMP-002 (page route).

- **`web/` (COMP-003, CON-005, ADR-0004).** Plain HTML, CSS and JavaScript with no build
  step, plus the Go file holding the `//go:embed` declaration that exports the embedded
  file system. One page (SCR-001, route `/`) shows the three columns `todo`, `doing`,
  `done` with their cards; every card has move left, move right, edit and delete
  controls; the page has an add form (FR-010). "Move left" on a `todo` card and "move
  right" on a `done` card are rendered but disabled (ADR-0008). After every action,
  succeeded or refused, the page shows any reason and then re-reads `GET /api/board` and
  re-renders; no polling (ADR-0009, ASM-0004). A refused action shows its reason on the
  page (FR-011). Styling only from `meta/design/tokens.css` (see Design context).
  Selectors `#edit-card` and `#add-submit` from `domain/screens.yml` recipes are
  placeholders the implementation honours.
- **`internal/httpapi` (COMP-002).** The page route `GET /` serving the embedded page
  from the same process.
- **`cmd/board` (COMP-004).** Wire the embedded file system into the server; the binary
  keeps `PORT` with default 8080 (CON-006, ADR-0003).
- **E2E (ADR-0010).** Headless Chrome driven from Go tests through chromedp (test-only
  dependency in `go.mod`), tests named `TestE2E_*` and run under `go test`, against the
  built binary or an httptest server. A Chrome/Chromium must be present where the suite
  runs; without it the suite skips or fails explicitly, never passes silently. Use the
  Go test names from the acceptance criteria below; the `e2e/root.spec.ts::<state>`
  names in `screens.yml`/`trace.yml` are state labels only (handoff "Known
  limitations").

Refusal reasons shown on the page (FR-011, the mechanism this card owns):

| state | output | channel | test |
| --- | --- | --- | --- |
| no refusal | RefusalNotice hidden, board re-read and re-rendered | page DOM | TestE2E_AddFormCreatesCardInTodo |
| wip_limit_exceeded | reason for `wip_limit_exceeded` in RefusalNotice, card stays | page DOM (`role="alert"`) | TestE2E_WipRefusalShownOnPage |
| validation_failed | reason for `validation_failed` in RefusalNotice, no card added | page DOM (`role="alert"`) | TestE2E_InvalidTitleReasonShownOnPage |
| invalid_move | reason for `invalid_move` in RefusalNotice, board re-read shows the real place | page DOM (`role="alert"`) | TestE2E_InvalidMoveReasonShownOnPage |
| not_found | reason for `not_found` in RefusalNotice, board re-read drops the card | page DOM (`role="alert"`) | TestE2E_NotFoundReasonShownOnStalePage |

**Checkpoint (wave 3):** with the binary running on `PORT=9090`, a browser run of the
planned screen states passes: board render, add form, edit saved, card deleted, card
moved, WIP-limit refusal, `invalid_move` and `not_found` reasons on a stale page,
blank-title refusal (as the `TestE2E_*` Go tests of the criteria below).

## Acceptance criteria

- AC-024 (FR-010): given the board holds todo=[A], doing=[B], done=[C]; when the page is opened in a browser; then three columns titled todo, doing, done are shown with cards A, B, C in them respectively — test: TestE2E_PageShowsThreeColumnsWithCards
- AC-025 (FR-010): given card A is on the board; when the page is opened in a browser; then card A shows move left, move right, edit and delete controls — test: TestE2E_CardShowsFourControls
- AC-026 (FR-010): given an empty board; when a title 'Ship it' is submitted through the add form; then a card 'Ship it' appears in the todo column — test: TestE2E_AddFormCreatesCardInTodo
- AC-027 (FR-011): given doing holds 3 cards and card A is in todo; when move right is clicked on A; then the page shows the reason for code wip_limit_exceeded and A stays in todo — test: TestE2E_WipRefusalShownOnPage
- AC-028 (FR-011): given the add form is open; when a blank title is submitted; then the page shows the reason for code validation_failed and no card is added — test: TestE2E_InvalidTitleReasonShownOnPage
- AC-034 (FR-011): given the page shows card A in doing, but another member has since moved A to todo; when move right is clicked on A (the page sends to 'done'); then the page shows the reason for code invalid_move and A stays in todo — test: TestE2E_InvalidMoveReasonShownOnPage
- AC-035 (FR-011): given card A is shown on the page and another member has already deleted A; when move right is clicked on A; then the page shows the reason for code not_found — test: TestE2E_NotFoundReasonShownOnStalePage
- AC-031 (NFR-003): given the binary is started with PORT unset; when GET / is requested on port 8080; then returns 200 with the board page — test: TestE2E_BinaryServesPageOnDefaultPort
- AC-032 (NFR-003): given the binary is started with PORT=9090; when GET / is requested on port 9090; then returns 200 with the board page — test: TestE2E_BinaryHonoursPortVariable

## Guardrails

- G-1: The API from Increments 1-2 is unaffected — the five CON-003 endpoints, their response bodies (ADR-0009) and error codes (CON-004) unchanged (rollback clause of Increment 3; test: TestGetBoard_GroupsCardsByColumnInOrder, TestMoveCard_IntoFullDoingRefused, TestEditCard_UpdatesTitleKeepsPlace, TestDeleteCard_RemovesAndClosesGap)

## System contract

- ADR-0002/R1 — The module has no third-party runtime dependencies; HTTP routing uses the standard net/http ServeMux (check: review-lens)
- ADR-0004/R1 — Package internal/board does not import net/http or any other HTTP concern (check: review-lens)
- ADR-0004/R2 — Handlers in internal/httpapi change cards only through the exported Board API (check: review-lens)
- ADR-0006/R1 — Every exported Board method locks the Board's mutex for its whole body; Snapshot returns a copy, never internal slices (check: review-lens)
- ADR-0007/R1 — A move refusal reports the first failing check in the order not_found, validation_failed, invalid_move, wip_limit_exceeded; same-column and unknown column names are invalid_move (check: review-lens)
- INV-001 — A card title is not empty after trimming whitespace and is at most 80 characters. (check: TestTitle_80CharsAcceptedOnEdit, TestTitle_81CharsRejected, TestTitle_81CharsRejectedOnEdit, TestTitle_WhitespaceOnlyRejectedOnEdit)
- INV-002 — The doing column holds at most 3 cards; a move into a full doing is refused and changes nothing. (check: PropertyTest_Doing_NeverExceedsWIPUnderConcurrency, TestMoveCard_20ConcurrentMovesKeepDoingAtMost3, TestMoveCard_IntoDoingAtTwoSucceeds, TestMoveCard_IntoFullDoingRefused, TestMoveCard_LeftIntoFullDoingRefused)
- INV-003 — A card moves only to a neighbouring column (todo-doing, doing-done, and back); any other move is refused and changes nothing. (check: TestMoveCard_DoneToTodoRefused, TestMoveCard_TodoToDoneRefused)
- INV-004 — Within every column card positions are exactly 0..n-1 without gaps or duplicates after any operation, including delete and move. (check: PropertyTest_DenseOrder_AnyOperationSequence, TestDeleteCard_RemovesAndClosesGap, TestDenseOrder_DeleteFirstShiftsUp, TestDenseOrder_MoveKeepsBothColumnsDense)
- INV-005 — The board has exactly three columns in fixed order todo, doing, done. (check: TestGetBoard_GroupsCardsByColumnInOrder)

## Readiness

- testable_acceptance_criteria — gap_free
- engineering_constraint_coverage — gap_free_by_non_applicability: not applicable — no requirement this candidate names declares `engineering_constraints:` and step 6c5 promoted none, so the card format's "omit the section entirely" applies.
- architecture_traceability — gap_free
- blocking_decisions_resolved — gap_free
- dependencies_resolved — gap_free
- design_seed_for_user_facing — gap_free
- open_subject_answered — gap_free

## Architecture context

- **FR:** FR-010, FR-011
- **NFR:** NFR-003
- **CON:** CON-005, CON-006, CON-008
- **ADR:** ADR-0001, ADR-0002, ADR-0003, ADR-0004, ADR-0008, ADR-0009, ADR-0010
- **Components:** COMP-003, COMP-004, COMP-002
- **Trace:** meta/architecture/trace.yml

## Design context

- **Design system:** meta/design/ — brief.md (direction "Soft depth / Humanist" + anti-patterns), tokens.css (all visual values), components.md (inventory + states)
- **UI components:** AppShell, Board, Column (incl. the WIP-limit accent outline), Card, CardActions (IconButton ×4, edge controls disabled per ADR-0008), AddCardForm, EditCardForm, RefusalNotice — all registered in components.md, none built yet; implement every listed state
- **Screens:** SCR-001 `/` (domain/screens.yml) — states: empty, populated, refreshed_after_action, add_form, edit_mode, edit_saved, card_deleted, card_moved, doing_at_wip_limit, edge_controls_disabled (FR-010); refused_wip_limit_exceeded, refused_invalid_move, refused_not_found, refused_validation_failed (FR-011)
- **Standards:** forge:engineering-standards §11 (tokens-only styling, all listed states, a11y minimum: WCAG AA, accessible names on the four controls, `role="alert"` on RefusalNotice)

## Failure matrix

Config values: `REQUEST_TIMEOUT_MS = 5000` (web/static/app.js, one bound per fetch);
an action is at most two requests (mutation, then `GET /api/board`), so a click settles
in at most `2 × REQUEST_TIMEOUT_MS = 10 s`. Server timeouts are unchanged
(`ReadTimeout 10 s`, `WriteTimeout 10 s`, cmd/board). E2E: `e2eTimeout = 30 s` per test,
`startTimeout = 10 s` for the built binary.

| operation / boundary | failure mode | declared behaviour | bound | test |
| --- | --- | --- | --- | --- |
| page fetch `/api/*` (any) | network error (server gone, connection refused) | RefusalNotice (`role="alert"`, `data-code="network"`) shows "Could not reach the board…"; the last rendered board stays; controls usable again | fails in ≤ REQUEST_TIMEOUT_MS; notice shown before the next frame | TestE2E_NetworkErrorReasonShown |
| page fetch `/api/*` | slow response | the request is aborted by an AbortController after REQUEST_TIMEOUT_MS and treated exactly as a network error (`data-code="network"`) | 5000 ms per request, ≤ 10 000 ms per action | TestE2E_SlowResponseTimesOutWithReason |
| page fetch `/api/*` | 4xx with `{"code":...}` | the code's reason is shown (table in "What to implement"); board re-read and re-rendered; reason stays visible after the re-render | one mutation + one re-read | TestE2E_WipRefusalShownOnPage, …InvalidTitle…, …InvalidMove…, …NotFound… |
| page fetch `/api/*` | 4xx/5xx with a non-JSON body or JSON without a string `code` (e.g. the stdlib 500 "internal error") | notice `data-code="unexpected"`: "The server sent an unexpected response (HTTP <status>)."; board re-read | same as above | TestE2E_PageLogicBranches (reason mapping) |
| page fetch `/api/*` | unknown `code` value | notice shows "The action was refused (<code>)." | — | TestE2E_PageLogicBranches |
| page fetch `/api/*` | 2xx with a non-JSON body | GET /api/board (or JSON without a `columns` array of `{name, cards[]}`): notice `unexpected` (HTTP 200), last board stays; mutations: every body is read to its end but its content is ignored (only the status matters, ADR-0009), so no failure mode | — | TestE2E_PageLogicBranches |
| re-read after an action | the re-read itself fails | notice replaced by the re-read's reason (network/unexpected); last rendered board stays | ≤ REQUEST_TIMEOUT_MS | TestE2E_NetworkErrorReasonShown (initial load path shares the code) |
| page actions | double click / a second action while one is in flight | at most ONE action in flight per page: a click while busy is ignored (no request sent); `#board[aria-busy="true"]` and the add/save submit buttons are `disabled` until the re-render | 1 in-flight action | TestE2E_EditSaveMoveAndDeleteFlow (two synchronous clicks → one move) |
| stale page (ASM-0004, no polling) | another member moved/deleted the card | the click is sent as rendered; the server's code (`invalid_move` / `not_found`) is shown; the re-read shows the real board. No polling: the page changes only after the user's own action | one mutation + one re-read | TestE2E_InvalidMoveReasonShownOnPage, TestE2E_NotFoundReasonShownOnStalePage |
| stale page while editing | the edited card was deleted meanwhile | Save → `not_found` reason; the re-render drops the card and the edit form with it | — | covered by not_found mapping (TestE2E_PageLogicBranches) |
| embedded FS | a file named in the `//go:embed` patterns is missing | compile error (go:embed fails on a pattern with no match) — cannot ship | build time | `go build ./...` |
| embedded FS | `index.html` absent from the fs.FS handed to httpapi.NewApp | NewApp returns an error; `cmd/board` prints it and exits 1 before listening | at startup, before Listen | TestNewApp_MissingIndexIsAnError |
| embedded FS | request for an asset not in the FS | `GET /static/<x>` → 404 from http.FileServerFS; no other route affected | — | TestNewApp_ServesPageAndAssets |
| tokens copy | web/static/tokens.css drifts from meta/design/tokens.css | `go test ./web` fails naming the fix command | every `make test` | TestTokensCSS_IsVerbatimCopyOfDesignTokens |
| e2e harness | no Chrome/Chromium found (CHROME_PATH unset, none on PATH or in the macOS app path) | the browser test FAILS (t.Fatal, not Skip) with "e2e: no Chrome/Chromium found…"; `make test-e2e` therefore exits non-zero | immediate | by construction (findChrome) |
| e2e harness | the built binary fails to start / never answers | test fails with the binary's stderr | `startTimeout = 10 s` | TestE2E_BinaryServesPageOnDefaultPort / …HonoursPortVariable |
| e2e harness | port 8080 / 9090 already in use (AC-031/AC-032 need the fixed ports) | test fails up front with "port N is in use" (probe before start), never attaches to a foreign server | immediate | same two tests |
| e2e harness | browser tests' app port | httptest server on 127.0.0.1:0 (free port), so no conflict is possible | — | all browser tests |
| e2e harness | a test hangs (page never reaches the expected state) | the chromedp context deadline fails the test with the step that timed out | `e2eTimeout = 30 s` per test | all browser tests |
| e2e harness | console error, uncaught exception, failed request, or an unexpected 4xx/5xx | test fails listing every event; a test that drives a refusal declares the expected `/api/*` status, anything else still fails | checked at the end of every test | all browser tests |

## Worktree notes

—
- Skill `golang-pro` is not available in this environment; implemented without it (forge:engineering-standards loaded).
- STRUCTURE: `web/` is a Go package `web` with `web.go` (`//go:embed index.html static/tokens.css static/app.css static/app.js`, unexported `files`, exported `func Page() fs.FS`), `index.html` at the root and every asset under `static/`. Why: explicit patterns turn a missing file into a compile error, and a function rather than an exported var keeps callers from reassigning the embed.
- STRUCTURE: page route is `httpapi.NewApp(b *board.Board, page fs.FS) (http.Handler, error)` in `internal/httpapi/page.go`. It takes the same ServeMux as `New` (factored into unexported `apiMux`) and adds `GET /{$}` → index.html (with CSP `default-src 'self'`) and `GET /static/{name}` → `http.ServeFileFS`. `New(b)` is unchanged, so the G-1 contract tests and their ~40 call sites stay as they were. Why: the C4 view has httpapi serving GET / from the fs.FS that cmd/board hands it; an error return makes a page without index.html stop the binary at startup.
- STRUCTURE: `cmd/board/main.go` builds `httpapi.NewApp(board.New(), web.Page())` before `net.Listen`; `addr`/`serve` are unchanged.
- STRUCTURE: the e2e suite is `cmd/board/e2e_test.go` (package main, `TestE2E_*`). Browser tests run `httpapi.NewApp` over a real Board and `web.Page()` on `httptest.NewServer` (free port), seed the board and play "another member" through the Board API. AC-031/AC-032 `go build` the real binary into `t.TempDir()` and start it on 8080/9090 (port probed first, fail if taken); the 9090 test also opens the page in Chrome (wave-3 checkpoint). Why: in-process seeding is deterministic, and the binary tests cover exactly what ships.
- STRUCTURE: Chrome is found via `CHROME_PATH`, then `/Applications/Google Chrome.app/...`, `/Applications/Chromium.app/...`, then `google-chrome`, `chromium`, `headless-shell` etc. on PATH. No Chrome means **t.Fatal, never t.Skip**, so `make test-e2e` cannot be green without a browser run (verified with `CHROME_PATH=/nonexistent`). It runs headless with chromedp's default throwaway user-data dir.
- STRUCTURE: console cleanliness is asserted in every browser test via `chromedp.Events` watchers (console error/warning/assert, uncaught exceptions, Log-domain warnings/errors such as CSP violations, `Network.loadingFailed`, any response ≥ 400). A test that drives a refusal whitelists only that `/api/*` status (`br.allow(409)`), and only network-failure tests set `allowNetworkFailure()`.
- STRUCTURE: settle signal is `#board[aria-busy="false"]` (a real a11y attribute, no test-only hook). The tests wait on "expected outcome && settled" with `chromedp.Poll`; there is no sleep.
- STRUCTURE: the tokens copy `web/static/tokens.css` is the marker line `/* derived from meta/design/tokens.css — do not edit … */` followed by `meta/design/tokens.css` byte for byte. `web/web_test.go::TestTokensCSS_IsVerbatimCopyOfDesignTokens` compares the whole file with marker+source exactly; the regeneration command is in web/README.md.
- STRUCTURE: selectors. `#add-submit` is the add button. `#edit-card` is the EditCardForm (only one card is in edit mode at a time, so the id is unique); the per-card Edit buttons are `[data-card-id=ID] button[data-action=edit]` (ids cannot repeat across cards). Cards carry `data-card-id`, columns `data-column`, controls `data-action=left|right|edit|delete`, and the notice carries `data-code`.
- Design: tokens only. Literals are the AppShell layout values from components.md (1100px → `68.75rem`, ≤760px → `47.5rem`; tokens.css has no layout role and media queries cannot read custom properties). The 28px icon button is `calc(var(--space-6) + var(--space-1))`. Rendered and inspected in light and dark at 1200px and 390px (screenshots in the session scratchpad); matches previews/key-screen.html.
- Deviation from the preview: the icon sprite is hidden with a CSS class, not the `hidden` attribute. Chrome ignores `hidden` on inline `<svg>` and reserved 300×150px; the CSP also forbids inline `style`.
- Bug found while testing: an unread 204 DELETE body was reported by Chrome as `net::ERR_ABORTED`. The page now reads every response body to its end.
- Tests added: e2e `TestE2E_` + PageShowsThreeColumnsWithCards, CardShowsFourControls, AddFormCreatesCardInTodo, WipRefusalShownOnPage, InvalidTitleReasonShownOnPage, InvalidMoveReasonShownOnPage, NotFoundReasonShownOnStalePage, BinaryServesPageOnDefaultPort, BinaryHonoursPortVariable (the 9 AC tests), plus EditSaveMoveAndDeleteFlow, NetworkErrorReasonShown, SlowResponseTimesOutWithReason, PageLogicBranches (the remaining FR-010 states and the failure matrix). Unit: TestNewApp_ServesPageAndAssets, TestNewApp_MissingIndexIsAnError, TestTokensCSS_IsVerbatimCopyOfDesignTokens, TestPage_LinksOnlyEmbeddedAssets. The e2e suite takes about 15 s and passed 6 runs in a row with no flake.
- go.mod: `github.com/chromedp/chromedp v0.21.0` and `github.com/chromedp/cdproto v0.158.4` are imported only from `cmd/board/e2e_test.go`; no runtime dependency (ADR-0002/R1). Note: chromedp v0.21 has the generic `Action[T]` API (`chromedp.Do`, `Run[T]`, `Events`).
- SCOPE+ cmd/board/README.md, internal/httpapi/README.md: keep the package doc true (NewApp, page wiring, e2e suite). New web/README.md (engineering-standards §8).
- Gates: `make test`, `make test-race`, `make test-e2e`, `go vet ./...` all pass. The G-1 tests are unchanged and green.
- [Build gate] cycle 1: `make test`, `make test-race`, `go vet ./...`, `make test-e2e` (13 TestE2E_* in real headless Chrome, ~16s, ports 8080/9090 free before and after) pass on 4259944. `make lint` (now working: golangci-lint 2.15.0 built with go1.27.3) reports 4 issues: 3 errcheck findings in tests of CARD-001 (cmd/board/main_test.go:51,:65, internal/httpapi/api_test.go:353; postponed on purpose, tracked in backlog.md) and 1 NEW staticcheck QF1001 in this card's cmd/board/e2e_test.go:335 (De Morgan), which belongs to this card.
- [Scope gate] cycle 1: in_scope (16 files; the only files outside Touches are the package READMEs cmd/board/README.md and internal/httpapi/README.md kept true by the docs step; no guardrail hit).
- [Review 1/3] score 8.6 — critical 0, important 0, minor 8, out-of-scope 2 (reviewer: forge 2026.10.10, model opus; independent capture: 20 frames, light/dark, 1200 and 390 px; 6 mutants killed). Report: meta/review/20261024T081929Z-CARD-003-cycle1.yml
- [Severity gate] cycle 1: PASS — 0 critical / 0 important (score 8.6 >= 8). Minor, non-gating and NOT fixed in this card (to follow-up): (1) web/static/app.js:244 an unsaved edit draft is lost when another action re-renders (reproduced); (2) cmd/board/e2e_test.go:335 staticcheck QF1001; (3) e2e_test.go:754 and :780 two more errcheck findings hidden by golangci-lint max-same-issues (5 errcheck in total, not 3); (4) failure-matrix notice texts drift from app.js:99-103 and a timeout-after-send row is missing; (5) screens.yml edit_mode recipe `click #edit-card` unreachable (route through forge:architect); (6) submitting / being-deleted states have no visual (components.md states); (7) page logic is tested only through the browser tier (accepted standing exception: no JS unit runner under CON-005/ADR-0010); (8) slow-response test spends ~6.5 s of real time. Out of scope: serve() Shutdown waits 5 s on a dialed-but-unused connection and exits 1 (CARD-001 code); the 3 postponed errcheck findings.
- [Review sync] 1 report(s) → meta/review/
- [Commit] 4259944 feat(web): add the board page and the browser check (single commit by the implementation agent via forge:commit; no attribution lines; no meta/ paths in the branch diff).
- [AC/EC check] All criteria/constraints ✓ (head-sha: 4259944cf5d2e5c815b2ae03049f6ae7591e41d6) (evidence):
  15 items, 15 demonstrated: AC-024..028, AC-031, AC-032, AC-034, AC-035, state table rows S1..S5 (TestE2E_AddFormCreatesCardInTodo, TestE2E_WipRefusalShownOnPage, TestE2E_InvalidTitleReasonShownOnPage, TestE2E_InvalidMoveReasonShownOnPage, TestE2E_NotFoundReasonShownOnStalePage), G-1. All 13 TestE2E_* run in real headless Chrome (`ok ... 21.4s`), ports 8080/9090 free before and after; G-1 tests green and no existing test edited (only 3 new test files); api.go change reviewed as additive (New split into New + apiMux).
- [Merge gate] dependency 2026-10-24T08:32:56Z PASS (deps:c5342f701863613f22d9114b18dcf48d374b1f312e014e3810964794793ed7f6, mode=interactive, scan=scanned) — 0 blocking, 0 allowlisted, 0 non-blocking finding(s)
- [Merge gate] PASS (tree-sha f17e33d759f0ab13a6f97a37cf6d7f631eeec2dc, scope=full, 17s, cmd="go build ./... && go test ./...")

