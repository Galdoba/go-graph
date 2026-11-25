# Design

## Context

The library produces node records; nothing writes them to disk yet. The vault is the first consumer that turns a record into an artifact a person can navigate. The record already carries everything the file needs: the package part, the node name, the receiver, the position and range, the kind, the semantic type and the properties. See `proposal.md` - Why for the motivation and `specs/vault/spec.md` for the behaviour each decision below has to satisfy.

Constraints that shape the approach:

- Only packages with a module enter the graph, so a package that reports no module never produces a file.
- Actions are discrete. Nothing observes the vault between them, so the check answers from the file and the code at the moment it is invoked.
- `golang.org/x/tools/go/packages` is the only way the vault reaches the code, and the mode the library uses already loads the module information the frontmatter needs.

## Goals / Non-Goals

**Goals:**

- A mapping in both directions: from a record to a location and name, and from a file back to the record it claims.
- Output that is reproducible for the same record, including field order.
- A delete that can only touch files the graph generated.
- A check that confirms or refutes one file on demand and states a reason when it refutes.

**Non-Goals:**

- Links and their form, human-owned content inside a generated file, hand moves and hand renames, the stdlib inside the graph, versioned dependencies, anything visual. These stay out of the change; see `proposal.md` - What Changes.

## Decisions

### The folder is built from the recorded package path, not from the printed form

`PackageID.String()` rewrites path separators as `~` and, when the module part is missing, prints only the name the package declares, so the path is not in the printed form. The folder therefore reads `PackageID.PackagePath` with normal separators (`pkg/graph/id.go:40`). Using the printed form as a folder would put `~` in a directory name and lose the path for a package without a module. The canonical form stays as data in the frontmatter, which keeps the `node-model` requirement that it is usable as a note name untouched.

### The prefix vocabulary is the semantic set

The file name uses the values of `SemanticType.String()` (`pkg/graph/semantic_type.go:22`): `structure`, `interface`, `function`, `method`, `constant`, `variable`, `unresolved`. No new vocabulary is introduced. The alternative was the kind set, which states the form of the declaration rather than what the node is, and which cannot separate a structure from an interface. The mismatch between the old `func_` prefix and `Kind.String()` returning `function` disappears with this choice, because the name no longer reads the kind.

Parsing rule: no value in the semantic set contains `_`, so the prefix is the text before the first `_`, and the rest is the node name. A node name may contain `_`, so the prefix is matched against the closed set rather than split blindly. A Go identifier cannot contain `.`, so the receiver separator inside the name is unambiguous.

Uniqueness inside a folder comes from the language: a package cannot declare two entities with the same name, so no disambiguation scheme is needed. Two packages with the same node name are in two folders.

### Reading a written value back is the path from a note to a node

A note states the kind and the semantic type as written values, so navigating from a note to the node it records means turning a value back into a member of its closed set. The graph exposes `ParseKind` (`pkg/graph/kind.go`) and `ParseSemanticType` (`pkg/graph/semantic_type.go`) as the inverse of `String()`, and the vault reads them rather than keeping its own table of names, because a table would have to be kept in sync with the set and a value added to the set extends the range it is read from. The check assembles the node from the name and the frontmatter, then resolves it with `Verify`.

### The mapping is a function of the record, not of the identifier

The semantic type is not part of the identifier, so the location of a note can move between builds for the same node. This is accepted knowingly: the rebuild action removes the files it owns by marker rather than by the locations the previous mapping produced, which is what keeps a moved node from leaving a file behind. The alternative, a name built only from the identifier, was rejected because the name should read as what the node is.

### The marker is a single boolean frontmatter field

The marker is one boolean field, `generated`, and it is the first field, because ownership is decided before anything else in the file is read. Only a value that reads as true claims graph ownership: a file with no such field, or a field whose value is not true, is not the graph's business and is never removed. A string naming the producer was considered and rejected - the question the delete action asks is who owns the file, not who wrote it, and a boolean answers it in one token, which Obsidian also shows as a property toggle. The generator always writes the canonical `true`, so the content stays reproducible; a person who writes `true` on their own note hands that file to the graph, which is what the marker means.

