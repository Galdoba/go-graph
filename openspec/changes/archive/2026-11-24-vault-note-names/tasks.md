# Tasks

## 1. The name form

- [x] 1.1 Build the file name from the node part of the identifier, the receiver when the node has one and then the name the package declares, and verify a test asserts the second ground node `Thing` lands at `<vault>/github.com/Galdoba/go-graph/testdata/second/Thing.md`, its method at `Thing.Read.md`, and the ground node `Box` at `Box.md`, with the folder still built from the recorded package path.
- [x] 1.2 Declare `Thing` in the main ground, which the second ground already declares, and declare `Read` at package level in the main ground together with a method named `Read` on that `Thing`, so the same-name cases exist as real code. Verify a test asserts that the two grounds give `Thing.md` in two different folders, and that the main folder holds both `Read.md` and `Thing.Read.md`. Resolve the two nodes named `Read` by the canonical form, since a lookup by name alone returns the first match and cannot tell a package-level node from a method.
- [x] 1.3 Read a name back as the receiver and the node name without matching any part against a closed set, remove the semantic-type part from `NameParts` and the gate that matches a prefix against the set, and verify a test asserts a node name containing an underscore reads whole, a dot separates the receiver from the name, a name with an empty part is not read as a produced name, and the parts read back to the identifier they came from. `Check` keeps reading the semantic type from the frontmatter through `graph.ParseSemanticType`.
- [x] 1.4 Document the name form and the read-back rule in `docs/vault.md`, align the wording in `docs/node-model.md` that states the canonical form is the name of the note with the requirement that it is usable as one, and verify the example paths in the documentation match the assertions of the mapping tests.

## 2. The check reason for the name

- [x] 2.1 Repurpose the name reason so the check compares the file name against the node part of the identifier the record read from the file states, and verify a test asserts a file whose name is not that part is refuted, that the other reasons still hold, and that a value the frontmatter states outside its closed set is still read as not a member of it.
- [x] 2.2 Document the reasons in `docs/vault.md` and verify the documented reason matches the value the test asserts and that no reason about a name prefix remains.

## 3. The rebuild action

- [x] 3.1 Verify a test that renders a type node, changes its declaration body from a struct to an interface, runs the action again, and asserts one file for that node under the same name as before. The assertion is about the mapping producing the same name for the same node part; the file is not kept from the earlier build, since the action removes the marked file before writing it.
- [x] 3.2 Verify a test that a marked file named in the form the earlier mapping produced is removed and written again under the node part of the identifier, so the change of form happens through the action and not by hand, and so that a name this build no longer produces is replaced rather than left behind.

## 4. Integration checks

- [x] 4.1 Run the round trip in file form on both test grounds through the action, which removes the marked files and writes them in the new form, read each file back and confirm it against the code, and verify `go test ./pkg/graph ./pkg/vault -count=1` passes and the files present in the two ground folders are exactly the names the mapping produces. Verify that the two files named `Thing.md` read back as two different identifiers and that each confirms against the declaration in its own package.
- [x] 4.2 Verify `gofmt -l pkg/graph pkg/vault testdata` reports nothing, that the node set the CLI prints for `./testdata` grows by exactly the declarations added to the ground while the set for `./testdata/second` is unchanged, and that the names in the round trip assertions match the names the mapping produces.

## Workflow follow-up

- Archive the change once the implementation and the project review requirements are satisfied; the archive applies the delta to `openspec/specs/vault/spec.md` by removing the name requirement and the check requirement, appending the requirements added here, and leaving the rebuild requirement as it is. The check requirement is restated rather than modified because one of its reasons reads a part of the name that no longer exists, and the validator will not let a MODIFIED block drop a scenario.
- The next change is the relation between notes. It depends on this one, because a link text is either the folder plus the name or the name alone, and the name is now the node part of the identifier.
- The wiring in `cmd/graphnodes` is test infrastructure for the vault change and not part of the `vault` capability, so the command is rewritten in a separate change, as is the vault root moving from the constant to configuration.
