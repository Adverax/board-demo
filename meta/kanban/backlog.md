# Backlog

## Deferred from Board brainstorm (2026-10-24)
- [ ] Archive: cards that stay in done for more than N days are archived and hidden from the board   @feature
- [ ] Drag and drop of cards between columns on the page   @feature

## Deferred during implementation
- [ ] Fix the 3 golangci-lint errcheck findings (unchecked `resp.Body.Close` in cmd/board/main_test.go:51 and :65, internal/httpapi/api_test.go:353), then confirm `make lint` and `make ci` are green. Deliberately postponed until CARD-003 is merged: both files are inside CARD-003's Touches.   @tech-debt
