# ADR-0013: Observability: no logging or metrics requirement

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** —

## Context

The brief asks for no logging or metrics (CON-006: one binary, run locally). The category needed an explicit answer. A minimal request log would be cheap, but it is a new capability that the frozen brief does not request.

## Decision

We will record observability as not applicable for the demo: no logging or metrics requirement is set. The frozen brief is not widened, and the operator asked for nothing beyond it.

## Alternatives considered

### basic_logging
The operator specifies a minimal request log, making the demo easier to debug. Rejected: it widens the frozen brief with a requirement nobody asked for.

## Consequences

### Positive
- Minimal scope, no logging code or format to maintain
- Nothing to configure or secure in the single binary

### Negative
- Refused actions and errors leave no server-side trace; a failure can only be seen in the HTTP response
- Debugging a surprising behavior during the live demo relies on tests and the client view alone

### Neutral
- Error codes from ADR-0007 still reach the client, so refusals are visible there
- Revisit if the board is ever run beyond a local demo

## References

- DEC-003 (resolved by this ADR)
- CTX-001
- CON-006, ADR-0007

## History

- 2026-10-24: Created — the operator confirmed no logging or metrics are expected.
