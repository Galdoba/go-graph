# Design

## Context

`pkg/graph` records `Kind` as a closed set of four values built from a `types.Object` type
switch, and before this change `Classify` had no branch for aliases, so a declared type alias
was recorded as a node. `go/types` is the source of truth for what an entity is: an alias name reports
`IsAlias()` true and its type is resolved through `Unalias()`, so the checker does not treat it
as a new type. Only one package is in scope, so nothing here can look at a node outside the
package being built. See proposal.md for motivation and specs/node-model/spec.md for the
requirements.

## Goals / Non-Goals

**Goals:**

- A semantic set that can be computed from the node alone, inside one package.
- Coverage over the test ground for every value, mirroring the existing kind coverage check.
- Alias behaviour that matches the requirement instead of contradicting it.

**Non-Goals:**

- Relations and the graph. A classification that needs another node is out.
- Package-level attributes, which the node model already deferred to the graph change.
- Role names based on naming convention (`constructor`, `getter`, `factory`, `marker`,
  `embedded base`).
- Deciding how to handle an alias attribute whose target is not a node or whose name is exported
  while the node is not. These are ownership questions rather than classification questions,
  because an alias declaration produces no node, so `unresolved` does not apply to them.
- Recording an alias attribute whose resolved target is a node of another package. Only one package
  is in scope, so the build has no node to own that attribute; the identifier of the foreign node is
  derivable from the object, but it can only be placed once the graph joins packages.

## Decisions

**The semantic type is a second field, not a refinement of `Kind`.**

`Kind` stays the closed set of four declaration forms, because the node-model spec states that
no other kind is produced and identity rules are written against it. The semantic set is a
separate field with its own closed set and its own coverage check. Alternative considered:
extend `Kind` with `structure` and `interface`; rejected, because it would break a requirement that
identity already depends on.

**Six decided values plus one reserved value.**

`structure`, `interface`, `function`, `method`, `constant`, `variable`, and `unresolved`. The six
decided values cover every node that the current rules classify, and `unresolved` is reserved for
a node that no rule classifies, so the set stays closed while individual cases are decided one at
a time and a value already decided keeps its meaning when a new one is added. `unresolved` is the
zero value, the way `KindUnknown` is, so an unset semantic type is never mistaken for a decided
one. No compiling entity produces `unresolved` today, because the mapping already covers every kind of
the closed set, which is the evidence that the rules are complete over what the model can produce. The
value stays as the default so that a case the mapping does not cover is reported as undecided instead of
being given a wrong name.

A package-level type declaration always produces a node, and the declared body decides which of
the two type values it gets: an interface body gives `interface`, anything else gives `structure`.
That covers `type Name string`, `type Set map[string]int` and a functional option type, so none of
them needs `unresolved`.

`structure` was chosen over `struct` because `struct` is the language's word for a declaration
body, and the value covers defined types whose body is not a struct. The alternative is the
language's own underlying-type taxonomy (basic, map, slice, signature), which is closed by
construction too, but it is finer than what was agreed and would report a functional option type
as `signature`.

**Derivation uses the node's own declaration plus fixed language facts.**

The facts `go/types` exposes are `TypeName.IsAlias()`, `Named.TypeParams()`, the `Interface`
predicates, and `Basic` information. `Underlying()` is used only to tell whether the declared body
is an interface, not to pick a value from the taxonomy of the underlying type. A value that
requires looking at another node is a relation, not a semantic type, and this change cannot
compute it. This is what keeps the semantic type buildable per package, which is how `Build` works
today.

**Semantic type is derived data, not identity.**

The identifier is built from module path, module version, package path, optional receiver and
name, and must not gain the semantic type. The value is recomputed on every rebuild, like the
declaration position. The rebuild comparison has to include it, because it can change without
identity or signature changing, for example when a defined type gains its first method and stops
being an empty interface body.

**`method` duplicates `Receiver` on purpose.**

