# Tasks

## 1. Alias declarations produce no node

- [x] 1.1 Add an alias branch to `Classify` that reports a package-level alias declaration as
      excluded with `ReasonAliasIsName` ("alias is a name, not an entity"). Verified over the test
      ground through the pairs `type Struct1` with `type Mirror = Struct1` and `type Name string`
      with `type NameAlias = Name`, which produce no node
- [x] 1.2 Record an alias name as an attribute of the node the alias resolves to, carrying the alias
      name and the identifier of the resolved node, and verify in the test ground that
      `type Mirror = Struct1` leaves the attribute on the `Struct1` node and creates no node
- [x] 1.3 Document in `docs/node-model.md` that an alias declaration names an existing entity
      instead of creating one, that the alias attribute is recorded only when the resolved target is
      a node of the same package, and that an alias whose target is not a node, or is a node of
      another package, is recorded as an excluded candidate with no owner, then verify the wording
      matches the test ground. An import name is recorded on the use-site node with a reference to
      the node it resolves to, which the documentation states separately
- [x] 1.4 Keep the existing import-alias assertions in `pkg/graph/build_test.go` unchanged, because
      that test reads a fixture that simulates an imported package, and add the declared-alias case
      as a check over the test ground
- [x] 1.5 Verify that a method written with an alias name as its receiver is attributed to the node
      the alias resolves to, and that a defined type with a non-interface body keeps its own method,
      both in the test ground

## 2. Semantic type of a node

- [x] 2.1 Define the semantic set as `structure`, `interface`, `function`, `method`, `constant`,
      `variable` plus the reserved `unresolved` as the zero value, following the `Kind` pattern with
      a `String()` form. The node field is `SemanticType` with json tag `semantic_type`. The test that
      asserts the set has exactly these values is listed under remaining verification
- [x] 2.2 Compute the semantic type from the entity itself: for a type node the declared body, so an
      interface body gives `interface` and any other declared type gives `structure`, receiver
      presence for a function node, and the object type for a constant and variable node, and verify
      that each decided value is produced by at least one case in the test ground
- [x] 2.3 Add the field to `Node` and keep it out of the identifier. The rebuild check that the
      semantic type changes while the identifier stays is listed under remaining verification
- [x] 2.4 Verify that `type Name string`, `type StringMap map[string]string` and `type Option
      func(*Struct1) error` each produce a node with the semantic type `structure`, and that a method
      name such as `NewStruct1` produces no role value
- [x] 2.5 Record that no test-ground case can produce `unresolved` as the evidence that the mapping is
      complete over the closed kind set, and verify that the test ground produces no `unresolved` node.
      The direct test of the default branch on a constructed object is listed under remaining
      verification
- [x] 2.6 Document the semantic set, the reserved `unresolved` value and the rule that the value is
      derived from the node alone in `docs/node-model.md`, and verify the documented values match the
      set

## 3. Properties

- [x] 3.1 Define `NodeProperties` as a struct with one boolean per property, `Exported`, `Generic` and
      `EntryPoint`, hold it on `Node` as `Properties NodeProperties` with json tags `properties`,
      `exported`, `generic` and `entry_point`, keep each value false when the property does not apply,
      keep the part separate from `SemanticType`, and verify a test shows that two nodes differing only
      by exported keep the same semantic type
- [x] 3.2 Verify that `type Box[T any] struct { Value T }` reports generic as a property with the
      semantic type `structure`, and that `func main()` in package main and `func init()` report the
      entry point property with the semantic type `function`
- [x] 3.3 Verify that a function named `main` in a package that is not `main` is classified as an
      ordinary function with no entry point property, in `pkg/graph/build_test.go`
- [x] 3.4 Document the properties in `docs/node-model.md` with the rule that a property never changes
      the semantic type, name the record fields `SemanticType` and `Properties` of type
      `NodeProperties` with their json tags there, correct the sentence that lists what a record
      reports so exported reads as a property rather than a field, and verify the documented examples
      match the tests
