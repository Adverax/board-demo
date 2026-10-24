# Board

A small kanban board (Go backend + one web page) built live with forge. Three fixed columns (`todo`, `doing`, `done`), at most 3 cards in `doing`, cards move one column at a time. The board lives in memory: it is empty after a restart. The scope is frozen in [`docs/board-service.md`](docs/board-service.md).

## Run

```
make run
```

Open <http://localhost:8080>. Another port: `PORT=9090 make run`. Without `make`: `go run ./cmd/board`.

## Develop

```
make test        # unit tests
make test-race   # the same under the race detector
make test-e2e    # browser tests in headless Chrome (needs Chrome; ports 8080 and 9090 must be free)
make ci          # lint, build, test
```

`make help` lists every target. Package overviews: [`cmd/board`](cmd/board/README.md), [`internal/board`](internal/board/README.md), [`internal/httpapi`](internal/httpapi/README.md), [`web`](web/README.md).

## API

| Route | Does |
|---|---|
| `GET /api/board` | the whole board, grouped by column |
| `POST /api/cards` | add a card (lands at the end of `todo`) |
| `PATCH /api/cards/{id}` | edit title and/or description |
| `POST /api/cards/{id}/move` | body `{"to": "<column>"}`, one column left or right |
| `DELETE /api/cards/{id}` | delete a card |

Refusals are JSON `{"code": ...}` with `validation_failed`, `not_found`, `invalid_move` or `wip_limit_exceeded`.

## Project documents

Architecture decisions: [`meta/architecture/decisions/adr/`](meta/architecture/decisions/adr/). Design system: [`meta/design/`](meta/design/). Release notes per wave: [`meta/releases/`](meta/releases/). Open follow-ups: [`meta/kanban/backlog.md`](meta/kanban/backlog.md). Licence: [MIT](LICENSE).
