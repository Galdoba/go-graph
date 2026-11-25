# Proposal

## Why

The node model records attributes as data a relation can later be built from, but nothing yet decides whether a pair of nodes is connected. Without a rule that answers for a pair, a note cannot state a relation, a test cannot assert one, and the difference between "there is no connection" and "no rule can answer yet" is lost. This change builds the linking infrastructure and the first rules, before any decision about how a link is written into a note.

## What Changes

- A link is a binary fact about one pair: a source node and a target identifier. Only the outgoing direction counts, so `A -> B` and `B -> A` are two answers, not one.
- A rule states one relation type and answers for one pair from the two members of the pair alone.
- A rule answers by the semantic type it covers. A rule that cannot answer for a pair does not answer no; it reports that it could not answer, and that case is named. The value `unresolved` is the signal that a rule is missing, so a pair with an undecided semantic value is not reported as having no connection.
- The answer for a pair is one of three outcomes: the connection exists, the connection does not exist, or it cannot be decided. The first rule that answers yes decides the answer; all rules that could answer saying no makes it not connected; otherwise the outcome is the third one, and the answer names the last rule in the fixed order that could not answer. The rule that produced a yes or a no is not reported, because a pair is a fact and not a fact with a reason, and which undecided rule is reported carries no meaning as long as it is reproducible.
- The rules decided here are two: a structure to the method whose parent it is, and a structure to the node one of its attributes refers to.
- The answer is reproducible for the same pair and the same rule set, and the set of connections can only grow as semantic values get decided: a rule may turn a no into a yes, never the reverse. A pair answered as existing keeps that outcome against a rule added later, not against a change to the source, because the outcome is read from the source at the moment it is answered.
- The test ground already contains pairs for both rules, so each rule is asserted on real code. No new entities are needed.

## Capabilities

### New Capabilities

- `node-links`: what a connection between two nodes is, what a rule may look at, the three outcomes an answer can have, and how the outcomes of rules combine into the answer for a pair.

### Modified Capabilities

None. The node model already records the data these rules read, and the vault spec does not read links: a program reads a note only to decide whether it owns the file, and the check answers from the record the file carries.

## Impact

- New package `pkg/link`: the rule set, the pair answer, and the three outcomes with a reason for the third.
- `pkg/graph` is read as data only. No change to the node record, to `Attributes`, or to the closed sets of kinds and semantic types.
- Tests assert each rule on pairs the test ground already produces (`Struct1` embedding `Reader`, `Thing` with its method `Read`) and the third outcome on a pair whose semantic value is undecided.
- `docs/node-links.md` documents the rule methodology, the outcomes and the aggregation rule. `docs/node-model.md` gains a pointer from its out-of-scope note to this capability.
- Out of scope: pairs whose target is an identifier without a record, writing links into a note, reading links back from a note, filtering links inside a note, the frontmatter field set and its order, the form a link takes inside a file, and the wiring in `cmd/graphnodes`.
