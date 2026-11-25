# Proposal

## Why

A node carries only the form of its declaration, which cannot separate entities the model
has to treat differently: a structure declaration and an interface declaration both report
`type`, and a declared type alias is recorded as a node even though it creates no entity. The
semantic type is what the node itself says about what the entity is, so a reader or a later
relation step does not have to re-derive it from the source or look at another node.

## What Changes

- Add a semantic type to the node record: a closed set, one value per node, derived from the
  entity's own declaration.
- Name the value for a defined type `structure` rather than `struct`, so the set does not use a
  word that is a Go term for a declaration body.
- Keep the set closed and extensible by reserving `unresolved` for a node that no rule
  classifies, so a value is added only when a case is decided and existing values never change.
- Set the derivation rule: a semantic type may use only the node's own declaration and what
  the language fixes for every package. A classification that needs another node is not a
  semantic type.
- Record node properties that never change the semantic type in one part of the node record:
  exported, generic, and the entry points the language assigns by name, `main` and `init`.
- **BREAKING** `Exported` moves from `Node` into `NodeProperties`, so a consumer reads every
  property from the same place; the printing path in `cmd/graphnodes` and the existing assertion in
  `pkg/graph/node_test.go` are updated in the same change.
- **BREAKING** A declared type alias is no longer a node. A declared alias is recorded as an
  attribute of the node it resolves to, as the test ground shows with `type Mirror = Struct1` on
  `Struct1` and `type NameAlias = Name` on `Name`, which is what the existing "Aliases do not create
  nodes" requirement says and the implementation now does.
- Extend the test ground with the cases the semantic set needs: the pair that separates a defined
  type from an alias (`type Name string` with `type NameAlias = Name`), a defined type with a
  collection body, a functional option type, an unnamed function held by a package-level variable, an
  empty package-level interface, `func init()`, a method with its own type parameters and a method
  written with an alias name as its receiver. An alias of a predeclared identifier is still missing.

## Capabilities

### New Capabilities

None. This change extends the existing node model rather than introducing a new boundary.

### Modified Capabilities

- `node-model`: adds the requirement that a node carries a semantic type and the rule that
  limits what a semantic type may be derived from; adds the requirement that properties do
  not affect the semantic type; narrows the alias requirement so a name created by a
  package-level alias declaration is an attribute of the resolved node rather than a node.

## Impact

- `pkg/graph`: a new closed set for the semantic type, a `SemanticType` field and a `NodeProperties`
  part on `Node`, an alias branch in `Classify` with the reason `ReasonAliasIsName`, property recording
  in `Build`, and json tags on the record so consumers read stable names.
- `pkg/graph/build_test.go` keeps its import-alias assertions, and the declared-alias case is checked
  over the test ground together with the closed-set test in `pkg/graph/node_test.go`, the rebuild test
  in `pkg/graph/identity_test.go` and the test that the entry point follows the package rather than the
  name.
- The printed form reads the package path as a segment of the module, so a module whose name is a
  prefix of another module name is read as a different module.
- `testdata` gained the cases listed above, and the coverage test over it holds a per-value check like
  the one that exists for kinds.
- `docs/node-model.md` gained the semantic set. The statement in the archived design that a declared
  alias records a node is superseded by this change and stays as a historical record; the living
  documentation states the rule.
- No new dependencies: `go/types` already exposes `TypeName.IsAlias()`, the `Interface`
  predicates and `Named.TypeParams()`.

## Out of scope for this change

The meaning of `structure` is settled: a package-level type declaration creates a new type, so
`type Name string`, `type Set map[string]int` and a functional option type are nodes with the
semantic type `structure`, while an interface body gives `interface`. `unresolved` keeps the set
closed while these points stay open, and nothing in the test ground produces it today, because
they are ownership questions rather than classification questions:

- an alias that resolves to something that is not a node (`int`, `map[string]int`, another
  alias), or to a node of another package, which no single-package build can own; such a
  declaration produces no node, so its attribute has no owner;
- a chain of aliases, which cannot be expressed because a reference points at a node and an
  alias is not a node;
- the exported flag for an alias name whose target node is unexported, where the name is
  exported and the node is not.
