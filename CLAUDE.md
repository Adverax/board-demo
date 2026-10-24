# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project state

A fresh Go project (module `github.com/adverax/board-demo`, Go 1.26) used for a live demonstration of forge. The only input is the design brief in `docs/board-service.md`; it is frozen, so do not widen its scope. There is no code yet. Update this file once real structure exists.

## Commands

- Build: `go build ./...`
- Test all: `go test ./...`
- Single test: `go test ./path/to/pkg -run '^TestName$'`
- Lint/vet: `go vet ./...`
- Race detector (relevant for the concurrency test): `go test -race ./...`

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
