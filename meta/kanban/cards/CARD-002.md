# CARD-002: Card editing and the full API contract

**Status:** done
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
**Readiness:** 7/7
**Review score:** 9.0
**Started:** 2026-10-24T07:21:48Z
**Closed:** 2026-10-24T07:31:56Z
**Actual:** 0.0d
**Merge commit:** f5e76a0c60ce8a39533cf361087d0f3f8cae09a5
**Blocked by:** —

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
- dependencies_resolved — gap_free
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
- Skill `golang-pro` is not available in this environment; implemented without it (forge:engineering-standards applied).
- `internal/board`: `Edit`/`Delete` already existed from CARD-001 with the ADR-0006 lock discipline; this card added `MaxTitleLength = 80` and the rune-count check (after trimming, ASM-0002) inside the shared `validTitle`, so Add and Edit enforce the full INV-001. Doc comments updated.
- `internal/httpapi`: `PATCH /api/cards/{id}` (pointer fields: absent/null leaves a field unchanged, ASM-0001; 200 with the card) and `DELETE /api/cards/{id}` (204, no body) added to the ServeMux; refusals go through the existing `writeError`. All five CON-003 routes now served. Decision: an unreadable PATCH body (malformed, oversized, non-string field) is `validation_failed` before the id lookup, like POST /api/cards; the ADR-0007 precedence is specified only for move and is unchanged.
- Tests: every AC test name in `internal/httpapi/edit_delete_test.go`, plus partial-update, malformed-body, rune-counting, route-set (405/404) tests and `TestPropertyTest_EditDelete_ContractHolds`. Invariants enumerated: INV-001 -> new `TestPropertyTest_Title_ValidOnAddAndEdit` and the model test (titles now include 80/81-rune values); INV-002/INV-003/INV-005 unchanged; INV-004 -> model and random-operation property tests plus `TestDelete_MiddleClosesGap`; EC ADR-0006/R1 -> `TestPropertyTest_Board_ConcurrentOpsSnapshotIsCopy` whose edit action now draws optional (valid and invalid) titles and descriptions and which checks that settled titles are INV-001-valid. A mutation run (limit off by one) failed four tests, so the properties bite.
- G-1: the existing tests were left untouched and pass (`make test`, `make test-race`, `go vet ./...` green).
- `make lint`: golangci-lint installed here is built for an older Go and cannot load Go 1.26 export data (typecheck error on stdlib imports) — environmental, not a code finding; `go vet` is clean.
- SCOPE+ internal/board/README.md, internal/httpapi/README.md — keeps the package doc true (title rule, PATCH/DELETE routes).
- [Build gate] cycle 1: `make test`, `make test-race`, `go vet ./...` pass on bee0526 (the Makefile is in the worktree base). `make lint` could not run on this host: the installed golangci-lint was built for an older Go and cannot read the Go 1.26 standard library data (tooling, not a code finding).
- [Scope gate] ⚠ grown: +0 components · 2 files outside Touches (internal/board/README.md, internal/httpapi/README.md: per-package docs kept true by the docs step, noted SCOPE+ by the implementation agent); no guardrail hit.
- [Review 1/3] score 9.0 — critical 0, important 0, minor 2, out-of-scope 0 (reviewer: forge 2026.10.10, model opus). Report: meta/review/20261024T072709Z-CARD-002-cycle1.yml
- [Severity gate] cycle 1: PASS — 0 critical / 0 important (score 9.0 >= 8). Minor (non-gating, triaged should_gate=false): api.go:105 PATCH precedence between not_found and an unreadable body is unspecified and differs from move (candidate follow-up: extend ADR-0007 to edit, or align the handler); edit_delete_test.go:71 AC-006 test does not set the given title 'Old'.
- [Review sync] 1 report(s) → meta/review/
- [Commit] bee0526 feat(httpapi): add card editing and complete the API contract (single commit by the implementation agent via forge:commit; no attribution lines; no meta/ paths in the branch diff).
- [AC/EC check] All criteria/constraints ✓ (head-sha: bee052601b4c708b074e2a19834eb634728f9555) (evidence):
  12 items, 12 demonstrated: AC-006..010, AC-019..021, AC-036, AC-037, EC(ADR-0006/R1), G-1. Each named test re-run fresh with -race; PropertyTest_* names implemented with the Go `Test` prefix; G-1: board_test.go additions only, property_test.go edits only strengthen the title rule, no assertion removed.
- [Merge gate] dependency 2026-10-24T07:31:16Z PASS (deps:b089475c9a8bfc7c4c7ef7bc0aa07a550a43e0f144491b2d88221e5ac799831f, mode=interactive, scan=scanned) — 0 blocking, 0 allowlisted, 0 non-blocking finding(s)
- [Merge gate] PASS (tree-sha 7410b7d797da715b09b8df49c8b95e4cf14f2f7f, scope=full, 2s, cmd="go build ./... && go test ./...")

