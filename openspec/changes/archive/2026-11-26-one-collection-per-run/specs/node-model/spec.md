# Spec Delta

## MODIFIED Requirements

### Requirement: A second pass produces a record for each name a reference names
The system SHALL produce a record for each node a reference of the first pass names in a package the
build did not read as a target, and SHALL produce a record only for such a name. The second pass SHALL run
over every package the run read as a target and SHALL name an entity once across them. The second pass SHALL
NOT enumerate a package it did not read as a target and SHALL NOT follow what that package refers to. A
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

#### Scenario: Two packages of the same run name each other
- **WHEN** a run reads two packages as targets and a reference of one names a node the other declares
- **THEN** the run holds one record for that node, the one read from source, because a record read from
  source states the full set its definition requires and the run names an entity once

#### Scenario: A reference names no node
- **WHEN** a reference names an entity the definition of a node does not cover
- **THEN** the second pass produces no record, because no node is named
