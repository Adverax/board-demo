# Component inventory — Board

Derived from stories US-001..US-006, capabilities CAP-001..003 and screen SCR-001 (`domain/screens.yml`, planned component COMP-003 `web/`). States are exhaustive on purpose. Behaviour that depends on an open decision is marked **pending** and not prescribed.

## AppShell
Used by: SCR-001 (the only page). Layout: centered column, max width 1100px, 16px side gutter; header row (title + add form), refusal notice, board.
Tokens: --color-bg, --color-text, --font-ui, --text-base, --leading-base, --space-4, --space-5.
States: default · narrow viewport (≤760px: columns stack in a single column).

## AddCardForm
Used by: US-002.
States: default (empty title, optional description) · focus · filled · submitting · validation-error (blank or over-80-character title: the reason is shown through RefusalNotice, the title input gets a danger outline) · disabled while submitting.
Tokens: --color-surface, --color-border, --color-accent (button), --color-on-accent, --radius-control, --text-base, --space-2, --space-3, --color-focus-ring.
Notes: primary button label "Add card"; title input is the first focusable field.

## RefusalNotice
Used by: US-006 (FR-011). Shows the reason of the last refused action.
States: hidden (nothing refused) · wip_limit_exceeded ("doing already holds 3 cards") · invalid_move (non-adjacent move) · not_found (the card is already gone, e.g. deleted by another member) · validation_failed (blank or too long title).
Tokens: --color-danger, --color-danger-bg, --radius-control, --space-2, --space-3, --weight-strong.
Notes: `role="alert"`; text is the server's reason, never colour alone. Decided by ADR-0009: after every action, succeeded or refused, the page shows any reason and then re-reads GET /api/board and re-renders; the re-render does not hide the reason just shown. No polling.

## Board
Used by: US-001 (FR-001, FR-010). A three-column grid in fixed order `todo`, `doing`, `done`.
States: populated · empty (all three columns empty) · re-rendered after an action (ADR-0009: after every action, succeeded or refused, the page re-reads GET /api/board and re-renders from it; no polling, the board changes only after the user's own action, ASM-0004).
Tokens: --space-4 (gap), --color-bg.

## Column
Used by: US-001.
States: default · empty ("No cards" in --color-text-secondary) · at WIP limit (`doing` holding 3 cards: 2px --color-accent outline, the single sharp anchor) · highlighted target (not used: drag and drop is out of scope).
Tokens: --color-column, --radius-container, --outline-emphasis, --color-accent, --text-base, --weight-heavy (column name), --space-3.
Notes: header = column name + a pill with the card count (--color-surface, --color-border, --radius-control, --text-sm). The limit indicator is the outline only; no "3 of 3" text.

## Card
Used by: US-001..US-005. Title (required), optional description, and CardActions.
States: default · hover (action buttons gain --color-text border) · focus-within · edit-mode (swaps to EditCardForm) · being-deleted · stale (rendered from an outdated page; a later action may be refused with `not_found` or `invalid_move`).
Tokens: --color-surface, --color-border, --radius-card, --shadow-raised, --color-text (title, --weight-strong), --color-text-secondary (description, --text-sm), --space-2, --space-3.
Notes: titles up to 80 characters wrap onto a second line, never truncate; description only when present.

## CardActions (IconButton ×4)
Used by: US-003, US-004, US-005 (FR-010). Four 28px icon buttons: move left, move right, edit, delete.
States: default · hover · focus-visible (--color-focus-ring, 2px, 2px offset) · pressed · disabled (edge controls only: "move left" on a `todo` card, "move right" on a `done` card).
Tokens: --color-text-secondary, --color-border, --color-text (hover), --radius-control, --space-1; disabled: icon and border both --color-border, no hover change, `cursor: not-allowed` (no new token).
Notes: each button has an accessible name ("Move left", "Move right", "Edit", "Delete"), icons are inline SVG at 14px, never emoji. Decided by ADR-0008: "move left" on a `todo` card and "move right" on a `done` card are rendered but disabled (native `disabled` attribute), so every card keeps four controls (AC-025) and no request is sent; the accessible name stays the same and the disabled state is exposed to assistive technology by the attribute, not by colour. Inactive controls are exempt from the WCAG 1.4.3/1.4.11 contrast minimums, so the faded --color-border icon is allowed; all other controls stay enabled.

## EditCardForm
Used by: US-003. Replaces the card body in place.
States: default (pre-filled title and description) · focus · validation-error (blank or over-80-character title, reason in RefusalNotice) · saving · saved (back to Card with the new title).
Tokens: same as AddCardForm; buttons "Save" (--color-accent) and "Cancel" (outlined, --color-border).
Notes: PATCH semantics (partial update, ASM-0001) are not visible to the user; saving an unchanged form is allowed.
