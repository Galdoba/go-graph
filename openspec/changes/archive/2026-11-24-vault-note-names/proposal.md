# Proposal

## Why

The file name is built from the semantic type, which the node model deliberately keeps outside node identity, so a note can be renamed between builds for a node that has not changed. The name states a classification rather than the node, and the classification is recomputed on every rebuild. The next step is relations between notes, where a name has to name a target reliably, so the name should express the node part of the identifier instead of what a rule decided about it.

Uniqueness inside a folder already comes from the language, not from the prefix, so removing the prefix costs nothing that the prefix was actually providing.

## What Changes

- The file name becomes the node part of the identifier: the receiver when the node has one, then the name the package declares. The semantic-type prefix is removed.
- The receiver part stays, because a package can declare a package-level entity and a method with the same name. A name made of the node name alone would need one file for two nodes in the same folder.
- The semantic type is stated only in the frontmatter, where it is already read as a member of its closed set.
- **BREAKING** the form of every existing note file name changes. Existing notes are renamed by the rebuild action rather than by hand, since the graph owns the files it marked.
- Reading a name back loses the rule about the prefix, so the name is read as the receiver and the node name, which is the inverse of the node part of the identifier.
- The check reason about a name becomes that the name is not the node part of the recorded identifier. The reason that compares a prefix with the semantic type the file records is removed, because the name no longer carries a classification.
- The name is determined by the node part of the identifier and by nothing a rule decides about the node, so two builds that disagree only about a classification produce the same name. Nothing carries a name from one build to the next and nothing monitors it against the code: the name is the one the current build produces from the source, and a correspondence a code change has invalidated is stated only by a check invoked for that file. The rebuild requirement does not change: the case where the name the mapping produces differs from the name of the marked file that holds a node is exactly the transition to this form, and the action handles it by ownership rather than by a migration.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `vault`: the requirement that states how a note file is named, the requirement that the name is determined by the node part of the identifier alone, and the closed set of check reasons.

## Impact

- `pkg/vault/mapping.go`: `FileName` builds the name from the identifier rather than from `SemanticType.String()`, and `ParseFileName` reads the receiver and the name without matching a prefix against the semantic set. `ParseSemanticType` stays used by the check for the value the frontmatter states.
- `pkg/vault/check.go`: one reason in the closed set is replaced.
- Tests for the mapping, the name read back, the check reasons, and the round trip change their expected names. The main ground gains a `Thing` that the second ground already declares, and a package-level `Read` beside a method named `Read` on that `Thing`, so the two same-name cases are asserted on real code; the lookups for the two nodes named `Read` use the canonical form, since a lookup by name alone returns the first match.
- `docs/vault.md`: the name form, the table of example names, the parsing rule and the reasons table.
- `docs/node-model.md` states that the canonical form is the name of the note, which is no longer true for a note in the vault; the wording is aligned with the requirement, which says the form is usable as a note name.
- The note files the test ground has in the vault are regenerated in the new form.
- Out of scope: links and how a link is written, which is the next change and depends on this one, the frontmatter field set and its order, the vault root moving to configuration, hand moves and renames, human-owned content inside a generated file, and the wiring in `cmd/graphnodes`.
