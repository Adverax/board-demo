# web

The board page (COMP-003, SCR-001): plain HTML, CSS and JavaScript with no build step,
no framework and no external resources (CON-005), embedded into the binary with
`go:embed` and served by the same process.

| file | role |
| --- | --- |
| `web.go` | `//go:embed` of the files below; `Page() fs.FS` hands them to `httpapi.NewApp` |
| `index.html` | the page shell: header with the add form, RefusalNotice (`#refusal`, `role="alert"`), the three columns, inline SVG icon symbols |
| `static/app.js` | fetches `/api/*`, renders the board, runs the actions |
| `static/app.css` | component styles (AppShell, Column, Card, CardActions, forms, RefusalNotice) using only design tokens |
| `static/tokens.css` | **derived** copy of `meta/design/tokens.css`, do not edit |

Served as `GET /` (index.html) and `GET /static/{name}` (internal/httpapi `NewApp`).

## Behaviour (ADR-0008, ADR-0009)

- After every action, succeeded or refused, the page shows any reason in the
  RefusalNotice (hidden on success), then re-reads `GET /api/board` and re-renders.
  There is no polling (ASM-0004).
- Only one action is in flight per page; a click while one runs is ignored.
- Every fetch is aborted after `REQUEST_TIMEOUT_MS` (5 s). A network failure or timeout
  shows "the board could not be reached", a response that is not the API's shape shows
  "unexpected response (HTTP n)", and the last rendered board stays.
- "Move left" on a todo card and "Move right" on a done card are rendered with the
  native `disabled` attribute and never send a request.
- The notice carries the code in `data-code` (`network` and `unexpected` are the page's
  own codes). The page words the reasons; the server sends only the code.
- A refused edit keeps the edit form (`#edit-card`) open with the user's text.

## Design tokens

`go:embed` cannot reach `meta/design/`, so `static/tokens.css` is the marker line
followed by `meta/design/tokens.css` byte for byte. `TestTokensCSS_IsVerbatimCopyOfDesignTokens`
fails when they differ. After `/forge:ui` changes the tokens, regenerate the copy from
the repository root:

    { printf '%s\n' "$(head -1 web/static/tokens.css)"; cat meta/design/tokens.css; } > web/static/tokens.css

`static/app.css` uses only `var(--…)` tokens. The exceptions are the AppShell layout
values from `meta/design/components.md` (max width 1100px, one column at ≤760px), which
tokens.css has no role for.

## Tests

- `web_test.go`: the tokens copy, and that every asset `index.html` links is embedded and
  same-origin.
- The browser tests (`TestE2E_*`, chromedp) live in `cmd/board/e2e_test.go`.