A method node already carries its receiver, so `method` is a rendering of a fact in the record
rather than a new fact. Accepted, with the rule that the value is computed from `Receiver` and
never stored independently, so the two cannot disagree. The same applies to the entry point, which is a
fact of the declared name. Exported is a fact of the name too, and it is stored once, in the properties
part, so there is no second copy that can drift.

**Declared type aliases produce no node.**

`Classify` gains an alias branch that reports the entity as excluded, and the name is recorded as
an attribute of the node the alias resolves to. Alternative considered: keep the alias as a node,
which is what the archived design says; rejected, because `types.Identical` proves the alias and
its target are one entity, and keeping both gives one entity two identifiers. This supersedes the
archived statement and the documentation has to be corrected.

**An alias owns nothing.**

The language resolves a method's receiver base type through an alias name, so a method written on
an alias belongs to the entity the alias resolves to, and a method cannot be written at all when
the alias resolves to a predeclared or foreign type. Interface satisfaction is attributed to the
resolved type in the same way, so the alias attribute is metadata only. This is the reason the
semantic set has no alias value: the alias has no shape of its own.

**Properties stay separate from the semantic type.**

`exported`, `generic` and the language-assigned entry points `main` and `init` are properties, and none
of them changes the semantic type value. Folding `generic` into the value would produce
`generic_structure`, `generic_interface` and more, so the set would stop being closed by construction.
Entry points are accepted as properties only because the language assigns them: a function named `main`
in a package that is not `main` is an ordinary function.

**The semantic set follows the pattern `Kind` already uses.**

A named enum with a `String()` form, zero value reserved for "not classified", same as
`KindUnknown` and `AttrUnknown`. A node prints its semantic value next to the kind and its properties as
one group, so a printed record stays predictable even when the property set grows.

**Names for the record, and json tags on it.**

Properties are held in one field on `Node`, `Properties NodeProperties`, where `NodeProperties` is a
struct with one boolean per property, so the record gains one field now and keeps its shape when the
set grows, while a consumer still reads a named value per property (`node.Properties.Generic`) when it
walks a collection of nodes or sends the whole graph somewhere. The name is qualified with `Node`
because `Attribute` already has a value part, and an unqualified `Properties` could be read as
belonging to that. The properties are `Exported`, `Generic` and `EntryPoint`, each false when it does not
apply, so the record reads `SemanticType` and `Properties` only. `Exported` moves from `Node` into
`NodeProperties` in this change: nothing inside the model reads the field, it is set in `Build`, printed
by `cmd/graphnodes` and asserted in one test, and exported is a fact of the node's own declaration, which
is what a property is. The move changes the shape of the record for a consumer, so the printing path and
that test are updated in the same change. The record types get json tags with snake_case names, because a
tag is what a consumer reads: `semantic_type`, `properties`, `properties.exported`,
`properties.generic` and `properties.entry_point`. The semantic constants follow the `Kind` shape, so the
values are `SemanticStructure` through `SemanticUnresolved`. The tags are labels only, and nothing inside
the model depends on them.

**Generic applies to any node that declares its own type parameters.**

Go 1.27 allows type parameters on a method, and `types.Signature.TypeParams()` reports them, so the
generic property is not limited to type nodes: a method such as `func (b Box) Pick[T any]() T` is a
method node with generic recorded. The property comes from the node's own declaration, which is what
the derivation rule allows, and the type parameter itself stays an excluded candidate.

**The module is read as a path segment.**

The printed form reads the package path as a segment of the module rather than as a plain prefix, so a
module named `project1` does not swallow the package path of a module named `project12`. The package
name is still printed as its own part even when it matches the last segment of the path, because the
spec states that the record and the printed form both state it. The form that prints it only when the
two differ is not part of this change.

**A reference points at a node.**

A reference is recorded only when the model records the referent as a node. An alias name is not a
node, so a use of an alias name references the node the alias resolves to, and an alias whose entity
is not a node, such as `type Counter = int`, carries its referent as the attribute value with no
reference. The json tags stay labels and the enum fields keep no marshalling rule, because an enum
carries its name in a `String()` form and nothing inside the model reads the tags.

