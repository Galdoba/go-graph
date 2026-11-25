# Spec Delta

## MODIFIED Requirements

### Requirement: Excluded entities are preserved as attributes
Information about what the node owns SHALL still be recorded as attributes of the node, whether or not the entity is itself a node, so that later relations can be built from them. An attribute SHALL carry its kind, name, a value and the reference to the node it points at, and the kinds the model records are field, method spec, method, embedding, alias and use. A reference SHALL point at a node, so a name the model does not record as a node carries its referent as the attribute value instead.

#### Scenario: Struct field
- **WHEN** a package-level type declares a field
- **THEN** the field is recorded as an attribute of the type node, together with the node that the field type refers to

#### Scenario: Predeclared identifier
- **WHEN** a node refers to a predeclared identifier such as `int` or `error`
- **THEN** the reference is recorded as an attribute value, and no node for the predeclared identifier is created

#### Scenario: Package declaration shadows a predeclared identifier
- **WHEN** a package declares its own entity named `min` at package level
- **THEN** that declaration is recorded as a node, and references in the package resolve to that node instead of to the predeclared function

#### Scenario: A declaration refers to another entity
- **WHEN** a declaration body refers to an entity that is a node
- **THEN** the reference is recorded as a use attribute of the node that owns the declaration

#### Scenario: A use of an alias name
- **WHEN** a declaration refers to a package-level alias of an entity the model records as a node
- **THEN** the use attribute references the node the alias resolves to, because the alias name itself is not a node

#### Scenario: A method declared on a type is both a node and an attribute
- **WHEN** a package-level type declares a method
- **THEN** the method is recorded as a node and as a method attribute of the type node that owns it

## ADDED Requirements

### Requirement: A node records the methods it owns
A structure node SHALL record each method declared on it as an attribute of kind method, carrying the method name, its signature and a reference to the method node, so that a relation can read the method set from the structure alone. The signature SHALL be recorded in the form the model uses for a method spec, without the receiver, so that a signature on a method spec and a signature on a method attribute are comparable as recorded.

#### Scenario: A structure carries its method set
- **WHEN** a package-level type declares a method and the build records the type node
- **THEN** the type node carries a method attribute naming the method, carrying its signature without the receiver, and referencing the method node

#### Scenario: The two sides of a comparison are recorded in the same form
- **WHEN** an interface records the method spec `Read` and a structure records the method `Read` declared on it with the same parameters and results
- **THEN** the two attribute values are the same text, because both are read from the signature without the receiver

#### Scenario: A method set that reaches the type only through embedding is not recorded
- **WHEN** a package-level type embeds another package-level type and a method is reachable only through that embedding
- **THEN** the outer type records the embedding as an attribute and no method attribute, because the reachable method is produced by a lookup rather than by a declaration
