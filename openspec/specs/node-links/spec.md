# node-links Specification

## Purpose

Defines what a connection between two nodes is, what a rule may look at when it answers for a pair, the outcomes an answer for a pair can have, and how the answers of rules combine into that outcome.

## Requirements

### Requirement: A connection is a fact about one pair
A connection SHALL be a binary fact about one pair of nodes, decided for the pair rather than for a single node, and SHALL be counted in the outgoing direction only. The answer for one pair SHALL NOT be taken as the answer for the reverse pair, and a pair answered yes by more than one rule SHALL be one fact rather than several. A pair SHALL have two different records as its members, and a record considered as both members of a pair SHALL NOT produce an answer.

#### Scenario: Only one direction carries the connection

- **WHEN** the pair `Struct1` to `Reader` is answered for the test ground, where `Reader` is recorded as an attribute of `Struct1`
- **THEN** the outcome for that pair is that the connection exists, and the outcome for the reverse pair is that the connection does not exist

#### Scenario: Several rules answer yes for the same pair

- **WHEN** two rules answer yes for the same pair
- **THEN** one fact exists for that pair, because the rule that answered is not part of the fact

#### Scenario: A record is not a pair with itself

- **WHEN** the same record is considered as both members of a pair
- **THEN** no answer is produced, because the members of a pair are two different records

### Requirement: The domain of the relation is the pairs a rule can read
A pair SHALL be answered only when both of its members are records that a rule can read. The records a rule can read SHALL be the records the run produced, covering every package the run read as a target, rather than the records of one package. A pair whose member has no record SHALL NOT be answered, and the domain SHALL grow by making more records readable rather than by moving a pair across a package boundary. A pair that is not in the domain SHALL become answerable once its member becomes a record, without changing any rule.

#### Scenario: A member of the pair has no record

- **WHEN** an attribute of a structure refers to an entity from a package the build never produced, so that entity has no record
- **THEN** no pair is answered for it, because the rule cannot read that member

#### Scenario: The domain grows by making records readable

- **WHEN** the entity that had no record becomes a record in the set the rules read
- **THEN** that pair can be answered, and the answers for pairs already in the domain are unchanged

#### Scenario: Two members come from two packages the same run read as targets

- **WHEN** a run read two packages as targets and a reference of a record of one names a record of the other
- **THEN** the pair is answered, because both members are records of the collection the run produced

### Requirement: A rule answers for one pair from the pair alone
A rule SHALL state one relation type and SHALL answer for one pair using only what the two members of the pair carry. A rule that needs information outside the pair SHALL NOT be part of the rule set. The answer of a rule for a pair SHALL be the same each time it is answered for that pair on the same source.

#### Scenario: A rule reads the record of the target as a member of the pair

- **WHEN** a rule answers for a pair whose target is a method that records its source as its parent
- **THEN** the rule answers from the record of the method, which is a member of the pair, and not from a lookup of other nodes

#### Scenario: The same pair is answered twice

- **WHEN** the same pair is answered twice with no change to the source
- **THEN** the two answers are the same outcome

### Requirement: The answer for a pair has three outcomes
The answer for a pair SHALL be one of three outcomes: the connection exists, the connection does not exist, or it cannot be decided. A rule that cannot answer SHALL NOT answer no. When the outcome is that the connection cannot be decided, the answer SHALL name the rule that could not answer, and an outcome of exists or does not exist SHALL NOT name a rule.

#### Scenario: A rule cannot answer for an undecided value

- **WHEN** a rule that reads the semantic type of a member meets a member whose semantic type is `unresolved`
- **THEN** the outcome is that the connection cannot be decided, and the answer names the rule that could not answer

#### Scenario: A rule answers no

- **WHEN** a rule that requires a structure as the source answers for a pair whose source is a variable node
- **THEN** the outcome is that the connection does not exist, and no rule is named

#### Scenario: A rule answers yes

- **WHEN** a rule answers yes for a pair
- **THEN** the outcome is that the connection exists, and the rule that answered is not named

### Requirement: Rules combine by the outcome and not by the rule
The rules for a pair SHALL be consulted in a fixed order, and the first rule that answers yes SHALL decide the answer and stop the consultation. When no rule answered yes, the outcome SHALL be that the connection does not exist when every rule that could answer answered no, and that the connection cannot be decided when at least one rule could not answer.

#### Scenario: The first yes stops the consultation

- **WHEN** the first rule consulted answers yes for a pair while a later rule could not answer for it
- **THEN** the outcome is that the connection exists, the later rule is not consulted, and no rule is reported

#### Scenario: Every rule that could answer said no

- **WHEN** no rule answered yes for a pair and every rule that could answer answered no
- **THEN** the outcome is that the connection does not exist

#### Scenario: No rule answered yes and one could not answer

- **WHEN** no rule answered yes for a pair and one rule could not answer for it
- **THEN** the outcome is that the connection cannot be decided, and that rule is named

### Requirement: The answer names the last rule that could not answer
The outcome that says the connection cannot be decided SHALL name the last rule in the fixed order that could not answer for the pair, because the outcome is the same for any of them and the answer is not read back.

#### Scenario: Several rules could not answer

- **WHEN** no rule answered yes for a pair and more than one rule could not answer for it
- **THEN** the outcome is that the connection cannot be decided, and the answer names the last rule in the fixed order that could not answer

