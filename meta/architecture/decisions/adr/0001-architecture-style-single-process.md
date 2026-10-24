# ADR-0001: Architecture style of the board service

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** —

## Context

The board service has one bounded context, CTX-001 (Board), holding one aggregate,
AGG-001, with the WIP limit, adjacency and dense-order rules. The platform baseline
needs its style axis recorded: how many deployables exist and how parts talk to each
other. The frozen brief already narrows the space: CON-006 requires one binary, and
CON-005 requires the page (plain HTML, CSS and JavaScript, no build step) to be embedded
with go:embed and served by the same process. CON-002 keeps all state in memory under a
mutex, so there is no shared store that could connect several processes. The forces are
small operational surface for a live demo against any independent deployability of the
API and the page.

## Decision

We will build the board service as a single-process monolith: one Go process holds the
board, serves the JSON API under /api and serves the embedded page, and every
interaction between its parts is a synchronous in-process function call, with no domain
events on a bus and no broker. This is the only style that satisfies CON-006 and CON-005
at the same time. With one context and one aggregate there is no seam that would earn a
second deployable, and in-memory state (CON-002) could not be shared between processes
anyway. This matches the prior pattern DP-006 (coarse start, seams enforced by package
structure); the package seams are settled separately in ADR-0004.

## Alternatives considered

### split_api_and_static_frontend
The JSON API would run as one process and the page would be served separately, from a
static host or a second process. Its only benefit is redeploying the page without the
API, which nothing in the brief asks for. It is rejected because it violates CON-006
(one binary) and CON-005 (the page is served by the same process), and because it adds
CORS configuration and a second deployable to build, start and keep in step.

## Consequences

### Positive
- One artifact to build, start and test; AC-031/AC-032 (PORT handling) and the NFR-003
  end-to-end check exercise exactly what is shipped.
- The page and the API share one origin, so no CORS handling is needed.
- No messaging infrastructure, no network hops between components, no partial-failure
  modes between deployables.
- The in-memory board under one mutex (CON-002) is reachable from every handler without
  any distribution concern.

### Negative
- The page and the API cannot be scaled, deployed or rolled back independently; any page
  change requires rebuilding and restarting the binary.
- A single process is a single point of failure, and a restart loses the board (already
  accepted by CON-002).
- Moving to more than one instance later would need a shared store and a different
  concurrency design (ADR-0006 relies on one process holding the only copy of the board).

### Neutral
- Internal seams are enforced by Go package boundaries rather than by deployables
  (ADR-0004).
- Unlocks the stack (ADR-0002) and runtime (ADR-0003) axes, which depend on this one.
- Recorded in `meta/architecture/platform.yml` as the `style` axis.

## References

- DEC-008 (resolved by this ADR)
- CTX-001 (Board), AGG-001
- CON-005, CON-006, CON-002
- Prior pattern DP-006 (coarse_start)
- ADR-0002, ADR-0003, ADR-0004

## History

- 2026-10-24: Created — single-process monolith chosen because CON-006 and CON-005 admit
  no other topology and one context earns no second deployable.
