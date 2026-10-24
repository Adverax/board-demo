# CARD-001: Moves under concurrency, end to end (walking skeleton)

**Status:** blocked
**Priority:** P1
**Category:** feature
**Estimate:** 2c
**Complexity:** architectural
**Closure:** closed
**Revision pending:** false
**Skill:** golang-pro
**TDD:** —
**States:** enumerable
**Branch:** card/001-moves-under-concurrency
**Worktree:** —
**Source:** meta/architecture/handoff.md#increment-1
**Idea:** — (no match in backlog-scored.yml)
**Wave:** 1
**Depends on:** —
**Touches:** internal/board/**.go, internal/httpapi/**.go, cmd/board/**.go, go.mod, go.sum
**UI surface:** none
**Readiness:** 6/7 — gaps: blocking_decisions_resolved
**Review score:** —
**Started:** —
**Closed:** —
**Actual:** —
**Merge commit:** —
**Blocked by:** readiness — blocking_decisions_resolved

## What to implement

Increment 1 of `meta/architecture/handoff.md` — the walking skeleton: the riskiest
invariants (WIP limit, adjacency, dense order under one lock) end to end through the
board read, add and move endpoints, served by the real binary.

Components: COMP-001, COMP-002 (board read, add, move), COMP-004.

- **`internal/board` (COMP-001, ADR-0004, ADR-0006).** The Board aggregate with the
  three fixed columns `todo`, `doing`, `done` (INV-005) and its own `sync.Mutex`; every
  exported method locks for its whole body and `Snapshot` returns a copy (ADR-0006/R1).
  Methods: `Snapshot`, `Add` (title trimmed and not blank, INV-001; lands at the end of
  `todo`), `Move` (one column left or right, lands at the end of the target, INV-003;
  `doing` holds at most 3, INV-002; refusal precedence per ADR-0007), and the
  aggregate-level `Edit` and `Delete` that the dense-order property test drives
  (INV-004; the HTTP edit/delete endpoints and the 80-character title rule belong to
  CARD-002, Increment 2). No `net/http` import (ADR-0004/R1).
- **`internal/httpapi` (COMP-002).** Standard `net/http` ServeMux (ADR-0002) with
  `GET /api/board`, `POST /api/cards` (201 with the created card, ADR-0009) and
  `POST /api/cards/{id}/move` with body `{"to": "<column>"}` (returns the moved card,
  ADR-0009). One JSON error helper emitting `{"code": ...}` with the four CON-004 codes;
  a missing/non-string `to` or a malformed body is `validation_failed`, a same-column or
  unknown column is `invalid_move` (ADR-0007, ASM-0003). Handlers change cards only
  through the exported Board API (ADR-0004/R2).
- **`cmd/board` (COMP-004).** package main: reads `PORT` (default 8080, CON-006), wires
  the board and the API and starts the server (ADR-0003).
- **Tests.** The 20-simultaneous-moves concurrency test (AC-018, AC-029, EC-001 per
  ADR-0005: polls snapshots while the moves run and checks the settled board) and the
  `pgregory.net/rapid` state-machine property test over add, move, edit and delete with a
  dense-order check after every step (EC-002, NFR-002, ADR-0011). rapid is a test-only
  dependency in `go.mod`. Everything passes under `go test -race ./...`.

Refusal codes of the move/add API (the mechanism this card owns, ADR-0007 precedence):

| state | output | channel | test |
| --- | --- | --- | --- |
| success (add) | 201 + created card JSON | HTTP response of `POST /api/cards` | TestAddCard_LandsAtEndOfTodo |
| success (move) | moved card JSON | HTTP response of `POST /api/cards/{id}/move` | TestMoveCard_RightLandsAtEnd |
| not_found | `{"code":"not_found"}` | HTTP error body | TestMoveCard_UnknownIdNotFound |
| validation_failed | `{"code":"validation_failed"}` (blank title; missing/non-string `to`; malformed body) | HTTP error body | TestAddCard_RejectsBlankTitle, PropertyTest_MoveRefusal_FirstFailingCheckWins |
| invalid_move | `{"code":"invalid_move"}` (non-adjacent, same-column, unknown column) | HTTP error body | TestMoveCard_TodoToDoneRefused |
| wip_limit_exceeded | `{"code":"wip_limit_exceeded"}` (adjacent move into a full doing) | HTTP error body | TestMoveCard_IntoFullDoingRefused |

**Checkpoint (wave 1):** `go test -race ./...` passes, including the test that sends 20
simultaneous moves into `doing` and never sees more than 3 cards, and the property test
that keeps positions dense after any random sequence of add, move, edit and delete;
`go run ./cmd/board` answers `GET /api/board`, `POST /api/cards` and
`POST /api/cards/{id}/move` on `PORT` (default 8080) with the stable JSON error codes.

## Acceptance criteria

- AC-001 (FR-001): given todo holds cards A (pos 0); doing holds B (pos 0) and C (pos 1); done is empty; when GET /api/board is called; then the response lists columns in order todo, doing, done with todo=[A], doing=[B, C], done=[] — test: TestGetBoard_GroupsCardsByColumnInOrder
- AC-002 (FR-001): given a board with no cards; when GET /api/board is called; then the response lists three columns todo, doing, done, each with 0 cards — test: TestGetBoard_EmptyBoardHasThreeEmptyColumns
- AC-003 (FR-002): given todo holds 2 cards; when POST /api/cards with title 'Write docs' and description 'draft'; then returns 201 and the new card is in todo at position 2 — test: TestAddCard_LandsAtEndOfTodo
- AC-004 (FR-002): given an empty board; when POST /api/cards with title 'Write docs' and no description; then returns 201 and the card has an empty description — test: TestAddCard_DescriptionIsOptional
- AC-005 (FR-002): given todo holds 2 cards; when POST /api/cards with title '   '; then returns code validation_failed and todo still holds 2 cards — test: TestAddCard_RejectsBlankTitle
- AC-011 (FR-005): given card A in todo, doing holds 1 card; when POST /api/cards/A/move with to 'doing'; then A is in doing at position 1 and no longer in todo — test: TestMoveCard_RightLandsAtEnd
- AC-012 (FR-005): given card A in doing, todo holds 2 cards; when POST /api/cards/A/move with to 'todo'; then A is in todo at position 2 — test: TestMoveCard_LeftLandsAtEnd
- AC-013 (FR-005): given no card with id 'missing'; when POST /api/cards/missing/move with to 'doing'; then returns code not_found — test: TestMoveCard_UnknownIdNotFound
- AC-014 (FR-006): given card A in todo; when POST /api/cards/A/move with to 'done'; then returns code invalid_move and A stays in todo — test: TestMoveCard_TodoToDoneRefused
- AC-015 (FR-006): given card A in done; when POST /api/cards/A/move with to 'todo'; then returns code invalid_move and A stays in done — test: TestMoveCard_DoneToTodoRefused
- AC-016 (FR-007): given doing holds 2 cards and card A is in todo; when POST /api/cards/A/move with to 'doing'; then the move succeeds and doing holds 3 cards — test: TestMoveCard_IntoDoingAtTwoSucceeds
- AC-017 (FR-007): given doing holds 3 cards and card A is in todo; when POST /api/cards/A/move with to 'doing'; then returns code wip_limit_exceeded, doing holds 3 cards and A stays in todo — test: TestMoveCard_IntoFullDoingRefused
- AC-018 (FR-007): given doing is empty and todo holds 20 cards; when 20 moves into doing are sent simultaneously; then exactly 3 moves succeed, 17 return wip_limit_exceeded and doing holds 3 cards — test: TestMoveCard_20ConcurrentMovesKeepDoingAtMost3
- AC-033 (FR-007): given doing holds 3 cards and card A is in done; when POST /api/cards/A/move with to 'doing'; then returns code wip_limit_exceeded, doing holds 3 cards and A stays in done — test: TestMoveCard_LeftIntoFullDoingRefused
- AC-022 (FR-009): given todo holds cards A, B, C at positions 0, 1, 2; when A is deleted; then B is at position 0 and C at position 1 — test: TestDenseOrder_DeleteFirstShiftsUp
- AC-023 (FR-009): given todo holds 3 cards and doing holds 1 card; when the first todo card is moved to doing; then todo positions are 0,1 and doing positions are 0,1 — test: TestDenseOrder_MoveKeepsBothColumnsDense
- AC-029 (NFR-001): given doing is empty and todo holds 20 cards; when 20 moves into doing run simultaneously; then max_doing_cards_under_concurrent_moves <= 3 — test: TestConcurrency_20SimultaneousMovesIntoDoing
- AC-030 (NFR-002): given a randomly generated sequence of add, move, edit and delete operations; when the sequence is applied to an empty board; then dense_order_violations == 0 after every operation — test: PropertyTest_DenseOrder_RandomOperations

## Engineering constraints

- EC-001 (FR-007): For any number of simultaneous moves into doing, every snapshot read of the board, and the board after the moves settle, shows at most 3 cards in doing — test: PropertyTest_Doing_NeverExceedsWIPUnderConcurrency
- EC-002 (FR-009): After any sequence of add, move, edit and delete operations, every column's positions are exactly 0..n-1 — test: PropertyTest_DenseOrder_AnyOperationSequence
- EC(ADR-0006/R1): every exported Board method of this card (Snapshot, Add, Move, Edit, Delete) holds the Board's mutex for its whole body, and a Snapshot is a copy that later operations and caller mutation never change — test: PropertyTest_Board_ConcurrentOpsSnapshotIsCopy
- EC(ADR-0007/R1): for any move request, the refusal reports the first failing check in the order not_found, validation_failed, invalid_move, wip_limit_exceeded; same-column and unknown column names are invalid_move — test: PropertyTest_MoveRefusal_FirstFailingCheckWins

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
- blocking_decisions_resolved — gap: blocking decisions: open and blocking: DEC-002 — the candidate waits on an unresolved DEC.
- dependencies_resolved — gap_free
- design_seed_for_user_facing — gap_free_by_non_applicability: not applicable — the card's own `UI surface:` machine stamp reads `none` (step 6b, FR-061); there is no surface to seed.
- open_subject_answered — gap_free

## Architecture context

- **FR:** FR-001, FR-002, FR-005, FR-006, FR-007, FR-009
- **NFR:** NFR-001, NFR-002
- **CON:** CON-001, CON-002, CON-004, CON-006, CON-007, CON-008
- **ADR:** ADR-0001, ADR-0002, ADR-0003, ADR-0004, ADR-0005, ADR-0006, ADR-0007, ADR-0009, ADR-0011
- **Components:** COMP-001, COMP-002, COMP-004
- **Trace:** meta/architecture/trace.yml

## Worktree notes

—
