# Design

## Context

`FileName` builds the name from `SemanticType.String()`, the receiver and the node name, and `ParseFileName` reads the name back by matching the prefix against the closed semantic set. The folder is built from the recorded package path, so the package part of the identifier is carried by the location rather than by the name, and the rest of the identifier is carried by the frontmatter. See proposal.md - Why for the motivation and the delta spec for the required behaviour.

The printed canonical form stays as data in the frontmatter, which is what the `node-model` requirement allows: it says the form is usable as a note name, not that a note must be named with it. `docs/node-model.md` states it as fact, so the wording is aligned with the requirement.

## Goals / Non-Goals

**Goals:**

- A name that depends only on node identity, so it does not move when a rule decides something different about the node.
- A name that is unique inside its folder without a disambiguation scheme, as the language already guarantees.
- A name that can be read back into the node parts without matching anything against a closed set.

**Non-Goals:**

- Links and how a link text is built, which is the next change and depends on this one.
- The frontmatter field set and its order, which do not change.
- Hand renames and hand moves, and content a person wrote inside a file the graph generated.

## Decisions

### The name is the node part of the identifier

`FileName` builds the receiver part and the name part from the record, the same two parts `graph.ID` prints, and leaves the package part in the folder. The alternative was to use the whole printed form as the name, which the `node-model` requirement permits: it writes path separators as `~`, drops the path when a package reports no module, and repeats the package part that the folder already carries. A name built from the printed form would need a folder that cannot be read back, so the name keeps only the part that the folder cannot carry.

### The receiver part stays

Go allows a package-level declaration named `Read` and a method named `Read` in the same package, so a name made of the node name alone would ask for one file for two nodes. The receiver part is what separates them. The alternative was a subfolder for methods, rejected because the location is the package path the record carries and a subfolder would break that.

A method always has a named receiver type, so the receiver part is never empty for a method node; an empty receiver only occurs for a package-level node, where the name alone is the whole node part. This is asserted rather than assumed.

### Reading a name back no longer matches a closed set

Without a prefix, the text before the dot is the receiver and the rest is the node name, so `ParseFileName` becomes the inverse of the node part. A name the mapping produces carries at most one dot, because neither part can contain one, so the split never has to choose between two dots, and a name that carries more than one dot is refused as a name the mapping could not produce. The rule that no value of the semantic set contains an underscore leaves with the prefix, and a node name containing an underscore is read whole. `graph.ParseSemanticType` and `graph.ParseKind` stay used for the values the frontmatter states, which is what the check compares against the code.

### The check reason for the name is repurposed rather than removed

The reason becomes that the name is not the node part of the recorded identifier. It stays computed from the record read out of the file, as it was, because the name is compared against the identifier the file states rather than against the code. `Verify` still supplies only the region reason.

### Nothing is carried between builds

The mapping recomputes the location, the name and the content for the current node set from the source on every build, and the action removes the marked files before writing them, so no name is preserved from one build to the next and nothing monitors a name against the code. The only stability the name has is a property of the mapping: for the same node part the name is the same, whatever a rule decides about the node. A correspondence that a change in the source has invalidated is stated only when a check is invoked for that file, which reads the code at that moment.

### The rebuild requirement stays out of the delta

Its case that a name differs from the name of the marked file that holds the node is what happens on the transition to this form, so the requirement is still true and is not restated. Only the fact that the name is determined by the node part of the identifier and by nothing a rule decides is new, and it is stated as its own requirement.

### Existing files are renamed by the action, not by hand

The note files the test ground has in the vault carry the marker, so the rebuild action removes them and writes them under the new name. A migration script was rejected: it would have to decide which files the graph owns, which is exactly what the marker already decides, and a hand rename would move a file out of the folder its recorded package path maps to.

### The same-name cases are asserted on real code in the test ground

The main ground declares `Thing`, which the second ground already declares, so the same node name appears in two ground packages and in two folders. The main ground also declares `Read` at package level and a method named `Read` on `Thing`, so the two forms of a name exist in one folder on real code rather than in two records assembled by hand. The lookups for these nodes use the canonical form, as `nodeByID` does, because `nodeByName` returns the first match and cannot tell a package-level `Read` from a method named `Read`. Adding entities to the ground does not change what the ground asserts for coverage, and the names the documentation shows stay names a real build produces.

## Risks / Trade-offs

- Existing files in the vault are removed and rewritten, so anything a person wrote inside a marked file is lost. Accepted knowingly: the marker claims full ownership, and this change makes that more visible.
- A file left in the vault can state an identifier that the source no longer has, because nothing monitors the correspondence between builds. Accepted: the correspondence is stated on demand by a check, which reads the code at that moment, and the action writes from the current source rather than from the state of an earlier build.
- The name no longer reads as what the node is in a file listing, so a folder of notes reads as a list of identifiers rather than as a list of kinds. The frontmatter and the body state the classification, and the check reads it from there.
- The folder still cannot express the module part or the version, so the same package path from two modules maps to one folder. This change does not make it worse, but a name that carries only the node part cannot help either, and a link that names a target inherits the folder. The pair of ground packages shows folders staying distinct, not this case, which no test ground can show.
- A file name containing a dot is read by Obsidian as one name, so `Thing.Read` is the display name. Nothing relies on the dot as a separator outside the program.

## Migration Plan

1. Declare the same-name entities in the main ground, then change the mapping and the name read back, with tests for the name form, the receiver separation, and the case where a package-level node and a method share a name.
2. Repurpose the check reason and update the tests that assert the reasons.
3. Run the rebuild action over both test grounds, which removes the marked files and writes them under the new names, then update the names the round trip asserts.
4. Update `docs/vault.md` and the wording in `docs/node-model.md`.

## Open Questions

- Whether a link text will be the path or the name is left to the next change. Nothing here depends on the answer: the path form uses the folder plus the name, and the name form uses the name alone.
