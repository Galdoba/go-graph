# Spec Delta

## Purpose

Defines what counts as a node in a package graph: the closed set of node kinds, the
entities that are excluded from being nodes, the identity rules that keep nodes unique,
and the information a node record must carry so per-package graphs can be joined. The
graph itself, cross-package resolution and graph validity are not part of this change;
they are deferred until the graph is implemented, and are documented in
`docs/node-model.md`.

## ADDED Requirements

### Requirement: Node is an entity created by a package-level declaration
The system SHALL record as a node only a named language entity created by a declaration located at the package level. An entity created inside a declaration body, or produced by a lookup rather than by a declaration, is not a node.

#### Scenario: Package-level declaration produces a node
- **WHEN** a package declares a type, a function, a variable or a constant at package level
- **THEN** each declared entity is recorded as exactly one node

#### Scenario: Entity created inside a declaration is not a node
- **WHEN** a function body declares a variable, a parameter or an anonymous function
- **THEN** those entities are not recorded as nodes

#### Scenario: Aliases do not create nodes
- **WHEN** an entity is reached through a name that is not its declaration name
- **THEN** the alias is recorded as an attribute of the node that the alias resolves to, and not as a separate node

### Requirement: Node kinds are a closed set
The system SHALL use a closed set of node kinds: type, function, variable and constant. No other kind SHALL be produced by the node model.

#### Scenario: Candidate outside the closed set
- **WHEN** an entity is encountered that is not one of the four node kinds
- **THEN** the entity is not recorded as a node

### Requirement: Method nodes carry a receiver
A function node declared with a receiver SHALL be recorded as a node whose receiver property identifies the node of the named type the receiver is based on. A function node without a receiver SHALL have an empty receiver property.

#### Scenario: Method on a named type
- **WHEN** a package declares `func (s *Struct1) Read() error`
- **THEN** one function node is recorded with receiver `Struct1`

#### Scenario: Interface method is not a node
- **WHEN** a package declares an interface type with a method spec
- **THEN** the interface type is recorded as a node and its method spec is recorded as an attribute of that node, not as a node

### Requirement: Excluded entities are preserved as attributes
Information about entities that are not nodes SHALL still be recorded as attributes of the node that owns them, so that later relations can be built from them.

#### Scenario: Struct field
- **WHEN** a package-level type declares a field
- **THEN** the field is recorded as an attribute of the type node, together with the node that the field type refers to

#### Scenario: Predeclared identifier
- **WHEN** a node refers to a predeclared identifier such as `int` or `error`
- **THEN** the reference is recorded as an attribute value, and no node for the predeclared identifier is created

#### Scenario: Package declaration shadows a predeclared identifier
- **WHEN** a package declares its own entity named `min` at package level
- **THEN** that declaration is recorded as a node, and references in the package resolve to that node instead of to the predeclared function

### Requirement: Node identity is unique per declaration
Every node SHALL have an identifier that is unique within the graph, derived from the entity itself rather than from its name alone, and qualified by module path, module version and package path. Nodes with the same name SHALL remain distinguishable when they are different entities.

#### Scenario: Same-named methods on different receivers
- **WHEN** a package declares `func (s1 *Struct1) Read()`, `func (s2 *Struct2) Read()` and `func (s2 *Struct2) ReadLoud(loud bool)`
- **THEN** three nodes are recorded with three distinct identifiers

#### Scenario: Same entity reached from several files
- **WHEN** the same package is analyzed through several files and the same entity is referenced from more than one file
- **THEN** the entity is recorded as one node, with one declaration site

#### Scenario: Identical package path in different modules
- **WHEN** two analyzed modules both contain a package path that declares an entity with the same name
- **THEN** the identifiers differ by module path, and the two entities are recorded as two nodes

#### Scenario: Same package at two module versions
- **WHEN** the same package path is recorded through two module versions
- **THEN** the identifiers differ by version, and the two sets of nodes are distinct

#### Scenario: Position is not part of identity
- **WHEN** an edit above a declaration shifts its declaration position without changing the entity
- **THEN** the identifier is unchanged and matching two builds uses identity rather than position

### Requirement: Node record carries package ownership
Every node SHALL record the package it belongs to and its declaration position. The package SHALL act as a namespace that groups nodes, not as a node itself.

#### Scenario: Record reports the package it belongs to
- **WHEN** a node is built from a package-level declaration
- **THEN** the record reports the module path, the module version and the package path that produced it, and the package carries no node of its own

### Requirement: Every node has exactly one declaration site
Every node SHALL have exactly one declaration site, recorded as the file path, the position of the declared identifier, and the line range of the declaration spec that created it, together with a reference to that declaration. These coordinates SHALL be recorded as provenance of the build, not as part of node identity. A candidate entity with no single declaration site SHALL be rejected rather than recorded as a node.

#### Scenario: Candidate without a declaration site
- **WHEN** an entity is encountered that is not produced by a package-level declaration, such as a method promoted through embedding
- **THEN** it is not recorded as a node and the candidate is reported as excluded

#### Scenario: Method promoted through embedding
- **WHEN** a package-level type embeds another package-level type and a method is reachable only through that embedding
- **THEN** no node is created for the reachable method, the outer type records the embedding as an attribute pointing to the embedded type node, and a use of that method resolves to the node of the type that declares it

#### Scenario: Verifying a node from the record
- **WHEN** a recorded node is opened at its recorded file path and line range
- **THEN** an entity with the recorded name and kind is declared there, and that declaration is the single site of the node

#### Scenario: Nodes created by one group declaration
- **WHEN** several entities are created by the same group declaration, such as two variables inside one `var` block
- **THEN** each node keeps the line range of its own declaration spec, not the range of the group
