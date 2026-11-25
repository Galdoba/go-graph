# Design

## Context

`pkg/graph` records a node with its attributes, and an attribute already carries the identifier of the node it points at. Nothing yet decides whether a pair of nodes is connected, and nothing states what the difference is between "no connection" and "no rule covers this case yet". See `proposal.md` for the motivation.

The rule methodology was settled before this change was written: rules read only the pair, the answer has three outcomes, and the rule identity survives only in the outcome that says the answer could not be given. This design records how that shape is expressed in code.

## Goals / Non-Goals

**Goals:**

- A rule is a decision about one pair, and the pair is the unit a test asserts on.
- The three outcomes are distinguishable in the answer, so a test can tell "no connection" from "no rule can answer yet".
- The pairs a build answers are reproducible without carrying anything between builds.

**Non-Goals:**

- A link whose target is an identifier without a record. The rule set answers for pairs of records, so this stays with the change that can form such a pair.
- Anything about a note: no link is written to a file, and no link is read from one. The program reads a note only to decide whether it owns the file, so links are not needed for that decision.
- A storage form for links. The answer is computed per pair; whether it is kept is a later decision.

## Decisions

### The answer for a pair is a three-outcome value, not a boolean

A rule answers yes, no, or "cannot answer". The rule that cannot answer does not answer no, because answering no would state a fact about a case the rule does not cover. The answer for a pair is the combination of the rules: any yes wins, and if no rule answered yes the outcome is no only when every rule that could answer answered no.

The Go shape is `(bool, error)`, where the error carries the name of the rule that could not answer. Nothing else can fail, because a rule reads only records. A rule that needs a loaded package is not part of the rule set, so there is no operational failure to distinguish from an answer, and one channel is enough.

Alternative considered: an integer outcome with a description. It names the three states directly, but it loses the fact that a yes or a no needs no description, and a test asserting the third outcome would compare text instead of a named rule.

### A rule is a decision over the pair, and the rule set is closed

Each rule states one relation type, so a rule that requires a structure as its source answers no for a pair whose source is a variable. Rules are consulted in a fixed order, the outcome does not depend on that order, and the rule named for the undecided outcome is the last one in that order, because the combination is a disjunction over answers rather than a choice of which rule won. The rule set is closed and grows only by adding a rule, which is the same stance the semantic set takes: a value already decided keeps its meaning when a new value is added.

The two rules decided here are the relation from a structure to the method that records it as its parent, and the relation from a structure to a node one of its attributes refers to. Both read the semantic types they cover, so an undecided value in either member gives the outcome that no rule can answer.

### The order decides which undecided rule is reported

The consultation stops at the first yes, and for the undecided outcome the reported rule is the last one that could not answer, overwriting the earlier one as the rules are consulted. Which undecided rule is reported carries no meaning, because any of them gives the same outcome and nothing reads the answer back, so a precedence among rules is not needed and a fixed order is enough to make the reported rule reproducible. The alternative, collecting every rule that could not answer, would put a list where one named rule is enough for a test to assert on.

### Rules read only the pair, and the pair is iterated

A rule may read what the two members carry, including the parent recorded on a method. Reading the parent of a method is pair-local, because the method is a member of the pair. A rule that needs a lookup outside the pair is not part of the rule set. An alias attribute is recorded on the node the alias resolves to (`pkg/graph/build.go:226`), so for such a pair the fact lives on the target; reading either member is allowed by the rule contract, while a rule for alias names is not decided here.

Iterating pairs is required rather than convenient: the relation from a structure to its method is visible from the record of the method, not from the attributes of the structure, so it cannot be reached by walking attributes alone.

The domain is defined by what a rule can read rather than by a package boundary. The iteration runs over the collection of records the caller passes, so no node container is added to `pkg/graph` for this change, and `pkg/link` answers only for pairs whose two members are records in that collection. The set grows by making more records readable, which is what a later change does for an identifier that has no record; no rule changes when the domain grows.

Pairs are ordered by the source identifier and then by the target identifier, so the answered set does not depend on the order in which attributes were recorded. `attributeOrder` sorts attributes inside a node, which is not enough for a set of pairs, and `sort.Slice` there is not stable; ordering by the pair is what makes the set reproducible.

### The relation lives above the records, not inside them

`pkg/link` adds no field to a node, because the relation type exists only while a rule is applied. It cannot therefore be part of `pkg/graph`, which builds one package and is documented as not a graph (`pkg/graph/build.go:66-71`). The relation is a layer over the records, in the same position as `pkg/vault`, which reads records and writes files, and a question about a pair always needs the whole collection the rules read rather than one node.

Nothing is kept between builds, so the answer for a pair is read from the records at the moment of the call, and answering the same collection twice gives the same ordered set. Because nothing is stored, the question of what can refute a link does not move into this change; it belongs to the change that writes links into a note.

### A pair is two different records

A record considered as both members of a pair produces no answer, so a field that points at the type that owns it is not answered as a connection. The rules cover relations between two records rather than a relation of a record to itself.

### Each rule is exercised by the test ground

The test ground already produces a pair for each of the two rules, so a rule is answered on real code without adding entities: `Struct1` embeds `Reader`, `Thing` has the method `Read`, `Outer` embeds `Base`, and `Mirror.DescribeAlias` shows that the receiver must be read through the node the alias resolves to rather than through the name written in the source. A rule answered yes by no pair of the ground is reported as not covered, which is the same reporting mechanism the node model uses for a semantic value the ground does not produce, with a rule as the subject instead of a value.

The third outcome is not reachable from the test ground, because the ground produces no `unresolved` node. It is asserted on a record whose semantic value is `unresolved`.

### Naming

`link.Exists(source, target)` answers one pair, where `link.Pairs` answers a collection of records and `link.Coverage` reports the rules no pair of that collection answers. The rules are named by the relation type they state, and that name is what the third outcome reports.

## Risks / Trade-offs

- A rule that answers no for a pair it cannot decide would hide a missing semantic value, and the answer would look like a fact → mitigation: a rule that reads a semantic value returns the third outcome for `unresolved`, and the answer names the rule.
- A pair may be answered yes by several rules, so the relation type cannot be attached to the fact → accepted, because the fact is the pair and the type is only used while applying a rule. No precedence order is needed, which is what storing the type would have required.
- Iterating all pairs is quadratic in the node set of a package, and a cross-package pair set is not reachable from one package graph → accepted for this change; the registry question belongs to the change that forms pairs with an identifier that has no record.
- A pair answered as existing keeps that outcome against a rule added later and not against a change to the source, so the growth rule could be read as a stored fact that survives a code change → mitigation: the requirement states that growth limits what a rule added later can change, while the outcome is read from the source at the moment it is answered.
- The test ground produces no record with a reserved semantic value, so the third outcome is asserted on a constructed record and is not exercised by real code → accepted, because the ground covers every semantic value; the outcome is observable from real code only when a record reports the reserved value.

## Open Questions

- What a link becomes inside a note, which depends on whether the relation type is carried at all. Filtering links inside a note is out of scope for this change.
- Whether a check ever answers about a pair. A check answers about a record and cannot refute what a note carries, so a link written into a file is outside its contract until that is decided.
