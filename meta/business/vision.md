# Vision — Board

## Problem

A team keeps its work items scattered and has no single shared view of what is
waiting, in progress and finished.

## Goal

A small kanban board: a Go backend and a single web page in front of it. The
team keeps its work items on three fixed columns (`todo`, `doing`, `done`) and
moves them along. There is one shared board; everyone sees the same state.

## Primary actor

- Team member — opens the page, adds cards, edits them, moves them between
  columns, deletes them. There are no other roles.

## Success metrics

- (none agreed yet) — the brief defines release acceptance checks, which are
  tracked as requirements in `raw-requirements.md`, not as outcome metrics.

## Non-goals

- Authentication, users, per-user boards, multiple boards.
- Persistence across restarts.
- Custom or renamable columns, configurable WIP limit.
- Drag and drop, keyboard shortcuts, animations.
- Comments, labels, assignees, due dates.

## Context constraints (given, not open questions)

- Go 1.26, standard library `net/http` for the HTTP layer.
- In-memory storage guarded by a mutex; the board is lost on restart.
- JSON over HTTP under `/api`; errors carry a stable `code` field.
- The page is plain HTML, CSS and JavaScript with no build step, embedded into
  the binary with `go:embed` and served by the same process.
- One binary, port from the `PORT` environment variable (default 8080).
- No authentication.

Source of truth: `docs/board-service.md` (frozen brief).
