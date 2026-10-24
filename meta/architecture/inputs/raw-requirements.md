# Raw requirements

# processed: 2026-10-24 → FR-001
- The board has three fixed columns in this order: todo, doing, done (FR). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → FR-001
- Reading the board returns every card grouped by column, in column order (FR). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → FR-002
- A card has a required title and an optional description (FR). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → FR-002
- A new card always lands at the end of todo (FR). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → FR-003, FR-004
- A card's title and description can be edited; a card can be deleted (FR). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → FR-005
- A card moves one column to the right or to the left and lands at the end of the target column (FR). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → FR-006
- A move to a non-neighbouring column (todo to done or back) is refused (FR). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → FR-007
- doing holds at most 3 cards; a move into a full doing is refused and changes nothing (FR). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → FR-008
- A title is not empty after trimming whitespace and is at most 80 characters (FR). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → FR-009
- Within every column card positions are 0..n-1 without gaps or duplicates after any operation, including delete and move (FR). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → FR-010
- One page shows the three columns and their cards; each card has move left, move right, edit and delete controls; the page has a form to add a card (FR). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → FR-011
- A refused action shows its reason on the page (FR). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → CON-003
- Endpoints: GET /api/board, POST /api/cards, PATCH /api/cards/{id}, POST /api/cards/{id}/move with body {"to": "<column>"}, DELETE /api/cards/{id} (CON). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → CON-004
- Errors are a JSON object with a stable code field: validation_failed, not_found, wip_limit_exceeded, invalid_move (CON). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → CON-001, CON-002
- Go 1.26 and standard library net/http for the HTTP layer; in-memory storage guarded by a mutex (CON). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → CON-005
- The page is plain HTML, CSS and JavaScript with no build step, embedded with go:embed and served by the same process (CON). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → CON-006, CON-007
- One binary; port from the PORT environment variable, default 8080; no authentication (CON). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → NFR-001
- 20 simultaneous moves into doing never leave the column above 3 cards (NFR, concurrency test). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → NFR-002
- A property test shows positions stay dense after any random sequence of add, move, edit and delete operations (NFR, verification). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → CON-008
- Every rule is covered by at least one named test (CON, verification). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → NFR-003
- The binary starts, serves the page, and every capability works end to end in a browser (NFR, release criterion). Source: docs/board-service.md, 2026-10-24.
# processed: 2026-10-24 → EC-001
- Narrow EC-001: for any number of simultaneous moves into doing, every snapshot read of the board, and the board after the moves settle, shows at most 3 cards in doing (replaces "never holds more than 3 cards at any observable instant"; the test name and the links to AC-018 and INV-002 stay) (NFR/EC). Source: ADR-0005 (DEC-006), 2026-10-24.
# processed: 2026-10-24 → nfr_coverage (performance, observability, compliance, cost)
- NFR categories performance, observability, compliance and cost are closed as not applicable for the demo: no latency or throughput target, no logging or metrics requirement, no regulatory or organizational rule, no hosting or resource budget (NFR coverage). Source: ADR-0012, ADR-0013, ADR-0014, ADR-0015 (DEC-002..005), 2026-10-24.
