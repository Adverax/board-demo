# ADR-0012: Performance: no latency or throughput target

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** —

## Context

The brief gives no latency or throughput target for the board API. The service is a single local process with in-memory storage (CON-002, CON-006) at demo scale, so a number would be speculative. The only performance-related guarantee in the requirements is the concurrency test of 20 simultaneous moves (NFR-001). The category needed an explicit answer so it is not left silently open.

## Decision

We will record performance as not applicable for the demo: no latency or throughput target is set. The only performance-related guarantee remains the 20-simultaneous-moves concurrency test (NFR-001). The operator supplied no threshold, and inventing one would put an unfounded number into the requirements.

## Alternatives considered

### set_target
The operator supplies a p95 threshold and measurement condition, giving an executable gate. Rejected: the threshold must come from the operator and none exists; a made-up figure would be a fake requirement.

## Consequences

### Positive
- No speculative numbers in the requirements
- Scope stays within the frozen brief

### Negative
- No regression guard: a change that makes the service much slower passes every test
- The demo cannot make any quantitative performance claim

### Neutral
- NFR-001 still guards correctness under concurrency, not speed
- Revisit if the board leaves demo scale or gains real users

## References

- DEC-002 (resolved by this ADR)
- CTX-001
- CON-002, NFR-001

## History

- 2026-10-24: Created — the operator confirmed no performance target exists for the demo scale.
