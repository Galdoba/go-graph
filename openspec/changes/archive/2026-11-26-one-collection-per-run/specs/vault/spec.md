# Spec Delta

## ADDED Requirements

### Requirement: The Related section is answered from the records of the whole run
The `## Related` section SHALL be answered from the records the run produced for every package it read as a
target, rather than from the records of one package. A note written in a folder that several packages of one
run write SHALL show what the run saw of its node as a whole.

#### Scenario: A note of a package referenced by another package of the same run
- **WHEN** a run read two packages as targets, one declares a node the other references, and the record
  that carries the reference is read from source
- **THEN** its `## Related` section lists the connection it points at, which the records of the package it
  names alone could not show, and the note of that entity is written from the record read from source

#### Scenario: A record read from source lists the connections it points at
- **WHEN** the section is answered for a record the first pass produced
- **THEN** it lists the pairs that have that record as their source, because a record read from source states
  the uses its declaration carries

#### Scenario: A record read from a name lists the connections that point at it
- **WHEN** the section is answered for a record produced from a reference
- **THEN** it lists the pairs that name that record as their target, because such a record states no uses and
  can be a member of a pair only as a target

## MODIFIED Requirements

### Requirement: The rebuild action removes the files it owns before writing them
The rebuild action SHALL remove the marked files the records of the current run map to before writing them,
and SHALL state which files it removed. For a package the run read as a target, that is every marked file in
the folder that package path maps to. For a record produced from a reference, the folder it maps to can hold
files another run wrote, so the action SHALL remove only the marked files the records of this run map to there
and SHALL NOT remove a marked file it did not produce.

#### Scenario: Node no longer exists
- **WHEN** a node present in an earlier build is absent from the current build
- **THEN** the marked file at the location that node mapped to is removed

#### Scenario: The name changes for a node that still exists
- **WHEN** a node is reclassified between builds so the name the mapping produces for it differs from the name
  of the marked file that holds it
- **THEN** the marked file is removed and the note is written under the name the current record produces, so no
  file is left for a node that no longer maps to that name

#### Scenario: File outside the files the graph owns
- **WHEN** a file sits in the package folder without the marker
- **THEN** that file is left untouched

#### Scenario: Two runs write in the same folder
- **WHEN** a run rebuilds a package and its records produced from reference map to a folder that another
  package's run also writes in
- **THEN** only the marked files this run produced are removed there, and a marked file written by the other run
  is left in place

#### Scenario: A run that read two packages as targets
- **WHEN** a run reads two packages as targets and writes the notes the records of the run map to
- **THEN** every marked file is removed in the folder of each of those packages, and nothing is removed in a
  folder that only records produced from reference map to
