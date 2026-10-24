# ADR-0004: Go package layout of the binary

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** —

## Context

Following ADR-0001 (single-process monolith, seams enforced by package structure) and
ADR-0002 (Go standard library stack), the single binary must be laid out in Go packages.
Three concerns share it: the board rules of AGG-001 (INV-001..INV-005), the HTTP/JSON
layer with the CON-003 routes and CON-004 error codes, and the embedded page of CON-005.
CON-008 requires every rule to be covered by a named test, and NFR-002 requires a
property test over random operation sequences, so where the rules live decides how
directly they can be tested. CON-006 requires one binary. The forces are a small,
readable layout for a live demo against a boundary that stops HTTP code from changing
cards without going through the rules.

## Decision

We adopt layered internal packages:

- `cmd/board`: package main; reads PORT (default 8080, CON-006), wires the packages and
  starts the server.
- `internal/board`: the Board aggregate and its rules (INV-001..INV-005); it must not
  import net/http or any HTTP concern.
- `internal/httpapi`: handlers, request decoding, the CON-004 JSON error helper, and the
  ServeMux routing of CON-003; it reaches cards only through the exported Board API.
- `web/`: the static HTML, CSS and JavaScript plus the Go file that holds the
  `//go:embed` declaration and exports the embedded file system.

This lets the rules be unit- and property-tested (CON-008, NFR-002) without HTTP, maps
the C4 component view one-to-one onto packages, and makes it structurally impossible for
a handler to bypass the aggregate's invariants, since the card state is unexported in
`internal/board`.

## Alternatives considered

### flat_main_package
Everything in package main at the module root (main.go, board.go, handlers.go, with the
page embedded alongside). It is the smallest layout and the fastest to read in a demo.
It is rejected because domain and HTTP share one namespace, so nothing stops a handler
from mutating cards directly and bypassing INV-002..INV-004, and the property tests would
live in package main next to startup code.

### domain_plus_main
`internal/board` holds the aggregate; package main holds handlers, routing, the embed
declaration and startup. It isolates the rules with only one extra package. It is
rejected because HTTP and startup concerns are mixed in main, and handler tests would
need main package internals instead of testing an importable `httpapi` package.

## Consequences

### Positive
- The rules in `internal/board` are tested directly, including the NFR-002 property test
  and the EC-001 concurrency test, with no HTTP server.
- Handlers in `internal/httpapi` can be tested with net/http/httptest against a real
  Board, independently of process startup.
- The C4 component view (synthesis) maps one-to-one onto the four packages.
- `internal/` prevents any other module from importing the packages.

### Negative
- Four packages for a small service: more files and import paths to show in a demo than
  a flat layout.
- `go:embed` cannot reference files in parent directories (no `..` in patterns), so the
  embed declaration must live in a Go file inside `web/` (or a directory above it);
  `cmd/board` and `internal/httpapi` cannot embed the page themselves and must receive
  the `fs.FS` from package `web`.
- The "no net/http in internal/board" boundary is a convention until a test or lint
  checks the import list; it can erode silently.

### Neutral
- `web` becomes an importable Go package that contains static assets as well as one Go
  file.
- Property tests (DEC-015) and end-to-end tests (DEC-014) attach to `internal/board` and
  to the binary respectively; their tooling stays open.
- The mutex placement (DEC-012) depends on this layout.

## References

- DEC-011 (resolved by this ADR)
- CTX-001 (Board), AGG-001, INV-001..INV-005
- CON-005, CON-006, CON-008, NFR-002
- ADR-0001 (style), ADR-0002 (stack), ADR-0006 (mutex placement)

## History

- 2026-10-24: Created — layered internal packages (cmd/board, internal/board,
  internal/httpapi, web) chosen so the rules are isolated and testable without HTTP.

## Rules
```yaml
- id: ADR-0004/R1
  statement: Package internal/board does not import net/http or any other HTTP concern
  scope: {contexts: [CTX-001], code: ["internal/board/**"]}
  check: {kind: review-lens}
  severity: mandatory
- id: ADR-0004/R2
  statement: Handlers in internal/httpapi change cards only through the exported Board API
  scope: {contexts: [CTX-001], code: ["internal/httpapi/**"]}
  check: {kind: review-lens}
  severity: mandatory
```
