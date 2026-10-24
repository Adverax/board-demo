<!-- forge:business ledger. How to read it:
     · entries with status: active/pending/unvalidated are current — the rest is history;
     · entries are never deleted: a content change = supersession (a new entry with a
       source: link to the old one), a status change = editing status + a reason line;
     · a human-readable overview of the whole layer — README.md alongside (generated). -->

## ASM-0001 · PATCH is a partial update
- status: unvalidated
- owner: roman.miakotin
- source: delta domain extraction 2026-10-24, FR-003
- date: 2026-10-24
- check: "Operator confirms by 2026-11-07 that PATCH /api/cards/{id} changes only the fields present in the body (description editable alone, title absent keeps the old title) and that a present but empty or blank title is validation_failed"
- fatal: false
- basis: guess
- closed_by: null

PATCH /api/cards/{id} updates only the fields present in the body, and an absent title is not an error.

## ASM-0002 · The 80-character limit counts runes after trimming
- status: unvalidated
- owner: roman.miakotin
- source: delta domain extraction 2026-10-24, FR-008
- date: 2026-10-24
- check: "Operator confirms by 2026-11-07 that the limit counts Unicode characters (runes) of the trimmed title, so 80 two-byte characters are accepted and 81 rejected"
- fatal: false
- basis: guess
- closed_by: null

The title limit of 80 is measured in Unicode characters after trimming surrounding whitespace, not in bytes.

## ASM-0003 · A malformed move body maps to validation_failed
- status: unvalidated
- owner: roman.miakotin
- source: delta domain extraction 2026-10-24, FR-005, CON-004
- date: 2026-10-24
- check: "Operator confirms by 2026-11-07 that unparseable JSON and a missing or non-string to on POST /api/cards/{id}/move return validation_failed (see DEC-001 for unknown column names)"
- fatal: false
- basis: guess
- closed_by: null

A move request with a malformed body, or a missing or non-string to, is answered with validation_failed.
