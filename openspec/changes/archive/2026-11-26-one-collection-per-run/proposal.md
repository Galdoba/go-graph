# Proposal

## Why

A run answers pairs and writes notes from the records of one package, so a note that several packages
can write shows only what the last of them saw: a record of a shared package lists the connections its
own package's collection holds, and a pair whose two members belong to two packages the same run read
as targets is never answered. The same entity is also recorded twice when one package reads it as a
target and another names it from a reference, with no rule saying which record a run writes.

## What Changes

- A run produces one collection of records covering every package it read as a target, instead of one
  collection per package.
- The collection is ordered by the identifier it holds, so the same set of packages produces the same
  collection whatever order the load returned them in.
- One entry per identifier. When the run both built a record for an entity and produced a record for it
  from a name, the record read from source is kept, because it states the full set its definition
  requires.
- The `## Related` section is answered from that single collection, so a note shows the whole run's view
  of its node. The direction the section lists is stated: a record read from source lists the
  connections it points at, a record read from a name lists the connections that point at it.
- Ownership is stated for the whole run: the run owns the folders of the packages it read as targets, and
  in any other folder it owns only the files its own records map to.
- Out of scope: the rules in `pkg/link` and how a pair is answered for one pair; nothing is carried
  between runs or persisted beyond the files.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `node-model`: the second pass is read across the whole run, so an entity recorded by both passes yields
  one record and the record read from source takes precedence.
- `node-links`: the collection the rules read is the collection of the whole run, which makes a pair whose
  members come from two target packages answerable, and the collection is ordered by identifier.
- `vault`: the `## Related` section is answered from the whole-run collection and its direction is stated;
  the files the rebuild action owns are stated for the whole run rather than for one package.

## Impact

- `pkg/graph`: a helper that merges the records of the two passes by identifier with source precedence and
  a fixed order.
- `pkg/vault`: the rebuild action takes the packages the run read as targets rather than one package; the
  `## Related` section keeps its rules but is answered over the whole-run collection.
- `cmd/graphnodes`: the loop runs the passes once per package and collects the records for the run, then
  writes the notes once for the whole collection.
- Docs: `docs/node-model.md`, `docs/node-links.md`, `docs/vault.md`.
- Tests: existing integration test in `pkg/vault` asserts the per-package view and is rewritten; new tests
  pin the precedence rule and the direction of the `## Related` section.
