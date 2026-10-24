# ADR-0011: Property-test tooling

**Status:** Accepted  
**Date:** 2026-10-24  
**Deciders:** @roman.miakotin  
**Revised:** —  
**Migration:** —  
**Pattern:** —  
**API-Posture:** —

## Context

NFR-002 (AC-030, EC-002) requires zero dense-order violations after every operation of a
random sequence of add, move, edit and delete operations, and EC-001 concerns the WIP
limit under concurrent moves. CON-008 requires every rule to be covered by a named test.
Following ADR-0004 (layered packages) and ADR-0006 (mutex inside the aggregate), the
property tests exercise `internal/board` through its locked API. CON-001 restricts only
the HTTP layer to the standard library, and ADR-0002 rules out third-party runtime
dependencies, so a test-only dependency is allowed but not required. The forces are zero
dependencies against readable, minimal failing cases when a property breaks, which
matters in a live demo.

## Decision

We will use pgregory.net/rapid: state-machine tests with add, move, edit and delete
actions and a dense-order check after each step, with automatic shrinking of a failing
sequence. Rapid's state-machine mode matches "after any operation" in NFR-002 directly,
and shrinking reduces a failure to the minimal sequence of operations, which is readable
on screen. Rapid is a test-only dependency; ADR-0002's "no third-party runtime
dependencies" still holds, because it governs the runtime only and nothing in the binary
imports rapid.

## Alternatives considered

### testing_quick
The standard library testing/quick would generate operation slices through a custom
Generate method, with no dependency. It is rejected because it has no shrinking, so a
failing sequence is reported at full length, and the package is frozen and documented as
such.

### seeded_math_rand
A hand-written loop over math/rand/v2 with a logged seed would apply N random
operations, with no dependency and full transparency. It is rejected because it has no
shrinking, reproducing a failure needs the logged seed, and the generator code becomes
test code of its own to maintain.

## Consequences

### Positive
- A failing property is reported as a minimal sequence of operations, easy to read and
  to turn into a regression test.
- The state-machine model checks dense order after every step, exactly as NFR-002 states.
- Generators for actions are declared, not hand-written loops.

### Negative
- go.mod gains a third-party test-only dependency (pgregory.net/rapid) to keep current.
- Rapid's state-machine API is one more thing for readers of the tests to learn.
- Property runs add test time; the number of checks must be tuned so `go test ./...`
  stays fast.

### Neutral
- CON-001 restricts only the HTTP layer, so a test-only dependency is allowed; ADR-0002's
  rule against third-party runtime dependencies is unchanged, because rapid is imported
  only from `_test.go` files.
- The EC-001 concurrency test (AC-029) still uses goroutines and `go test -race`
  (ADR-0006); rapid is used where random sequences are needed.

## References

- DEC-015 (resolved by this ADR)
- CTX-001 (Board), AGG-001, INV-004
- NFR-002, AC-030, EC-001, EC-002, CON-001, CON-008
- ADR-0002 (stack, runtime dependencies), ADR-0004 (package layout), ADR-0006 (mutex)
- rapid: https://pkg.go.dev/pgregory.net/rapid

## History

- 2026-10-24: Created — property tests use pgregory.net/rapid state machines with shrinking.
