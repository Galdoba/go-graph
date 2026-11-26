# Design

## Context

`Build` produces records for one package, and it needs the package to be loaded as a target: type
information, type info for its syntax, and the syntax itself. `packages.Load` fills the syntax only for the
packages that match the pattern, so an imported package arrives with type information and no syntax. A
reference across a package boundary is already resolved today: `referentObj` builds the identifier of the node
the name denotes, and `namesNode` decides node-ness for a referent from another package by looking the name up
in that package's scope. `pkg/link` answers a pair from the records it is passed, and `pkg/vault` maps a record
to a file. See `proposal.md` for the motivation.

The load mode asks for syntax for every package it type-checks, so an imported package arrives with type
information and with the syntax of its files as well. A record for a foreign name therefore states what a name
and a type state, and also the region and the comment that the reported syntax shows. What it cannot state is
what a body does, because no foreign declaration body is read, so a foreign record carries no uses.

## Goals / Non-Goals

**Goals:**
- Make the records the rules read include the foreign names the first pass already references.
- Keep the rule set and the domain rule exactly as they are, so growth happens by adding records.
- Keep the growth bounded by what our code names.
- Keep every record inside the identifier rules the model already has, so a note for a foreign record reads
  back to the node it records with no name space of its own.

**Non-Goals:**
- Enumerating a foreign package as a build target. The second pass resolves only the entities a reference names,
  and the region and the comment it states for them come from the syntax the same load already reports.
- Reading a foreign declaration body, and so recording its uses.
- Enumerating the transitive closure of imports, or answering relations from a foreign node through its body.
- Changing the canonical form of an identifier, which the second pass uses as it is.
- Filling a foreign note with more than the values its record states, which is decided after the first run.

## Decisions

**Pass 2 is demand-driven and reads no body.** The first pass already records what each declaration refers to,
so pass 2 turns the names those references already carry into records. Alternative considered: enumerate every
exported name of every direct import, which would give the same answers for referenced names and far more
records for names nothing references, so it was rejected.

**A qualified reference names the entity it denotes.** A reference through an import is recorded as a use
attribute, so `strings.ToLower` names the node the `strings` package owns. The gap the spike found is that
today only a renamed import produces an attribute, which leaves the most common form of a cross-package
reference out of the domain and out of the demand set. The rule that a renamed import name is recorded as an
alias attribute of the node that owns the declaration stays as it is, so both parts are recorded: the name the
import wrote and the entity the reference names.

**The second pass is its own producer.** `Build` keeps producing the records of one package, and the second
pass produces the records for the entities the first pass named in packages it did not read as a target. The
collection the rules read is the two sets together. Alternative considered: produce the foreign records inside
`Build`, which would mix two packages into the result of one build and change every test that asserts the node
set of a package; rejected as a larger change than the requirement asks for.

**A record for a foreign name is built from the load that read the reference, not from a second load.**
`p.Imports[path]` gives the object, its package and the module that package was loaded from, which is enough for
identity, kind, semantic type, receiver, parent, the exported property, the signature and the shape a type
states, and the same load reports the syntax that declares it, which gives the region, the declaration reference
and the comment. When a load reports no syntax for that package, the record states nothing there, and the
provenance says so. Alternative considered: loading each used package as a pattern, which would repeat work the
load already did.

**Node-ness for a foreign entity is decided by the definition, not by the package scope.** The scope holds
package-level declarations only, so a method declared on a foreign type is invisible to the current lookup. The
definition already covers a method declared on a package-level named type, so the check follows the receiver of
the method object. Alternative considered: keep the scope lookup and accept that calls to foreign methods name
no node; rejected, because that is the most common way a package uses an external dependency.

**Alias resolution is not changed.** A reference names the node the alias resolves to, so a name an exported
package alias points at produces a record for the entity the package declares, even when that entity is not
exported. The exported filter is a condition on what a reference from our package can name, which the language
already decides, and it stays a property of the record. Alternative considered: filter after resolution, which
would fork the alias rule for a second pass; rejected as a larger model change.

**The identifier rules are not forked for the second pass.** A record produced from a reference carries an
identifier built by the same rules as any other record, from the entity itself, so the same entity named from
several declarations is one record and a foreign note reads back to the identifier its record states.
Alternative considered: a separate identifier or name space for foreign records; rejected, because a note is
already read back to a node through its location and its name, and a second space would need a table that maps
one space to the other.

**A record states where its values came from.** A rule that reads a value a foreign record does not carry must
report that it cannot answer, so the record needs to distinguish "no uses recorded" from "the body was never
read". The provenance also says whether the region and the comment were read from a package the build read as a
target or from the syntax reported for a package it did not, so a record that states no region is read as one
that could not state one rather than as one whose declaration has no region. Alternative considered: read the
absence as no, which would let a foreign node answer no for relations it never read; rejected because it
conflicts with the three outcomes.

