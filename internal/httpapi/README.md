# internal/httpapi

The JSON API (COMP-002) on the standard `net/http` ServeMux (ADR-0002). `New(*board.Board)`
returns the handler.

| route | success | refusals |
| --- | --- | --- |
| `GET /api/board` | 200 `{"columns":[{"name","cards":[card...]}...]}` | — |
| `POST /api/cards` `{"title","description"?}` | 201 card | 400 `validation_failed` |
| `POST /api/cards/{id}/move` `{"to"}` | 200 card | 404 `not_found`, 400 `validation_failed`, 409 `invalid_move`, 409 `wip_limit_exceeded` |

A card is `{"id","title","description","column","position"}`. Errors are `{"code": ...}`
(CON-004) from the single `writeError` helper. Bodies are capped at 1 MiB.
On move, a body with no readable target (malformed, oversized, missing or non-string `to`)
is `validation_failed`; any string `to` that is not an adjacent column, the empty string
included, is `invalid_move`; `not_found` wins over both (ADR-0007). Handlers change
cards only through the exported Board API (ADR-0004/R2). Contract tests: `api_test.go`.
