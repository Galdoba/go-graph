// Package graph holds the node model of a package graph: what counts as a node, the
// closed set of node kinds, the entities excluded from being nodes, and the identity
// rules that keep nodes unique once per-package graphs are joined.
//
// A node is a named language entity created by a declaration at the package level.
// The kinds are a closed set: type, function, variable and constant. Entities that
// are not nodes - struct fields, interface method specs, generic type parameters,
// function parameters, local variables, anonymous function literals, predeclared
// identifiers and import aliases - are recorded as attributes of the node that owns
// them. The package acts as a namespace that groups nodes, not as a node itself.
//
// The rules and their exclusions are documented in docs/node-model.md. Edges and the
// graph that holds the nodes are a separate change, so an excluded entity is stored
// as data a later relation can be built from rather than as a graph vertex.
package graph
