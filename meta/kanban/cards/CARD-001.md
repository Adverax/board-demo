# CARD-001: Moves under concurrency, end to end (walking skeleton)

**Status:** done
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
**Readiness:** 7/7
**Review score:** 9.5
**Started:** 2026-10-24T06:46:19Z
**Closed:** 2026-10-24T07:13:22Z
**Actual:** 0.1d
**Merge commit:** 9c0229ddd327d360f1d34ea1311f9e42744ca10d
**Blocked by:** —

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
- blocking_decisions_resolved — gap_free
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

## Failure matrix

Config values: `maxBodyBytes` = 1 MiB (httpapi), `ReadHeaderTimeout` = 5 s, `ReadTimeout` = 10 s,
`WriteTimeout` = 10 s, `IdleTimeout` = 60 s, `shutdownTimeout` = 5 s (cmd/board). `n` = cards on the board.

| operation / boundary | failure mode | declared behaviour | numeric bound |
| --- | --- | --- | --- |
| Board.* (Snapshot/Add/Move/Edit/Delete) | concurrent callers | serialised on the Board's one `sync.Mutex`; each critical section is pure in-memory work, no I/O, no callbacks, no JSON encoding under the lock | lock hold time O(n); wait time ≤ (queued callers) × O(n); doing ≤ 3 in every snapshot and after settle |
| Board.* | panic inside a critical section | `defer Unlock` releases the lock; the state change is applied only after every check passed (checks first, single mutation last), so a refused call changes nothing | 0 partial mutations on refusal |
| Board.* | non-cooperative callee | none exists: the aggregate calls no external code while holding the lock | — |
| HTTP request (all routes) | poison input: oversized body | `http.MaxBytesReader` stops reading; request refused `validation_failed` (400); on move the handler passes a nil target, so `not_found` wins if the id is unknown (ADR-0007) | ≤ 1 MiB read per request |
| HTTP request | poison input: malformed JSON, non-string `to`/`title`, missing field | `validation_failed` (400); on move the handler passes a nil target to `Board.Move`, so `not_found` wins if the id is unknown (ADR-0007) | board unchanged |
| POST /api/cards/{id}/move | present but unknown `to` string, including `""` | a non-nil target: `invalid_move` (409) after `not_found`; never `validation_failed` (ADR-0007/R1) | board unchanged |
| HTTP request | slow client (slowloris / slow body) | server read deadlines close the connection | header ≤ 5 s, whole request ≤ 10 s, response write ≤ 10 s, idle keep-alive ≤ 60 s |
| HTTP request | client timeout / disconnect mid-request | the handler is not cancelled (no context checks are needed: work under the lock is O(n) in memory and finishes); the mutation may be applied although the client never sees the response | ≤ 1 abandoned mutation per abandoned request |
| POST /api/cards/{id}/move | retry / duplicate delivery | not idempotent by key, but self-detecting: a duplicate of a succeeded move names the card's current column, so it is refused `invalid_move` and changes nothing | ≤ 1 effective move per (card, target) pair while the card stays put |
| POST /api/cards | retry / duplicate delivery | not idempotent: a retried add creates a second card (no idempotency key; the frozen brief has none) | 1 extra card per duplicate request |
| HTTP handler | panic | `defer Unlock` in the aggregate frees the lock; net/http's per-connection recovery logs the stack to the server error log and closes that connection; the server keeps serving. No custom recover middleware or metric (ADR-0013: no observability) | 1 connection lost per panic; 0 other requests affected |
| Process | crash mid-operation | the in-memory board is lost (CON-002, accepted); no torn state can be persisted because nothing is persisted | board state after restart = empty |
| Process | graceful shutdown (SIGINT/SIGTERM) | stop accepting, `Server.Shutdown` waits for in-flight handlers; on deadline `Server.Close` force-closes the rest | exit ≤ 5 s after the signal (+ ≤ 10 s WriteTimeout bound on a stuck writer is cut by Close) |
| Process | port already in use | `ListenAndServe` fails, the binary exits non-zero with the error | immediate |

## Worktree notes

