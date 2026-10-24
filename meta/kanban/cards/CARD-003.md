# CARD-003: The board page and the browser check

**Status:** blocked
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
**Readiness:** 6/7 — gaps: dependencies_resolved
**Review score:** —
**Started:** —
**Closed:** —
**Actual:** —
**Merge commit:** —
**Blocked by:** readiness — dependencies_resolved

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
- dependencies_resolved — gap: dependencies: CARD-002 (status 'blocked' does not resolve) — a declared dependency does not resolve.
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

## Worktree notes

—
