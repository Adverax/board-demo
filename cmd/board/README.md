# cmd/board

The server entry (COMP-004): reads `PORT` (default 8080, CON-006), wires `board.New()`
into `httpapi.New` and serves with read/write/idle timeouts until SIGINT/SIGTERM, then
drains for at most 5 s (ADR-0003).

Run: `go run ./cmd/board` (or `PORT=9000 go run ./cmd/board`).
