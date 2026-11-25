## ADDED Requirements

### Requirement: Node carries a semantic type
Every node SHALL record a semantic type: one value from a closed set that states what the
entity is, not which form of declaration created it. A semantic type SHALL be recorded for
every node, and no value outside the closed set SHALL be produced.

#### Scenario: Two nodes of the same kind with different semantic types
- **WHEN** a package declares a struct type and an interface type
- **THEN** both nodes have kind type, and their semantic types are structure and interface

#### Scenario: Defined type with a basic underlying
- **WHEN** a package declares `type Name string`
- **THEN** the semantic type of the node is structure, because the entity is a defined type and
  not an interface

#### Scenario: Declared type with a collection or a signature body
- **WHEN** a package declares `type Set map[string]int` and `type Option func(*Client) error`
- **THEN** each declaration produces one node, and the semantic type of each is structure, because
  a package-level type declaration creates a new type

#### Scenario: Function node with and without a receiver
- **WHEN** a package declares a method on a package-level type and a package-level function
  without a receiver
- **THEN** the method node has semantic type method and the function node has semantic type
  function

#### Scenario: Constant and variable nodes
- **WHEN** a package declares a constant and a variable at package level
- **THEN** their semantic types are constant and variable, and being typed or untyped does not
  change the value

### Requirement: The semantic set is closed and extensible
The semantic set SHALL be closed, and it SHALL reserve `unresolved` as the value for a node that
no rule classifies. While the node kind set stays closed, the rules SHALL cover every kind, so
`unresolved` is reported only when the mapping does not cover a kind. A value SHALL be added to the set
only when a case is decided, and a value already decided SHALL NOT change its meaning when a new value
is added.

#### Scenario: Every kind is covered
- **WHEN** a compiling package produces nodes of every kind
- **THEN** each node reports a decided value, so `unresolved` appears only when the mapping is
  incomplete

#### Scenario: No rule classifies the node
- **WHEN** a node is produced and no rule decides its semantic type
- **THEN** the semantic type is `unresolved`

#### Scenario: A case is decided later
- **WHEN** a node that reported `unresolved` is classified by a rule added in a later change
- **THEN** the node reports the new value and the values decided earlier keep their meaning

#### Scenario: The test ground produces no unresolved node
- **WHEN** the build runs against the test ground
- **THEN** no node reports `unresolved`, because every entity in the test ground is classified by
  a rule already decided

### Requirement: Semantic type is derived from the node alone
A semantic type SHALL be derived from the node's own declaration and from what the language
fixes for every package. A classification that requires another node SHALL NOT be recorded as
a semantic type.

#### Scenario: Name-based role is not a semantic type
- **WHEN** a package-level function is named `NewStruct1` and returns a pointer to a declared
  type
- **THEN** its semantic type is function and no value derived from its name is recorded

#### Scenario: Classification that needs another node
- **WHEN** a defined type node satisfies the method set of an interface node in the same package
- **THEN** the semantic type of the defined type node is unchanged and the satisfied interface
  is not recorded as part of it

### Requirement: Semantic type is not part of node identity
The node identifier SHALL NOT include the semantic type. The semantic type SHALL be recomputed
on rebuild rather than matched between builds.

#### Scenario: Rebuild changes the semantic type without changing identity
- **WHEN** a node's declaration changes from a struct body to an interface body between two
  builds while the declared name stays the same
- **THEN** the identifier is unchanged and the semantic type changes from structure to interface

### Requirement: Alias attribute names the node it resolves to
A name created by a package-level alias declaration SHALL be recorded as an attribute of the node
the alias resolves to, when that target is a node of the package being built, carrying the alias name
and the identifier of the resolved node. An alias SHALL NOT own a method set or a semantic type of
its own.

#### Scenario: Alias to a declared type in the same package
- **WHEN** a package declares `type Struct1 struct` and `type Mirror = Struct1`
- **THEN** `Struct1` is one node, no node is created for `Mirror`, and the node `Struct1`
  carries an alias attribute named `Mirror`

#### Scenario: Method written on a defined type
- **WHEN** a package declares `type Struct1 string` with a method and `type Mirror = Struct1`
- **THEN** the method node is owned by the node `Struct1`, which has semantic type structure, and
  `Mirror` produces no node

#### Scenario: Method written with an alias as receiver
- **WHEN** a package declares `type Mirror = Struct1` and writes a method with `Mirror` as its
  receiver
- **THEN** the method node is owned by the node `Struct1`, because the receiver base type is
  resolved through the alias name, and the alias keeps no method set of its own

### Requirement: Node record carries a properties part
The system SHALL record node properties as named values inside one part of the node record, and a
property value SHALL report false when the property does not apply to the node. The shape of that part
SHALL NOT change when the set of properties grows.

#### Scenario: Selecting nodes by property
- **WHEN** a consumer reads a collection of nodes and selects the nodes that have a property
- **THEN** it reads the named value of that property on each record instead of searching a list of
  values

#### Scenario: A node with no property
- **WHEN** a node is declared with no property
- **THEN** every property value in the properties part reports false

### Requirement: Properties do not change the semantic type
A property SHALL NOT change the semantic type value. Every property, exported included, SHALL be
recorded in the properties part of the node record. The properties defined by this change are exported,
generic and the entry points the language assigns by name, and generic applies to any node that declares
its own type parameters, including a method.

