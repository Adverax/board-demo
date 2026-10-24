# ADR-0007: Move error codes and check precedence

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** —

## Context

CON-004 fixes the error object of every refusal to a JSON object with a stable `code`
field and lists exactly four codes: `validation_failed`, `not_found`,
`wip_limit_exceeded` and `invalid_move`. CON-003 fixes the move request as
`POST /api/cards/{id}/move` with body `{"to": "<column>"}`. The brief names the codes
but not which one a move returns when the target is the card's own column, an unknown
column name, a missing or non-string `to`, or a body that is not JSON at all (FR-005,
FR-006, FR-007). Several checks can also apply to one request at once, for example a
non-adjacent move whose target is a full doing, and the brief does not say which reason
wins. The answer shapes the move handler in `internal/httpapi` and the Move method of the
Board aggregate (ADR-0004, ADR-0006), and every rule here must be covered by a named test
(CON-008). The forces are reuse of the four listed codes versus a sharper split between
bad input and a bad move, and a literal reading of "a move into a full doing is refused"
versus reporting the most fundamental reason first.

## Decision

We will return `invalid_move` for a same-column move and for an unknown column name, and
`validation_failed` for a missing or non-string `to` or a malformed body; when several
checks apply, the first one in this order wins: `not_found`, then `validation_failed`,
then `invalid_move`, then `wip_limit_exceeded`. This keeps to the four codes of CON-004
and gives each a single meaning: `validation_failed` means the request could not be read
as a move at all, `invalid_move` means it was read but names a target that is not an
adjacent column, and `wip_limit_exceeded` is reserved for an otherwise valid adjacent
move into a full doing (for example done to doing). Ordering the checks from "does the
card exist" to "is there room" means a refusal always reports the most fundamental
reason, and the order is a fixed sequence that tests can pin.

## Alternatives considered

### split_codes
A same-column move would be `invalid_move`, while an unknown column name, a missing or
non-string `to` and a malformed body would all be `validation_failed`, with the same
precedence. It distinguishes malformed input from a well-formed but forbidden move. It is
rejected because it creates two rules for what a client sees as one question ("is this a
column I may move to?"): an unknown name and a non-neighbouring name are both targets
the card cannot reach, and splitting them adds a rule to document and test for no
requirement in the brief.

### wip_first
A target of doing that is full would return `wip_limit_exceeded` before adjacency is
checked. It matches the most literal reading of "a move into a full doing is refused".
It is rejected because a move that is invalid for another reason but targets a full
doing, such as a same-column doing to doing move, would be reported with the wrong
reason, and it needs a special case for that same-column move; the user would be told
"doing is full" when the move was never allowed.

## Consequences

### Positive
- Only the four codes of CON-004 are used; no new code is invented.
- Each code has one meaning, so the page (FR-011) can map a code to one human reason.
- The precedence order is deterministic and directly testable, which serves CON-008.
- `wip_limit_exceeded` is never reported for a move that would have been refused anyway.

### Negative
- An unknown column name is arguably malformed input, yet it is reported as
  `invalid_move`; a client that wants to tell typos from forbidden moves cannot do so
  from the code alone.
- The precedence order is a contract that must be enforced in two places (handler for
  decoding, aggregate for the rest) and pinned by tests for every pair of overlapping
  checks, which adds test cases beyond the brief's examples.
- `not_found` before `validation_failed` means the card is looked up even when the body
  is malformed, so the handler must decode lazily or the aggregate must check existence
  first.

### Neutral
- ASM-0003 (a malformed move body returns `validation_failed`) is consistent with this
  decision; it stays an assumption until tests validate it, and assumptions.md is not
  changed by this ADR.
- The edge controls of the page (DEC-007) never produce an unknown or same-column target
  in normal use; this ADR still defines what the API returns if a client sends one.

## References

- DEC-001 (resolved by this ADR)
- CTX-001 (Board), AGG-001, INV-002, INV-003
- FR-005, FR-006, FR-007, FR-011, CON-003, CON-004, CON-008
- ASM-0003
- ADR-0004 (package layout), ADR-0006 (mutex inside the aggregate)

## History

- 2026-10-24: Created — same-column and unknown column are `invalid_move`, malformed
  input is `validation_failed`, precedence not_found, validation_failed, invalid_move,
  wip_limit_exceeded.

## Rules
```yaml
- id: ADR-0007/R1
  statement: A move refusal reports the first failing check in the order not_found, validation_failed, invalid_move, wip_limit_exceeded; same-column and unknown column names are invalid_move
  scope: {contexts: [CTX-001], code: ["internal/board/**", "internal/httpapi/**"]}
  check: {kind: review-lens}
  severity: mandatory
```
