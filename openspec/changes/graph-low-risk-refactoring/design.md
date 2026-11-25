# Design

## Context

`pkg/graph` is the node model: the definition of what a node is, the closed sets, the exclusion reasons,
the identity rules and the enumeration that produces a record per package. It is covered at 87.4%, passes
`go vet` and `golangci-lint`, and its tests read as one check per spec scenario. See `proposal.md - Why`
for the motivation. The constraints that shape this change:

- The package's value is that the code states the definition once. Anything that puts the same rule in two
  places is the problem being removed.
- The printed forms are file names for a vault, so they are behavior and are not touched.
- Every helper touched here is unexported, so a step can be proven by build, vet and test alone.

## Goals / Non-Goals

**Goals:**

- One place decides each rule that a test already asserts.
- Test helpers live in one file and are used, not re-declared per test.
- Every commit in this change is structural only, with a green suite after each.

**Non-Goals:**

- No change to what a node is, to the closed sets, to a reason, to a printed form, or to the order of
  recorded attributes. Determinism is guaranteed in two places today (the file order in attribute recording
  and the attribute sort at the end of the build); keeping both costs nothing and removing one changes
  observable output, so neither is touched.
- No reformatting of tests into table-driven form: a test name is the spec scenario it checks, and the
  mapping would be lost.

## Decisions

**Receiver normalization has one decision point.** `receiverName` keeps its own job (report the name) but
resolves the receiver through `baseNamed`, which is the primitive the rest of the enumeration already uses
for the same pointer-strip-and-alias-resolution rule. The alternative was to keep both copies, which is
what allows the two to drift: the rule that a pointer wrapper is a property of the node rather than of its
identity is stated in the model once, and two implementations of it can disagree later.

**The unused `reason` argument is dropped, not used.** `add` decides a reason from `Classify` and from the
declaration index; the argument is only consulted when a caller passes one, and no caller does. The
alternative was to have the method enumeration pass the lookup reason, which would change which candidates
are reported, so it is a behavior change and belongs to a different commit. Dropping the argument removes
a branch that cannot be reached.

**Functions that build a record sit with the record type.** `position` and `lineRange` move to the file
that holds `Position` and `LineRange`. The alternative was to leave them where they are used, but the pair
is small and its only job is to fill those two types, so reading the type and the builder together is the
shorter path.

**Test helpers move to the shared test file, and the repeated lookup guard becomes one helper.** The
alternative was table-driven tests, rejected above. The helper that fails on a missing node keeps the same
message shape the tests print today, so a failure still names the node that was expected.

**Each step is a separate commit.** A rename and a helper move are proven by the tooling plus the suite; a
step that removes a branch needs the suite to show the branch was unreachable. Mixing them forces one
reviewer to check both, and this package has no reason to accept that.

## Risks / Trade-offs

- A helper rename can surface in a test that asserts a printed form → run the full suite, not only the
  tests for the touched file.
- The branches being removed are not covered today, so no test proves they were unreachable → the check is
  that the call sites no longer pass a value that could reach them, and the suite stays green from a
  committed baseline before each step.
- The single decision point makes `receiverName` depend on `baseNamed`, so a future change to the alias rule
  now changes both the identifier and the receiver name. That is the intended coupling, since the model
  states the rule once, but it means a test that expects them to differ would need to be written on purpose.
