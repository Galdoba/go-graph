# vault Specification

## Purpose

Defines how a node record becomes a note file in the vault: where the file is placed, how it is named, which parts of the record the file carries, which files the rebuild action removes, and how a check invoked on demand confirms or refutes that a file matches the node it records.

## Requirements

### Requirement: A note file is written for each node the graph built
The system SHALL write one note file for each record the graph produced, including a record produced from a reference to a package the build did not read as a target. Such a record belongs to the graph of the package that declares the entity, so its file is written in the folder that package path maps to. A note produced from a reference holds the information that record can state, and no field the record does not state is written.

#### Scenario: Node in a package the graph built
- **WHEN** the graph produces a node for a package loaded as a target
- **THEN** exactly one note file exists for that node

#### Scenario: Reference points at a package the graph never built
- **WHEN** a recorded reference resolves to a node of a package the graph never read as a target, as the test ground reference from `main` to `strings.ToLower`
- **THEN** one note file exists for that node, in the folder that package path maps to, because the second pass produced a record for it

#### Scenario: A record the second pass did not produce
- **WHEN** a reference names an entity for which the second pass produced no record
- **THEN** no file is written for it, because no record exists for that entity

### Requirement: The location mirrors the package path in the record
The path from the vault root to a note file SHALL be the package path the node records, so the location alone identifies the package the node belongs to. When the record states a module version, that version SHALL be a part of the location, so two records that differ only by version do not map to the same folder.

#### Scenario: Package under the vault root
- **WHEN** a note is written for a node of `github.com/Galdoba/go-graph/testdata/second`
- **THEN** it is placed under `<vault>/github.com/Galdoba/go-graph/testdata/second`

#### Scenario: Directory name differs from the declared package name
- **WHEN** a note is written for a node of the test ground package whose directory is `testdata` and whose declared name is `main`
- **THEN** the location uses the recorded package path, and the declared name is read from the file

#### Scenario: A record that states a module version
- **WHEN** a note is written for a record produced from a reference to a package loaded from a module that has a version
- **THEN** the location states that version, so the folder the record maps to is not the folder the same package path maps to at another version

### Requirement: The frontmatter carries the identifier parts the location cannot carry
The frontmatter SHALL carry the module part, the module version, the package path and the name the package declares, so the location together with the frontmatter gives the identifier of the node the note represents. The recorded package path and the recorded version are the claim a location is read against, so a check compares the folder a file sits in against the parts the record states. A note written for a record produced from a reference carries the same parts, because the identifier rules are the same for it.

#### Scenario: Identifier recovered from the file alone
- **WHEN** a note is read from its location and its frontmatter without the build that produced it
- **THEN** the identifier assembled from both is the identifier of the node that produced the file

#### Scenario: Same node name in two packages
- **WHEN** two nodes in two packages carry the same node name and so the same file name
- **THEN** the two notes are still read as two nodes, because their identifiers differ

#### Scenario: A value written in a note is read as a member of its closed set
- **WHEN** the kind or the semantic type written in a note is read back
- **THEN** it is read as a value of the closed set that wrote it, and a value outside the set is not read as one, so a note can be navigated back to the node it records without the build that produced it

#### Scenario: A foreign note states the parts its record carries
- **WHEN** a note written for a record produced from a reference is read
- **THEN** the frontmatter states the module part, the version, the package path and the name the package declares, and it states no part the record does not carry

### Requirement: The frontmatter carries the code region the node was declared at
The frontmatter SHALL carry the file, the line, the column and the declaration range recorded in the node, so a note names the code region it came from without the build that produced it. A record that states no region, because the load reported no syntax for the package that declares the entity, writes none of those fields, and a consumer reading them reads nothing rather than an empty value.

#### Scenario: Note names its declaration site
- **WHEN** a note is read for a record that states a region
- **THEN** the recorded region names the file and the lines the declaration spans

#### Scenario: Note for a record produced from a name carries the region the load showed
- **WHEN** a note is read for a record produced from a reference, and the load reported the syntax that declares the entity
- **THEN** the frontmatter carries the file, the line, the column and the declaration range of that declaration

#### Scenario: Note for a record that states no region
- **WHEN** a note is read for a record produced from a reference and the load reported no syntax for the package that declares the entity
- **THEN** the frontmatter carries the identifier parts and the values the record states, and carries no file, line, column, start line or end line

