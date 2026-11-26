# Spec Delta

## MODIFIED Requirements

### Requirement: The domain of the relation is the pairs a rule can read
A pair SHALL be answered only when both of its members are records that a rule can read. The records a rule
can read SHALL be the records the run produced, covering every package the run read as a target, rather than
the records of one package. A pair whose member has no record SHALL NOT be answered, and the domain SHALL grow
by making more records readable rather than by moving a pair across a package boundary. A pair that is not in
the domain SHALL become answerable once its member becomes a record, without changing any rule.

#### Scenario: A member of the pair has no record

- **WHEN** an attribute of a structure refers to an entity from a package the build never produced, so that
  entity has no record
- **THEN** no pair is answered for it, because the rule cannot read that member

#### Scenario: The domain grows by making records readable

- **WHEN** the entity that had no record becomes a record in the set the rules read
- **THEN** that pair can be answered, and the answers for pairs already in the domain are unchanged

#### Scenario: Two members come from two packages the same run read as targets

- **WHEN** a run read two packages as targets and a reference of a record of one names a record of the other
- **THEN** the pair is answered, because both members are records of the collection the run produced

### Requirement: The pairs a build answers are reproducible
The set of pairs answered as existing for a source SHALL be identical for the same source regardless of the
order in which its attributes were recorded, and SHALL be ordered by the pair so that the same source produces
the same list. The collection the run produces SHALL be ordered by the identifier each record states, so the
same packages answered in a different order produce the same collection.

#### Scenario: Attributes recorded in a different order

- **WHEN** the same source is answered for the same set of targets with its attributes in a different order
- **THEN** the list of pairs answered as existing is the same, because the outcome is read from the pair and not from the order of the attributes

#### Scenario: The same packages read in a different order

- **WHEN** a run produces records for the same set of packages in a different order
- **THEN** the collection is the same, because it is ordered by the identifier each record states
