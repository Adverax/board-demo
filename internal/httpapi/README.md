# internal/httpapi

The JSON API (COMP-002) on the standard `net/http` ServeMux (ADR-0002). `New(*board.Board)`
returns the handler.

| route | success | refusals |
| --- | --- | --- |
| `GET /api/board` | 200 `{"columns":[{"name","cards":[card...]}...]}` | — |
| `POST /api/cards` `{"title","description"?}` | 201 card | 400 `validation_failed` |
| `PATCH /api/cards/{id}` `{"title"?,"description"?}` | 200 card | 400 `validation_failed`, 404 `not_found` |
| `POST /api/cards/{id}/move` `{"to"}` | 200 card | 404 `not_found`, 400 `validation_failed`, 409 `invalid_move`, 409 `wip_limit_exceeded` |
| `DELETE /api/cards/{id}` | 204, no body | 404 `not_found` |

A card is `{"id","title","description","column","position"}`. Errors are `{"code": ...}`
(CON-004) from the single `writeError` helper. Bodies are capped at 1 MiB.
A title is 1..80 runes after trimming (INV-001, enforced by the Board).
PATCH is a partial update (ASM-0001): an absent or `null` field stays as it is; a body
that cannot be read as an edit (malformed, oversized, non-string field) is
`validation_failed`.
On move, a body with no readable target (malformed, oversized, missing or non-string `to`)
is `validation_failed`; any string `to` that is not an adjacent column, the empty string
included, is `invalid_move`; `not_found` wins over both (ADR-0007). Handlers change
cards only through the exported Board API (ADR-0004/R2). Contract tests: `api_test.go`
(board, add, move) and `edit_delete_test.go` (edit, delete, title limits, the route set).
