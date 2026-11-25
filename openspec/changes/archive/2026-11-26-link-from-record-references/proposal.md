# Proposal

## Why

A node already records every entity its declaration refers to, but no rule reads that record for any source other than a structure, so a function, a method, a variable or a constant that names a package-level entity produces no connection. Enforcing what the node model already requires for a reference also prevents two connections the widened rule would answer about the wrong entity.

## What Changes

- A rule answers for a use reference the source carries, so a declaration connects to the node its parameters, its results and its body name. The relation is no longer restricted to a structure as the source, and the target is not filtered by kind.
- **BREAKING** A reference is recorded only when the entity it names is a node. An entity the model does not record as a node — a field, an interface method spec, a parameter, a local variable, a type parameter, a predeclared identifier — keeps its referent as the attribute value instead of a reference.
- The receiver name is no longer recorded as a reference, because the receiver is already recorded as the receiver property, so a method does not connect to the type it is declared on.
- Only entities visible at package level are nodes, so a reference to a local entity names no node and produces no pair.
- Primitives produce no pair by themselves: an entity with no record cannot be a member of a pair, so the rule answers nothing for it.
- The answered pair set on the test ground grows, which changes the `## Related` section the vault writes for a note.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `node-links`: the relation a rule answers is stated for any node as the source rather than for a structure only, and a rule that reads a reference decides a pair without reading a semantic value.
- `node-model`: a reference names a node or carries its value, and the receiver is recorded as the receiver property rather than as a reference.

## Impact

- `pkg/link`: one rule's source condition, and the fixed order the rules are consulted.
- `pkg/graph`: `recordUses` and `referentObj` in `pkg/graph/build.go` — references recorded only for nodes, receiver section skipped.
- Tests: `pkg/link` pins the size of the rule set and the rule named for an undecided pair; `pkg/graph` asserts recorded references.
- Output: `pkg/vault/note.go` builds the related section from `link.Pairs`, so `examples/Thing.md` and the vault tests change with the answered set.
- `docs/node-links.md` documents the rules the change decides.
