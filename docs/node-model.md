# Node model

A node is a named language entity created by a declaration at the package level. The
enumeration is the definition: package scope gives the package-level declarations, and the
methods declared on package-level named types give the method nodes. There is no filter
step, so an entity that is not a node is reported as an excluded candidate with a reason
instead of being dropped, and its information survives as an attribute of the node that
owns it.

Implementation: `pkg/graph`. `Build` produces the nodes of one package; `Classify` applies
the definition to a single entity.

## Closed set of node kinds

Node kinds are a closed set, and no other kind is produced:

| Kind | What produces it |
| --- | --- |
| `type` | a package-level `type` declaration, including an interface type |
| `function` | a package-level function declaration, and a method declared on a package-level named type |
| `variable` | a package-level `var` declaration |
| `constant` | a package-level `const` declaration |

A candidate whose kind falls outside this set is excluded (`kind outside the closed set`).

## Semantic type

Every node carries a semantic type, which states what the entity is, not which form of
declaration created it. The semantic type is derived from the node's own declaration and
from what the language fixes for every package. A classification that requires another
node is not a semantic type.

The semantic set is closed and extensible:

| Semantic type | What produces it |
| --- | --- |
| `structure` | a package-level type declaration whose body is not an interface (including defined types with basic, collection, or signature underlying types) |
| `interface` | a package-level interface type declaration |
| `function` | a package-level function declaration without a receiver |
| `method` | a method declared on a package-level named type |
| `constant` | a package-level constant declaration |
| `variable` | a package-level variable declaration |
| `unresolved` | reserved for a node that no rule classifies; the zero value, so an unset semantic type is never mistaken for a decided one |

No compiling entity produces `unresolved` today, because the mapping covers every kind of
the closed set. `unresolved` stays as the default so that a case the mapping does not
cover is reported as undecided instead of being given a wrong name.

The semantic type is not part of node identity. The identifier is built from module path,
module version, package path, receiver named type, and name, and must not gain the semantic
type. The semantic type is recomputed on every rebuild, because it can change without
identity or signature changing (for example, when a defined type gains its first method
and stops being an empty interface body).

## Node record

Every node record carries the following fields:

| Field | Type | Description |
| --- | --- | --- |
| `id` | `ID` | unique identifier built from declaration components |
| `kind` | `Kind` | one of the four declaration forms |
| `name` | `string` | the declared identifier |
| `package` | `PackageID` | the package the node belongs to |
| `parent` | `ID` | the node that owns it; the package namespace when empty |
| `receiver` | `string` | the named type a function node is based on, empty for other kinds |
| `semantic_type` | `SemanticType` | one value from the closed semantic set |
| `properties` | `NodeProperties` | named boolean properties that never change the semantic type |
| `pos` | `Position` | the position of the declared identifier |
| `range` | `LineRange` | the lines of the declaration spec that created it, read inside the file of the position |
| `decl` | `ast.Decl` | the declaration that created the node |
| `attributes` | `[]Attribute` | entities that are not nodes |

The json tags name what a consumer reads, and they are labels: the enum fields (`kind`,
`semantic_type` and the attribute `kind`) are read as numbers, because an enum carries its name in a
`String()` form and has no marshalling rule, and `decl` is a reference to a syntax node rather than
data a consumer can read on its own.

### Properties

The `NodeProperties` struct holds properties that never change the semantic type value:

| Property | Type | Description |
| --- | --- | --- |
| `exported` | `bool` | true when the entity is exported (visible outside the package) |
| `generic` | `bool` | true when the entity declares its own type parameters (applies to any node kind) |
| `entry_point` | `bool` | true when the language assigns the name `main` or `init` to a function in package `main` |

Each property reports false when it does not apply to the node. The shape of the properties
part does not change when the set of properties grows.

### Alias declarations

A package-level alias declaration creates a name for an existing entity, not a new entity,
so it does not produce a node. The alias name is recorded as an attribute of the node the
alias resolves to, carrying the alias name and the identifier of the resolved node. This
applies only when the resolved target is a node of the same package.

An alias declaration whose target is not a node of the same package (such as a predeclared
identifier, a collection type, or a node of another package) is recorded as an excluded candidate
with no owner, because the build holds one package and has no node to place the attribute on. An
alias owns no method set and no semantic type of its own.

## Exclusions

| Entity | Node? | Reason | Where it survives |
| --- | --- | --- | --- |
| Entity created inside a declaration body (local variable, parameter, receiver, anonymous function literal) | no | `created inside a declaration body` | attribute of the node that owns the declaration |
| Struct field | no | `part of a declaration spec rather than a declaration` | `field` attribute of the type node, with the node its type refers to |
| Interface method spec | no | `part of a declaration spec rather than a declaration` | `method spec` attribute of the interface type node |
| Embedded field | no | `part of a declaration spec rather than a declaration` | `embeds` attribute pointing at the embedded type node |
| Import name or alias | no | `the package groups nodes and is not a node` | `alias` attribute of the use-site node, with the reference to the node the name resolves to |
| Predeclared identifier (`int`, `error`, `len`) | no | `predeclared identifier, never created by a package-level declaration` | attribute value, with no referent node |
| Package-level declaration that shares a name with a predeclared identifier | yes | - | a node; references in the package resolve to that node |
| Generic type parameter | no | `generic type parameter, not a package-level declaration` | reported as excluded |
| Blank identifier `_` | no | `blank identifier names no entity` | reported as excluded |
| Method reachable only through embedding | no | `produced by a lookup rather than by a declaration` | the outer type keeps the embedding, and a call resolves to the node of the type that declares the method |
| Entity declared by another package | no in this package | `declared by another package and owned by that package graph` | a reference resolves to the node owned by that package |
| The package itself | no | `the package groups nodes and is not a node` | the package namespace in the identifier |
| Declared type alias (`type Mirror = Struct1`) | no | `alias is a name, not an entity` | alias attribute of the resolved node (when target is a same-package node) |