—
Skill note: no `golang-pro` skill is available in this environment; implemented without it
(forge:engineering-standards loaded).
Summary: walking skeleton done. `internal/board` (aggregate, mutex, Snapshot copy, Add,
Move, Edit, Delete), `internal/httpapi` (GET /api/board, POST /api/cards,
POST /api/cards/{id}/move, one `writeError` helper), `cmd/board` (PORT/8080, server
timeouts, graceful shutdown). `go build ./...`, `go vet ./...`, `go test ./...` and
`go test -race ./...` are green. The binary was smoke-run on a PORT and answered all three
routes. Mutation check (in a scratch copy): releasing the lock between the WIP check and the
state change makes TestConcurrency_20SimultaneousMovesIntoDoing and
TestPropertyTest_Doing_NeverExceedsWIPUnderConcurrency fail.
- STRUCTURE: `internal/board` has one file, `board.go`: exported value types `Column`, `Card`,
  `Lane`, `Snapshot{Lanes [3]Lane}`, and an unexported `card` stored in `lanes [3][]card`. The
  position is the slice index, so dense order (INV-004) holds by construction, and `Card.Position`
  is computed when a card is copied out. Why: no position bookkeeping to get wrong; the property
  tests still check positions through `Snapshot`.
- STRUCTURE: refusals are package-level sentinel errors (`ErrNotFound`, `ErrValidation`,
  `ErrInvalidMove`, `ErrWIPLimit`), one per CON-004 code. httpapi maps them in a single
  `writeError` switch (404/400/409/409). Why: the aggregate stays free of HTTP (ADR-0004/R1),
  and CARD-002 reuses the same mapping for PATCH/DELETE.
- STRUCTURE: the ADR-0007 precedence is enforced only in `Board.Move`. A move body that cannot be
  read (malformed, missing, null or non-string `to`, or an oversized body) is passed in as
  `board.NoColumn` (""), which Move refuses with ErrValidation after the card lookup. Why: not_found
  beats validation_failed within one critical section, with no Exists() pre-check and its TOCTOU.
  Consequence: `{"to":""}` is `validation_failed`, not `invalid_move`, because an empty string
  names no column.
