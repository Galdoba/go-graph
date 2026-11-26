# node-model Specification

## Purpose

Defines what counts as a node in a package graph: the closed set of node kinds, the closed semantic
set every node records, the entities that are excluded from being nodes, the identity rules that
keep nodes unique, and the information a node record must carry so per-package graphs can be
joined. The graph itself, cross-package resolution and graph validity are not part of this change;
they are deferred until the graph is implemented, and are documented in `docs/node-model.md`.

## Requirements

### Requirement: Node is an entity created by a package-level declaration
The system SHALL record as a node only a named language entity created by a declaration located at
the package level. An entity created inside a declaration body, or produced by a lookup rather than
by a declaration, is not a node. A package-level alias declaration creates a name for an existing
entity, not an entity, so it does not produce a node.

#### Scenario: Package-level declaration produces a node
- **WHEN** a package declares a type, a function, a variable or a constant at package level
- **THEN** each declared entity is recorded as exactly one node

#### Scenario: Entity created inside a declaration is not a node
- **WHEN** a function body declares a variable, a parameter or an anonymous function
- **THEN** those entities are not recorded as nodes

#### Scenario: Aliases do not create nodes
- **WHEN** an entity is reached through a name that is not its declaration name
- **THEN** the alias is recorded as an attribute of the node that the alias resolves to, and not as
  a separate node

#### Scenario: Alias declaration produces no node
- **WHEN** a package declares `type Mirror = Struct1` where `Struct1` is a package-level declared
  type
- **THEN** `Mirror` is recorded as an alias attribute of the node `Struct1`, and no node is created
  for `Mirror`

#### Scenario: Import name is recorded on the use-site
- **WHEN** a declaration refers to a package through a name that is not the name that package
  declares
- **THEN** the import name is reported as an excluded candidate and recorded as an alias attribute of
  the node that owns the declaration, with a reference to the node the name resolves to

### Requirement: Node kinds are a closed set
The system SHALL use a closed set of node kinds: type, function, variable and constant. No other
kind SHALL be produced by the node model.

#### Scenario: Candidate outside the closed set
- **WHEN** an entity is encountered that is not one of the four node kinds
- **THEN** the entity is not recorded as a node

### Requirement: Method nodes carry a receiver
A function node declared with a receiver SHALL be recorded as a node whose receiver property
identifies the node of the named type the receiver is based on. A function node without a receiver
SHALL have an empty receiver property. The name written as the receiver SHALL NOT also be recorded as
a use attribute of the method node, because the receiver property already names that node.

#### Scenario: Method on a named type
- **WHEN** a package declares `func (s *Struct1) Read() error`
- **THEN** one function node is recorded with receiver `Struct1`

#### Scenario: Interface method is not a node
- **WHEN** a package declares an interface type with a method spec
- **THEN** the interface type is recorded as a node and its method spec is recorded as an attribute
  of that node, not as a node

#### Scenario: The receiver is recorded once
- **WHEN** a package-level type declares a method whose body refers only to its receiver
- **THEN** the method node carries no use attribute for the receiver name, and the relation to the type
  is read from the receiver property

### Requirement: Excluded entities are preserved as attributes
Information about what the node owns SHALL still be recorded as attributes of the node, whether or not
the entity is itself a node, so that later relations can be built from them. An attribute SHALL carry
its kind, name, a value and the reference to the node it points at, and the kinds the model records are
field, method spec, method, embedding, alias and use. A reference SHALL point at a node, so a name the
model does not record as a node carries its referent as the attribute value instead.

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

#### Scenario: A reference to a name created inside a declaration
- **WHEN** a declaration refers to a field, a parameter, a local variable or a type parameter
- **THEN** the use attribute carries its referent as the value and no reference, because that name is
  not a node

#### Scenario: A reference to a method spec
- **WHEN** a declaration uses a method through an interface value
- **THEN** the use attribute carries the method as its value and no reference, because a method spec is
  not a node and has no receiver to tell it apart from a package-level declaration of the same name

#### Scenario: A use of an alias name
- **WHEN** a declaration refers to a package-level alias of an entity the model records as a node
- **THEN** the use attribute references the node the alias resolves to, because the alias name itself
  is not a node

