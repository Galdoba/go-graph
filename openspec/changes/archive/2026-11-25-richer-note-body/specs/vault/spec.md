# Spec Delta

## MODIFIED Requirements

### Requirement: The note states what its node is
The note SHALL state the kind, the semantic type and the properties of the node it holds, in a form that can be read without the build that produced the file, because the file name states the semantic type only. The body SHALL include a `# Name` heading, a `## Code` section with the source code snippet and an optional doc comment, and a `## Related` section listing incoming connections as Obsidian wikilinks.

#### Scenario: Generic entry point reads as such
- **WHEN** a note is read for a node that reports `generic` and `entry_point` as properties
- **THEN** both properties are stated in the file

#### Scenario: Struct node with doc comment and related method
- **WHEN** a note is written for a struct node that has a declaration doc comment and an incoming parent-method link
- **THEN** the `## Code` section includes the doc comment above the code snippet, and the `## Related` section lists the method as a wikilink

#### Scenario: Node with no incoming links
- **WHEN** a note is written for a node with no incoming connections
- **THEN** the `## Related` section is present and contains the text "None detected"
