# Node links

A link is a binary fact about one pair of node records, read in the outgoing direction only. `A->B` and `B->A` are separate answers. Several rules answering yes for the same pair produce one fact.

Implementation: `pkg/link`. `Exists` answers for one pair, `Pairs` answers for a collection of records, `Coverage` reports rules with no pair in the test ground, and `UndecidedError` carries the rule that could not answer.

## The three outcomes

An answer for a pair has one of three shapes:

| Outcome | What it states |
| --- | --- |
| yes | the connection exists, and no rule is named |
| no | the connection does not exist, and no rule is named |
| a rule could not answer | every rule that could answer answered no, and the last rule in the fixed order that could not answer is named |

A rule that cannot answer a pair does not answer it no.

## What a rule may read

A rule is a decision over one pair that reads only what the two members carry, including the parent and the receiver recorded on the target. A rule that needs a lookup outside the pair is not part of the rule set, because the answer is read from the pair at the moment of the call rather than from a lookup of other nodes.

## The rules decided here

The rule set holds four rules, consulted in the fixed order below. The relation from a node to the node
a reference it carries names is the rule added here, and it is consulted last.

| Order | Rule | The ground pair it answers |
| --- | --- | --- |
| 1 | a structure and the method that records it as its parent | `Thing` to `Read`, with the same rule also covering `Base` to its method `Read` |
| 2 | a structure and the node one of its attributes refers to | `Struct1` to `Reader`, with `Outer` to `Base` as the second covered case |
| 3 | an interface and a structure that records every method spec it states | `Reader` to `Base`, with `Adder` to `Struct1` as the second covered case |
| 4 | a node and the node a use attribute it carries names | `describe` to `Reader`, with `NewStruct1` to `Struct1`, `zeroStruct` to `Struct1` and `main` to `Five` as the other covered cases |

A method written with an alias as its receiver is answered through the node the alias resolves to, so `Struct1` to `DescribeAlias` is a pair of the same rule.

The rule added here reads the references the source carries, so it answers for a source of any kind, and
a pair is answered no when the reference names an entity the model records as an attribute rather than a
node. Because it reads no semantic value, it is never the rule named for an undecided pair for a record
the build read as a target, so the rule named for such a pair stays the interface rule. A record produced
from a reference is the case where it is named, because the value it reads is one that record cannot
carry. The name written as the receiver records no reference, so
a method answers nothing for the type it is declared on, and the pair `Base` to its method `Read` is
answered by the first rule alone.

## Records produced from a reference

The domain grows through the records the second pass produces for the entities the references of the
first pass name, and the rule set is unchanged. A pair whose members come from two packages is answered
by the same rules as a pair whose members come from one package.

The exported filter reaches only records produced from a reference, because a reference from the package
the build read as a target can name only a name that package exports. A pair inside that package is
answered whatever the exported value of its target is, so `main` to the entity of an imported package is
answered while `main` is not exported.

A pair is answered only when both members have a record of their own. A foreign method whose receiver
type no reference named produces no parent pair, because the type is not a member of the pair.

A rule cannot answer for a value a record produced from a reference does not carry, and it does not
answer no for it. The rule that reads uses reports that it cannot answer for such a source, because no
declaration body of that package was read and the absence of uses is not a claim that the entity refers
to nothing. A rule that reads only what such a record can carry, such as the method specs an interface
states, answers for it as it does for any other record.

## Combining rules and answering a set of pairs

The consultation runs the rules in a fixed order and stops at the first yes. A yes wins over a rule that could not answer. A no outcome requires every rule that could answer to have answered no. When no rule answers yes and some rules could not answer, the answer names the last rule in that order that could not answer.

The outcome of the combination does not depend on the order the rules are consulted; only the rule reported for the undecided outcome depends on that order.

The iteration runs over the pairs the collection produces, ordered by the source identifier and then by the target identifier, so the same source with its attributes in a different order produces the same answered set.

A pair answered yes by two rules produces one fact. The reverse pair is a separate answer, so direction is preserved.

A pair answered as existing becomes no when the source stops carrying what the rule reads, so growth is scoped to the rule set and not to the code.

The domain of the relation is the pairs a rule can read. A pair is answered only when both members are records in the collection passed in; a reference to an entity with no record produces no pair, and a reference that names an entity the model records as an attribute rather than as a node names no pair either. The answers for pairs already in the collection are unchanged when the collection grows, and no node container is added to `pkg/graph`. The domain grows by making more records readable rather than by moving a pair across a package boundary.

The answer for a pair is read from the records passed in at the moment of the call rather than from a stored structure. Answering the same collection twice gives the same ordered set, and nothing is kept between calls.

A record is not a pair with itself: a record considered as both members of a pair produces no answer.

The answered set is ordered by the source identifier and then by the target identifier.

## Coverage of the rules by the test ground

Coverage is decided by which pairs the test ground produces for each rule. All four rules are covered by
the test ground, and a rule with no pair in the ground is reported as not covered. The rule that reads
references is covered by a pair whose members come from two packages, `main` to `strings.ToLower`, where
the second member is a record the second pass produced for an entity of a package the build reads from
type information only.