#### Scenario: A method declared on a type is both a node and an attribute
- **WHEN** a package-level type declares a method
- **THEN** the method is recorded as a node and as a method attribute of the type node that owns it

### Requirement: Node identity is unique per declaration
Every node SHALL have an identifier that is unique within the graph, derived from the entity itself
rather than from its name alone, and qualified by module path, module version, package path and the
name the package declares. Nodes with the same name SHALL remain distinguishable when they are
different entities.

#### Scenario: Same-named methods on different receivers
- **WHEN** a package declares `func (s1 *Struct1) Read()`, `func (s2 *Struct2) Read()` and
  `func (s2 *Struct2) ReadLoud(loud bool)`
- **THEN** three nodes are recorded with three distinct identifiers

#### Scenario: Same entity reached from several files
- **WHEN** the same package is analyzed through several files and the same entity is referenced from
  more than one file
- **THEN** the entity is recorded as one node, with one declaration site

#### Scenario: Identical package path in different modules
- **WHEN** two analyzed modules both contain a package path that declares an entity with the same
  name
- **THEN** the identifiers differ by module path, and the two entities are recorded as two nodes

#### Scenario: Same package at two module versions
- **WHEN** the same package path is recorded through two module versions
- **THEN** the identifiers differ by version, and the two sets of nodes are distinct

#### Scenario: Position is not part of identity
- **WHEN** an edit above a declaration shifts its declaration position without changing the entity
- **THEN** the identifier is unchanged and matching two builds uses identity rather than position

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

### Requirement: Package name is recorded as a part of its own
The record SHALL store the name the package declares as a part of its own, beside the package path,
because the directory a package sits in can be named differently from the package it declares.

#### Scenario: Directory name differs from the package name
- **WHEN** the test ground directory `testdata` holds package `main`
- **THEN** the record reports `main` as the name the package declares and `testdata` as the last
  segment of the package path

#### Scenario: Name matches the last path segment
- **WHEN** the name the package declares is the same as the last segment of its package path
- **THEN** the record states both, and the canonical form prints both

### Requirement: Node record carries package ownership
Every node SHALL record the package it belongs to. For a record produced from a reference, the module
part SHALL be the module the load reported for the package that declares the entity, which carries no
version when that load reported no module. The package SHALL act as a namespace that groups nodes, not
as a node itself.

#### Scenario: Record reports the package it belongs to
- **WHEN** a node is built from a package-level declaration
- **THEN** the record reports the module path, the module version and the package path that produced
  it, and the package carries no node of its own

#### Scenario: A record read from a name reports the package it belongs to
- **WHEN** a record produced by the second pass is read
- **THEN** it reports the package that declares the entity, and the module part it reports is the module
  that package was loaded from, which has no version for a package that reports no module

### Requirement: Every node has exactly one declaration site
Every node SHALL have exactly one declaration site, recorded as the file path, the position of the
declared identifier, the line range of the declaration spec that created it, and a reference to that
declaration, when the load reports the syntax that declares it, whether or not the build read that
package as a target. A candidate entity with no single declaration site SHALL be rejected rather than
recorded as a node. The coordinates are provenance of the build, not part of node identity.

#### Scenario: Candidate without a declaration site
- **WHEN** an entity is encountered that is not produced by a package-level declaration, such as a
  method promoted through embedding
- **THEN** it is not recorded as a node and the candidate is reported as excluded

#### Scenario: Method promoted through embedding
- **WHEN** a package-level type embeds another package-level type and a method is reachable only
  through that embedding
- **THEN** no node is created for the reachable method, the outer type records the embedding as an
  attribute pointing to the embedded type node, and a use of that method resolves to the node of the
  type that declares it

#### Scenario: Verifying a node from the record
- **WHEN** a recorded node is opened at its recorded file path, position and line range, together with
  the reference to its declaration
- **THEN** an entity with the recorded name and kind is declared there, and that declaration is the
  single site of the node

#### Scenario: Nodes created by one group declaration
- **WHEN** several entities are created by the same group declaration, such as two variables inside
  one `var` block
- **THEN** each node keeps the line range of its own declaration spec, not the range of the group

#### Scenario: The file is recorded once
- **WHEN** a record is read
- **THEN** the file appears once, in the position, and the line range is read as lines inside that file

