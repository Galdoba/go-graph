## MODIFIED Requirements

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
