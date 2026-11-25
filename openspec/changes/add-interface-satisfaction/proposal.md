# Proposal

## Why

The link rules currently cover two relation types, both from a structure to another node. A connection between an interface and the structures that satisfy it is a distinct relation that the rule set does not answer. Adding it extends the graph without changing existing rules, and it is the first rule whose source is not a structure, which exercises the rule set in a new way.

## What Changes

- A new rule `interface to a struct that satisfies it` is added to the rule set, consulted after the existing rules in a fixed order.
- The rule answers yes when every method spec recorded as an attribute of the interface is present as a method recorded on the structure, with the same name, the same receiver and the same signature.
- A structure records the methods declared on it as attributes, so that a relation can read the method set from the structure alone instead of looking up other records.
- The rule set in the spec now covers three relation types instead of two.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `node-links`: The requirement on the rule set coverage gains a third relation type. A new requirement defines interface satisfaction: a connection exists when every method spec of the interface source is present as a method recorded on the structure target, matched by name, receiver and signature.
- `node-model`: The requirement on preserved attributes gains the kind `method`, and a new requirement states that a structure records each method declared on it as an attribute carrying the signature in the same form a method spec uses. Without this the rule cannot answer from the pair alone, because its method set lives in records outside the pair.

## Impact

- `pkg/graph/attribute.go`: a new attribute kind for a method declared on a type node.
- `pkg/graph/build.go`: each declared method is recorded as an attribute of the node that owns it, with the signature rendered without the receiver.
- `pkg/link/link.go`: a new rule variable and an `answerSatisfiedInterface` function that compares name, receiver and signature, and skips method attributes when reading referents.
- `pkg/link/testhelper_test.go`: a `stubInterface` helper for tests that need an interface record.
- `pkg/link/link_test.go`: test cases for satisfied, not satisfied, signature mismatch and undecided.
- `pkg/graph/build_test.go`: a test that a type node carries its declared method as an attribute.
- `testdata/main.go`: no change needed — `Reader` with `Read() error`, `Base` with `Read() error` and `Thing` with `Read() int` already cover a match, a missing method and a signature mismatch.
