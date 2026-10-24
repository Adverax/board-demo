# ADR-0006: Placement and scope of the board mutex

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** —

## Context

CON-002 requires in-memory storage guarded by a mutex. A move is a check-then-act
operation: it checks adjacency (INV-003) and the WIP limit (INV-002), then removes the
card from its column and appends it to the target, re-densifying positions (INV-004).
Under 20 simultaneous moves into doing (NFR-001, AC-018) the check and the act must be
atomic, or several moves can each see a non-full doing and all succeed. Following
ADR-0005 (EC-001 narrowed to settled snapshots), every snapshot read must also be
consistent, and following ADR-0004 (layered packages) the rules live in `internal/board`
and the handlers in `internal/httpapi`. NFR-002 requires a property test of dense order
over random operation sequences. The open question is where the mutex lives and what one
critical section covers.

## Decision

We will put the lock inside the aggregate: the Board type owns a sync.Mutex, and every
exported method (Snapshot, Add, Edit, Delete, Move) locks it for its whole body, so
validation, the decision and the state change of each command happen in one critical
section; Snapshot returns a copy of the columns, never a reference to internal slices.
AC-018 then holds by construction: moves are serialised, so at most 3 of them can see a
non-full doing. Because the lock cannot be skipped by any caller, the guarantee does not
depend on handler discipline. This also makes every read take the lock, so the
settled-snapshots guarantee of ADR-0005 is met and, in practice, exceeded.

## Alternatives considered

### pure_aggregate_locked_store
Board would be a plain value with pure methods returning a new state or a refusal, and a
separate store type would hold the mutex and run each command as lock, apply, unlock.
The rules would be pure and directly property-testable, with locking policy in one
place. It is rejected because it adds a type, and any caller holding a bare Board can
bypass the lock unless the type is carefully unexported, which reintroduces the risk the
lock is meant to remove.

### lock_in_http_handlers
Handlers would share a package-level mutex and lock around each request's call into an
unsynchronised Board. The domain would have no concurrency concern at all. It is rejected
because every new handler must remember the lock and a single missed one breaks EC-001,
and because the domain is then unsafe for any non-HTTP caller, including the property
and concurrency tests.

## Consequences

### Positive
- The aggregate cannot be used without the lock; there is no unsynchronised path to the
  cards from handlers or tests.
- AC-018 / NFR-001 hold by construction, and Snapshot reads are consistent (ADR-0005).
- One type to reason about in the demo: rules and their atomicity sit together.
- Snapshot copies prevent handlers from holding references into state that a later move
  changes.

### Negative
- Rule code and locking are interleaved in the same methods; the NFR-002 property test
  exercises the rules through the locked API rather than as pure functions.
- Exported methods must not call each other while holding the lock (sync.Mutex is not
  reentrant); shared logic must go into unexported, lock-free helpers, which is a
  discipline to keep.
- All operations, reads included, are serialised on one lock; fine at demo scale, but a
  throughput ceiling if the brief ever grew.
- Snapshot allocates a copy of the board on every read.

### Neutral
- The concurrency test for EC-001 must run with `go test -race`.
- The lock is process-local; this decision assumes the single process of ADR-0001.
- The C4 component view shows the Board component as the only owner of card state.

## References

- DEC-012 (resolved by this ADR)
- CTX-001 (Board), AGG-001, INV-002, INV-003, INV-004
- CON-002, NFR-001, NFR-002, EC-001, AC-018
- ADR-0004 (package layout), ADR-0005 (EC-001 guarantee), ADR-0001 (style)

## History

- 2026-10-24: Created — the Board aggregate owns the mutex and every exported method
  locks for its whole body, so each command is one critical section.

## Rules
```yaml
- id: ADR-0006/R1
  statement: Every exported Board method locks the Board's mutex for its whole body; Snapshot returns a copy, never internal slices
  scope: {contexts: [CTX-001], code: ["internal/board/**"]}
  check: {kind: review-lens}
  severity: mandatory
```
