# Proposal

## Why

A pair is answered only when both of its members are records, and a build produces records only for the
package it loads as a target. A reference from our code to another package already names a node, but that
node has no record, so the graph answers nothing for the relations a graph is read for: an unexported
function of our package calling an external dependency, our structure satisfying an interface of another
package, a field whose type belongs to another module. The node model already resolves such a name across
the package boundary; what it does not do is turn the resolved name into a record the rules can read.

## What Changes

- Add a second pass. The first pass stays as it is: the nodes of the package the build reads as a target.
  The second pass produces a record for each node a reference of the first pass names, in a package the
  build did not read as a target.
- The second pass is demand-driven. It produces a record only for a name the first pass actually names, so
  a package that uses one function of an imported package produces one record and not the whole package.
  It reads no body, because the first pass already recorded what each body refers to.
- Node-ness for an entity declared by another package is decided by the definition of a node rather than
  by whether the name appears in that package's scope, so a method called on a type of another package
  names a node.
- The exported filter applies to records the second pass produces. Nodes of the package the build reads as
  a target are recorded whether exported or not, so `doJob` stays a node and still answers for the pair it
  produces through its references.
- A record produced by the second pass carries what a name can state without the source of the package it
  belongs to: identity, kind, semantic type, receiver, parent, the exported property, and the shape a type
  states (its declared methods and its method specs). It states no declaration region, no uses and no doc
  comment.
- A rule that needs a value a second-pass record does not carry cannot answer for that pair, so the pair
  keeps the third outcome instead of answering no.
- A note for a second-pass record is lighter than a note for a record the build read from source: it holds
  what the record can state, and the formatting rules of a note are not applied to it at this stage. The
  check answers for such a note against the package, not against a region. What such a note should hold
  beyond the values the record states is decided after this change is implemented, from what a run produces.
- A record produced from a reference is identified by the same rules that identify any other node, so its
  note reads back to that identifier from its location and its name. The module version a record states
  becomes a part of the location, so two module versions of one package path cannot map to one file.

No rule in the rule set changes, and the domain rule is unchanged: the domain grows by making more records
readable, which is what the second pass does.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `node-model`: a second producer of node records, the identifier rules that apply to a record produced
  from a reference, what such a record carries, and node-ness for an entity declared by another package,
  including a method.
- `node-links`: the exported filter for a target that belongs to another package, and what a rule answers
  when a member carries no value it reads.
- `vault`: which records become notes, what a note states when its record carries no declaration region,
  the location a note for such a record maps to so that it reads back to the identifier the record states,
  and what the check can answer for such a note.

## Impact

- `pkg/graph`: `namesNode` and `referentObj` for a referent from another package, and a new producer that
  builds a record from a name the first pass named.
- `pkg/link`: no rule changes, only the collection it is passed grows.
- `pkg/vault`: the render and the check, for records that state no region, and the folder mapping, so a
  record that states a module version maps to a folder that states it.
- Test ground: the existing `foreign` and `alias` fixtures already show a reference crossing a package
  boundary; the ground needs a use of a method declared on a type of an imported package.
- Unknowns to confirm before implementing: what a load that does not read a package as a target preserves
  for a declared entity (positions, type parameters, aliases, and whether a method was declared or promoted).
