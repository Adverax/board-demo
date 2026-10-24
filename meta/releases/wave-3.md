# Wave 3 — 2026-10-24   (tag: wave-3)

## Shipped
- CARD-003 (feature): Added the board page (embedded in the binary, styled only from the design tokens) with the add form, edit, delete, move and refusal reasons, served at `/`, and a headless-Chrome browser suite covering every page state.   score 8.6 (1 review cycle)   FR-010, FR-011 · NFR-003

## Requirements closed
- FR-010 ✓ (goal-backward), FR-011 ✓, NFR-003 ✓
- Checkpoint of the interval passed: with the binary on `PORT=9090` (and on the default 8080), a real-browser run of the planned screen states passes: board render, add form, edit saved, card deleted, card moved, WIP-limit refusal, `invalid_move` and `not_found` reasons on a stale page, blank-title refusal. Verified twice: by the 13 `TestE2E_*` tests and by an independent drive of the built binary in headless Chrome (no console errors, both colour schemes render).

## Convergence
- ✅ converged — every target FR/NFR of this wave demonstrated (API + UI).
- Cross-wave closure: the user-facing requirements of waves 1 and 2 that had no UI surface now have one (FR-001, FR-002, FR-005, FR-003, FR-004, FR-008, see the notes appended to `wave-1.md` and `wave-2.md`).
- Not verified at runtime: nothing (no performance NFR applies, ADR-0012). No axe accessibility scan and no pixel baseline: the project has no Playwright/axe toolchain (browser driver is chromedp, ADR-0010); accessibility was checked statically and by an independent screenshot review in both themes at 1200 and 390 px.

## Known gaps / escalations
- No escalations; no fix round was needed.
- Eight non-gating review findings were left for follow-up and are in `meta/kanban/backlog.md`: the page loses an unsaved edit draft when another action re-renders (the only one a user can notice), failure-matrix wording drift and a missing timeout-after-send row, no visual for the submitting / being-deleted states, an unreachable `edit_mode` recipe in `screens.yml`, one staticcheck and two errcheck findings (fixed after the merge together with the three from wave 1; `make lint` and `make ci` are green), the page logic is tested only at the browser tier (accepted exception, no JS unit runner under CON-005), and a 6.5 s real-time slow-response test.
- Carried: `serve()` Shutdown waits its full 5 s on a dialed-but-unused connection and exits 1 (CARD-001 code, worked around in the e2e tests); unknown routes and wrong methods answer `text/plain` (CON-004 has no code for them); the PATCH check order is not fixed by any decision.
- Observations from the goal check (no requirement covers them): the refusal notice stays until the next successful action and has no dismiss control; a refused move gives no feedback on the card itself; the move buttons are icon-only (aria-label and title, no visible text).
- No deferrals captured by the scan (hits were the word `todo` and test stubs).

## Migrations
- none

## Metrics
- snapshot not captured: `forge:metrics` refuses without the signal definitions registry `meta/metrics/signal-definitions.yml` (same as waves 1 and 2).