### Requirement: The note states what its node is
The note SHALL state the kind, the semantic type and the properties of the node it holds, in a form that can be read without the build that produced the file, because the file name states the semantic type only. The body SHALL include a `# Name` heading, a `## Code` section with the source code snippet and an optional doc comment, and a `## Related` section listing connections as Obsidian wikilinks. A rule of this requirement is not answered against a record that does not state the value it asks for, so a note for a record that states no region has no code snippet.

#### Scenario: Generic entry point reads as such
- **WHEN** a note is read for a node that reports `generic` and `entry_point` as properties
- **THEN** both properties are stated in the file

#### Scenario: Struct node with doc comment and related method
- **WHEN** a note is written for a struct node that has a declaration doc comment and an incoming parent-method link
- **THEN** the `## Code` section includes the doc comment above the code snippet, and the `## Related` section lists the method as a wikilink

#### Scenario: A node that is a member of no pair
- **WHEN** a note is written for a record that no pair names
- **THEN** the `## Related` section is present and contains the text "None detected"

#### Scenario: A foreign note is read by these rules
- **WHEN** a note is written for a record produced from a reference and the load reported the syntax that declares the entity
- **THEN** the file carries the region, the code snippet and the comment above the declaration, and its `## Related` section lists the links that point at that record

#### Scenario: A link from a note the build read points at a foreign note
- **WHEN** a node of the package the build read as a target has a connection to a record produced from a reference
- **THEN** its `## Related` section lists that record as a wikilink pointing at the folder the referenced package path maps to

### Requirement: Every generated note carries an ownership marker
The frontmatter of every note the graph writes SHALL carry a single boolean marker stating that the file was generated by the graph, so a rebuild can tell a file it owns from a file a person wrote. Only a marker that reads as true claims graph ownership.

#### Scenario: File written by the graph
- **WHEN** the graph writes a note for a node
- **THEN** the frontmatter of that file carries the marker as true

#### Scenario: File without the marker is never removed
- **WHEN** a file in a package folder carries no marker
- **THEN** the system does not remove it

#### Scenario: A marker that does not claim graph ownership
- **WHEN** a file in a package folder carries a marker whose value is not true
- **THEN** the file is not treated as generated by the graph and is not removed

#### Scenario: Frontmatter cannot be read
- **WHEN** a file that should be a note cannot be read as frontmatter
- **THEN** the system does not remove it and reports it as unreadable

### Requirement: Generation is deterministic
The output of the mapping SHALL be reproducible: the same node record produces the same location, the same file name and the same file content, including the order of the frontmatter fields.

#### Scenario: Same node rendered twice
- **WHEN** a node is rendered twice with no change to the source
- **THEN** the two renderings are identical

### Requirement: The rebuild action removes the files it owns before writing them
The rebuild action SHALL remove the marked files the records of the current run map to before writing them, and SHALL state which files it removed. For the package it built, that is every marked file in the folder that package path maps to. For a record produced from a reference, the folder it maps to can hold files another package's run wrote, so the action SHALL remove only the marked files the records of this run map to there and SHALL NOT remove a marked file it did not produce.

#### Scenario: Node no longer exists
- **WHEN** a node present in an earlier build is absent from the current build
- **THEN** the marked file at the location that node mapped to is removed

#### Scenario: The name changes for a node that still exists
- **WHEN** a node is reclassified between builds so the name the mapping produces for it differs from the name of the marked file that holds it
- **THEN** the marked file is removed and the note is written under the name the current record produces, so no file is left for a node that no longer maps to that name

#### Scenario: File outside the files the graph owns
- **WHEN** a file sits in the package folder without the marker
- **THEN** that file is left untouched

#### Scenario: Two runs write in the same folder
- **WHEN** a run rebuilds a package and its records produced from reference map to a folder that another package's run also writes in
- **THEN** only the marked files this run produced are removed there, and a marked file written by the other run is left in place

### Requirement: The file name carries the node part of the identifier
The file name SHALL be the node part of the identifier the record carries: the receiver named type when the node has one, then the name the package declares. The name SHALL NOT state a classification, because the semantic type is not part of node identity. A Go identifier cannot contain a dot, so a name carries at most one dot, and that dot separates the receiver from the node name; a name that carries more than one dot cannot come from two identifiers and SHALL NOT be read as a node part.

#### Scenario: Package-level node
- **WHEN** a note is written for the type node `Thing` of the second test ground
- **THEN** the file is named `Thing.md` in the folder that its recorded package path maps to

#### Scenario: Method node keeps its receiver
- **WHEN** a note is written for the method `Read` declared on `Thing`
- **THEN** the file is named `Thing.Read.md`

