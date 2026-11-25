# Tasks

## 1. Record a reference only when it names a node

- [x] 1.1 Record the reference of a use attribute only when the object it names is a node, and keep the referent as the attribute value otherwise. Verify with a build test that a use of a field, a parameter, a local variable and a type parameter carries a value with no reference.
- [x] 1.2 Record a use for a method through an interface value as a value with no reference, since a method spec has no receiver to tell it apart from a package-level declaration of the same name. Verify with a build test for `describe` in the test ground.
- [x] 1.3 Stop recording a use for the name written as the receiver, since the receiver is recorded as the receiver property. Verify with a build test that a method node carries no use attribute for its receiver type while the receiver property still names the node.
- [x] 1.4 Document the reference rule in `docs/node-model.md` and verify the wording matches what the build records for a field, a method spec and a receiver.

## 2. The rule that reads references

- [x] 2.1 Add the rule to the rule set at the end of the fixed order, reading the use references of any source kind with no filter on the target kind. Verify against the test ground pairs: `describe` to `Reader`, `NewStruct1` to `Struct1`, `zeroStruct` to `Struct1`, `main` to `Five`.
- [x] 2.2 Verify the rule answers no for a source whose uses name no node and for alias and method attributes, so the existing exclusions keep answering no.
- [x] 2.3 Verify the rule answers for a pair whose member reports `unresolved` without naming a rule, and that the rule named for an undecided pair is unchanged.
- [x] 2.4 Update the assertion that pins the size of the rule set and verify the coverage run reports no uncovered rule for the test ground.
- [x] 2.5 Document the rule and its position in the order in `docs/node-links.md`, and verify the documented ground pairs match what `Pairs` returns.

## 3. Integration checks

- [x] 3.1 Run `go test ./...` and verify the whole suite is green with the answered set grown.
- [x] 3.2 Run `graphnodes` against the test ground and inspect a generated note file for a function node, verifying its related section lists the node its parameter names and that `examples/Thing.md` still matches the generated format.

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied.
- Verify the archived result.
- Rename the covered-relations requirement in a follow-up change once the specs are synced, because its title names two relations while its body enumerates four; the archive matches a MODIFIED requirement by the name it already has, so the rename cannot happen in this change.
