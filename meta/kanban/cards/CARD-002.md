# CARD-002: Card editing and the full API contract

**Status:** blocked
**Priority:** P1
**Category:** feature
**Estimate:** 1c
**Complexity:** standard
**Closure:** closed
**Revision pending:** false
**Skill:** golang-pro
**TDD:** —
**Branch:** card/002-card-editing-api-contract
**Worktree:** —
**Source:** meta/architecture/handoff.md#increment-2
**Idea:** — (no match in backlog-scored.yml)
**Wave:** 2
**Depends on:** CARD-001
**Touches:** internal/board/**.go, internal/httpapi/**.go
**UI surface:** none
**Readiness:** 6/7 — gaps: dependencies_resolved
**Review score:** —
**Started:** —
**Closed:** —
**Actual:** —
**Merge commit:** —
**Blocked by:** readiness — dependencies_resolved

## What to implement

Increment 2 of `meta/architecture/handoff.md` — card editing and the full API contract
on top of CARD-001's walking skeleton.

Components: COMP-001, COMP-002 (edit, delete, title rules, error mapping).

- **`internal/board` (COMP-001).** The full title rule INV-001 on add and on edit: not
  empty after trimming whitespace and at most 80 characters (ASM-0002: the limit counts
  runes after trimming). `Edit` changes title and/or description without changing the
  card's column or position; `Delete` removes the card and re-densifies its column
  (INV-004). Both keep the ADR-0006 lock discipline.
- **`internal/httpapi` (COMP-002).** `PATCH /api/cards/{id}` (partial update, ASM-0001;
  returns the edited card, ADR-0009) and `DELETE /api/cards/{id}` (204 with no body,
  ADR-0009); unknown id is `not_found`, a title breaking INV-001 is `validation_failed`
  (CON-004). After this card all five endpoints of CON-003 are served by the standard
  ServeMux (ADR-0002).
- **Assumptions validated by tests (Increment 2 collapses them):** ASM-0001 (PATCH is a
  partial update), ASM-0002 (80 limit counts runes after trimming), ASM-0003 (malformed
  move body is `validation_failed` — the behaviour and its test
  `PropertyTest_MoveRefusal_FirstFailingCheckWins` land in CARD-001; this wave's
  checkpoint re-runs it).

**Checkpoint (wave 2):** every API acceptance criterion passes (title rules incl. 80/81
characters on add and edit, delete closes the gap, unknown id is `not_found`), and the
five endpoints of CON-003 are all served by the standard ServeMux.

## Acceptance criteria

- AC-006 (FR-003): given card A in doing at position 1 with title 'Old'; when PATCH /api/cards/A with title 'New'; then the card has title 'New' and stays in doing at position 1 — test: TestEditCard_UpdatesTitleKeepsPlace
- AC-007 (FR-003): given no card with id 'missing'; when PATCH /api/cards/missing with title 'New'; then returns code not_found — test: TestEditCard_UnknownIdNotFound
- AC-008 (FR-003): given card A with title 'Old'; when PATCH /api/cards/A with title ''; then returns code validation_failed and the title remains 'Old' — test: TestEditCard_RejectsEmptyTitle
- AC-009 (FR-004): given todo holds cards A, B, C at positions 0, 1, 2; when DELETE /api/cards/B; then B is gone and A is at position 0, C at position 1 — test: TestDeleteCard_RemovesAndClosesGap
- AC-010 (FR-004): given no card with id 'missing'; when DELETE /api/cards/missing; then returns code not_found and the board is unchanged — test: TestDeleteCard_UnknownIdNotFound
- AC-019 (FR-008): given an empty board; when POST /api/cards with a title of exactly 80 characters; then returns 201 and the card is created — test: TestTitle_80CharsAccepted
- AC-020 (FR-008): given an empty board; when POST /api/cards with a title of 81 characters; then returns code validation_failed and no card is created — test: TestTitle_81CharsRejected
- AC-021 (FR-008): given card A with title 'Old'; when PATCH /api/cards/A with title of spaces and a tab; then returns code validation_failed and the title remains 'Old' — test: TestTitle_WhitespaceOnlyRejectedOnEdit
- AC-036 (FR-008): given card A with title 'Old'; when PATCH /api/cards/A with a title of 81 characters; then returns code validation_failed and the title remains 'Old' — test: TestTitle_81CharsRejectedOnEdit
- AC-037 (FR-008): given card A with title 'Old'; when PATCH /api/cards/A with a title of exactly 80 characters; then returns 200 and the title is the 80-character value — test: TestTitle_80CharsAcceptedOnEdit

## Engineering constraints

- EC(ADR-0006/R1): the Edit and Delete methods this card completes hold the Board's mutex for their whole body, and a Snapshot taken around them is a copy that they never change — test: PropertyTest_Board_ConcurrentOpsSnapshotIsCopy (extended with edit and delete actions)

## Guardrails

- G-1: Increment 1 stays usable — board read, add and move behaviour, the WIP limit under concurrency and dense order unchanged (rollback clause of Increment 2; test: TestGetBoard_GroupsCardsByColumnInOrder, TestMoveCard_20ConcurrentMovesKeepDoingAtMost3, PropertyTest_DenseOrder_AnyOperationSequence)

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
- engineering_constraint_coverage — gap_free
- architecture_traceability — gap_free
- blocking_decisions_resolved — gap_free
- dependencies_resolved — gap: dependencies: CARD-001 (status 'blocked' does not resolve) — a declared dependency does not resolve.
- design_seed_for_user_facing — gap_free_by_non_applicability: not applicable — the card's own `UI surface:` machine stamp reads `none` (step 6b, FR-061); there is no surface to seed.
- open_subject_answered — gap_free

## Architecture context

- **FR:** FR-003, FR-004, FR-008
- **CON:** CON-001, CON-003, CON-004, CON-008
- **ADR:** ADR-0002, ADR-0004, ADR-0006, ADR-0009
- **Components:** COMP-001, COMP-002
- **Trace:** meta/architecture/trace.yml

## Worktree notes

—