**A foreign method set holds only declared methods.** `types.NewMethodSet` returns promoted methods as well, so
the model keeps a method whose receiver resolves to the type itself. This keeps the existing exclusion for a
method reachable only through an embedding, and it is decided from the object rather than from the source.

**The location states the version the record carries.** The file name carries the node part of an identifier and
the folder carries the package part, so two module versions of one package path would map to one file, which
cannot be read back as two nodes. The version a record states is therefore written as a part of the folder. A
package the build reads as a target reports no version, so the folders the test ground produces today do not
move. Alternatives considered: keep the folder as the package path and accept that one folder holds one version,
which leaves a second version unreadable as a node, and write the version in the file name, which conflicts with
the rule that a name carries the node part and nothing else; both rejected.

**A foreign note is written and holds what its record states.** A foreign note is still written, so a link from a
note of the built package points at a file rather than at nothing. The rules of a note are answered against it as
against any other note, because the record states a region and a comment, and a rule that asks for a value the
record does not state is not answered against it: a record that states no region writes no region fields and no
code snippet. The `## Related` section lists the links that point at the record, which the rules answer from the
records that reference it.

**The check answers a foreign note against its recorded region.** A foreign record states a region when the load
reported the syntax that declares it, so the check answers it at that region as it answers any other record. For
a record that states no region, the check answers against the entity the package declares. The reason that a
package cannot be read is stated as its own reason, because "cannot be read" is not the same answer as "the
recorded region no longer declares the node", and the closed set needs the two apart. The rebuild owns only the
files the records of the current run map to, because several runs can write in the same foreign folder.

## Risks / Trade-offs

- **The folder changes shape for a record that states a version.** A note for a record read from a versioned
  module sits in a folder that states the version, which is a shape a reader of `docs/vault.md` does not expect
  from the rule that the folder is the package path. Mitigation: the version is written only when the record
  states one, so the folders the test ground produces do not move, and a test shows two versions mapping to two
  folders.
- **The syntax a load reports for an imported package is not guaranteed.** The region and the comment of a
  foreign record come from the syntax the load reports for that package, which the load mode asks for but a
  loader that only type-checks does not provide. Mitigation: the record states nothing where the syntax is
  absent and the provenance says the values were read from a name, so a smaller result does not change the rules.
- **The same entity can get two identifiers from two loads.** The module part a record carries is the one the
  load reports, so a load that reports no version for a module that has one gives an identifier different from
  the one a build that reads the same package as a target produces, which would read as two nodes for one
  entity. Mitigation: the spike in group 1 records what a load reports for an imported package, including the
  version, and a test compares the identifier of a record produced from a reference with the identifier the
  first pass produces for the same entity.
- **Volume.** A package that references many foreign names writes many notes, including standard library
  folders. Mitigation: demand-driven scope; a limit on which records are rendered can be added later without
  changing the rules.
- **A foreign node answers nothing through its body.** Rules that read uses cannot answer for a foreign source,
  so a foreign node is a target and answers only through shape. This is stated as a requirement rather than
  worked around.

## Open Questions

### What the spike answered

A fixture package that imports one package of the repository and one standard library package was loaded with
the load mode the build uses. For a referent that the build does not read as a target:

- The load gives syntax for an imported package as well as type information: one file for a package of the
  repository, eight for a standard library package. `CompiledGoFiles` is empty for an imported package, so the
  files the load reports cannot tell a foreign record from a target record.
- A package-level object reports a usable position, and a method object reports one too, in the file of the
  package that declares it. A position is available for a foreign entity, so a record that states no region is
  a choice about what a record reports.
- A generic named type reports its type parameters.
- An exported alias of an unexported named type reports that it is an alias and resolves to the entity the
  package declares, which is the alias rule the model already has.
- A method object identifies the type that declares it through its receiver, and a method is not in the package
  scope, so node-ness for a foreign method follows the definition rather than the scope.
- The load reports a module for an imported package only when it resolved one. A dependency in the module cache
  reports `golang.org/x/tools@v0.53.0`, the repository's own module reports a path with an empty version, and a
  standard library package reports no module. So the version belongs to a location only for a record that was
  named from a module, and the folders a target package maps to do not move.

The fields a record produced from a reference carries are the identifier, the kind, the name, the package, the
parent, the receiver, the exported property, the signature and the shape the type states, the region and the
comment of the declaration when the load reports the syntax that declares it, and the provenance of its values.
It carries no use attributes, because no declaration body of that package was read.

Decisions held until this change is implemented and decided from what the first run produces:

- Whether a limit on which foreign records are rendered is needed once the set grows.
