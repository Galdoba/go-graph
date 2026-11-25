## ADDED Requirements

### Requirement: A connection from a node reaches the node a reference names
A rule SHALL answer for a pair when the source carries a use attribute whose reference names the target, whatever kind the source is. The target SHALL be read as the node the reference names, with no filter on its kind. A reference that names no node SHALL NOT answer for a pair, because an entity the model does not record as a node is not a member of a pair. A rule that reads a reference SHALL NOT read a semantic value.

#### Scenario: A function reaches the node its parameter names

- **WHEN** the pair `describe` and `Reader` is answered for the test ground, where `describe` declares `Reader` as the type of a parameter
- **THEN** the outcome is that the connection exists

#### Scenario: A function reaches a node its body names

- **WHEN** the pair `NewStruct1` and `Struct1` is answered for the test ground, where `Struct1` appears only inside the body of `NewStruct1`
- **THEN** the outcome is that the connection exists, because the reference is recorded on the node that owns the declaration

#### Scenario: A variable reaches the node its value names

- **WHEN** the pair `zeroStruct` and `Struct1` is answered for the test ground
- **THEN** the outcome is that the connection exists

#### Scenario: A target of any kind is a member of the pair

- **WHEN** the pair `main` and `Five` is answered for the test ground, where `Five` is a constant node
- **THEN** the outcome is that the connection exists, because the rule does not filter the target by kind

#### Scenario: An entity with no record answers nothing

- **WHEN** a function whose parameters name only predeclared identifiers is answered for the test ground, such as `min` against any node
- **THEN** the outcome is that the connection does not exist, because a predeclared identifier has no reference to read

#### Scenario: A reference that names no node answers no

- **WHEN** the pair `describe` and the package-level function `Read` is answered for the test ground, where the body of `describe` uses `Read` through the interface `Reader`
- **THEN** the outcome is that the connection does not exist, because the reference names a method spec rather than a node and carries its referent as a value

#### Scenario: An attribute that records a name rather than a referent answers no

- **WHEN** a node that carries an alias attribute or a method attribute is answered for this relation against the node that attribute points at
- **THEN** the outcome is that the connection does not exist, because only a use attribute records an entity a declaration refers to

#### Scenario: The rule answers for an undecided member

- **WHEN** a pair whose source reports `unresolved` carries a reference to the target
- **THEN** the outcome is that the connection exists and no rule is named, because the rule reads a reference and not a semantic value

### Requirement: The receiver names no reference
A name written as a receiver SHALL NOT be recorded as a reference, because the receiver is recorded as the receiver property of the node. A method SHALL NOT answer for the node its receiver names.

#### Scenario: A method does not reach the type it is declared on

- **WHEN** the pair `Read`, declared on `Base`, and `Base` is answered for the test ground
- **THEN** the outcome is that the connection does not exist, because `Base` is recorded as the receiver rather than as a reference

#### Scenario: The reverse pair still carries the method

- **WHEN** the pair `Base` and the method `Read` declared on it is answered for the test ground
- **THEN** the outcome is that the connection exists, through the method recorded as an attribute of `Base`

## MODIFIED Requirements

### Requirement: The rule set covers the two relations decided here
The rule set SHALL cover the relation from a structure to the method that records it as its parent, the relation from a structure to a node one of its attributes refers to, the relation from an interface to a structure that satisfies it, and the relation from a node to a node a reference it carries names. No other relation type SHALL be answered by this change. A method whose receiver is written with an alias name SHALL be answered through the node the alias resolves to.

#### Scenario: A structure to the method that records it as its parent

- **WHEN** the pair `Thing` and the method `Read` declared on it is answered for the test ground
- **THEN** the outcome is that the connection exists

#### Scenario: A structure to the node an attribute of it refers to

- **WHEN** the pair `Struct1` and the node `Reader`, recorded as an embedded attribute of `Struct1`, is answered for the test ground
- **THEN** the outcome is that the connection exists

#### Scenario: An interface to a structure that satisfies it

- **WHEN** the pair `Reader` and `Base` is answered for the test ground
- **THEN** the outcome is that the connection exists, because `Base` owns the method `Read` that `Reader` records as a method spec

#### Scenario: A node to a node a reference it carries names

- **WHEN** the pair `describe` and `Reader` is answered for the test ground
- **THEN** the outcome is that the connection exists, because `describe` records `Reader` as the type of its parameter

#### Scenario: A method written with an alias as its receiver

- **WHEN** the pair `Struct1` and the method `DescribeAlias`, written with the alias name `Mirror` as its receiver, is answered for the test ground
- **THEN** the outcome is that the connection exists, because the receiver resolves to the node `Struct1`

#### Scenario: A pair outside the covered relations

- **WHEN** a pair is answered for a relation type that none of the rules covers
- **THEN** the outcome is that the connection does not exist, because no rule answers for that relation type
