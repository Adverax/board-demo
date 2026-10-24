# ADR-0002: Language, HTTP and storage stack

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** server-rendered — the only consumer of /api is the page embedded in and served by the same binary (plain JavaScript with no build step) so nothing generates code from a spec

## Context

Following ADR-0001 (single-process monolith), the stack axis of the platform baseline
must be recorded. CON-001 fixes Go 1.26 and standard library net/http for the HTTP
layer; CON-002 fixes in-memory storage guarded by a mutex; CON-003 fixes five endpoints
(GET /api/board, POST /api/cards, PATCH /api/cards/{id}, POST /api/cards/{id}/move,
DELETE /api/cards/{id}); CON-004 fixes JSON error objects with a stable `code` field;
CON-005 requires the page to be embedded with go:embed. The open questions are the
routing mechanism and whether any third-party runtime dependency is justified.

The same decision must settle the API posture required by forge:engineering-standards §2
(OpenAPI-first). The HTTP surface is a JSON API, and its only consumer is the page that
CON-005 embeds into the same binary. That page is plain HTML, CSS and JavaScript with no
build step, is released in lockstep with the handlers, and has no toolchain that could
generate a client or types from a spec. There is no external client, no mobile app and no
sibling service (CON-007: no authentication, no other users of the API are in scope).

## Decision

We will use Go 1.26 with net/http and the standard ServeMux, using method and path
wildcard patterns (for example `POST /api/cards/{id}/move`) for all five CON-003 routes,
encoding/json for request and response bodies and for the CON-004 error object, embed
for the page, and an in-memory board guarded by sync.Mutex; the module has no
third-party runtime dependencies. This is exactly what CON-001 and CON-002 state, and
the Go 1.22+ ServeMux pattern syntax covers every route of CON-003, including method
matching and the `{id}` wildcard, so a router would add nothing.

API posture: we record the OpenAPI-first rule as **N/A by scope**. Rule applied:
forge:engineering-standards §2a, the scope test "is there a consumer that should be
generating code from your spec?". The answer is no: the only consumer is the embedded
page, which is part of the same binary and, by CON-005, has no build step in which
generated code could exist; it is not a "separate frontend" in the sense of §2a. The
header uses the `server-rendered` token, the nearest value in the fixed vocabulary of
§2b: the browser is the only consumer and the server ships the page itself. The token is
not literal (the page fetches JSON rather than HTML fragments); the gloss on the header
line says so. The CON-003/CON-004 shapes are instead pinned by named Go tests (CON-008).

## Alternatives considered

### go_nethttp_with_router
Go 1.26 with a net/http-compatible third-party router (chi class) and the same in-memory
storage. It would bring middleware chaining and route grouping out of the box. It is
rejected because it contradicts CON-001 (standard library net/http for the HTTP layer)
and adds a dependency for five routes the standard ServeMux already handles.

## Consequences

### Positive
- The runtime `require` block of go.mod stays empty: nothing to audit, update or pin.
- Routing, JSON and embedding are all standard library and stable across Go releases.
- The board lives in process memory, so reads and writes have no I/O and no failure mode
  beyond the process itself.
- No OpenAPI spec and generator toolchain to maintain for a single in-binary consumer.

### Negative
- The JSON error helper for the CON-004 `code` field, request decoding and validation
  are hand-written, and must be covered by tests rather than by a framework.
- Middleware (if ever needed, for example request logging per DEC-003) has to be
  written as plain http.Handler wrappers.
- Without a spec, the request and response shapes are documented only by code and tests;
  the exemption must be revisited (flip to `spec-required`) the day a separate consumer
  appears, such as a CLI client or a second page with a build step.
- The `server-rendered` token is an approximation of this project's posture and may
  mislead a reader who sees only the header; the gloss and this section carry the real
  reasoning.

### Neutral
- Requires a Go 1.22+ toolchain for ServeMux patterns; Go 1.26 is fixed by CON-001.
- Test-only dependencies (DEC-014 browser driver, DEC-015 property-test library) are not
  decided here; CON-001 restricts only the HTTP layer.
- Recorded in `meta/architecture/platform.yml` as the `stack` axis.

## References

- DEC-009 (resolved by this ADR)
- CTX-001 (Board)
- CON-001, CON-002, CON-003, CON-004, CON-005, CON-007, CON-008
- ADR-0001 (style), forge:engineering-standards §2a/§2b
- DEC-014, DEC-015 (test tooling, still open)

## History

- 2026-10-24: Created — Go 1.26 standard library stack with in-memory board chosen as
  stated by CON-001/CON-002; OpenAPI-first recorded as N/A by scope (single embedded
  consumer).

## Rules
```yaml
- id: ADR-0002/R1
  statement: The module has no third-party runtime dependencies; HTTP routing uses the standard net/http ServeMux
  scope: {contexts: [CTX-001]}
  check: {kind: review-lens}
  severity: mandatory
```