#### Scenario: A record produced from a name states the site the load reports
- **WHEN** the second pass produces a record for a node of a package the build did not read as a target
  and the load reported the syntax that declares it
- **THEN** the record names the file, the position and the line range of that declaration, and its
  provenance states that the values were read from a name rather than from a package the build read as a
  target

### Requirement: Node carries a semantic type
Every node SHALL record a semantic type: one value from a closed set that states what the entity is,
not which form of declaration created it. A semantic type SHALL be recorded for every node, and no
value outside the closed set SHALL be produced.

#### Scenario: Two nodes of the same kind with different semantic types
- **WHEN** a package declares a struct type and an interface type
- **THEN** both nodes have kind type, and their semantic types are structure and interface

#### Scenario: Defined type with a basic underlying
- **WHEN** a package declares `type Name string`
- **THEN** the semantic type of the node is structure, because the entity is a defined type and not an
  interface

#### Scenario: Declared type with a collection or a signature body
- **WHEN** a package declares `type Set map[string]int` and `type Option func(*Client) error`
- **THEN** each declaration produces one node, and the semantic type of each is structure, because a
  package-level type declaration creates a new type

#### Scenario: Function node with and without a receiver
- **WHEN** a package declares a method on a package-level type and a package-level function without a
  receiver
- **THEN** the method node has semantic type method and the function node has semantic type function

#### Scenario: Constant and variable nodes
- **WHEN** a package declares a constant and a variable at package level
- **THEN** their semantic types are constant and variable, and being typed or untyped does not change
  the value

### Requirement: The semantic set is closed and extensible
The semantic set SHALL be closed, and it SHALL reserve `unresolved` as the value for a node that no
rule classifies. While the node kind set stays closed, the rules SHALL cover every kind, so
`unresolved` is reported only when the mapping does not cover a kind. A value SHALL be added to the
set only when a case is decided, and a value already decided SHALL NOT change its meaning when a new
value is added.

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
- **THEN** no node reports `unresolved`, because every entity in the test ground is classified by a
  rule already decided

### Requirement: Semantic type is derived from the node alone
A semantic type SHALL be derived from the node's own declaration and from what the language fixes for
every package. A classification that requires another node SHALL NOT be recorded as a semantic type.

#### Scenario: Name-based role is not a semantic type
- **WHEN** a package-level function is named `NewStruct1` and returns a pointer to a declared type
- **THEN** its semantic type is function and no value derived from its name is recorded

#### Scenario: Classification that needs another node
- **WHEN** a defined type node satisfies the method set of an interface node in the same package
- **THEN** the semantic type of the defined type node is unchanged and the satisfied interface is not
  recorded as part of it

### Requirement: Semantic type is not part of node identity
The node identifier SHALL NOT include the semantic type. The semantic type SHALL be recomputed on
rebuild rather than matched between builds.

#### Scenario: Rebuild changes the semantic type without changing identity
- **WHEN** a node's declaration changes from a struct body to an interface body between two builds
  while the declared name stays the same
- **THEN** the identifier is unchanged and the semantic type changes from structure to interface

### Requirement: Alias attribute names the node it resolves to
A name created by a package-level alias declaration SHALL be recorded as an attribute of the node the
alias resolves to, when that target is a node of the package being built, carrying the alias name and
the identifier of the resolved node. An alias SHALL NOT own a method set or a semantic type of its
own.

#### Scenario: Alias to a declared type in the same package
- **WHEN** a package declares `type Struct1 struct` and `type Mirror = Struct1`
- **THEN** `Struct1` is one node, no node is created for `Mirror`, and the node `Struct1` carries an
  alias attribute named `Mirror`

#### Scenario: Method written on a defined type
- **WHEN** a package declares `type Struct1 string` with a method and `type Mirror = Struct1`
- **THEN** the method node is owned by the node `Struct1`, which has semantic type structure, and
  `Mirror` produces no node

#### Scenario: Method written with an alias as receiver
- **WHEN** a package declares `type Mirror = Struct1` and writes a method with `Mirror` as its
  receiver
