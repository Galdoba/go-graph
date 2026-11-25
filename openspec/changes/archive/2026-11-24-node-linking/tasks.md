# Tasks

## 1. The answer for a pair

- [x] 1.1 Create `pkg/link` with the three outcomes an answer can have and the rule identity reported only for the outcome that says no rule could answer, and verify a test that asserts the three shapes: yes with no rule named, no with no rule named, and the third outcome naming the rule that could not answer.
- [x] 1.2 Write the section of `docs/node-links.md` that states the three outcomes and that a rule that cannot answer does not answer no, and verify that the documented case matches the test from 1.1.

## 2. The rule contract

- [x] 2.1 Define a rule as a decision over one pair that reads only what the two members carry, and verify a test where a rule answers for a pair whose target records the source as its parent, showing that the answer comes from the pair and not from a lookup of other nodes.
- [x] 2.2 Verify with a test that the same pair answered twice on the same source gives the same outcome, and that a rule whose condition does not match answers no rather than the third outcome.
- [x] 2.3 Write the section of `docs/node-links.md` that states what a rule may read and why a rule that needs a lookup outside the pair is not part of the rule set, and verify it against the tests from 2.1 and 2.2.

## 3. The two rules decided here

- [x] 3.1 Implement the rule that answers for a structure and the method that records it as its parent, and verify it against the test ground pair `Thing` to `Read` while asserting that a pair whose source is not a structure answers no.
- [x] 3.2 Implement the rule that answers for a structure and the node one of its attributes refers to, and verify it against the test ground pair `Struct1` to `Reader`.
- [x] 3.3 Verify with a test that a method written with an alias as its receiver is answered through the node the alias resolves to, on the ground pair `Struct1` to `DescribeAlias`.
- [x] 3.4 Verify with a test that a pair for a relation type neither rule covers answers no, and that a pair whose source or target reports `unresolved` gives the third outcome naming the rule, on a record built with the reserved value since the test ground produces no `unresolved` node.
- [x] 3.5 Write the section of `docs/node-links.md` that lists the rules decided here and the alias receiver rule, and verify that each documented pair matches the test that asserts it.

## 4. Combining rules and answering a set of pairs

- [x] 4.1 Implement the consultation in a fixed order that stops at the first yes, and verify with tests that a yes wins over a rule that could not answer, that a no outcome requires every rule that could answer to have answered no, and that no yes plus rules that could not answer gives the third outcome naming the last rule in that order that could not answer.
- [x] 4.2 Verify with a test that the outcome of the combination does not depend on the order in which the rules are consulted, while the rule reported for the undecided outcome is the last one that could not answer in that order.
- [x] 4.3 Implement the iteration over pairs ordered by the source identifier and then by the target identifier, and verify with a test that the answered set for a source is identical when its attributes are recorded in a different order.
- [x] 4.4 Verify with a test that a pair answered yes by two rules produces one fact and that the reverse pair is answered separately, so direction is preserved.
- [x] 4.5 Verify with a test that a pair answered as existing becomes no when the source stops carrying what the rule reads, so growth is scoped to the rule set and not to the code.
- [x] 4.6 Answer a pair only when both of its members are records in the collection passed in, and verify with a test that a reference to an entity with no record produces no pair, that the answers for pairs already in the collection are unchanged, and that no node container is added to `pkg/graph`.
- [x] 4.7 Read the answer for a pair from the records passed in at the moment of the call rather than from a stored structure, and verify with a test that answering the same collection twice gives the same ordered set while nothing is kept between calls.
- [x] 4.8 Answer nothing for a record considered as both members of a pair, and verify with a test that such a case produces no answer.
- [x] 4.9 Write the section of `docs/node-links.md` that states the combination rule, which undecided rule the answer reports, the monotone growth of the answered set, the growth that a code change does not protect, the domain of the relation, that the answer is read at the moment of the call rather than kept, that a record is not a pair with itself, and the ordering of the answered pairs, and verify it against the tests from 4.1 to 4.8.

## 5. Coverage of the rules by the test ground

- [x] 5.1 Report a rule that no pair of the test ground answers as not covered, and verify with a test that the current ground reports both rules as covered and that adding a rule with no pair in the ground reports it as not covered.
- [x] 5.2 Write the section of `docs/node-links.md` that states how coverage is decided and lists the ground pairs for each rule, and verify the listed pairs against the test from 5.1.

## 6. Integration check

- [x] 6.1 Run `go test ./...` and verify that the whole suite passes with the new package in place, that `pkg/graph` is unchanged, and that `docs/node-model.md` points from its out-of-scope note to `docs/node-links.md`.

## Workflow follow-up

- Archive the change once the review requirements are satisfied; the archive creates `openspec/specs/node-links/spec.md` from the delta.
- Verify the archived result, and keep the deferred work as its own change: a pair whose target is an identifier without a record, and how a link is written into a note.
