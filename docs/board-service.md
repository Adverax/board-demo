# Board — design brief

A small kanban board: a Go backend and a single web page in front of it. A
team keeps its work items on three columns and moves them along.

This brief is frozen for the live demo. Anything not written here is out of
scope.

## Users

- **Team member** — opens the page, adds cards, edits them, moves them between
  columns, deletes them. There is one shared board; everyone sees the same state.

## Capabilities

1. **Board.** The board has three fixed columns in this order: `todo`,
   `doing`, `done`. Reading the board returns every card grouped by column, in
   column order.
2. **Cards.** A card has a title (required) and an optional description. A new
   card always lands at the end of `todo`. A card's title and description can
   be edited; a card can be deleted.
3. **Moving.** A card moves one column to the right or to the left. It lands
   at the end of the target column.
4. **The page.** One page shows the three columns and their cards. Each card
   has "move left", "move right", "edit" and "delete" controls, and the page has
   a form to add a card. A refused action shows its reason on the page.

## Rules that must always hold

- **Title.** A title is not empty after trimming whitespace and is at most 80
  characters.
- **WIP limit.** `doing` holds at most 3 cards. A move into a full `doing` is
  refused and changes nothing.
- **Adjacent moves only.** A card moves only to a neighbouring column. A move
  from `todo` straight to `done` (or back) is refused.
- **Dense order.** Within every column the card positions are 0..n-1 without
  gaps or duplicates after any operation, including delete and move.

## Decisions already made

- Go 1.26, standard library `net/http` for the HTTP layer.
- In-memory storage guarded by a mutex; the board is lost on restart.
- JSON over HTTP under `/api`; errors are a JSON object with a stable `code`
  field (`validation_failed`, `not_found`, `wip_limit_exceeded`,
  `invalid_move`).
- Endpoints: `GET /api/board`, `POST /api/cards`, `PATCH /api/cards/{id}`,
  `POST /api/cards/{id}/move` (body `{"to": "<column>"}`),
  `DELETE /api/cards/{id}`.
- The page is plain HTML, CSS and JavaScript with no build step, embedded into
  the binary with `go:embed` and served by the same process.
- One binary, port from the `PORT` environment variable (default 8080).
- No authentication.

## Out of scope

- Authentication, users, per-user boards, multiple boards.
- Persistence across restarts.
- Custom or renamable columns, configurable WIP limit.
- Drag and drop, keyboard shortcuts, animations.
- Comments, labels, assignees, due dates.

## Later, not now

- **Archive.** Cards that stay in `done` for more than N days are archived and
  hidden from the board.
- **Drag and drop** between columns on the page.

## Success criteria for the first release

- The binary starts, serves the page, and every capability works end to end
  in a browser.
- A property test shows that positions stay dense after any random sequence of
  add, move, edit and delete operations.
- A concurrency test sends 20 simultaneous moves into `doing` and the column
  never exceeds 3 cards.
- Every rule above is covered by at least one named test.