- **THEN** the method node is owned by the node `Struct1`, because the receiver base type is resolved
  through the alias name, and the alias keeps no method set of its own

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
recorded in the properties part of the node record. The properties defined by this change are
exported, generic and the entry points the language assigns by name, and generic applies to any node
that declares its own type parameters, including a method.

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
excluding the reserved `unresolved`, and it SHALL include the pair that separates a defined type from
an alias.

#### Scenario: A decided value is not exercised
- **WHEN** a decided value of the semantic set is produced by no entity in the test ground
- **THEN** the run against the test ground reports that value as missing

#### Scenario: Defined type and alias pair
- **WHEN** the test ground declares `type Name string` and `type NameAlias = Name`
- **THEN** the defined type produces one node with semantic type structure and the alias produces no
  node

### Requirement: A node records the methods it owns
A structure node SHALL record each method declared on it as an attribute of kind method, carrying the
method name, its signature and a reference to the method node, so that a relation can read the method
set from the structure alone. The signature SHALL be recorded in the form the model uses for a method
spec, without the receiver, so that a signature on a method spec and a signature on a method attribute
are comparable as recorded. For a record produced by the second pass, the method set SHALL hold only the
methods the package declares on that type, so a method reachable only through an embedding is not
recorded.

#### Scenario: A structure carries its method set
- **WHEN** a package-level type declares a method and the build records the type node
- **THEN** the type node carries a method attribute naming the method, carrying its signature without
  the receiver, and referencing the method node

#### Scenario: The two sides of a comparison are recorded in the same form
- **WHEN** an interface records the method spec `Read` and a structure records the method `Read`
  declared on it with the same parameters and results
- **THEN** the two attribute values are the same text, because both are read from the signature without
  the receiver

#### Scenario: A method set that reaches the type only through embedding is not recorded
- **WHEN** a package-level type embeds another package-level type and a method is reachable only
  through that embedding
- **THEN** the outer type records the embedding as an attribute and no method attribute, because the
  reachable method is produced by a lookup rather than by a declaration

#### Scenario: A method reachable through an embedding is not recorded on a foreign type
- **WHEN** a record produced from a name is read for a type that reaches a method only through an
  embedded type
- **THEN** that type carries the embedding and no method attribute for the reachable method, because the
  method set of a record produced from a name holds only the methods that package declares

### Requirement: A second pass produces a record for each name a reference names
The system SHALL produce a record for each node a reference of the first pass names in a package the
build did not read as a target, and SHALL produce a record only for such a name. The second pass SHALL run
over every package the run read as a target, and the run SHALL name an entity once across them, because the
merge keeps one entry per identifier. The second pass SHALL NOT enumerate a package it did not read as a
target and SHALL NOT follow what that package refers to. A reference from the package the build reads as a
target names only a name that package exports, so every record the second pass produces comes from an
exported name.

#### Scenario: A package used through one name
- **WHEN** the build reads one package as a target and one of its nodes references one entity of a package
  it did not read as a target
- **THEN** the second pass produces one record, for the node that reference names, and no record for the
  other entities of that package

#### Scenario: A name that resolves to an entity the package does not export
- **WHEN** a reference names a name that the other package exports and that name is an alias of an entity
  that package does not export
- **THEN** the record is produced for the entity the name resolves to, because a reference names the node
  the alias resolves to, and the record states the exported value that entity has

#### Scenario: The same entity is read by both passes
- **WHEN** a run reads as a target a package that a reference of another package names
- **THEN** one record exists for that node, the one the first pass produced, because a node has one
  identifier and one declaration site

#### Scenario: Two packages of the same run name each other
- **WHEN** a run reads two packages as targets and a reference of one names a node the other declares
- **THEN** the run holds one record for that node, the one read from source, because a record read from
  source states the full set its definition requires and the run names an entity once

#### Scenario: A reference names no node
- **WHEN** a reference names an entity the definition of a node does not cover
- **THEN** the second pass produces no record, because no node is named

### Requirement: Node-ness of an entity declared by another package is decided by the definition
An entity declared by another package SHALL be read as a node of that package's graph when the definition
of a node covers it, including a method declared on a package-level named type. Node-ness SHALL NOT depend
on whether the name can be written from the referencing package, or on whether the name appears in that
package's scope.

#### Scenario: A method called on a type of another package
- **WHEN** a node of the package the build reads as a target calls a method declared on a package-level
  type of a package it did not read as a target