**The alias exclusion reason is a sentence like the other reasons.**

`ReasonAliasIsName = "alias is a name, not an entity"`, in the style the existing reasons use. The
alias branch reports every package-level alias declaration with it, so the closed list of reasons
gains one entry rather than a special case for aliases.

**The test ground stays one place.**

Every case this change needs is added to the test ground at the repository root, because that is the
package the coverage check reads. `pkg/graph/testdata/fixtures` is left for tests that need a second
package, such as the import-alias test that simulates an imported package, which is outside this
cycle.

**Coverage is a requirement, not a convention.**

The test ground gets one entity per decided semantic value, and `unresolved` is excluded from
that check, because it is produced only when no rule classifies a node. The test ground must
therefore produce no `unresolved` node, which keeps the reserved value from silently swallowing a
case that a rule already covers. The test ground now has the pair that proves the alias rule, a
defined type with a collection body, a functional option type and an unnamed function held by a
package-level variable, and the pair that separates a defined type from an alias. It also holds an
alias of a predeclared identifier, which is the case that shows an alias name with no attribute owner.

## Risks / Trade-offs

- [`structure` covers defined types whose body is not a struct, so a reader may still expect a
  struct body for `type Name string`] → the definition is stated in `docs/node-model.md` and tested
  by cases for a defined type with a basic, a collection and a signature body.
- [A semantic type can change on an edit that changes neither identity nor signature] → a recorded
  graph is stale unless the comparison between two builds includes the semantic type; the rebuild
  trigger has to be extended when the graph change is designed.
- [An alias whose target is not a node, or is a node of another package, has no attribute owner]
  (`type Counter = int`, `type SetAlias = map[string]int`, `type Mirror = foreign.Type`) →
  deferred; the information is reported as an excluded candidate with no owner, so facts such as
  "the package exports a name for `int`" are lost for now.
- [`Moving `Exported` into `NodeProperties` changes the shape of the record`] → accepted as part of this
  change, because the property set is expected to grow and a consumer that reads a collection of nodes
  should read every property from the same place. The printing path and the existing assertion are
  updated in the same change, so no reader is left with one property outside the group.
- [A chain of aliases cannot be expressed] → an attribute reference points at a node, and an alias
  is not a node, so `type A = B` with `type B = C` has nowhere to attach. Deferred.
- [An exported alias of an unexported node is the only case where the name is exported and the
  entity is not] → the attribute has no exported flag in this change, so the fact is lost. Deferred.
- [A reserved `unresolved` value can hide a missing rule] → the coverage check requires that the
  test ground produces no `unresolved` node, so a case that is classified by an existing rule cannot
  fall into the reserved value by accident. The mapping covers every kind of the closed set, so
  `unresolved` can only appear when a kind is added or a rule is missing, and a direct test of the default
  branch keeps that path visible.
- [The alias change breaks the current test expectation] → the alias test in `pkg/graph/build_test.go`
  covers import aliases only, so it stays, but a new case for declared aliases changes the node
  count and needs its own assertion.

## Migration Plan

No stored graphs exist, so the alias change is a behaviour change for `Build` and its tests only.
Order: correct the `Classify` alias branch, then the semantic type field, then the properties part with
`Exported` moved into it, then the test ground, then the documentation and the coverage check. The move
is done in the same step as the properties part, so the printing path and the assertion that reads
exported change once rather than twice. Rollback is not needed beyond reverting the change, because
nothing consumes the semantic field yet.

## Open Questions

- Whether the property set grows beyond generic, exported and entry point, for example zero-size
  struct or explicit-versus-inferred variable type. Answerable later, because no requirement in
  this change depends on it.
- Whether the receiver recorded on a method node keeps the name as written in the source or is
  normalised to the node the receiver resolves to. Answerable later, because the ownership rule is
  already settled and nothing in this change depends on the rendering.