A boolean field that carries no value, or a value that cannot be read as a boolean, is not a claim of ownership, so such a file is reported as unreadable rather than removed.

### Frontmatter format and field order

YAML, because Obsidian reads frontmatter as properties; a JSON block was rejected as not readable as properties. The order is fixed, which is what makes the output reproducible:

1. `generated` - the boolean marker
2. `module`, `version`, `package_path`, `package` - the identifier parts the folder cannot carry on its own
3. `kind`, `semantic_type`, `properties`
4. `file`, `line`, `column`, `start_line`, `end_line` - the code region

The recorded package path is written as a field even though the folder mirrors it, because the location check has to compare the folder a file sits in against a claim the file itself carries. Without it, a file read from any folder takes its package path from that folder, so the folder always reproduces itself and a moved file cannot be refuted.

Optional fields are written only when the record has them, in the same order, and an absent field is not written as an empty value. The node name and the receiver are not repeated in the frontmatter: the file name carries them, and the check reads them from the name. The version is written as the record holds it, so a module with no version produces no `version` field.

### The check needs `Verify` to accept a record without a declaration reference

`Verify` already resolves a record by coordinates alone (`pkg/graph/verify.go:29-40`), which is how a position is meant to be read. The obstacle is the guard that requires `n.Decl` (`pkg/graph/verify.go:23`), and a record read from a file has no declaration reference at all. So the guard becomes conditional and the coordinate path becomes the path for a file. A separate position-based function was rejected: it would duplicate the comparison of name, kind and range that `Verify` already performs.

The check reports confirm or refute plus a reason from a small closed set: unreadable frontmatter, location that does not match the recorded identifier, a region that no longer declares the recorded node, a name that does not match the record the file carries. The name reason cannot come from `Verify`, because the semantic type a file records is not read from the code; the location check is computed from the record read out of the file, so a prefix that disagrees with the recorded semantic type is that name reason. Verify supplies the region reason. Free-text reasons were rejected because the reasons are what the tests assert.

### The vault root is a constant

A constant for now, moved to configuration later, as the proposal states. The package path is joined to it with the platform separator; nothing else in the path is rewritten.

## Risks / Trade-offs

- A node reclassified between builds gets a different name, so its file moves. Inside the rebuild action this is absorbed by removing the marked file first; the cost falls on links a person wrote by name, which are deferred. If links are later written by name, the file names would have to become the full canonical form, and requirement 3 of the vault spec would become a MODIFIED delta with a rename of the whole vault.
- The marker asserts full ownership, so prose a person added inside a generated file is removed on rebuild. The boundary is that human content lives in files without the marker; a protected region inside a generated file is a separate decision.
- A file whose frontmatter cannot be read cannot be judged, so it is reported rather than removed. Such a file can survive a rebuild as stale until it is repaired by hand.
- The full canonical form can exceed the 255-byte file name ceiling, which is a further reason the name is the short form and the canonical form stays as data.
- The whole vault is graph output for now, so a rebuild that removes every marked file in the affected folder is safe. Once human content exists in marked files, that safety depends on the ownership rule holding.

## Migration Plan

The vault does not exist yet, so the first run creates the folders and the files for the current node set; nothing is migrated. Rollback is the removal of the marked files, which leaves anything a person wrote in place.

## Open Questions

- How properties are rendered in the body of the note, as an Obsidian property or as a plain list. The check does not depend on it.
- Whether the frontmatter carries a hash of the rendered content to make a comparison cheap. Determinism already makes the content identical, so this is a performance choice, not a behaviour.
- Whether the check result is printed per file or as a summary by the CLI. The existing CLI prints nodes to stderr, so either form is consistent.