- **THEN** the use attribute carries a reference to the node that package's graph owns, and the second pass
  produces a record for it

#### Scenario: A method reached through an embedding
- **WHEN** a call resolves to a method that a type reaches only through an embedding in another package
- **THEN** the record is produced for the method the embedded type declares, because the reference names
  that declaration, and no record is produced for the reachable position, which a lookup rather than a
  declaration produces

### Requirement: A record produced from a reference carries what the load states
A record produced by the second pass SHALL carry the identifier, the kind, the semantic type, the receiver,
the parent, the exported property, the signature of the entity itself and the shape the type states. When
the load reports the syntax that declares the entity, the record SHALL also carry the declaration region,
the reference to that declaration and the comment above it. It SHALL NOT carry a use attribute, because no
declaration body of that package was read.

#### Scenario: A record read from a name carries the region the load shows
- **WHEN** the second pass produces a record for an entity of a package whose syntax the load reports
- **THEN** the record states the file, the position, the line range and the comment above the declaration

#### Scenario: A record read from a name states no uses
- **WHEN** a record produced by the second pass is read
- **THEN** it carries no use attribute, because a use is a fact of a declaration body the build did not read

#### Scenario: The shape a type states is carried
- **WHEN** the second pass produces a record for an interface of another package
- **THEN** the record carries the method specs that interface states, so a relation that reads them can be
  answered for a pair with that member

#### Scenario: A load that reports no syntax states no region
- **WHEN** the second pass produces a record for an entity of a package whose syntax the load did not report
- **THEN** the record carries no file, position, line range or comment, and its provenance states that the
  values were read from a name

### Requirement: A qualified reference names the entity the package declares
A reference written through an import SHALL be recorded as a use attribute of the node that owns the
declaration, when the entity it names is a node of the imported package, whatever name the import wrote. A
name an import wrote that is not the name that package declares SHALL still be recorded as an alias
attribute, so a record keeps both the name the import wrote and the entity the reference names.

#### Scenario: A reference through the name a package declares
- **WHEN** a declaration refers to an entity of an imported package through the name that package declares
- **THEN** the reference is recorded as a use attribute of the node that owns the declaration, pointing at
  the node that package's graph owns

#### Scenario: A reference through a name the import wrote
- **WHEN** a declaration refers to an entity of an imported package through a name that is not the name
  that package declares
- **THEN** the import name stays an alias attribute of the node that owns the declaration and the entity the
  reference names is recorded as a use attribute

### Requirement: A record states whether its values were read from source or from a name
A record SHALL state whether its information was read from a declaration the build read as a target or from
a name it did not read, so a consumer can tell a value that is absent from a value that was never read. A
record read from source SHALL state the full set its definition requires.

#### Scenario: Absence is not a claim
- **WHEN** a consumer reads a record that states its values came from a name and finds no uses recorded
- **THEN** it reads that absence as not read, and not as a claim that the entity refers to nothing

#### Scenario: A record read from source is complete
- **WHEN** a record produced by the first pass is read
- **THEN** it states that its values were read from a declaration the build read as a target, with the
  region and the attributes that definition requires

### Requirement: A record produced from a reference is identified by the same rules
The identifier of a record produced by the second pass SHALL be built by the rules that build the
identifier of any other node, from the entity itself, qualified by the module part, the module version, the
package path, the name the package declares and, for a method, the receiver named type. The second pass
SHALL NOT keep an identifier space of its own, so two references that name the same entity produce one
record.

#### Scenario: Two references name the same entity
- **WHEN** two declarations of the package the build reads as a target reference the same entity of a
  package it did not read as a target
- **THEN** one record exists for that entity, because both references carry the same identifier

#### Scenario: A foreign method and a foreign package-level name
- **WHEN** a package the build did not read as a target declares an entity named `Read` and a method named
  `Read` on a type named `Thing`, and both are referenced
- **THEN** two records exist, because the identifier of the method carries the receiver

#### Scenario: A foreign record joins to a build that reads its package as a target
- **WHEN** a run produces a record from a reference and a run that reads the same package as a target
  records the same entity
- **THEN** the two records carry the same identifier, so the entity is one node rather than two entries
  that have to be matched by name
