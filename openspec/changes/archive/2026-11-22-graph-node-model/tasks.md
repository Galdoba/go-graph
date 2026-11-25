# Tasks

## 1. Node record and closed kind set

- [x] 1.1 Define the node record with id, kind, name, package, parent, receiver, exported flag and declaration position, and verify a test builds one node from a package-level declaration and reports all fields.
- [x] 1.2 Restrict node kinds to type, function, variable and constant, and verify a test reports a candidate of any other kind as excluded.

## 2. Enumeration produces exactly the node set

- [x] 2.1 Load one package through `go/packages` and enumerate the package scope, and verify a test with a type, a function, a variable and a constant declaration produces exactly one node per declaration.
- [x] 2.2 Add methods of package-level named types to the node set, and verify a test where `func (s *Struct1) Read() error` yields a node whose parent is the `Struct1` type node and whose receiver is `Struct1`.
- [x] 2.3 Store struct fields and interface method specs as attributes of the owning type node, and verify a test where a field and an interface method spec appear as attributes and never as nodes.
- [x] 2.4 Record import aliases as attributes of the resolved node and create no node for imports or predeclared identifiers, and verify a test where an aliased reference resolves to a foreign node and no package or universe node exists.

## 3. Identity and the declaration-site invariant

- [x] 3.1 Build the canonical identifier from module path, package path, receiver named type and name for function nodes and from the qualified name for other kinds, independently of printed output, and verify a test where two modules with the same package path and entity name produce two distinct identifiers.
- [x] 3.2 Add the same-name scenario test for `func (s1 *Struct1) Read()`, `func (s2 *Struct2) Read()` and `func (s2 *Struct2) ReadLoud(loud bool)` and verify it records three nodes with three distinct identifiers.
- [x] 3.3 Enforce the single declaration site invariant, and verify a test where an entity produced by a lookup rather than a declaration, such as a method promoted through embedding, is reported as excluded rather than recorded.
- [x] 3.4 Treat the declaration position as provenance rather than identity, and verify a test where editing code above a declaration shifts its position while the identifier stays the same.
- [x] 3.5 Record the file path, the declaration position, the line range of the declaration spec and a reference to that declaration, and verify a test that resolves a node back to its declaration and that two variables inside one `var` block keep different ranges.

## 4. Package boundary and documented rules

- [x] 4.2 Document the node definition, the closed kind set and the exclusion rules in `docs/node-model.md`, and verify the documented exclusions match the scenarios in `specs/node-model/spec.md`.

Deferred until the graph itself is implemented, in a separate change:

- 4.1 Represent a package graph as nodes scoped to one package plus references resolved by package path, where a reference to an exported node of another package resolves to that graph and an unexported node is reported as unreachable.
- 4.3 Record graph validity as identity, signature and relations rather than positions, where a comment-only edit leaves the recorded node set unchanged while a rename is reported as requiring a rebuild.

Both rules are documented in `docs/node-model.md` as rules of the model without being enforced, and the requirements that state them are not part of this change's spec delta.

## Workflow follow-up

- Archive the change: the node model is implemented and every scenario in the spec delta has a test.
- Round trip on a real repository, `/home/galdoba/go/src/github.com/Galdoba/lazyam3`: the nodes of every package are built separately, and each record resolves back to the declaration its recorded coordinates point at, matched by package, name and kind. This is the check that a record carries enough information to locate its declaration from the node alone, with no access to the build that produced it. Run through `cmd/graphnodes`, which prints the nodes of one package and the round trip for each record.
- The test ground at the repository root, `testdata`, declares one entity of every node kind and one entity of every kind the model excludes, and a test asserts that coverage.
