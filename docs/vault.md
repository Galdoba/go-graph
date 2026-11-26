# Vault note files

A node record becomes one note file in the vault. A note is written for every record the graph
produced, both the records of the package the build read as a target and the records a second pass
produced from a reference. The mapping works in both directions: from a
record to a location and a name, and from a file back to the record it claims. Actions are
discrete, so nothing observes the vault between them; a check answers from the file and the code
at the moment it is invoked, without the build that produced the file.

Implementation: `pkg/vault`. `Folder` and `FileName` are the mapping, `ParseFileName` reads a name
back, `Render` writes a note, `Read` reads its record, `Rebuild` is the action over one run,
and `Check` answers for one file.

## Location

The path from the vault root to a note is the package path the record carries, joined to the root
with the platform separator. The folder is built from the recorded path rather than from the
printed canonical form, which writes separators as `~` and drops the path when a package reports
no module.

The test ground and its second package map to:

```
<vault>/github.com/Galdoba/go-graph/testdata/Box.md
<vault>/github.com/Galdoba/go-graph/testdata/second/Thing.md
```

A directory is not always named after the package it holds, so the name the package declares is
read from the frontmatter rather than from the folder: the ground package sits in `testdata` and
declares `main`.

When the record states a module version, that version is a part of the folder. One entity named from
two versions of the same package path therefore maps to two files, because a folder that holds two
identifiers cannot be read back to one node. The version is written in the location rather than in the
name, which carries the node part and nothing else. A record that states no version maps to the folder
the package path alone gives, which is the shape every note the test ground produces keeps:

```
<vault>/mod.test/shape/Shape.md
<vault>/mod.test/shape/v1.2.3/Shape.md
<vault>/mod.test/shape/v1.3.0/Shape.md
```

The version is a part of the folder for the same reason the package path is: it is a part of the
identifier, so a check reads the folder against it. A note whose folder states a different version
than the record states is refuted as a location that does not match the record.

## File name

A name states the node part of the identifier alone: the receiver when the node has one and then the
name the package declares.

```
[<receiver>.]<name>.md
```

The brackets mark an optional part and are not written. Nothing in a name states a classification, so
a rule that decides differently about a node does not change its name, and the name is the same for a
node classified as a structure and for the same node left unresolved.

| Node | Name |
| --- | --- |
| the type `Thing` of the second ground | `Thing.md` |
| the method `Read` declared on `Thing` | `Thing.Read.md` |
| the function `New` of the second ground | `New.md` |
| the `Read` the main ground declares at package level, sharing its name with the method above | `Read.md` |
| a node no rule classifies | `Shape.md` |

The receiver stays in the name because a package can declare an entity and a method with the same
name, which would otherwise ask for one file for two nodes. Inside one folder the names are unique:
a package cannot declare two entities with the same name, and a method is named with the type it is
declared on. Two packages with the same node name are in two folders, which is the case the test
ground shows by declaring `Thing` in both.

## Parsing a name back

No part of a name is matched against a closed set, because no part of a name states a classification.
A Go identifier cannot contain a dot, so a name the mapping produces carries at most one dot, and that
dot separates the receiver from the node name while the node name keeps any underscore it has:
`Foo_Bar.md` reads as the node name `Foo_Bar` with no receiver, and `Thing.Read.md` reads as receiver
`Thing` with node name `Read`. A name whose receiver or node part is empty is not read as a name the
mapping produced, and neither is a name that carries more than one dot, because two identifiers cannot
produce one.

This is the programmatic path from a note to the node it records: the name gives the node part of the
identifier, the frontmatter gives the package part and the rest of the record, and a value read from
the frontmatter is matched against its closed set by `graph.ParseSemanticType` and `graph.ParseKind`,
so no separate table of names is kept in the vault. A name read on its own does not say which node the
package declares: a node named `Read` and a method named `Read` share the node name, so the receiver
the name carries is what tells them apart.

## Frontmatter

The frontmatter is read by Obsidian as properties, and its field order is fixed, which is what
makes a rendering reproducible. A field the record does not carry is not written as an empty
value, so the fields written differ between records.

| Order | Field | What it carries |
| --- | --- | --- |
| 1 | `generated` | the boolean marker, the claim that the file was generated by the graph |
| 2 | `module` | the module part of the identifier |
| 3 | `version` | the module version, written only when the module has one |
| 4 | `package_path` | the package path, the claim the location is read against |
| 5 | `package` | the name the package declares |
| 6 | `kind` | the kind of the node |
| 7 | `semantic_type` | the semantic type the record states |
| 8 | `properties` | the properties the node reports |
| 9 | `file` | the file the node was declared in |
| 10 | `line`, `column` | the position of the declared identifier |
| 11 | `start_line`, `end_line` | the declaration range |

The node name and the receiver are not repeated, because the file name carries them.

Only a marker that reads as true claims graph ownership. A file with no marker, or a value that
does not read as a boolean, is not the graph's business and is never removed.

