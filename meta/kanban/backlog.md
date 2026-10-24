# Backlog

## Deferred from Board brainstorm (2026-10-24)
- [ ] Archive: cards that stay in done for more than N days are archived and hidden from the board   @feature
- [ ] Drag and drop of cards between columns on the page   @feature

## Deferred during implementation
- [x] Fix the golangci-lint findings (6 in total once the default cap of 3 per linter was lifted: 5 errcheck and 1 staticcheck QF1001; done after wave 3, `make lint` and `make ci` are green). Originally recorded as the 3 errcheck findings (unchecked `resp.Body.Close` in cmd/board/main_test.go:51 and :65, internal/httpapi/api_test.go:353), then confirm `make lint` and `make ci` are green. Deliberately postponed until CARD-003 is merged: both files are inside CARD-003's Touches.   @tech-debt

## Follow-ups from the wave 1-3 reviews
- [ ] The page loses an unsaved edit draft when another action re-renders (web/static/app.js:244, reproduced in Chrome): keep the draft live and add an e2e assertion.   @tech-debt
- [ ] serve(): Shutdown waits its full 5 s on a dialed-but-unused connection and the binary exits 1 with "context deadline exceeded" (cmd/board/main.go:72, CARD-001 code; the e2e tests work around it with a no-keep-alive client).   @tech-debt
- [ ] PATCH check order is not fixed by any decision: `not_found` vs an unreadable body differs from move (ADR-0007). Extend ADR-0007 to edit, or align the handler.   @tech-debt
- [ ] Unknown routes and wrong methods answer `text/plain` from the standard ServeMux while CON-004 defines no code for them: needs a product reading (ADR addendum) or a JSON 404/405 wrapper.   @tech-debt
- [ ] requirements.yml names property tests `PropertyTest_*` while the Go functions are `TestPropertyTest_*`: rename the refs through a forge:architect delta (the AC/EC test refs are dead links for the validator).   @tech-debt
- [ ] domain/screens.yml: the `edit_mode` recipe `click #edit-card` is unreachable (the id is on the edit form); the component `COMP-003` is a hand edit that a fresh `derive_screens.py derive` reverts; planned e2e names are Playwright style while the tests are Go `TestE2E_*`. Fix through forge:architect.   @tech-debt
- [ ] CARD-003's failure matrix wording drifts from web/static/app.js (notice texts, one missing row: timeout after the request was sent); the "submitting" and "being deleted" states of components.md have no visual (`aria-busy` / token-based disabled style).   @tech-debt
- [ ] The slow-response e2e test spends ~6.5 s of real time (40% of the e2e suite) because REQUEST_TIMEOUT_MS is a fixed constant.   @tech-debt
- [ ] Forge process notes for the retrospective: the dependency gate runs `make audit` in the worktree (a card cut before the harness cannot pass it without a workaround); `wave_branch.py commit-bookkeeping` does not run the CI gate; `readiness_score.py` blocks on open DECs with `blocks: [none]`; golangci-lint hides issues beyond 3 per linter by default.   @tech-debt

