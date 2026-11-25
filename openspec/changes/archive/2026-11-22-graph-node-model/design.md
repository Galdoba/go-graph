# Design

## Context

The change covers a single package, with the intent that each package builds its own graph and those graphs are later joined into one. The node model is implemented in `pkg/graph`; the joined graph is a separate change. See proposal.md for motivation.

The approach is shaped by one constraint: `go/types` is the source of truth for entity identity, because it is the only layer that resolves an identifier to a single entity. `go/ast` sees declarations but cannot tell which entity an identifier denotes, so it cannot decide node identity on its own.

## Goals / Non-Goals

**Goals:**

- A definition of a node that is executable: enumeration produces exactly the node set, with no filtering step.
- A node record shape that stays valid once edges are added.
- Identity rules that hold across a joined multi-package graph.

**Non-Goals:**

- The graph itself: joining the node sets of several packages, resolving a reference that points at a node of another package, and deciding when a recorded graph stops being valid. These are tasks 4.1 and 4.3 and belong to the graph change.
- Edge construction, traversal and query APIs.
- SSA, call graphs, and data-flow. Not needed to decide nodes, since nodes come from declarations.
- Package-level dependency aggregation, which needs edges or package-level attributes.

## Decisions

**Definition axis: package-level declaration, not lexical scope.**

`types.Package.Scope()` is a clean container holding exactly type names, constants, variables and functions, but methods and struct fields have no scope at all — `Object.Parent()` is documented as nil for them. Defining a node through scope therefore needs two exceptions. Defining it through "created by a package-level declaration" needs none: a method declaration is a package-level function declaration with a receiver, and a field is part of a type spec rather than a declaration, so the exclusion follows from the rule itself. Alternative considered: keep the scope axis and accept that methods are outside the node set; rejected, because call relations are a stated use case.

**Interface method specs and struct fields are not nodes.**

A method spec inside an interface is not a package-level declaration, and a field is not a declaration under the language's definition of a declaration. Extending the rule to "objects created by a package-level type declaration" would put both back in, so they cannot be separated by derivation; keeping fields out is a deliberate choice recorded here. Consequence: `implements` later becomes a relation between a concrete type node and an interface type node carrying method names, not a relation between two method nodes, which also survives implicit satisfaction where no declaration exists at all.

**The package is a namespace, not a node.**

There is no operation in Go where a package is an operand: `fmt.Println` is qualification, not a reference to an entity. So the package groups nodes and provides their namespace; any package-level graph view is computed by aggregation over nodes. Alternative considered: a package node in the joined graph; deferred until package-level dependency queries are designed.

**Import is not a node.**

Import is a file-level declaration, and an alias is a name, not an entity: the same package can be imported under different aliases in different files of one package. Aliases are stored as attributes of the use-site. Files are treated as a visual division of a package, so they carry no structural role and appear only as the declaration position of a node.

**Identity: receiver-qualified name for functions, qualified name for other kinds.**

`Object.Id()` returns the plain name for an exported object, so `Read` on `Struct1` and `Read` on `Struct2` collide. For function nodes the identifier is built from the package path, the receiver's named type and the name, which yields three distinct identifiers for `func (s1 *Struct1) Read()`, `func (s2 *Struct2) Read()` and `func (s2 *Struct2) ReadLoud(loud bool)`. For other kinds the qualified name is sufficient, and it is already globally unique, which is what makes the joined graph collision-free. Alternatives considered: position-based identity, which breaks on any edit above the declaration, and a hash of the declaration text, which breaks on any body change. Both remain candidates for a separate identity investigation, not for this change.

**Parent chain is derived for methods.**

For package-level nodes the parent is the package namespace. For methods the parent is the node of the named type behind the receiver, because the language model exposes no scope for them. The pointer wrapper in a receiver is a property of the node, not part of the parent link.

**Enumeration is the definition.**

Nodes are produced by walking the package scope plus the methods of package-level named types. That produces exactly the node set defined above, so there is no filter step to keep in sync with the definition, and the invariant "one declaration site per node" becomes a testable check rather than a convention.

**The graph is static and rebuilt on edit, so identifiers do not need to survive renames.**

