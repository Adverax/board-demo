# Wave 1 — 2026-10-24   (tag: wave-1)

## Shipped
- CARD-001 (feature): the board aggregate with its own lock, the JSON API for reading the board, adding a card and moving it between columns under the WIP limit, and the server entry point; the invariants are covered by unit, property and concurrency tests.   score 9.5 (2 review cycles)   FR-001, FR-002, FR-005, FR-006, FR-007, FR-009 · NFR-001, NFR-002

## Requirements closed
- FR-001 ✓ (goal-backward), FR-002 ✓, FR-005 ✓, FR-006 ✓, FR-007 ✓, FR-009 ✓ · NFR-001 ✓ (observed live: 20 simultaneous moves into doing, exactly 3 accepted and 17 refused, three runs), NFR-002 ✓
- Checkpoint of the interval passed: `go test -race` incl. the 20-simultaneous-moves test and the dense-order property tests; the real binary answers `GET /api/board`, `POST /api/cards` and `POST /api/cards/{id}/move` with the stable error codes.

## Convergence
- ⚠ NOT converged on the user-facing side — ui-gaps: FR-001, FR-002, FR-005 are satisfied at the API level but have no UI surface yet. Deliberate backend-first order: the page is Increment 3 (follow-up: CARD-003, wave 3).
- Not verified at runtime: nothing (no performance NFR applies, ADR-0012).
- Caveat from the goal check: a random operation sequence over HTTP was not run because edit and delete have no routes until CARD-002; the dense-order property is proved at the aggregate level (property tests, `-race`).

## Known gaps / escalations
- No escalations.
- Review finding F-003 left open on purpose: unknown routes and wrong methods answer `text/plain` from the standard ServeMux, and CON-004 defines no code for them. Needs a product reading (ADR-0007 or ADR-0009 addendum, or a JSON 404/405 wrapper).
- Out of scope, noted by the reviewer: model test names `PropertyTest_*` lack the Go `Test` prefix (requirements.yml follow-up delta); the board size is unbounded (the frozen brief sets no limit).
- No deferrals captured to the backlog (the scan hit only the column name `todo`).

## Migrations
- none

## Metrics
- snapshot not captured: `forge:metrics` refused (exit 2) because the signal definitions registry `meta/metrics/signal-definitions.yml` does not exist; the project has not adopted the metrics layer. Nothing was lost: the registry has to exist before any wave can be captured.
