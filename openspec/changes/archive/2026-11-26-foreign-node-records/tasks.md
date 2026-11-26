# Tasks

## 1. Confirm what a package that is not a target carries

- [x] 1.1 Load a package that imports another package of the repository and one standard library package, and record for a referent that is not a target: whether the load gives syntax or type info, whether a package-level object reports a usable position, whether a generic named type reports its type parameters, whether an exported alias to an unexported named type reports that it is an alias and resolves to its target, whether a method object identifies the type that declares it, and whether the load reports the module and the module version for the imported package. Verify by writing the answers in design.md under the open question and using them to fix the list of fields a foreign record can carry.

## 2. Node-ness for an entity declared by another package

- [x] 2.1 Decide node-ness for a referent from another package by the definition of a node rather than by that package's scope, so a method declared on a package-level type of another package names a node, and keep the exclusion for a method reachable only through an embedding. Verify with a test on a fixture that calls a method declared on a type of an imported package and asserts the use attribute carries the reference to that node.
- [x] 2.2 Add the test ground case that a declaration references a node of a package the build reads from type information only, and assert the reference rule is covered by that pair. Verify with the coverage run reporting no rule as uncovered.
- [x] 2.3 Update the exclusion table in docs/node-model.md for an entity declared by another package, so it states that node-ness is decided by the definition and a reference can name a foreign method. Verify by reading the row against the tests added in 2.1.

## 3. The second pass

- [x] 3.1 Produce one record for each node a reference of the first pass names, as a second pass of its own that leaves `Build` producing the records of one package, with the identifier, kind, semantic type, receiver, parent, exported property, signature and the shape the type states, and no record for any other name of that package. Verify with a test that a package using one name of an imported package produces exactly one record, and that a run which loads that package as a target keeps one record for the same identifier.
- [x] 3.2 Record a plain qualified reference as a use attribute of the node that owns the declaration, keeping the alias attribute for a name an import wrote that is not the name the package declares, so the most common cross-package reference is in the domain and in the demand set. Verify with a test on a fixture that refers to an entity of an imported package through the name that package declares and asserts the use attribute and its reference, and that a renamed import keeps its alias attribute beside it.
- [x] 3.3 Record on each record whether its values were read from a declaration the build read as a target or from a name it did not read, and carry the region, the declaration and the comment that the syntax the load reports shows. Verify with a test that a foreign record states no uses, states the region and the comment when the load reports the syntax that declares the entity, and states nothing where it does not, so a consumer can tell an absent value from a value that was never read.
- [x] 3.4 Add the second pass to docs/node-model.md, stating what a record produced from a reference carries, including the region and the comment, and what it cannot carry. Verify by reading the section against the tests from 3.1 to 3.3.
- [x] 3.5 Build the identifier of a foreign record by the rules that build the identifier of any other record, so two references naming the same entity produce one record. Verify with a test that two references from two declarations to the same foreign entity produce one record, that a foreign package-level name and a foreign method of the same name produce two records, and that the identifier of a foreign record equals the identifier the first pass produces for the same entity when its package is read as a target.

## 4. Answers for pairs with a foreign member

- [x] 4.1 Answer the cross-package pair from the collection that now holds the foreign records, and confirm the rules are unchanged. Verify with a test that a pair from an unexported node of the built package to a foreign record is answered as existing, that a pair whose foreign member carries no uses leaves the reference rule unable to answer, and that a pair needs a record for both members, so a foreign method whose receiver type was not named produces no parent pair.
- [x] 4.2 Document in docs/node-links.md that the domain grows through the records the second pass produces, that the exported filter reaches only records produced from a reference, and that a rule cannot answer for a value a record does not carry. Verify by reading the section against the tests from 4.1.

## 5. The minimal note, the location, the check and the rebuild

- [x] 5.1 Write a note for a record produced from a reference that holds the values its record states, including the region and the comment when the load reported the syntax, and verify with a render test that such a note is created, that no field the record does not state is written, and that a rule which asks for a value the record does not state is not answered against it.
- [x] 5.2 Confirm or refute a foreign note against the region its record states, against the entity the package declares when it states none, and state the reason that a package cannot be read as its own reason rather than as a reason about a region. Verify with a check test that a foreign note is confirmed while the package declares the entity, refuted when it no longer declares it, refuted for a package that cannot be read, and still refuted for a location or a name that does not match the record.
- [x] 5.3 Scope the rebuild to the files the records of the current run map to, so a folder another run also writes in keeps the files that run produced. Verify with a test that two runs writing in the same foreign folder remove only their own marked files.
- [x] 5.4 Write the module version a record states as a part of the folder its note maps to, so two records that differ only by version do not map to one file. Verify with a mapping test that two versions of one package path produce two folders, that a record with no version maps to the folder it maps to today, and that a foreign note read from its location and its name gives back the identifier its record states.
- [x] 5.5 Update docs/vault.md for which records become notes, what a foreign note holds while the formatting rules are not applied to it, the folder a record that states a version maps to, the check for such a note, and the rebuild ownership. Verify by reading the sections against the tests from 5.1 to 5.4.

## 6. Integration check

- [x] 6.1 Run the command against a package that imports a package of the repository and one standard library package, and verify that the vault holds one note per referenced foreign name, that the `## Related` section of an unexported node of the built package lists those links, and that no other foreign name gets a file.
- [x] 6.2 Run the round trip for a foreign note and verify that its location and its name read back to the identifier the record states, that the check reads the folder against the version the record states, and that two records for one package path at two module versions give two files rather than one.

## Workflow follow-up

- Sync the delta specs into the main specs and archive the change once the project's review requirements are satisfied.
- Decide what a foreign note holds beyond the values its record states, including the doc comment and the `## Related` section, in a separate change, after this change is implemented and from what the first run produces.
