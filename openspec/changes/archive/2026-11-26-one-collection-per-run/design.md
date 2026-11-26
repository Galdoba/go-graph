# Design

## Context

See proposal.md - Why. The command loads a target, builds the first pass per package, runs the second
pass per package, and writes notes per package, so each note is rendered with the collection of its own
package. `pkg/link` reads whatever collection it is passed and keeps nothing between calls; the pairs it
answers are already ordered by the pair. `pkg/vault.Rebuild` takes one package, which is the folder it owns
in full, plus the files its records map to in other folders. The identifier of a record produced from a name
is built by the same rules as the identifier of a record built from source, so the same entity from both
passes already carries the same identifier.

## Goals / Non-Goals

**Goals:**

- One collection per run, covering every package the run read as a target.
- The collection is the same for the same set of packages, whatever order they arrive in.
- One entry per identifier, with the record read from source kept when both passes produced one.
- The `## Related` section of a note is answered from that collection.

**Non-Goals:**

- No change to how a rule answers one pair, to the rule set, or to the outcomes.
- No state carried between runs and no index beyond the files themselves.
- No change to which entity a run writes a note for: a note is written only for a record the run produced.

## Decisions

**Merge the records of the run in one place.** A helper in `pkg/graph` takes the records the run produced and
returns one entry per identifier, ordered by it, keeping the record read from source for an entity that
appears in both. The alternative was to tell `SecondPass` which packages are targets so it skips the ones the
run already built; that spreads the run's shape into a pass that is called per package and cannot see the
other packages, and it does not help when two packages of the same run produce the same record from their
own references.

**Order the collection by identifier rather than by package.** The load returns packages in an order the
loader chooses, so a collection assembled package by package is not reproducible. Sorting by the identifier
each record states is the same rule the model already uses for its own lists and for attributes, so the run
and everything read from the collection - the `## Related` section and the removal report - becomes
reproducible without a new ordering rule.

**Keep the direction of the `## Related` section as it is.** A record read from source lists the connections
it points at; a record read from a name lists the connections that point at it, because it states no uses and
can be a member of a pair only as a target. Obsidian already renders the incoming side as backlinks, so the
section is more useful as the outgoing side for a record that can state its uses. This is the wording the
previous change settled, and it becomes a test rather than a comment.

**Ownership is stated for the run.** The rebuild action takes the packages the run read as targets, so it owns
the folders of all of them, and in any other folder it owns only the files its records map to. The alternative
was to keep one call per package; that is what makes a shared folder show the last writer's view.

## Risks / Trade-offs

- A note for an entity that two packages of the same run can write changes content when the source record
  wins, because the region and the uses come from the record read from source. → Acceptable: the file is the
  view of the package that declares the entity, which is what a check reads it against.
- Across runs, the last run to write a file still decides what the `## Related` section says. → The fixed order
  makes this deterministic inside a run; a folder that two runs write is still the last writer's view, and the
  docs will say so.
- Keeping the source record means a run no longer writes a name-record for a package it also read as a target,
  so a run that previously wrote those files writes them from the fuller record. → The mapping is a function of
  the identifier alone, so the file name and location do not change.

## Migration Plan

No stored state to migrate. The next run over the affected packages rewrites the notes it owns; the vault is
rewritten rather than patched. Rollback is reverting the merge helper and the command loop, since the rules in
`pkg/link` and the mapping are untouched.

## Open Questions

- Whether `pkg/graph` should expose the merge as a function the command calls or as a method on the records it
  already returns; this does not change the behavior, so it can be settled during implementation.
