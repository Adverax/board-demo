# Design brief — Board

## Personality
Audience: a small team at their desks; the board stays open all day and is checked in short, frequent glances ("what is in doing?"). One shared board, one role (team member), no onboarding flow.
Sliders: serious ●───○─ playful · dense ─●──○─ airy (a notch denser than medium) · classic ──○──● experimental (familiar layout, one bold decision) · warm ─○──●─ cool (neutral background with a warm cast)
Mood words: calm, warm, legible, trustworthy, unhurried
References liked: Basecamp, Trello
References disliked: Jira
Hard constraints: light and dark themes (via `prefers-color-scheme`, plus a manual toggle in previews); WCAG AA contrast; plain HTML/CSS/JS with no build step (CON-005), so design tokens are plain CSS custom properties; no brand identity exists.

## Direction: Soft depth / Humanist (chosen 2026-10-24)
Warm cream surfaces with a single terracotta accent, rounded controls sitting inside squarer containers, and one level of tinted soft shadow on cards. It fits a board that people glance at all day: warmth lowers the visual noise of a work tool without making it a toy, and the rounded-humanist face keeps card titles readable at 15px. It satisfies the sliders: serious-with-warm-accents (restrained palette, one accent), slightly dense (4px rhythm, 15px base), familiar layout (three columns of cards) with one bold decision, and a neutral-warm background.

The one bold decision: a column that has reached its WIP limit (`doing` at 3 cards) gets a crisp 2px accent outline. Everything else is soft, so that single sharp edge is what the eye finds first when a move is refused.

Base direction(s): Soft depth / Humanist (style-directions.md entry 3).

Deliberate deviation from the direction bank: the bank asks for gentle ease-out motion and micro-feedback on every interaction. The frozen brief lists animations as out of scope, so all motion tokens are 0. A later brief revision may restore them without renaming any token.

### Rejected candidates
- **Swiss / International** — precise and legible, but its cold grid-and-hairline character sits closest to the Jira look the team dislikes, and it ignores the warmth the Basecamp/Trello references ask for.
- **Warm print** — the most distinctive on sight, but serif card titles slow scanning in short glances, hairline cards leave the move/edit/delete controls with the least visible affordance, and the period cue risks kitsch on a work tool.

## Anti-patterns (this product must NOT look like)
- Jira: grey corporate chrome, dense toolbars and badges everywhere, visual noise around the cards.
- A consumer toy: saturated pastel column headers, mascots, confetti. The audience is a working team.
- An AI demo: purple gradients, emoji as icons, a generic rounded-xl card with a soft grey shadow on every element (the global anti-defaults list in style-directions.md applies).
- Colour used for decoration: the accent is reserved for the primary action, focus, and the WIP-limit frame.

## Open points that the design deliberately does not decide
- DEC-007: how the "move left" control on a `todo` card and the "move right" control on a `done` card behave (disabled, hidden, or sending the request). The previews render both enabled; the tokens and components do not prescribe a state for them.
- DEC-013 / ASM-0004: how the page refreshes after an action. Components describe observable outcomes only.

## History
- 2026-10-24 — initial direction "Soft depth / Humanist" chosen (interactive); candidates "Swiss / International" and "Warm print" kept in previews/ for comparison