### Requirement: The set of connections grows only by adding
A rule SHALL be able to turn an outcome of does not exist into exists when a semantic value gets decided, and SHALL NOT be able to turn an outcome of exists into does not exist. A value already decided SHALL keep its meaning when a new value is added to the semantic set.

#### Scenario: A value is decided by a rule added later

- **WHEN** a pair whose source reported `unresolved` is answered after a rule decides that value
- **THEN** the outcome for that pair becomes that the connection exists, and the outcome of a pair answered by an earlier rule is unchanged

#### Scenario: The source stops producing what the rule reads

- **WHEN** a pair answered as existing is answered again after the source no longer carries what the rule reads
- **THEN** the outcome for that pair is that the connection does not exist, because the outcome is read from the source at that moment, while the growth rule limits only what a rule added later can change

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

### Requirement: Every rule is exercised by the test ground
The test ground SHALL produce a pair for every rule in the rule set, so that each rule is answered on real code. A rule answered yes by no entity of the test ground SHALL be reported as not covered. The test ground SHALL include a declaration that references a node of a package the build reads from type information only, so that a rule is answered on a pair whose members come from two packages.

#### Scenario: A rule has no pair in the test ground

- **WHEN** a rule in the rule set is answered yes by no pair the test ground produces
- **THEN** the run against the test ground reports that rule as not covered

#### Scenario: A pair whose members come from two packages

- **WHEN** the test ground declares a package that calls a method declared on a type of another package the build reads from type information only
- **THEN** the rule that reads references is answered yes for that pair, and the run reports it as covered

### Requirement: The pairs a build answers are reproducible
The set of pairs answered as existing for a source SHALL be identical for the same source regardless of the order in which its attributes were recorded, and SHALL be ordered by the pair so that the same source produces the same list. The collection the run produces SHALL be ordered by the identifier each record states, so the same packages answered in a different order produce the same collection.

#### Scenario: Attributes recorded in a different order

- **WHEN** the same source is answered for the same set of targets with its attributes in a different order
- **THEN** the list of pairs answered as existing is the same, because the outcome is read from the pair and not from the order of the attributes

#### Scenario: The same packages read in a different order

- **WHEN** a run produces records for the same set of packages in a different order
- **THEN** the collection is the same, because it is ordered by the identifier each record states

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

- **WHEN** the pair `Adder` and `Struct1` is answered for the test ground, where `Adder` records the method spec `Add5` and `Struct1` records the method `Add5`, written with a pointer receiver
- **THEN** the outcome for that pair is that the connection exists, because the receiver is recorded as the named type the receiver is based on

#### Scenario: A method reachable only through an embedding does not satisfy an interface

- **WHEN** the pair `Reader` and `Outer` is answered for the test ground, where `Outer` records `Base` as an embedding attribute and records no method
- **THEN** the outcome for that pair is that the connection does not exist, because the model records the embedding and not the method that is reachable only through it

#### Scenario: Two methods with the same name and different signatures are different methods

- **WHEN** the pair `Reader` and `Thing` is answered for the test ground, where `Reader` records the method spec `Read` with signature `func() error` and `Thing` records the method `Read` with signature `func() int`
- **THEN** the outcome for that pair is that the connection does not exist, because the signatures are not the same

#### Scenario: Method attributes are not read as referents

- **WHEN** a structure that records its methods as attributes is answered for the relation from a structure to a node one of its attributes refers to
- **THEN** that rule answers no for a method attribute, because a method attribute names the method set of the structure rather than a referent

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

### Requirement: The domain holds a pair with a record produced from a reference
A pair SHALL be answered when one of its members is a record the second pass produced, because that member is a record the rules can read, and a pair whose members come from two packages SHALL be answered by the same rules as a pair whose members come from one package. The exported filter SHALL reach only records produced from a reference, so a pair inside the package the build read as a target is answered whatever the exported value of its target is.

#### Scenario: An unexported source reaches an exported node of another package

- **WHEN** a node of the package the build read as a target is not exported and carries a use whose reference names a record the second pass produced
- **THEN** the pair is answered as existing, because the filter reaches only records produced from a reference

#### Scenario: An unexported target inside the package

- **WHEN** a source of the package the build read as a target carries a use that names an unexported node of that same package
- **THEN** the pair is answered as it is answered today, because no record of that package was produced from a reference

#### Scenario: A member comes from a package the second pass produced no record for

- **WHEN** a reference names a node for which no record was produced
- **THEN** no pair is answered for it, because a member of a pair must be a record

#### Scenario: A pair that needs two records from the same foreign package

- **WHEN** the second pass produced a record for a method declared on a type of another package, and the code that produced that record named no record for the type the method is declared on
- **THEN** the relation from that type to its method is not answered, because the pair would need a record for the type as its other member

### Requirement: A rule cannot answer for a value a record does not carry
A rule that needs a value a record produced from a reference does not carry SHALL report that it cannot answer for that pair, and SHALL NOT answer no, because the record does not claim that the value is absent. A rule that reads only what a record produced from a reference can carry SHALL answer for that pair as it does for any other record.

#### Scenario: The rule that reads uses meets a record that states no uses

- **WHEN** a pair whose source is a record produced from a reference is answered by the rule that reads the uses the source carries
- **THEN** that rule reports that it cannot answer, and the outcome for the pair names it when no rule answered yes

#### Scenario: A rule that reads shape answers for a record produced from a reference

- **WHEN** a record produced from a reference carries the method specs of the interface it states, and a structure of the package the build read as a target records a method for each of them
- **THEN** the pair is answered as existing, because the rule reads only what the two members carry
