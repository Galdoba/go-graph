# Spec Delta

## REMOVED Requirements

### Requirement: The file name states the semantic type of the node
**Reason**: The name no longer carries a classification. The semantic type is not part of node identity and is recomputed on every rebuild, so the same node can be written under two different names by two builds that disagree only about its classification, which a link text that later points at that node cannot rely on.
**Migration**: Replaced by the requirement added below, which names a note with the node part of the identifier. The semantic type is read from the frontmatter, where the note already states it and where a value is read as a member of its closed set.

### Requirement: A check invoked on demand confirms or refutes a note
**Reason**: The rule itself does not change, but one reason reads a part of the name that no longer exists. A name has no prefix to compare with the semantic type the file records, so the reason becomes that the name is not the node part of the recorded identifier.
**Migration**: The rule continues under the requirement restated in the ADDED section, with the same unreadable case and the name reason in its new form.

## ADDED Requirements

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

### Requirement: The name depends on the identifier alone
The name the mapping produces for a node SHALL be determined by the node part of its identifier and by nothing else, including no value a rule decides about the node. No state is carried from one build to the next, and nothing monitors a name against the code: the name is the one this build produces from the current source, and a correspondence that a code change has invalidated is stated only by a check invoked for that file.

#### Scenario: A rule deciding differently does not change the name
- **WHEN** the same node record is rendered while a rule decides a different semantic type for it
- **THEN** the name is unchanged, because no part of the name comes from a rule's decision

#### Scenario: The code changes the node part
- **WHEN** the source is changed so the node part of the identifier differs from the name of the marked file that holds the node
- **THEN** this build produces a different name for that node, and no earlier name is carried over

### Requirement: A check confirms a note against the identifier its record states
When invoked for one file, the system SHALL confirm or refute that the file matches the node it records, using the file and the code at that moment and without the build that produced the file. A file that cannot be read as frontmatter SHALL be reported as unreadable rather than as matching.

#### Scenario: File matches the node it records
- **WHEN** a check is invoked for a file whose location is the location its recorded package path maps to, whose recorded region still declares a node with the recorded name and kind, and whose name is the node part of the identifier the file records
- **THEN** the file is confirmed as matching that node

#### Scenario: Recorded region no longer declares the node
- **WHEN** the code region recorded in the frontmatter no longer declares a node with the recorded name and kind
- **THEN** the check refutes the file and states that reason

#### Scenario: Location does not match the recorded package path
- **WHEN** a check is invoked for a file at a location the package path written inside it does not map to
- **THEN** the check refutes the file and states that reason

#### Scenario: Name is not the node part of the recorded identifier
- **WHEN** the name of a file is not the receiver and the name the identifier it records states
- **THEN** the check refutes the file and states that the name does not match the record the file carries

#### Scenario: Check runs on a file it cannot read
- **WHEN** a check is invoked for a file that cannot be read as frontmatter
- **THEN** the system reports it as unreadable
