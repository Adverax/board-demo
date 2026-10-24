# ADR-0015: Cost: no hosting or resource budget constraint

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** —

## Context

The brief sets no hosting or resource budget. The service is one binary running locally with in-memory storage (CON-002, CON-006). The category needed an explicit answer.

## Decision

We will record that no hosting or resource budget constraint applies to the demo, on the operator's confirmation. The service is a local single process, so there is no recurring cost to bound.

## Alternatives considered

### budget_set
The operator supplies a resource budget as a checkable constraint. Rejected: none was given, and adding one would widen the frozen brief.

## Consequences

### Positive
- The category is closed with a stated reason
- No sizing or cost work in the plan

### Negative
- Memory use is unbounded in principle: nothing limits the number of cards held in memory
- No basis exists for comparing alternatives on cost if the deployment changes

### Neutral
- Revisit if the service is ever hosted or run at a scale where resources are billed

## References

- DEC-005 (resolved by this ADR)
- CTX-001
- CON-002, CON-006

## History

- 2026-10-24: Created — the operator confirmed no budget constraint exists for the demo.