- STRUCTURE: `Edit(id string, title, description *string)`: a nil pointer leaves the field
  unchanged (the partial-update shape ASM-0001 needs for CARD-002's PATCH). For now the title rule
  is only "trimmed, not blank". CARD-002 adds the 80-rune limit inside `validTitle`, which both Add
  and Edit share.
- STRUCTURE: card ids are decimal strings from a per-board counter ("1", "2", ...). Why: simple,
  unique within the process, and readable in a demo. There is no persistence (CON-002).
- STRUCTURE: httpapi takes the concrete `*board.Board`, with no port interface. Why: it is the same
  bounded context and the only implementation (engineering-standards §9). `httpapi.New(b)` returns
  an `http.Handler` (a ServeMux). CARD-002 appends its PATCH/DELETE routes there, and the page card
  can add `GET /` the same way.
- STRUCTURE: JSON shapes. A card is `{"id","title","description","column","position"}`, the board is
  `{"columns":[{"name","cards":[...]}]}` with three columns always present and `cards` never null,
  and an error is `{"code":...}` only. Bodies are decoded with `io.ReadAll` (capped at 1 MiB) plus
  `json.Unmarshal`, so trailing garbage is `validation_failed`. Unknown fields are ignored.
- STRUCTURE: test naming. Go requires a `Test` prefix, so a traced name such as
  `PropertyTest_DenseOrder_AnyOperationSequence` becomes the function
  `TestPropertyTest_DenseOrder_AnyOperationSequence`. Run one with `go test ./internal/board -run PropertyTest_DenseOrder`.
- STRUCTURE: `cmd/board` splits into `addr(getenv)` and `serve(ctx, ln, h)` so the PORT default
  and graceful shutdown can be tested in-process without a fixed port.
- Tests beyond the named ones: TestBoard_DoesNotImportHTTP (pins ADR-0004/R1, which the ADR says
  can erode silently), TestAddCard_OversizedBodyRejected, TestMove_Refusals (a precedence table at
  aggregate level), TestServe_AnswersRoutesAndShutsDownOnCancel.
- Property tests: rapid v1.2.0 is a test-only `require` in go.mod and appears only in `_test.go`
  files. They use the default 100 checks per test, and the whole suite runs in about 2 s under -race.
- Not done (out of scope): CLAUDE.md still says "no code yet". It sits outside this card's
  Touches, so updating it is left to the merge or a later card.
[Commit] untracked outside Touches: cmd/board/README.md, internal/board/README.md, internal/httpapi/README.md

- [Build gate] cycle 1: `go build ./... && go vet ./... && go test ./... && go test -race ./...` pass on 51fc75b (full suite; Makefile not yet in the worktree base).
- [Scope] cycle 1: in_scope (10 files, none outside Touches, no guardrail hit).
- [Review 1/3] score 8.5 — critical 0, important 1, minor 3, out-of-scope 2 (reviewer: forge 2026.10.10, model opus). Report: meta/review/20261024T065900Z-CARD-001-cycle1.yml
- [Severity gate] cycle 1: BLOCKING — 1 Important: api.go:101 empty-string move target returns validation_failed instead of invalid_move (violates ADR-0007/R1; NoColumn sentinel "" collides with a present empty "to"). Note: the YAML report records it as severity `medium`, the text verdict as Important; the gate follows the text sections.
- [Review sync] 1 report(s) → meta/review/
- [Fix 1] FIXED F-001 (tests: TestMoveCard_EmptyTargetIsInvalidMove, TestPropertyTest_MoveRefusal_FirstFailingCheckWins, TestMove_Refusals), FIXED F-002 (TestSnapshot_IsACopy), FIXED F-004 (TestMoveCard_OversizedBodyRejected), SKIPPED F-003 (CON-004 ambiguity for unknown routes/wrong methods: needs a product reading, not a code defect). DECLARATIONS F-001: board.Move/NoColumn docs, httpapi.moveCard doc, READMEs, failure-matrix rows re-derived. Discriminator: presence (`*Column` nil = no readable target), not value.
- [Fix pre-gate] cycle 1: the five named tests re-run independently with -race: PASS; full build/vet/test pass; scope in_scope.
- [Review sync] 1 report(s) → meta/review/
- [Review 2/3] score 9.5 (confirmation mode) — critical 0, important 0, minor 2, out-of-scope 3. VERIFY: F-001 ✓ resolved, F-002 ✓ resolved, F-004 ✓ resolved, F-003 ✗ still present (deliberately skipped: CON-004 gives no code for unknown routes/wrong methods; needs a product reading). Report: meta/review/20261024T070532Z-CARD-001-cycle2.yml
- [Severity gate] cycle 2: PASS — 0 critical / 0 important.
- [Correction] the STRUCTURE bullet above that says an unreadable move body is passed as `board.NoColumn` and that `{"to":""}` is `validation_failed` is SUPERSEDED by Fix 1: the sentinel is removed, `Board.Move` takes `*Column` (nil = no readable target -> validation_failed after the not_found lookup); any present string, including "", reaches the adjacency check -> invalid_move (ADR-0007/R1).
- [Commit] 18ef343 fix(board): classify empty move target per ADR-0007 and add tests and READMEs (on top of 51fc75b). Made by a plain `git commit` in a commit agent, not through the forge:commit skill (deviation disclosed by the agent); no attribution lines, no meta/ paths in the branch diff.
- [AC/EC check] All criteria/constraints ✓ (head-sha: 18ef34328313ac6ac48fffa3de4ce4cff388f8b9) (evidence):
  21 items, 21 demonstrated: AC-001..005, AC-011..018, AC-033, AC-022, AC-023, AC-029, AC-030, EC-001, EC-002, EC(ADR-0006/R1), EC(ADR-0007/R1). Each named test re-run fresh with -race; PropertyTest_* names implemented with the Go `Test` prefix.
- [Review sync] 2 report(s) → meta/review/
- [Merge gate] dependency 2026-10-24T07:10:41Z PASS (deps:13523bdbf26c32764f0013736c73b1426f091d599cbb36333cf7fda6c5729f0b, mode=interactive, scan=scanned) — 0 blocking, 0 allowlisted, 0 non-blocking finding(s)
- [Merge gate] PASS (tree-sha 16c6367faa2c8056b7220f642f683f99b1e048e7, scope=full, 1s, cmd="go build ./... && go test ./...")