#### Scenario: Generic type node
- **WHEN** a package declares `type Box[T any] struct { Value T }`
- **THEN** the semantic type is structure and generic is recorded as a property

#### Scenario: Generic method node
- **WHEN** a package declares a method with its own type parameters, such as
  `func (b Box) Pick[T any]() T`
- **THEN** the semantic type is method and generic is recorded as a property

#### Scenario: Entry point assigned by the language
- **WHEN** package main declares `func main()`
- **THEN** the semantic type is function and the entry point is recorded as a property

#### Scenario: Exported does not change the value
- **WHEN** two nodes differ only in whether they are exported
- **THEN** they have the same semantic type, and the difference is recorded as the exported property in
  the properties part of the record

### Requirement: Test ground exercises every decided semantic type
The test ground SHALL declare at least one entity for every decided value of the semantic set,
excluding the reserved `unresolved`, and it SHALL include the pair that separates a defined type
from an alias.

#### Scenario: A decided value is not exercised
- **WHEN** a decided value of the semantic set is produced by no entity in the test ground
- **THEN** the run against the test ground reports that value as missing

#### Scenario: Defined type and alias pair
- **WHEN** the test ground declares `type Name string` and `type NameAlias = Name`
- **THEN** the defined type produces one node with semantic type structure and the alias produces
  no node

## MODIFIED Requirements

### Requirement: Node is an entity created by a package-level declaration
The system SHALL record as a node only a named language entity created by a declaration located
at the package level. An entity created inside a declaration body, or produced by a lookup
rather than by a declaration, is not a node. A package-level alias declaration creates a name
for an existing entity, not an entity, so it does not produce a node.

#### Scenario: Package-level declaration produces a node
- **WHEN** a package declares a type, a function, a variable or a constant at package level
- **THEN** each declared entity is recorded as exactly one node

#### Scenario: Entity created inside a declaration is not a node
- **WHEN** a function body declares a variable, a parameter or an anonymous function
- **THEN** those entities are not recorded as nodes

#### Scenario: Aliases do not create nodes
- **WHEN** an entity is reached through a name that is not its declaration name
- **THEN** the alias is recorded as an attribute of the node that the alias resolves to, and not
  as a separate node

#### Scenario: Alias declaration produces no node
- **WHEN** a package declares `type Mirror = Struct1` where `Struct1` is a package-level declared
  type
- **THEN** `Mirror` is recorded as an alias attribute of the node `Struct1`, and no node is
  created for `Mirror`

#### Scenario: Import name is recorded on the use-site
- **WHEN** a declaration refers to a package through a name that is not the name that package
  declares
- **THEN** the import name is reported as an excluded candidate and recorded as an alias attribute of
  the node that owns the declaration, with a reference to the node the name resolves to

### Requirement: Canonical form is usable as a note name
The canonical form of an identifier SHALL be usable as the name of the note that holds the node, so
the characters a note name cannot contain SHALL NOT appear in it. The path part SHALL be written with
tildes instead of slashes, and the package, the receiver and the name SHALL be separated so that the
last segment is always the name. The package path SHALL be read as a path segment of the module
rather than as a plain prefix.

#### Scenario: The module is named once
- **WHEN** the package path already begins with the module path
- **THEN** the canonical form prints the module part once, with the version written after it when the
  module has one

#### Scenario: A module name that is a prefix of another module name
- **WHEN** a package path begins with text that matches a module name without being a path segment of
  that module
- **THEN** the package path keeps its own segment, so the two modules are read as two modules and the
  printed form states both parts

#### Scenario: A package reports no module
- **WHEN** a package is loaded without module information, as the standard library is reported
- **THEN** the canonical form is the name the package declares, and the package path is read from the
  record rather than from the printed form

#### Scenario: A package path that ends in a name
- **WHEN** the package part of a canonical form ends in a word that could be read as a receiver
- **THEN** the last segment is read as the name, and the two identifiers stay distinct

#### Scenario: The package namespace has no name
- **WHEN** a node's parent is the package namespace
- **THEN** the package part alone is the canonical form and does not end in a separator

### Requirement: Excluded entities are preserved as attributes
Information about entities that are not nodes SHALL still be recorded as attributes of the node that
owns them, so that later relations can be built from them. An attribute SHALL carry its kind, a name,
a value and the reference to the node it points at, and the kinds the model records are field, method
spec, embedding, alias and use. A reference SHALL point at a node, so a name that the model does not
record as a node carries its referent as the attribute value instead of as a reference.

#### Scenario: Struct field
- **WHEN** a package-level type declares a field
- **THEN** the field is recorded as an attribute of the type node, together with the node that the
  field type refers to

#### Scenario: Predeclared identifier
- **WHEN** a node refers to a predeclared identifier such as `int` or `error`
- **THEN** the reference is recorded as an attribute value, and no node for the predeclared
  identifier is created

#### Scenario: Package declaration shadows a predeclared identifier
- **WHEN** a package declares its own entity named `min` at package level
- **THEN** that declaration is recorded as a node, and references in the package resolve to that
  node instead of to the predeclared function

#### Scenario: A declaration refers to another entity
- **WHEN** a declaration body refers to an entity that is a node
- **THEN** the reference is recorded as a use attribute of the node that owns the declaration

#### Scenario: A use of an alias name
- **WHEN** a declaration refers to a package-level alias of an entity the model records as a node
- **THEN** the use attribute references the node the alias resolves to, because the alias name itself
  is not a node