## What a note for a record produced from a reference holds

A record produced from a reference belongs to the package that declares the entity, so its note is
written in the folder that package path maps to rather than in the folder of the package that
referenced it. Its location and its name read back to the identifier that record states, as any other
note does, because the identifier rules are the same for it and the vault keeps no name space of its
own for these notes. The test ground pair shows the shape: a reference from `use` to `shape` writes

```
<vault>/github.com/Galdoba/go-graph/pkg/graph/testdata/fixtures/shape/Shape.md
```

Such a note holds the values its record states and nothing more: the identifier parts, the kind, the
semantic type, the properties, the region and the comment above the declaration when the load reported
the syntax that declares the entity. The rules that shape a note are answered against it as against any
other note, but a rule that asks for a value the record does not state is not answered against it, so
the formatting rules are not applied to a part the record cannot state. A record that states no region,
which is what a record produced from a name states when the load reported no syntax, writes no `file`,
`line`, `column`, `start_line` or `end_line`, and its `## Code` section states its heading and nothing
else, because there is no snippet to write. Nothing is invented to fill a part the record cannot state,
so a consumer reading the file reads nothing rather than an empty value. The field order stays the same
for these notes, and the pass that produced a record is not written in the file, so a note is read as
the record it claims without knowing which pass produced it.

The `## Related` section lists the connections that point at the record, because a record produced from
a reference carries no uses to answer from. The link that names it is the declaration that referenced it,
so the section of `Shape.md` lists the node that used it:

```
- [[github.com/Galdoba/go-graph/pkg/graph/testdata/fixtures/use/Use.md|Use]]
```

The section is answered from the records of the run, which hold every package the run read as a target,
rather than from the records of one package. The side it lists depends on how the record was read: a
record the build produced states the uses its declaration carries, so its section lists the connections it
points at, while a record produced from a name states no uses, so the connections it can be a member of
are the ones that name it as their target. Read from the two sides of the same pair, `Use.md` lists
`Shape.md` and `Shape.md` lists `Use.md`, and a pair whose members come from two packages of the run is
answered only when the collection holds both of them. A folder that two runs write shows the view of the
run that last wrote a file in it, so the section of a note is the view of the run that wrote it and not
the view of a later run that reads the same records.

What such a note should hold beyond the values its record states is decided in a later change, from
what the first run produces.

## The rebuild action

The action runs over the records of one run, which covers every package the run read as a target. It
removes the marked files in the folder that each of those packages maps to before writing the
files the mapping produces for the records of the run, and it states which files it removed. Nothing
is carried between builds: the action does not look at what a file was named before, so a name this
build no longer produces is replaced rather than left behind. A file whose frontmatter cannot be read
is reported as unreadable rather than removed, because ownership cannot be confirmed from it.

Ownership has two levels across the run: a folder one of its packages maps to is owned in full, and a
folder the run reaches only through a reference is owned file by file. In the second kind of folder
ownership follows the records of the run rather than the folder, because several runs write in the same
folder for different packages. There the action removes only the marked files the records of this run
map to, so a marked file another run wrote stays, and a file this run does not map to is not judged at
all, which is why it is not reported as unreadable. The folders and the entries in them are visited in
sorted order, so the removals the action states are the same for the same records.

## The check and its reasons

A check invoked for one file confirms or refutes it against the record the file carries and the
code at that moment. A record that states a region is answered at that region, and a record that
states none is answered against the entity the package declares, because nothing in the file names a
place to read it at. The folder is read against the package parts the record states, so a note whose
folder states another version than the record states is refuted as a location. The reasons are a closed
set, and they are what a test asserts:

| Reason | What refutes the file |
| --- | --- |
| `frontmatter cannot be read` | the file has no frontmatter block, its marker does not read as a boolean, or a value a field states is not a member of the closed set that wrote that field, which a value of the other set is not because the two sets share no members |
| `location is not the folder the recorded package path maps to` | the file sits in a folder its own recorded package path does not map to, including a folder that states a version the record does not state |
| `file name does not match the record the file carries` | the name is not the node part of the identifier the record states: a record that states a method names an identifier whose node part carries a receiver, a record that states any other semantic type names one that carries none, and a name with an empty part is not a node part |
| `the recorded region no longer declares the node` | the file and the lines the record names no longer declare a node with the recorded name and kind |
| `the package the record names cannot be read` | the check cannot read the package the record names, which is what a record that states no region is answered against |
| `the package the record names no longer declares the node` | the package the record names no longer declares an entity with the recorded name, kind and exported value, which is the answer for a record that states no region |

A package that cannot be read is its own reason rather than a reason about a region, because a record
that states no region carries no region for such a reason to be read against. A foreign note is
confirmed while the package declares the entity it records, and refuted when that package no longer
declares it, which is the answer a note read from its folder alone can be given without the build that
wrote it.
