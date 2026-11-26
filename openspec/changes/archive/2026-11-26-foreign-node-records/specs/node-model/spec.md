# Spec Delta

## ADDED Requirements

### Requirement: A second pass produces a record for each name a reference names
The system SHALL produce a record for each node a reference of the first pass names in a package the build
did not read as a target, and SHALL produce a record only for such a name. The second pass SHALL NOT
enumerate a package it did not read as a target and SHALL NOT follow what that package refers to. A
reference from the package the build reads as a target names only a name that package exports, so every
record the second pass produces comes from an exported name.

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
- **THEN** the record is produced for the method the embedded type declares, because the reference names that
  declaration, and no record is produced for the reachable position, which a lookup rather than a declaration
  produces

### Requirement: A record produced from a reference carries what the load states
A record produced by the second pass SHALL carry the identifier, the kind, the semantic type, the receiver, the
parent, the exported property, the signature of the entity itself and the shape the type states. When the load
reports the syntax that declares the entity, the record SHALL also carry the declaration region, the reference to
that declaration and the comment above it. It SHALL NOT carry a use attribute, because no declaration body of
that package was read.

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
- **THEN** the record carries no file, position, line range or comment, and its provenance states that the values
  were read from a name

### Requirement: A qualified reference names the entity the package declares
A reference written through an import SHALL be recorded as a use attribute of the node that owns the declaration,
when the entity it names is a node of the imported package, whatever name the import wrote. A name an import
wrote that is not the name that package declares SHALL still be recorded as an alias attribute, so a record keeps
both the name the import wrote and the entity the reference names.

#### Scenario: A reference through the name a package declares
- **WHEN** a declaration refers to an entity of an imported package through the name that package declares
- **THEN** the reference is recorded as a use attribute of the node that owns the declaration, pointing at the
  node that package's graph owns

#### Scenario: A reference through a name the import wrote
- **WHEN** a declaration refers to an entity of an imported package through a name that is not the name that
  package declares
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

## MODIFIED Requirements

### Requirement: Node record carries package ownership
Every node SHALL record the package it belongs to. For a record produced from a reference, the module part SHALL
be the module the load reported for the package that declares the entity, which carries no version when that load
reported no module. The package SHALL act as a namespace that groups nodes, not as a node itself.

#### Scenario: Record reports the package it belongs to
- **WHEN** a node is built from a package-level declaration
- **THEN** the record reports the module path, the module version and the package path that produced it,
  and the package carries no node of its own

#### Scenario: A record read from a name reports the package it belongs to
- **WHEN** a record produced by the second pass is read
- **THEN** it reports the package that declares the entity, and the module part it reports is the module
  that package was loaded from, which has no version for a package that reports no module

### Requirement: Every node has exactly one declaration site
Every node SHALL have exactly one declaration site, recorded as the file path, the position of the
declared identifier, the line range of the declaration spec that created it, and a reference to that
declaration, when the load reports the syntax that declares it, whether or not the build read that package as a
target. A candidate entity with no single declaration site SHALL be rejected rather than recorded as a node. The
coordinates are provenance of the build, not part of node identity.

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
- **WHEN** the second pass produces a record for a node of a package the build did not read as a target and the
  load reported the syntax that declares it
- **THEN** the record names the file, the position and the line range of that declaration, and its provenance
  states that the values were read from a name rather than from a package the build read as a target

### Requirement: A node records the methods it owns
A structure node SHALL record each method declared on it as an attribute of kind method, carrying the
method name, its signature and a reference to the method node, so that a relation can read the method set
from the structure alone. The signature SHALL be recorded in the form the model uses for a method spec,
without the receiver, so that a signature on a method spec and a signature on a method attribute are
comparable as recorded. For a record produced by the second pass, the method set SHALL hold only the
methods the package declares on that type, so a method reachable only through an embedding is not recorded.

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
- **WHEN** a record produced from a name is read for a type that reaches a method only through an embedded
  type
- **THEN** that type carries the embedding and no method attribute for the reachable method, because the
  method set of a record produced from a name holds only the methods that package declares
