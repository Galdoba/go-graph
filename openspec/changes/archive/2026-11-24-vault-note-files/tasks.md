# Tasks

## 1. Mapping from a record to a location and a name

- [x] 1.1 Create `pkg/vault` with the vault root constant and the mapping from a node record to the folder built from the recorded package path, and verify a test asserts the second ground node `Thing` lands at `<vault>/github.com/Galdoba/go-graph/testdata/second/structure_Thing.md` while the ground node `Box` lands at `<vault>/github.com/Galdoba/go-graph/testdata/structure_Box.md` with the declared name read from the file.
- [x] 1.2 Build the file name from the semantic type, the receiver when the node has one, and the node name, taking the prefix from the closed semantic set, and verify a test asserts `function_New.md`, `method_Thing.Read.md`, and the `unresolved_` form for a node no rule classifies.
- [x] 1.3 Read a file name back into the record parts by matching the prefix against the closed semantic set, which is the programmatic path from a note back to the node it records, and verify a test covers a node name that contains an underscore and a name whose prefix is not in the set, which is reported as not produced by the mapping, and that a value of each closed set reads back to itself.
- [x] 1.4 Document the mapping, the name form and the parsing rule in `docs/vault.md` and verify the example paths in it match the assertions of the mapping tests.

## 2. The note file

- [x] 2.1 Write the frontmatter in the fixed field order with the boolean marker first, writing only the fields the record carries, and verify a test asserts the order, the marker value, and that an absent field is not written as an empty value.
- [x] 2.2 State the kind, the semantic type and the properties of the node in the note in a form readable without the build, and verify a test reads a ground node that reports two properties and finds both stated, and a test for a node that reports generic and entry_point.
- [x] 2.3 Render the same node twice with no change to the source, and verify a test asserts the location, the file name and the bytes of the two renderings are identical.

## 3. The rebuild action

- [x] 3.1 Remove the files the graph owns, those whose marker reads as true, in the folder of the package it builds before writing the files the mapping produces, and verify a test asserts a marked file is replaced while an unmarked file in the same folder survives, including a file whose marker reads as false.
- [x] 3.2 State which files the action removed, and verify a test asserts the reported removals are exactly the files the action removed.
- [x] 3.3 Verify a test that renders a type node, changes its declaration body from a struct to an interface, runs the action again, and asserts one file for that node under the name the new semantic type produces and no file left under the old name.

## 4. The check invoked on demand

- [x] 4.1 Let `Verify` resolve a record that carries no reference to its declaration by using the recorded coordinates, and verify the existing graph tests still pass and a new test resolves a record read back from a file.
- [x] 4.2 Confirm or refute one file from the file and the code at the moment the check is invoked, with reasons from the closed set, reading the kind and the semantic type the file states back into their closed sets so the record read from a file can be compared against the code, and verify a test for each reason: a location the recorded identifier does not map to, a name that does not match the record the file carries, and a recorded region that no longer declares the node.
- [x] 4.3 Report a file whose frontmatter cannot be read as unreadable and never remove it, and verify a test asserts such a file survives a run of the rebuild action, including a value the file states that is not a member of its closed set.
- [x] 4.4 Document the check and the closed set of reasons in `docs/vault.md` and verify the documented reasons match the values the tests assert.

## 5. Integration checks

- [x] 5.1 Run the round trip in file form on both test grounds: render every node, read each file back, and confirm it against the code, and verify `go test ./pkg/graph ./pkg/vault -count=1` passes.
- [x] 5.2 Verify a test that the ground reference from `main` to `strings.ToLower` produces no file, because that package was never built as a target.
- [x] 5.3 Verify `gofmt -l pkg/graph pkg/vault testdata/second` reports nothing and the CLI prints the same node set for `./testdata` and `./testdata/second` as before the change.
- [x] 5.4 Verify a test that two records carrying the same node name in two packages give the same file name in two folders and read back as two identifiers that differ by the module part and the package path, so the two notes are read as two nodes.

## Workflow follow-up

- Archive the change once the implementation and the project review requirements are satisfied; the archive creates `openspec/specs/vault/spec.md` and copies the Purpose from the delta.
- The wiring in `cmd/graphnodes` calls the action and the check for a live run; it is test infrastructure for this change and not part of the `vault` capability, so the command is rewritten in a separate change.
- The vault root moves from the constant to configuration in a separate change, and the link form stays a separate change; if links are later written by name, the file names become the full canonical form and requirement 3 of the vault spec becomes a MODIFIED delta.
