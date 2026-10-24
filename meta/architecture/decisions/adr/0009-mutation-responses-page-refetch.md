# ADR-0009: Mutation responses and page refresh

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** —

## Context

CON-003 fixes the five routes and the body of a move request but not the response bodies
of the mutating endpoints (POST /api/cards, PATCH /api/cards/{id},
POST /api/cards/{id}/move, DELETE /api/cards/{id}). CON-004 fixes the error object of a
refusal. The page (FR-010) must show the board after each action, and FR-011 requires a
refused action to show its reason. AC-034 and AC-035 describe a stale page: another
member has moved or deleted a card since the page was rendered, and a click on it must
show `invalid_move` or `not_found`. ASM-0004 assumes there is no auto-refresh. Following
ADR-0002 (stdlib, API posture server-rendered: the embedded page is the only consumer),
the response shape is a contract only between the handlers and that page. The forces are
round trips per action against one source of truth for the rendered board, and
conventional REST bodies against a single render path.

## Decision

We will have POST /api/cards, PATCH /api/cards/{id} and POST /api/cards/{id}/move return
the affected card as JSON and DELETE /api/cards/{id} return 204 with no body; after every
action, whether it succeeded or was refused, the page shows any reason and then re-reads
GET /api/board and re-renders. The page therefore always shows server truth after a
click, which makes the outcomes of AC-034 and AC-035 visible (the moved or deleted card
appears in its real place or disappears), and the page never needs its own copy of the
dense-order or adjacency rules. The bodies follow common REST practice, so the API stays
readable without the page.

## Alternatives considered

### board_response
Every mutation would return the full board in the GET /api/board shape and the page would
re-render from that response, one round trip and one render path. It is rejected because
the bodies are unconventional for REST, and a refusal still returns only the CON-004
error object, so the page would need a second path after a refusal anyway, losing the
single-path benefit.

### card_response_local_patch
Mutations would return the affected card and the page would patch its DOM from the
response, never re-reading the board until reload, for minimal traffic. It is rejected
because the page drifts from server state whenever another member acts, and the page
would have to replicate the dense-order placement logic (INV-004) to put a moved card at
the right position.

## Consequences

### Positive
- After every click the page shows the board as the server holds it, including changes
  made by other members since the last render.
- The page contains no copy of the board rules; it renders whatever GET /api/board says.
- Response bodies are conventional and easy to test at the API level.
- One render function (from GET /api/board) serves first load and every action.

### Negative
- Two requests per action (the mutation, then GET /api/board); harmless at demo scale but
  doubles traffic per click.
- A short window exists between the mutation response and the re-read in which another
  change can land; the page then shows the newer state, which is correct but can surprise
  a user watching a specific card.
- The card returned by a successful mutation is not used by the page for rendering, so
  that response shape is tested by API tests only.

### Neutral
- This decision does not introduce polling: the page refreshes only after the user's own
  action. ASM-0004 (no auto-refresh) therefore stays an unvalidated assumption, and the
  stale-page scenarios AC-034 and AC-035 remain valid.
- Follow-up, not done by this ADR: the screens model state `refresh_pending_dec013` in
  `meta/architecture/domain/screens.yml` is now decidable and must be updated.

## References

- DEC-013 (resolved by this ADR)
- CTX-001 (Board), INV-004
- CON-003, CON-004, FR-010, FR-011, AC-034, AC-035
- ASM-0004
- ADR-0002 (stack and API posture)

## History

- 2026-10-24: Created — mutations return the affected card (DELETE 204); the page re-reads
  GET /api/board after every action, succeeded or refused.
