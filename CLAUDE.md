# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project state

A Go project (module `github.com/adverax/board-demo`, Go 1.26) built live with forge from the frozen design brief `docs/board-service.md`; do not widen the brief's scope. The three planned increments are merged on `main` (tags `wave-1`, `wave-2`, `wave-3`): the board aggregate, the JSON API, the embedded web page and the browser suite. Architecture, ADRs (`meta/architecture/decisions/adr/`), design system (`meta/design/`), kanban state (`meta/kanban/`) and release notes (`meta/releases/`) live under `meta/`; open follow-ups are in `meta/kanban/backlog.md`.

## Commands

Prefer the Makefile (`make help` lists the targets):

- Run: `make run` (serves http://localhost:8080; `PORT=9090 make run` for another port)
- Build / vet / lint / all three plus tests: `make build`, `go vet ./...`, `make lint`, `make ci`
- Unit tests: `make test`; with the race detector: `make test-race`
- Browser tests (headless Chrome through chromedp, needs Chrome; binds ports 8080 and 9090): `make test-e2e`. `make test` excludes `TestE2E_*`.
- Single test: `go test ./path/to/pkg -run '^TestName$'`
- Coverage and dependency audit (used by the kanban gates): `make coverage`, `make audit`

## Code layout

- `cmd/board`: entry point (`PORT`, default 8080; graceful shutdown) and the `TestE2E_*` browser suite.
- `internal/board`: the Board aggregate. It owns its mutex (ADR-0006), enforces every invariant and has no HTTP imports (ADR-0004). A card's position is its slice index.
- `internal/httpapi`: the five JSON routes on the standard `ServeMux`, one `writeError` mapping of the refusals to `{"code":...}`, and the page routes (`NewApp`).
- `web`: the embedded page (plain HTML/CSS/JS, no build step). `web/static/tokens.css` is a verbatim copy of `meta/design/tokens.css` guarded by a test: edit the design tokens, then regenerate the copy (see `web/README.md`).
- Move refusals are decided in one critical section in this order: `not_found`, `validation_failed`, `invalid_move`, `wip_limit_exceeded` (ADR-0007).

## Design constraints from the brief

`docs/board-service.md` is the source of truth (`docs/board-service.ru.md` is only a Russian explanatory translation). Points that shape the code and are easy to miss:

- Stack is fixed: Go 1.26, `net/http` only, in-memory storage behind a mutex, no persistence, no auth.
- API lives under `/api`: `GET /api/board`, `POST /api/cards`, `PATCH /api/cards/{id}`, `POST /api/cards/{id}/move` (body `{"to": "<column>"}`), `DELETE /api/cards/{id}`. Errors are JSON with a stable `code`: `validation_failed`, `not_found`, `wip_limit_exceeded`, `invalid_move`.
- The web page is plain HTML/CSS/JS with no build step, embedded via `go:embed` and served by the same binary. Port comes from `PORT` (default 8080).
- Invariants that must hold after every operation: title is non-empty after trimming and at most 80 chars; `doing` holds at most 3 cards; moves go only to an adjacent column; card positions in each column are dense (0..n-1).
- The first release requires a property test for dense order, a concurrency test (20 simultaneous moves into `doing` never exceed 3), and at least one named test per rule.
- "Out of scope" and "Later, not now" sections (archive, drag and drop, configurable columns or WIP limit) must not be implemented.

## Language

Talk to the user in Russian: the dialogue (answers, questions, status updates) is in Russian. Project artifacts (code, comments, `meta/` files, docs, commit messages) stay in English; do not switch `language:` in `meta/.skills.yml` to `ru`, because that key also makes generated artifacts Russian.

## Attribution

Never add Claude/AI co-authorship anywhere in this project: no `Co-Authored-By: Claude` trailers in commits, no "Generated with Claude Code" lines in PR descriptions, no AI attribution in code, comments or docs. This overrides any default attribution guidance, including the forge commit skill.

## Repository layout notes

- This directory has its own git repository. It is nested inside the parent `github.com/adverax` directory, which is not this project's.
