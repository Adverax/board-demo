# internal/board

The Board aggregate (COMP-001, AGG-001): three fixed columns `todo`, `doing`, `done`
and the cards on them, with the rules INV-001..INV-005.

- `Board` owns a `sync.Mutex`; every exported method (`Snapshot`, `Add`, `Move`, `Edit`,
  `Delete`) locks it for its whole body (ADR-0006). Lock-free helpers assume the lock.
- `Snapshot` returns a value copy (`Lanes` in column order); mutating it never reaches
  the board.
- Positions are slice indices, so dense order (INV-004) holds by construction.
- Refusals are the sentinel errors `ErrNotFound`, `ErrValidation`, `ErrInvalidMove`,
  `ErrWIPLimit`; `Move` checks them in the ADR-0007 order. `Move` takes the target as
  `*Column`: `nil` means the request named no readable target (`ErrValidation`); any
  non-nil value, the empty string included, is a named target, and an unknown one is
  `ErrInvalidMove`.
- No HTTP imports (ADR-0004/R1, pinned by `TestBoard_DoesNotImportHTTP`).

Tests: unit tests in `board_test.go`; rapid property and concurrency tests in
`property_test.go` (run with `go test -race`).
