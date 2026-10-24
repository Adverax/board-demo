# Wave 2 — 2026-10-24   (tag: wave-2)

## Shipped
- CARD-002 (feature): Added editing of cards (partial PATCH) and deletion through the API, and the 80-character title limit on add and edit; all five documented routes are now served.   score 9.0 (1 review cycle)   FR-003, FR-004, FR-008

## Requirements closed
- FR-003 ✓ (goal-backward), FR-004 ✓, FR-008 ✓
- Checkpoint of the interval passed: every API acceptance criterion passes (title rules incl. 80/81 characters on add and edit, delete closes the gap, unknown id is `not_found`) and the five endpoints of CON-003 are served by the standard ServeMux. Verified on the live binary, including 80 and 81 multibyte characters (the limit counts runes, ASM-0002) and dense positions after deleting the first, a middle and the last card of a column.

## Convergence
- ⚠ NOT converged on the user-facing side — ui-gaps: FR-003, FR-004, FR-008 are satisfied at the API level but have no UI surface yet. Deliberate backend-first order: the page is Increment 3 (follow-up: CARD-003, wave 3).
- Assumptions exercised: ASM-0001 (PATCH is a partial update) and ASM-0002 (the 80 limit counts runes after trimming) now hold in code and tests; they stay `unvalidated` in the ledger until their validation plans are closed.
- Not verified at runtime: nothing (no performance NFR applies, ADR-0012).

## Known gaps / escalations
- No escalations; no fix round was needed.
- Review minor (non-gating, triaged): the PATCH check order between `not_found` and an unreadable body is not fixed by any decision and differs from move (ADR-0007): `PATCH /api/cards/missing` with an unreadable body answers `validation_failed`, with a readable invalid title it answers `not_found`. Candidate follow-up: extend ADR-0007 to edit, or align the handler.
- Review minor: the AC-006 test does not set the given title "Old".
- Carried from wave 1: review finding F-003 (unknown routes and wrong methods answer `text/plain`, CON-004 defines no code for them) is still open; the new route set test pins 404 and 405 for them.
- Host tooling: `make lint` cannot run here (the installed golangci-lint was built for an older Go and cannot read the Go 1.26 standard library data); `go vet` and the test targets do run.
- No deferrals captured to the backlog.

## Migrations
- none

## Metrics
- snapshot not captured: `forge:metrics` refuses without the signal definitions registry `meta/metrics/signal-definitions.yml` (same as wave 1).

## Closure notes
  ↳ FR-003, FR-004, FR-008 converged in wave 3 (2026-10-24): the UI surface they lacked shipped with CARD-003.
