# Proposal

## Why

The `graph` package has no behavior problem to solve, so this change is not a rewrite: it removes the few
places where the same rule is written twice and where test helpers are scattered, so that the next change
that joins per-package graphs can add to the package without re-deriving rules that already exist. Only
the steps a build, vet and test run can prove are behavior-preserving are in scope.

## What Changes

- `receiverName` reads the receiver through `baseNamed` instead of repeating the pointer strip, alias
  resolution and named-type check that `baseNamed` already performs.
- The enumeration helper `add` loses its `reason` argument, which every caller passes as an empty string
  and which therefore never affects a decision.
- Helpers that build a record (`position`, `lineRange`) move next to the record types they produce.
- Test helpers are consolidated in one file, and the repeated "find the node or fail" guard becomes one
  helper, so a failure message stays the same while the boilerplate disappears.

Out of scope: splitting the enumeration state by responsibility, folding the exclusion rule shared by
`Classify` and the inner-entity reason into one decision, replacing `sort` with `slices`, any change to
an exported name, the redundant determinism guarantee in attribute recording, and any test that asserts
behavior the specs already state.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. This change is purely structural: the node set, the closed kind and semantic sets, the exclusion
reasons, the canonical printed form and the attribute contents are unchanged, so no requirement changes
and no spec file is written (`skip_specs: true`).

## Impact

- `pkg/graph`: two private helpers and one unexported method signature; one pair of functions moves files.
- `pkg/graph` tests: helper placement and one shared test lookup helper; no test loses an assertion.
- No exported API changes, so `pkg/vault`, `pkg/link` and `cmd/graphnodes` are unaffected.
- Verification stays `go build ./... && go vet ./... && go test ./...` after each step, with the current
  87.4% package coverage kept as the safety net.