A graph is valid from the moment it is built until the package is edited in a way that touches a node signature or a relation between nodes. A rename of a node invalidates the graph and forces a rebuild, so identity does not have to be rename-stable, and a content-based identifier is not needed. Two consequences:

- The rebuild trigger is a comparison of node identity, signature and relations, not of positions. A body edit that does not create or remove a reference to a node keeps the graph valid; a body edit that does change a relation invalidates it, because relations are what the graph stores. Comparing two builds is the graph change's work; what this change settles is that the comparison must use identity, since positions move with any edit.
- The declaration position stays as provenance of the build moment, not as part of identity. Any edit above a declaration shifts its stored position without changing the node, so matching two builds must use identity, not position.

**The identifier carries the module path and the module version.**

`github.com/user/project1/pkg1.Read` and `github.com/user/project2/pkg1.Read` are different nodes, and the same package at two versions is two different nodes. The identifier is built from module path, module version, package path, optional receiver type and name. Consequence: a dependency version change invalidates the subgraph of that package even though its source did not change, because identity is defined by the version. This is consistent with the static graph rule, but the rebuild trigger must include version changes, not only edits.

**Predeclared identifiers are not nodes; package declarations that shadow them are.**

Universe objects have no package (`Object.Pkg()` is documented as nil for them) and no declaration position, so they cannot satisfy the node record, and their kinds are outside the closed set. `panic`, `error`, `iota` and `nil` are never nodes. A package that declares its own `min` is a declaration at package level, so `pkg.min` is a node, and the reference to it resolves to that node rather than to the builtin. An alias declared in the package is also a node: `type Counter = int` records `pkg.Counter` as a type node, while `int` stays an attribute value. The rule that separates them is the same as the definition: a node is created by a package-level declaration, and a predeclared identifier never is.

**Package-level variables are nodes, and they are the only kind whose value changes without changing identity.**

`var GlobalCounter int` is recorded as `pkg.GlobalCounter`. Membership is not the open point — it follows from the definition. What it changes for the future relation model is that a variable is read and written, while a type, a constant and a function are only referenced. So the edge model will need at least two relation kinds for variables, and for the static graph rule a change in a variable initializer does not change its signature, so it invalidates the graph only when it adds or removes a relation.

**Physical coordinates stay in this change as provenance.**

The declaration-site invariant cannot be verified without them, so the record keeps the file path, the position of the declared identifier, the line range of the declaration spec and a reference to that declaration. Two limits:

- the range is per declaration spec, not per enclosing group declaration, because one group declaration can create several nodes and a shared range makes verification ambiguous;
- the range covers the declaration of the node, including parts of it that are not nodes, such as a struct body, and it is recomputed on rebuild rather than matched between builds, which is what keeps coordinates outside identity.

Semantic attributes — fields, method specs, embedding, aliases and referenced entities — are recorded here as data a relation is built from, because an excluded entity would otherwise be lost. The relations themselves belong to the next change. Coordinates are kept here because they carry the verification of this change's own invariant, not because they describe a node.

## Risks / Trade-offs

- [Identifier derived from a display-oriented string form can change with formatting or importers] → build the identifier from package path, receiver named type and name as an internal canonical value, independent of printed output.
- [Unexported nodes are unreachable from other package graphs] → intended behaviour, documented in `docs/node-model.md`; enforcing it needs cross-package resolution, which the graph change will add.
- [Node-only model cannot answer "which package depends on X"] → needs package-level attributes for imports that reference no node, including blank and unused imports. Deferred to the edge change.
- [Excluding fields loses field-level relations] → field names and their types are kept as attributes, so the relation type-to-type survives; only field-level traversal is lost.
- [Generic declarations versus their instantiations] → the node model records the declaration; instantiation is deferred, and a later decision to record instantiations would change identity.

## Open Questions

None remain for the node model. The remaining unknowns belong to the graph change, not to this one: which relation kinds a variable needs, how a package-level dependency is recorded when an import references no node (blank and unused imports), whether a generic declaration or each instantiation is the node, and how a reference to another package is resolved and when a recorded graph has to be rebuilt.
