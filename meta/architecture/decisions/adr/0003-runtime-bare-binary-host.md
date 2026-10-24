# ADR-0003: Runtime environment of the binary

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** —

## Context

Following ADR-0001 (single-process monolith), the runtime axis of the platform baseline
must be recorded: where the binary runs for the live demo and for the end-to-end release
check of NFR-003 ("the binary started with no PORT set, every capability exercised in a
browser"). CON-006 fixes one binary with the port taken from the PORT environment
variable, default 8080, and CON-002 keeps all state in memory, so there is nothing to
mount, persist or discover. What remains open is whether the binary runs directly on the
host or is wrapped in a container image. The forces are minimal tooling for a demo
against reproducibility on machines without a Go toolchain.

## Decision

We will run the binary directly on the host: it is built with `go build` (or started
with `go run`) and listens on PORT, or 8080 when PORT is unset; there is no container
image and no orchestrator. This matches CON-006 literally, lets AC-031/AC-032 and NFR-003
test the very process that is demonstrated, and needs nothing beyond the Go 1.26
toolchain already required by CON-001 (ADR-0002).

## Alternatives considered

### container_image
The same single binary would be packaged into a minimal container image (distroless or
scratch) and run with `docker run -e PORT=...`. It would give a reproducible run without
a local Go toolchain and would be ready for any container host later. It is rejected
because it adds a Dockerfile and a container runtime the brief never mentions, and
because port mapping introduces a second place where the listening port can disagree
with PORT, which weakens the AC-031/AC-032 checks.

## Consequences

### Positive
- Nothing to install beyond the Go toolchain; build and run are one command each.
- The PORT/8080 behaviour is observed exactly as CON-006 states it, with no port mapping
  layer in between.
- The end-to-end tests (NFR-003) can start the binary or an httptest server in-process
  without a container daemon.

### Negative
- Every machine that builds the service needs a matching Go 1.26 toolchain; there is no
  prebuilt, self-contained artifact to hand to someone without Go.
- The run is not isolated from the host: a port already in use on the host makes startup
  fail, and the environment is whatever the host provides.
- Deploying to a container platform later would need a new decision and a Dockerfile.

### Neutral
- Process supervision (restart on crash) is out of scope; a restart loses the board
  anyway (CON-002).
- forge:harness reads this axis to generate a Makefile without docker-compose.
- Recorded in `meta/architecture/platform.yml` as the `runtime` axis.

## References

- DEC-010 (resolved by this ADR)
- CTX-001 (Board)
- CON-006, CON-002, CON-001, NFR-003
- ADR-0001 (style), ADR-0002 (stack)

## History

- 2026-10-24: Created — bare binary on the host chosen because CON-006 fixes one binary
  with PORT/8080 and in-memory state needs no container concerns.
