# ADR-0005: Observable guarantee of the WIP limit under concurrency

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** —

## Context

INV-002 limits the doing column to 3 cards. The brief turns this into a release
criterion, captured as NFR-001: 20 simultaneous moves into doing never leave the column
above 3 cards (AC-018). Domain extraction formalised the edge case EC-001 more strongly:
"the doing column never holds more than 3 cards at any observable instant". "Any
observable instant" is not defined by the brief: it could mean every state a reader can
obtain through the API, or every intermediate state inside a move, which no black-box
test can observe without instrumenting internals. CON-002 fixes in-memory storage
guarded by a mutex but says nothing about reads. The forces are faithfulness to the
frozen brief and a testable statement against the strongest possible guarantee.

## Decision

We adopt the settled-snapshots guarantee: EC-001 is narrowed to "every snapshot read
(GET /api/board, Board.Snapshot), and the state after the moves settle, shows at most 3
cards in doing". This is what the brief and NFR-001/AC-018 actually ask for, and it is
testable from outside the aggregate: the concurrency test fires 20 simultaneous moves,
polls snapshots while they run, and checks the settled state, with no access to internal
state. It deliberately makes no claim about a transient state inside one move, which no
caller can observe.

Independently of this decision, ADR-0006 (DEC-012, lock inside the aggregate) makes every
read, including Snapshot, take the aggregate's mutex. In practice no read can therefore
see more than 3 cards in doing even mid-move; that is a property of the chosen
implementation, not a requirement stated by EC-001.

**Follow-up (not done by this ADR):** the wording of EC-001 in
`meta/architecture/requirements.yml` still reads "at any observable instant". A
requirement's wording is changed only through the domain-extraction path (raw entry in
`meta/architecture/inputs/raw-requirements.md`, then a delta run of
forge:architect-domain-extraction with its impact-radius gate), never by hand from an
ADR. The delta must restate EC-001 as: "For any number of simultaneous moves into doing,
every snapshot read of the board, and the board after the moves settle, shows at most 3
cards in doing." The test name PropertyTest_Doing_NeverExceedsWIPUnderConcurrency and
the links to AC-018 and INV-002 stay unchanged.

## Alternatives considered

### reads_under_mutex
Keep the current EC-001 wording and add a stated requirement that every read takes the
same mutex as moves, so no read can ever see more than 3 cards. It gives a strong
guarantee and lets a property test poll reads during moves. It is rejected as a stated
requirement because it is stronger than the brief and turns an implementation choice
(where the lock sits and what reads take it) into a requirement; that choice belongs to
DEC-012 / ADR-0006, which happens to deliver it anyway.

## Consequences

### Positive
- EC-001 matches the brief and NFR-001; no requirement is invented beyond the frozen
  scope.
- The guarantee is checkable black-box, through the public Board API or GET /api/board,
  which suits CON-008's named-test rule.
- The implementation keeps freedom over how reads are synchronised, as long as every
  snapshot is consistent.

### Negative
- Until the domain-extraction delta lands, requirements.yml and this ADR disagree on
  EC-001's wording; trace and review must read this ADR as authoritative meanwhile.
- The requirement says nothing about transient internal states, so a future
  implementation that releases the lock mid-move would not violate EC-001 on paper even
  if it exposed a torn state to internal callers.
- One more follow-up step in the pipeline before synthesis can treat EC-001 as final.

### Neutral
- The concurrency test (PropertyTest_Doing_NeverExceedsWIPUnderConcurrency) polls
  snapshots during the 20 moves and checks the settled state; it must run under the Go
  race detector to be meaningful.
- ADR-0006 depends on this decision for the read guarantee it has to deliver.

## References

- DEC-006 (resolved by this ADR)
- CTX-001 (Board), AGG-001, INV-002
- EC-001, AC-018, NFR-001, CON-002, CON-008
- ADR-0006 (mutex placement)

## History

- 2026-10-24: Created — EC-001 narrowed to settled snapshots to match the brief; the
  wording change in requirements.yml is deferred to a domain-extraction delta.
