# cmd/board

The server entry (COMP-004): reads `PORT` (default 8080, CON-006), wires `board.New()`
and the embedded page `web.Page()` into `httpapi.NewApp`, and serves with
read/write/idle timeouts until SIGINT/SIGTERM, then drains for at most 5 s (ADR-0003).
If the page cannot be served (no `index.html`), it exits 1 before it listens.

Run: `go run ./cmd/board` (or `PORT=9000 go run ./cmd/board`), then open http://localhost:8080/.

## Tests

- `main_test.go`: the `PORT` default and graceful shutdown, in-process.
- `e2e_test.go`: the browser suite (ADR-0010), run by `make test-e2e` (`make test` skips
  `TestE2E_*`). Headless Chrome through chromedp with a throwaway profile, against
  `httpapi.NewApp` on an httptest server bound to a free port. Tests seed the board and
  play "another member" through the Board API. `TestE2E_BinaryServesPageOnDefaultPort`
  and `TestE2E_BinaryHonoursPortVariable` build the real binary and start it on 8080
  and 9090. Each fails up front if its port is already taken.
  - Chrome is looked up at `CHROME_PATH`, then the macOS app paths, then
    `google-chrome`/`chromium`/`headless-shell` on `PATH`. If none is found the browser
    tests **fail**. They never skip, so a green `make test-e2e` always means the page ran
    in a browser.
  - Every browser test fails on a console error or warning, an uncaught exception, a
    failed request, or a 4xx/5xx response the test did not drive on purpose.
  - Each browser test is bounded by 30 s, and binary startup by 10 s.
