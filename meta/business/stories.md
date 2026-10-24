# Stories

## Board (2026-10-24)

US-001: As a TeamMember, I want to see all cards grouped into `todo`,
        `doing` and `done` on one shared board, so that I know where the
        team's work stands.
        Screen: Board at / [queries: GetBoard] [fr: FR-010]
          State empty: Empty board with three empty columns
          State populated: Board with cards in todo, doing and done
          State refreshed_after_action: Board re-read after every action, succeeded or refused, no polling (ASM-0004)
US-002: As a TeamMember, I want to add a card with a title and an optional
        description that lands at the end of `todo`, so that new work is
        captured in the queue.
        Screen: Board at / [fr: FR-010]
          State add_form: Add form visible with title and description inputs
US-003: As a TeamMember, I want to edit a card's title and description, so
        that the card stays accurate.
        Screen: Board at / [fr: FR-010]
          State edit_mode: Card in edit mode [click: #edit-card]
          State edit_saved: Edited card shows its new title
US-004: As a TeamMember, I want to delete a card, so that dropped work
        leaves the board.
        Screen: Board at / [fr: FR-010]
          State card_deleted: Card deleted, remaining cards keep dense positions
US-005: As a TeamMember, I want to move a card one column left or right so
        that it lands at the end of the target column, so that the board
        reflects progress.
        Screen: Board at / [fr: FR-010]
          State card_moved: Card moved to the adjacent column
          State doing_at_wip_limit: Doing column holds 3 cards
          State edge_controls_disabled: Move left on todo card and move right on done card rendered disabled
US-006: As a TeamMember, I want a refused action to show its reason on the
        page, so that I understand why nothing changed (WIP limit,
        non-adjacent move, invalid title).
        Screen: Board at / [fr: FR-011]
          State refused_wip_limit_exceeded: Reason shown for wip_limit_exceeded
          State refused_invalid_move: Reason shown for invalid_move on a stale page
          State refused_not_found: Reason shown for not_found on a stale page
          State refused_validation_failed: Reason shown for blank title in add form [click: #add-submit]