- [x] 3.5 Verify that a method declaring its own type parameters reports generic as a property with
      the semantic type `method`, and that its type parameter stays an excluded candidate
- [x] 3.6 Update `Build` so the property part is filled in one place, and read exported through
      `NodeProperties` in the printing path and in `pkg/graph/node_test.go`, so no reader keeps the old
      field

## 4. Test ground coverage

- [x] 4.1 Add to `testdata` the cases the semantic set needs: a defined type, an alias of that type, an
      alias of a declared type, a defined type with a map underlying, a functional option type, an
      unnamed function held by a package-level variable, `func init()`, an empty package-level
      interface, a method with its own type parameters, and a method written with an alias name as its
      receiver, and verify each new case is produced by the enumeration. The alias of a predeclared
      identifier is listed under remaining verification
- [x] 4.2 Add a coverage check over the test ground that fails when a decided value of the semantic set
      is not exercised, excluding the reserved `unresolved`, following the shape of
      `TestTestgroundCoversEveryNodeKind`
- [x] 4.3 Verify the defined-type and alias pair in the test ground: the defined type produces one node
      with the semantic type `structure`, and the alias produces no node

## 5. Integration check

- [x] 5.1 Run `go test ./pkg/graph` and `go run ./cmd/graphnodes ./testdata` and verify the node count
      matches the test ground, every node reports a semantic type from the set and none reports
      `unresolved`, and the round-trip verification still passes for every recorded node

## 6. Canonical form of the identifier

- [x] 6.1 Write the canonical form so it can be used as the name of the note that holds the node:
      tildes inside the package part, dots between the package, the receiver and the name, the module
      named once, the version written after the module part
- [x] 6.2 Store the name the package declares as a part of its own and print it as its own part, and
      print the package namespace as the package part alone
- [x] 6.3 Read the module as a path segment rather than as a plain prefix, so that a module whose name
      is a prefix of another module name is not read as the same module, and cover it in
      `pkg/graph/identity_test.go`
- [x] 6.4 Decide what the printed form does for a package that reports no module: the printed form
      keeps the name the package declares alone, because it has no module part to carry the path, and
      the package path stays in the record as a component

## 7. Remaining verification

- [x] 7.1 A test that the semantic set has exactly the values documented, in `pkg/graph/node_test.go`
- [x] 7.2 A test that a rebuild changes the semantic type from `structure` to `interface` while the
      identifier stays the same, in `pkg/graph/identity_test.go`
- [x] 7.3 A direct test of the `unresolved` default branch on a constructed object the mapping does not
      cover, in `pkg/graph/node_test.go`
- [x] 7.4 A test that a function named `main` in a package that is not `main` is an ordinary function
      with no entry point property, in `pkg/graph/build_test.go`
- [x] 7.5 The test ground case for an alias of a predeclared identifier, with the use of that name
      recorded as an attribute value that carries no reference to a node
- [x] 7.6 Decide that the enum fields keep no marshalling rule: a consumer reads them as numbers and
      reads the name through the `String()` form, so the json tags stay labels as the documentation
      states

## Workflow follow-up

- The delta at `specs/node-model/spec.md` has been synced into `openspec/specs/node-model/spec.md`, and
  the main spec now states the canonical form, the package name as a part of its own, the use attribute
  kind and the placement of an import-name attribute. Because the requirements are already in the main
  spec, archiving reports them as already existing rather than merging them; archive then only verifies
  the result.
- The archived design of `2026-11-22-graph-node-model` still states that a declared alias is a node. It
  is kept as a historical record; the living documentation states the rule that a declared alias
  produces no node.
- Keep the deferred cases out of the implementation: an alias whose target is not a node, alias chains,
  and the exported flag for an alias name whose target node is unexported. Each one gets a value in the
  semantic set only when it is decided, so the set can be finished before those questions are answered.