#### Scenario: A package-level entity and a method with the same name
- **WHEN** the main test ground declares `Read` at package level and a method named `Read` on the type `Thing` it declares, which the second ground also declares
- **THEN** two files exist in the folder of that package, `Read.md` and `Thing.Read.md`, and the file named `Thing.md` sits in the folder of each ground package, because a name made of the node name alone would need one file for two nodes in the same folder

#### Scenario: A name is read back as the node part of the identifier
- **WHEN** a note name is read back without the build that produced it
- **THEN** it gives the receiver and the name the package declares, and a node name that contains an underscore stays intact, since no part of a name is matched against the semantic set, while a name that carries more than one dot is refused because no pair of identifiers can produce one

### Requirement: A note maps to one file per identifier
A note written for a record produced from a reference SHALL read back to the identifier that record states, as any other note does: the location and the file name together give it, with the file name carrying the node part. The system SHALL NOT write two notes that map to one file, and SHALL NOT keep a name space of its own for notes produced from references.

#### Scenario: A foreign note is read back to its identifier

- **WHEN** a note written for a record produced from a reference is read from its location and its name
- **THEN** the identifier assembled from both is the identifier the record states, so the note is read as the node of the package the record names rather than as a note of the package that referenced it

#### Scenario: The same package path at two module versions

- **WHEN** two records name the same package and the same entity but come from two module versions
- **THEN** they map to two different files, because a file that holds two identifiers cannot be read back to one node

### Requirement: The name depends on the identifier alone
The name the mapping produces for a node SHALL be determined by the node part of its identifier and by nothing else, including no value a rule decides about the node. No state is carried from one build to the next, and nothing monitors a name against the code: the name is the one this build produces from the current source, and a correspondence that a code change has invalidated is stated only by a check invoked for that file.

#### Scenario: A rule deciding differently does not change the name
- **WHEN** the same node record is rendered while a rule decides a different semantic type for it
- **THEN** the name is unchanged, because no part of the name comes from a rule's decision

#### Scenario: The code changes the node part
- **WHEN** the source is changed so the node part of the identifier differs from the name of the marked file that holds the node
- **THEN** this build produces a different name for that node, and no earlier name is carried over

### Requirement: A check confirms a note against the identifier its record states
When invoked for one file, the system SHALL confirm or refute that the file matches the node it records, using the file and the code at that moment and without the build that produced the file. A file that cannot be read as frontmatter SHALL be reported as unreadable rather than as matching. A record that states a region is answered at that region, and a record that states none is answered against the entity the package declares. The reason that a package cannot be read SHALL be its own reason, not the reason that a recorded region no longer declares the node. The location is read against the package part the record states, including a version the record states.

#### Scenario: File matches the node it records
- **WHEN** a check is invoked for a file whose location is the location its recorded package path maps to, whose recorded region still declares a node with the recorded name and kind, and whose name is the node part of the identifier the file records
- **THEN** the file is confirmed as matching that node

#### Scenario: Recorded region no longer declares the node
- **WHEN** the code region recorded in the frontmatter no longer declares a node with the recorded name and kind
- **THEN** the check refutes the file and states that reason

#### Scenario: A note for a record that states no region is confirmed
- **WHEN** a check is invoked for a file whose record states no region, and the package the record names still declares an entity with the recorded name, kind and exported value
- **THEN** the file is confirmed as matching that node

#### Scenario: The package no longer declares the entity
- **WHEN** a check is invoked for a file whose record states no region, and the package the record names no longer declares an entity with the recorded name and kind
- **THEN** the check refutes the file and states that reason

#### Scenario: The package a record names cannot be read
- **WHEN** a check is invoked for a note whose record names a package the check cannot read
- **THEN** the check refutes the file and states that the package cannot be read, and states no reason that needs a region the record does not carry

#### Scenario: Location does not match the recorded package path
- **WHEN** a check is invoked for a file at a location the package path written inside it does not map to
- **THEN** the check refutes the file and states that reason

#### Scenario: The folder states another version than the record
- **WHEN** a check is invoked for a note whose record states a module version and whose folder states a different version
- **THEN** the check refutes the file, because the location does not match the package part the record states

#### Scenario: Name is not the node part of the recorded identifier
- **WHEN** the name of a file is not the receiver and the name the identifier it records states
- **THEN** the check refutes the file and states that the name does not match the record the file carries

#### Scenario: Check runs on a file it cannot read
- **WHEN** a check is invoked for a file that cannot be read as frontmatter
- **THEN** the system reports it as unreadable
