# Changelog

## 2026-10-24
- CARD-003 (feature): Added the board page (embedded in the binary, styled only from the design tokens) with the add form, edit, delete, move and refusal reasons, served at `/`, and a headless-Chrome browser suite covering every page state.
- CARD-002 (feature): Added editing of cards (partial PATCH) and deletion through the API, and the 80-character title limit on add and edit; all five documented routes are now served.
- CARD-001 (feature): Added the board aggregate with its lock, the JSON API for reading the board, adding a card and moving it between columns under the WIP limit, and the server entry point; the invariants are covered by unit, property and concurrency tests.