## Attributes

Entities that are not nodes are recorded as attributes of the node that owns them, so a
relation can later be built from them. An attribute carries its kind, a name, a value, and
the identifier of the node it points at. The reference is empty when the referent is not a
node, in which case the value carries it, as the predeclared type `int` does for a field. The
kinds the code records are `field`, `method spec`, `embeds`, `alias` and `use`, where `use` is an
entity a declaration refers to.

The use-site is the node that owns the declaration the reference appears in. A reference
through a name that is not the declaration name is recorded as an alias, not as a second
node for the same entity. A declared alias is recorded on the node the alias resolves to; an
import name is recorded on the use-site node with a reference to the node the name resolves to,
which can be a node of another package graph. A use of an alias name references the node the alias
resolves to, because the alias name is not a node, and an alias that resolves to something the model
does not record as a node carries no reference.

## Identity

Every node has an identifier unique within the graph, built from the components of the
declaration rather than from printed output: module path, module version, package path, the
name the package declares, receiver named type, and name. The canonical form is
`packagepath~packagename[.receiver].name`, where the brackets mark an optional part and are not
printed. The node part of the form is usable as the name of the note that holds the node, with the
package part carried by the folder instead, and the form as a whole cannot contain the characters a
note name cannot contain (`#`, `[`, `]`, `\`, `|`, `/`, `^`): the code rewrites the slash
of a path to a tilde, and the other characters cannot appear in a module path, a version or a Go
identifier. The receiver and the name are separated by dots. A Go identifier contains
neither a dot nor a tilde, so the last segment is always the name and a package path that ends in
a name cannot be read as a receiver. The version is written after the module part, and the
version and the receiver are left out when they have none.

The package name is a part of its own, because the directory a package sits in can be named
differently from the package it declares: the test ground is the directory `testdata` holding
package `main`. The name is stored even when it matches the last segment of the package path, so
the record always states both and the printed form states both too, as
`github.com~user~project1@v1.0.0~pkg1~pkg1.Read` shows.

The module is named once rather than twice: a package path already names the module it
belongs to, so the module part is printed only when the package path does not begin with it, and
the module path stays in the record as a component for a consumer that groups nodes by module,
where it is read as a property of the collection rather than as part of a name. The package path is
read as a path segment of the module rather than as a plain prefix, so a module whose name is a
prefix of another module name is read as a different module and keeps its own part in the printed
form. When the package reports no module, which
is what a load of the standard library reports, the printed package part is the name the package
declares alone: the package path stays in the record and is not printed. A node whose parent is the
package namespace prints as the package part alone, because the namespace carries no name of its
own.

For the test ground the form reads `github.com~Galdoba~go-graph~testdata~main.Box.PickVal` for
the method `PickVal` on `Box`, and `github.com~Galdoba~go-graph~testdata~main` for the namespace
that owns a package-level node.

The identifier keeps entities distinct that a name alone cannot separate:

- the same package path in two modules differs by module path;
- the same package path at two module versions differs by version;
- `func (s1 *Struct1) Read()`, `func (s2 *Struct2) Read()` and
  `func (s2 *Struct2) ReadLoud(loud bool)` are three identifiers, because the receiver
  named type is part of the identifier for function nodes.

The same entity referenced from several files is one node with one declaration site.

## Declaration site

Every node has exactly one declaration site, recorded as the file path, the position of the
declared identifier, the line range of the declaration spec that created it, and a reference
to that declaration. The file is recorded once, in the position, and the range is read as
lines inside it, because a declaration spec and the identifier it declares are always in the
same file. The line range is per declaration spec rather than per group declaration, so two
variables inside one `var` block keep their own ranges.

These coordinates are provenance of the build that produced the record, not part of node
identity: an edit above a declaration moves the position without changing the identifier, so
comparing two builds uses identity. `Verify` reads a record back from the recorded file, position
and line range together with the reference to the declaration, and checks that an entity with the
recorded name and kind is declared there, which is the declaration site invariant seen from the
other side.

A candidate with no single declaration site is rejected rather than recorded, which is what
the excluded reasons are for.

## Package as namespace

The package groups nodes and is never a node itself. A node records the package it belongs
to and its properties. The parent of a package-level node is the package namespace,
which carries no name of its own; the parent of a method node is the node of the named type
its receiver is based on.

## Out of scope for this change

Relations, the graph that holds the nodes, and the rules that decide when a recorded graph
stops being valid are a separate change. This change records nodes of a single package and
the attributes a later relation can be built from. Links between nodes are documented in
`docs/node-links.md`. Two consequences are documented here without being enforced:

- a reference to an exported node of another package resolves to that package's nodes, and
  an unexported node is not reachable from another package;
- a rename makes an identifier stop matching the source and requires a rebuild, while an
  edit that changes no signature and no relation leaves the recorded node set valid.
