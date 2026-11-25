# Design

## Context

See proposal.md for motivation. The rule set in `pkg/link` consults rules in a fixed order, stopping at the first yes. Two rules already exist, both with a structure as the source. The new rule has an interface as the source, which is the first rule of that shape.

## Goals / Non-Goals

**Goals:**
- Define how a method spec on an interface is matched to a method node on a structure
- Keep the rule read-only against existing records, consistent with the existing rules

**Non-Goals:**
- Cross-package interface satisfaction — a method is matched only when it is recorded on the structure in the same record set
- Interface-to-interface satisfaction (an interface that embeds another interface)
- Methods reachable only through embedding — the model records the embedding, not the promoted method, so the rule answers no for a type that satisfies an interface through an embedded type

## Decisions

**Matching by name, receiver and signature.** A method spec on an interface records its name in `Name` and its signature in `Value` (`pkg/graph/build.go:295`), which for a method spec is `func() error`. A method declared on a type renders as `func (b Base) Read() error` through `types.TypeString`, so the receiver and the name are part of that text. The rule compares name, receiver and the signature without the receiver, which means the method side must be recorded in the same form as the method spec. Alternative: compare `types.Signature` values structurally. Rejected for this change, because the rule reads records rather than the type checker, and both sides already carry the same rendering.

**The method set is read from attributes, not from other records.** A declared method is a node, so a structure record currently carries nothing about its methods — `Base` has no attributes and its methods are separate records with `Parent = Base`. Reading those records is a lookup outside the pair, which the rule set forbids, so the build records each declared method as an attribute of the type node that owns it, with kind `method`, the signature without the receiver as the value, and a reference to the method node. Alternative: relax the requirement that a rule answers from the pair alone. Rejected, because it would let every rule reach records outside the pair for one relation.

**The referent rule must not read method attributes.** `answerReferentNode` scans attributes without filtering by kind, so a structure that carries method attributes would answer yes for the pair from that structure to its own method. The outcome for that pair is unchanged, because the method rule is consulted first, but the relation it names would change, so method attributes are excluded from the referent relation.

**Empty interfaces answer no.** An interface with no method specs has nothing to check. The rule answers no rather than yes, because a rule that cannot answer cannot answer no, and answering yes without any method to check is not a satisfaction decision — it is the absence of a test. Alternative: answer yes for any structure, matching the language semantics that every type satisfies `interface{}`. This is rejected because the rule must answer from the pair, and the interface node carries no method specs to compare.

**Pointer receivers need no extra work.** The receiver is recorded as the base named type: `pkg/graph/id.go:123` strips the pointer and resolves aliases, so `func (s1 *Struct1) Add5()` is recorded with receiver `Struct1`. The rule compares the recorded receiver, which is what the language counts for both the value and the pointer.

**Scenarios are checked on real code, not on hand-built records.** The rule's scenarios name the pairs the test ground produces, so each one is answered against a real declaration. The ground gains `type Adder interface { Add5() int }`, which turns the pointer receiver case into the pair `Adder` to `Struct1`, and the existing `Outer` turns the embedding case into the pair `Reader` to `Outer`. Hand-built records stay only for the outcome the ground cannot produce: the undecided outcome, because an existing requirement states that the test ground reports no unresolved node.

**Rule position in the order.** The new rule is consulted after `parentMethod` and `referentNode`. The first rule in order is consulted first, so an existing structure-to-method pair is not affected by this rule, which only fires when the source is an interface. Placing it last keeps the consultation order stable for existing pairs.

## Risks / Trade-offs

- **Text comparison of signatures depends on both sides being rendered the same way.** The build renders a signature through `types.TypeString` with the package qualifier, which resolves aliases, so `Counter` is written as `int`; a value taken from the syntax instead of from the type checker would keep the written name and the comparison would fail silently. Mitigation: record the method attribute value through the same rendering the method spec already uses.
- **Receiver names can diverge from the name a method spec carries.** A method declared on `Mirror` is recorded with receiver `Struct1` because the alias resolves. The rule compares the recorded receiver, so an interface method spec has to name the resolved type. Mitigation: the test ground has no embedded interfaces, and the scenario that answers no covers a spec the structure does not carry.
- **Recording a node as an attribute is new for the model.** A declared method is a node, and attributes have so far recorded entities that are not nodes. The change keeps the node record unchanged and adds the method attribute alongside it, so the same method appears twice, once as a record and once as an attribute of its owner.
