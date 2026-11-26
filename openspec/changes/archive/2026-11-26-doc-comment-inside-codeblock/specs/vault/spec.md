# Spec Delta

## MODIFIED Requirements

### Requirement: The note states what its node is
The note SHALL state the kind, the semantic type and the properties of the node it holds, in a form that can be read without the build that produced the file, because the file name states the semantic type only. The body SHALL include a `# Name` heading, a `## Code` section with the source code snippet and an optional doc comment, and a `## Related` section listing connections as Obsidian wikilinks. A rule of this requirement is not answered against a record that does not state the value it asks for, so a note for a record that states no region has no code snippet.

#### Scenario: Generic entry point reads as such
- **WHEN** a note is read for a node that reports `generic` and `entry_point` as properties
- **THEN** both properties are stated in the file

#### Scenario: Struct node with doc comment and related method
- **WHEN** a note is written for a struct node that has a declaration doc comment and an incoming parent-method link
- **THEN** the `## Code` section includes the doc comment inside the code block, before the code snippet, and the `## Related` section lists the method as a wikilink

#### Scenario: A node that is a member of no pair
- **WHEN** a note is written for a record that no pair names
- **THEN** the `## Related` section is present and contains the text "None detected"

#### Scenario: A foreign note is read by these rules
- **WHEN** a note is written for a record produced from a reference and the load reported the syntax that declares the entity
- **THEN** the file carries the region, the code snippet and the doc comment inside the code block, and its `## Related` section lists the links that point at that record

#### Scenario: A link from a note the build read points at a foreign note
- **WHEN** a node of the package the build read as a target has a connection to a record produced from a reference
- **THEN** its `## Related` section lists that record as a wikilink pointing at the folder the referenced package path maps to
