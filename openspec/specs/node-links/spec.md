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
A pair SHALL be answered only when both of its members are records that a rule can read. A pair whose member has no record SHALL NOT be answered, and the domain SHALL grow by making more records readable rather than by moving a pair across a package boundary. A pair that is not in the domain SHALL become answerable once its member becomes a record, without changing any rule.

#### Scenario: A member of the pair has no record

- **WHEN** an attribute of a structure refers to an entity from a package the build never produced, so that entity has no record
- **THEN** no pair is answered for it, because the rule cannot read that member

#### Scenario: The domain grows by making records readable

- **WHEN** the entity that had no record becomes a record in the set the rules read
- **THEN** that pair can be answered, and the answers for pairs already in the domain are unchanged

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
The rule set SHALL cover the relation from a structure to the method that records it as its parent, and the relation from a structure to a node one of its attributes refers to. No other relation type SHALL be answered by this change. A method whose receiver is written with an alias name SHALL be answered through the node the alias resolves to.

#### Scenario: A structure to the method that records it as its parent

- **WHEN** the pair `Thing` and the method `Read` declared on it is answered for the test ground
- **THEN** the outcome is that the connection exists

#### Scenario: A structure to the node an attribute of it refers to

- **WHEN** the pair `Struct1` and the node `Reader`, recorded as an embedded attribute of `Struct1`, is answered for the test ground
- **THEN** the outcome is that the connection exists

#### Scenario: A method written with an alias as its receiver

- **WHEN** the pair `Struct1` and the method `DescribeAlias`, written with the alias name `Mirror` as its receiver, is answered for the test ground
- **THEN** the outcome is that the connection exists, because the receiver resolves to the node `Struct1`

#### Scenario: A pair outside the covered relations

- **WHEN** a pair is answered for a relation type that neither of the two rules covers
- **THEN** the outcome is that the connection does not exist, because no rule answers for that relation type

### Requirement: Every rule is exercised by the test ground
The test ground SHALL produce a pair for every rule in the rule set, so that each rule is answered on real code. A rule answered yes by no entity of the test ground SHALL be reported as not covered.

#### Scenario: A rule has no pair in the test ground

- **WHEN** a rule in the rule set is answered yes by no pair the test ground produces
- **THEN** the run against the test ground reports that rule as not covered

### Requirement: The pairs a build answers are reproducible
The set of pairs answered as existing for a source SHALL be identical for the same source regardless of the order in which its attributes were recorded, and SHALL be ordered by the pair so that the same source produces the same list.

#### Scenario: Attributes recorded in a different order

- **WHEN** the same source is answered for the same set of targets with its attributes in a different order
- **THEN** the list of pairs answered as existing is the same, because the outcome is read from the pair and not from the order of the attributes
