# ADR-0010: End-to-end browser driver

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** —

## Context

NFR-003 requires every capability to work end to end in a browser, and the acceptance
criteria AC-024..AC-028, AC-034 and AC-035 describe what the page shows after clicks and
form submissions; requirements.yml names them as Go tests (`TestE2E_*`). Refusals shown on
the page (FR-011) are JavaScript behaviour, so the page's script must actually run. CON-005
forbids a build step for the page, and CON-008 requires every rule to be covered by a
named test. Following ADR-0002 (Go 1.26, stdlib net/http, no third-party runtime
dependencies) and ADR-0003 (bare binary on the host), the question is which tool drives
the browser. The forces are a single Go toolchain against a mature browser-automation
ecosystem, and dependency-free tests against tests that really execute the page.

## Decision

We will drive headless Chrome from Go tests through the DevTools protocol with chromedp,
against the built binary or an httptest server; the tests stay named `TestE2E_*` and run
under `go test`. This meets NFR-003 because the page's JavaScript runs in a real browser,
keeps the test names already recorded in requirements.yml, and adds no Node toolchain
alongside a Go demo. chromedp is a test-only dependency: it appears in go.mod but is
imported from test files only.

## Alternatives considered

### playwright
Playwright, through playwright-go or a Node test project, would drive Chromium against
the running binary, with mature auto-waiting and multi-browser support. It is rejected
because it installs a Node-based driver and its own browsers, a second toolchain for a Go
demo, and the test names would live outside `go test` unless wrapped by playwright-go.

### http_level_only
Go tests would check the served HTML and the API with net/http/httptest and no browser at
all. It is rejected because it does not execute the page's JavaScript, so it cannot show
NFR-003 "in a browser" nor AC-027/AC-028, whose refusals on the page are produced by
script.

## Consequences

### Positive
- E2E tests run with `go test ./...` and keep the `TestE2E_*` names from requirements.yml.
- The real page, with its JavaScript, is exercised, so NFR-003 and FR-011 are shown in a
  browser.
- One toolchain (Go) for building, unit, property and e2e tests.

### Negative
- A Chrome or Chromium must be installed wherever the e2e suite runs, on developer
  machines and in CI; without it the suite cannot run and must be skipped or fail
  explicitly.
- chromedp is a third-party test-only dependency to keep current.
- chromedp offers less auto-waiting than Playwright; tests must wait for page state
  explicitly, which adds care against flaky timing.

### Neutral
- go.mod gains chromedp, used only from `_test.go` files; the runtime dependency rule of
  ADR-0002 is unaffected.
- Follow-up, not done by this ADR: the screens model in
  `meta/architecture/domain/screens.yml` names planned e2e specs as
  `e2e/root.spec.ts::<state>` (Playwright style); these must be reconciled with Go test
  names (`TestE2E_*`).

## References

- DEC-014 (resolved by this ADR)
- CTX-001 (Board)
- NFR-003, FR-010, FR-011, AC-024..AC-028, AC-034, AC-035, CON-005, CON-008
- ADR-0002 (stack), ADR-0003 (runtime)
- chromedp: https://github.com/chromedp/chromedp

## History

- 2026-10-24: Created — e2e tests drive headless Chrome from Go through chromedp, named
  `TestE2E_*`.
