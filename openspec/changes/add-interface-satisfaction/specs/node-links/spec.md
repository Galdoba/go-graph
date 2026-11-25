# Spec Delta

## ADDED Requirements

### Requirement: A connection from an interface reaches the struct that satisfies it
The relation from an interface to a structure SHALL be answered from the method specs recorded as attributes of the interface and the methods recorded as attributes of the structure. A method spec SHALL match a method by its name, its receiver and its signature, not by its name alone, and a signature SHALL be read as recorded, without the receiver. A structure with no recorded method for one method spec of the interface SHALL NOT produce a connection.

#### Scenario: A structure satisfies an interface

- **WHEN** the pair `Reader` and `Base` is answered for the test ground, where `Reader` records the method spec `Read` and `Base` owns the method `Read` with receiver `Base`
- **THEN** the outcome for that pair is that the connection exists

#### Scenario: A structure does not satisfy an interface

- **WHEN** the pair `Reader` and `Struct1` is answered for the test ground, where `Reader` records the method spec `Read` and `Struct1` owns no method named `Read`
- **THEN** the outcome for that pair is that the connection does not exist

#### Scenario: An interface has no method specs

- **WHEN** the pair `EmptyInterface` and `Base` is answered for the test ground, where `EmptyInterface` records no method spec
- **THEN** the outcome for that pair is that the connection does not exist, because a rule with no method specs to check answers no

#### Scenario: A method with a pointer receiver satisfies an interface

- **WHEN** a structure with a pointer receiver method is answered as the target of a pair whose source is an interface whose method spec has the same name
- **THEN** the outcome for that pair is that the connection exists, because the language counts a pointer receiver method in the method set of both the value and the pointer

#### Scenario: Two methods with the same name and different signatures are different methods

- **WHEN** the pair `Reader` and `Thing` is answered for the test ground, where `Reader` records the method spec `Read` with signature `func() error` and `Thing` records the method `Read` with signature `func() int`
- **THEN** the outcome for that pair is that the connection does not exist, because the signatures are not the same

#### Scenario: Method attributes are not read as referents

- **WHEN** a structure that records its methods as attributes is answered for the relation from a structure to a node one of its attributes refers to
- **THEN** that rule answers no for a method attribute, because a method attribute names the method set of the structure rather than a referent

## MODIFIED Requirements

### Requirement: The rule set covers the two relations decided here
The rule set SHALL cover the relation from a structure to the method that records it as its parent, the relation from a structure to a node one of its attributes refers to, and the relation from an interface to a structure that satisfies it. No other relation type SHALL be answered by this change. A method whose receiver is written with an alias name SHALL be answered through the node the alias resolves to.

#### Scenario: A structure to the method that records it as its parent

- **WHEN** the pair `Thing` and the method `Read` declared on it is answered for the test ground
- **THEN** the outcome is that the connection exists

#### Scenario: A structure to the node an attribute of it refers to

- **WHEN** the pair `Struct1` and the node `Reader`, recorded as an embedded attribute of `Struct1`, is answered for the test ground
- **THEN** the outcome is that the connection exists

#### Scenario: An interface to a structure that satisfies it

- **WHEN** the pair `Reader` and `Base` is answered for the test ground
- **THEN** the outcome is that the connection exists, because `Base` owns the method `Read` that `Reader` records as a method spec

#### Scenario: A method written with an alias as its receiver

- **WHEN** the pair `Struct1` and the method `DescribeAlias`, written with the alias name `Mirror` as its receiver, is answered for the test ground
- **THEN** the outcome is that the connection exists, because the receiver resolves to the node `Struct1`

#### Scenario: A pair outside the covered relations

- **WHEN** a pair is answered for a relation type that none of the three rules covers
- **THEN** the outcome is that the connection does not exist, because no rule answers for that relation type
