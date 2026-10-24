# ADR-0014: Compliance: no regulatory or organizational rule applies

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** —

## Context

The brief names no regulatory or organizational rule. The service stores no personal data and has no authentication or users (CON-007), so the usual drivers of compliance requirements are absent. The category needed an explicit answer.

## Decision

We will record that no regulatory or organizational rule applies to the demo, on the operator's confirmation. This holds because there is no personal data and no authentication (CON-007).

## Alternatives considered

### rule_applies
The operator names a rule to formalize as a CON or NFR. Rejected: no rule was named, and inventing one would widen the frozen brief.

## Consequences

### Positive
- The category is closed with a stated reason rather than left open
- No compliance work (retention, consent, audit) enters the backlog

### Negative
- Card text is free-form user input and is stored as given; if real data were ever entered, nothing would protect it
- The conclusion depends on the board staying personal-data-free

### Neutral
- Revisit if authentication, users or personal data are ever added

## References

- DEC-004 (resolved by this ADR)
- CTX-001
- CON-007

## History

- 2026-10-24: Created — the operator confirmed no rule applies.
