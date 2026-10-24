# ADR-0008: Edge move controls on the page

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** —

## Context

FR-010 says each card on the page has move left, move right, edit and delete controls,
and AC-025 tests that a card shows all four. A card in todo has no column to its left and
a card in done has no column to its right, and the brief is silent on what those two
controls do at the edges. FR-006 refuses a move to a non-neighbouring column, and FR-011
requires a refused action to show its reason on the page. CON-004 fixes the error codes
the API can return; CON-005 requires the page to be plain HTML, CSS and JavaScript with no
build step. The forces are a literal reading of "four controls on every card" against a
control that can never succeed, and uniform behaviour (every click is a request) against
inventing a target column that does not exist.

## Decision

We will render "move left" on a todo card and "move right" on a done card but disabled;
all other controls stay enabled. This keeps the four controls of FR-010 on every card, so
AC-025 holds literally, and it means an impossible move never leaves the page, so no
target column name has to be invented for a move off the board edge. A disabled control
is a refusal the user can see before clicking, so no request and no reason message are
needed for it.

## Alternatives considered

### hidden
The control would not be rendered at the edge, so there would be no dead control on the
page. It is rejected because it contradicts the literal requirement that each card has
move left and move right controls (FR-010, AC-025): a todo card would show three
controls.

### sends_request
The control would stay active, send a move request and show the `invalid_move` reason on
the page, giving uniform behaviour and exercising the FR-011 reason display. It is
rejected because a move off the board edge has no target column name to send; the page
would have to invent one (an unknown name, or the card's own column), turning a
predictable UI state into a deliberate API error.

## Consequences

### Positive
- Every card shows four controls, so AC-025 holds on every column.
- No request is sent for a move that cannot succeed; the page never invents a column.
- The edge state is visible before the click, which is clearer than an after-the-fact
  error message.

### Negative
- A disabled control is not covered by the FR-011 refusal-reason rule: the user sees the
  control greyed out but gets no stated reason, unless a tooltip or label is added.
- The page must compute the disabled state from the card's column on every render, a
  small piece of rule knowledge (the column order) duplicated in the page.
- An extra e2e check is needed to show the edge controls are disabled, beyond AC-025.

### Neutral
- Follow-up, not done by this ADR: the screens model state `edge_controls_pending_dec007`
  in `meta/architecture/domain/screens.yml` and the CardActions component in
  `meta/design/components.md` still mark this as pending and must be updated to the
  disabled behaviour.
- ADR-0007 still defines what the API returns if another client sends such a move.

## References

- DEC-007 (resolved by this ADR)
- CTX-001 (Board)
- FR-006, FR-010, FR-011, AC-025, CON-004, CON-005
- ADR-0007 (move error codes)

## History

- 2026-10-24: Created — edge move controls are rendered but disabled; no request is sent.
