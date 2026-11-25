# Proposal

## Why

The node model describes a node but gives it no physical form. A record is printed and verified against a fresh load, yet nothing stores a node as an artifact a person can navigate or link. One note file per node gives the graph a place outside the library where a node can be identified from the file alone and connected to the code it came from.

## What Changes

- A new capability that turns a node record into a note file: where the file lives, what it is named, and what it carries.
- A note is created for a node owned by a package the graph built. A node the graph never built gets no file, so a package outside the graph contributes references only.
- The file location and the file name together identify the node exactly, and the file name states the semantic type of the node rather than the form of declaration that created it. The frontmatter carries what the name cannot carry, plus the code region the node was declared at.
- Every generated note carries an ownership marker in its frontmatter, so a rebuild can tell the files the graph owns from the files a person wrote.
- The rebuild action removes the files the graph owns in the folder of the package it builds, states what it removed, and writes the files again, so a note for a node that no longer exists does not survive and a file the graph did not generate stays intact.
- A note and its node can be confirmed or refuted by a check that is invoked on demand for one file and answers from the file and the code at that moment, without the build that produced the file. Nothing observes the vault between actions.
- Only packages that belong to a module are included, since a package with no module cannot be compiled.
- Out of scope: relations between nodes, how a link is written, human-owned content inside a file the graph generated, moving or renaming a note by hand, and any form of visualization.

## Capabilities

### New Capabilities

- `vault`: how a node record becomes a note file - the location, the name, the frontmatter including the ownership marker, the content written on generation, the files a rebuild action removes, and the check that confirms or refutes a file against its node on demand.

### Modified Capabilities

None. `node-model` requirements do not change. The canonical form stays usable as a note name; the vault keeps that form as data in the frontmatter and uses a shorter file name for navigation, which is a choice inside the vault rather than a change to the identity rules.

## Impact

- New package `pkg/vault`, depending on the node record only.
- `graph.Verify` requires a reference to the declaration (`pkg/graph/verify.go:23`), so the check that starts from a file needs a form that takes the recorded position instead of an `ast.Decl`.
- The file name is derived from the semantic type, which is recomputed on rebuild, so a reclassified type node can get a different name in the next run. Removing the files the graph owns by marker rather than by the locations the mapping produced is what keeps no file left behind.
- Reading a file as frontmatter becomes a precondition for any decision about it, because ownership is decided from the marker.
- The vault root is a constant for now and moves to configuration later.
- The test ground gains expected locations and names, and the round trip gains a file form.
- No change to node identity, semantic types, properties, or attributes.
