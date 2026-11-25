# Proposal

## Why

A graph of Go code cannot be built until the unit of the graph is settled. Today "node" is ambiguous: a struct method, an interface method, an import alias and a package name all look like candidates, and they produce different results for identity and for collisions. Every later decision — edges, traversal, storage, incremental rebuild — depends on this definition, so it must be fixed before any edge work starts.

This change settles the node definition for a single package and the shape of a node record, which is also what makes the planned per-package graphs joinable into one graph.

## What Changes

- Define a node as a `types.Object` created by a package-level declaration.
- Fix a closed node kind set: `TypeName`, `Func`, `Var`, `Const`. No other kind is permitted.
- State exclusion rules as part of the definition: struct fields, interface method specs, generic type parameters, function parameters and local variables, anonymous function literals, predeclared identifiers, import aliases. The package itself is a namespace that groups nodes, not a node.
- Require node identity to be unique per declaration site, including three methods with the same name on different receivers.
- Define the node record shape and which information is kept as an attribute of a node rather than as a node (method sets, fields, aliases).
- Record an excluded entity as an attribute of the node that owns it, so a later relation can be built from it without making it a graph vertex.
- Require each node to carry the package path, so a per-package graph can reference a node owned by another package graph.
- Deferred to the change that implements the graph itself: cross-package reference resolution and reachability of unexported nodes, graph validity across an edit, the edge model, dependency produced by blank or unused imports, and generic instantiation.

## Capabilities

### New Capabilities

- `node-model`: what counts as a node in a package graph, the closed set of node kinds, exclusion rules, node identity and uniqueness, and the attributes a node record must carry.

### Modified Capabilities

None — this project has no existing specs.

## Impact

- Implemented in `pkg/graph`: the node record, the closed kind set, the reasons an entity is excluded, the identifier, the attribute model, the enumeration of one package, and the check that a record resolves back to its declaration. `cmd/graphnodes` prints the nodes of one package and runs that round trip, and `testdata` is the test ground that declares one entity of every node kind.
- The definition relies on `go/types` through `go/packages` from `golang.org/x/tools`, already a project dependency. It does not require SSA, which stays out of scope until edges are designed.
- Establishes the contract that later edge work must satisfy: an edge can only point at a node, so excluded entities must be represented as attributes or as package-level facts, not as graph vertices.
- The rules are documented in `docs/node-model.md`, and every scenario in the spec delta is covered by a test. What is left for the graph change is cross-package resolution for unexported nodes and generic instantiation.
