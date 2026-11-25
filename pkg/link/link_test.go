package link

import (
	"errors"
	"testing"

	"github.com/Galdoba/go-graph/pkg/graph"
)

// methodNode finds the method node a receiver's named type carries, which is how the method
// that records a structure as its parent appears in the test ground.
func methodNode(nodes []*graph.Node, receiver, name string) *graph.Node {
	for _, n := range nodes {
		// The receiver is part of a method identifier, so a method on Struct1 is one record rather
		// than a package-level function of the same name.
		if n.Receiver == receiver && n.Name == name {
			return n
		}
	}
	return nil
}

func TestExistsYesNamesNoRule(t *testing.T) {
	nodes := groundNodes(t)
	struct1, reader := nodeByName(nodes, "Struct1"), nodeByName(nodes, "Reader")
	if struct1 == nil || reader == nil {
		t.Fatal("the test ground has no Struct1 or Reader node")
	}
	exists, err := Exists(struct1, reader)
	if !exists {
		t.Errorf("pair Struct1 to Reader answers %v, want the connection to exist", exists)
	}
	if err != nil {
		t.Errorf("pair Struct1 to Reader names %v, want no rule named with a yes", err)
	}
}

func TestExistsNoNamesNoRule(t *testing.T) {
	nodes := groundNodes(t)
	struct1, reader := nodeByName(nodes, "Struct1"), nodeByName(nodes, "Reader")
	if struct1 == nil || reader == nil {
		t.Fatal("the test ground has no Struct1 or Reader node")
	}
	// Only the outgoing direction carries the connection, so the reverse pair is a separate answer.
	exists, err := Exists(reader, struct1)
	if exists {
		t.Error("the reverse pair answers the connection as existing, want it not to exist")
	}
	if err != nil {
		t.Errorf("the reverse pair names %v, want no rule named with a no", err)
	}
}

func TestExistsThirdNamesTheLastRuleThatCouldNotAnswer(t *testing.T) {
	nodes := groundNodes(t)
	struct1 := nodeByName(nodes, "Struct1")
	if struct1 == nil {
		t.Fatal("the test ground has no Struct1 node")
	}
	// The test ground produces no unresolved node, so the third outcome is asserted on a record
	// built with the reserved value.
	unresolved := stubNode(struct1.ID.Package, "StubUnresolved", graph.SemanticUnresolved)
	exists, err := Exists(struct1, unresolved)
	if exists {
		t.Error("the pair with an unresolved member answers the connection as existing")
	}
	var u *UndecidedError
	if !errors.As(err, &u) {
		t.Fatalf("the pair with an unresolved member answers %v without naming a rule", err)
	}
	if u.Rule != satisfiedInterface.Name {
		t.Errorf("the answer names %q, want the last rule in the fixed order that could not answer", u.Rule)
	}
}

func TestRuleAnswersForThePairItIsGiven(t *testing.T) {
	nodes := groundNodes(t)
	thing := nodeByName(nodes, "Thing")
	read := methodNode(nodes, "Thing", "Read")
	if thing == nil || read == nil {
		t.Fatal("the test ground has no Thing node or method Read on Thing")
	}
	// The relation from a structure to its method is visible from the method's own record,
	// which is a member of the pair, not from a lookup of other nodes.
	exists, err := Exists(thing, read)
	if !exists || err != nil {
		t.Errorf("pair Thing to Read answers %v %v, want the connection to exist with no rule named", exists, err)
	}
}

func TestTheSamePairAnsweredTwiceGivesTheSameOutcome(t *testing.T) {
	nodes := groundNodes(t)
	struct1, reader := nodeByName(nodes, "Struct1"), nodeByName(nodes, "Reader")
	if struct1 == nil || reader == nil {
		t.Fatal("the test ground has no Struct1 or Reader node")
	}
	first, firstErr := Exists(struct1, reader)
	second, secondErr := Exists(struct1, reader)
	if first != second {
		t.Errorf("the same pair answered twice gives %v and %v, want one outcome", first, second)
	}
	if (firstErr == nil) != (secondErr == nil) {
		t.Errorf("the same pair answered twice names %v and %v, want one outcome", firstErr, secondErr)
	}
	// A variable is read by the rule that follows references, so a pair is answered no only when
	// the source carries no reference to the target.
	zero, base := nodeByName(nodes, "zeroStruct"), nodeByName(nodes, "Base")
	if zero == nil || base == nil {
		t.Fatal("the test ground has no zeroStruct or Base node")
	}
	exists, err := Exists(zero, base)
	if exists || err != nil {
		t.Errorf("a variable that carries no reference to its pair answers %v %v, want no connection with no rule named", exists, err)
	}
}

func TestTheMethodRuleAnswersItsGroundPair(t *testing.T) {
	nodes := groundNodes(t)
	thing := nodeByName(nodes, "Thing")
	read := methodNode(nodes, "Thing", "Read")
	base := nodeByName(nodes, "Base")
	baseRead := methodNode(nodes, "Base", "Read")
	if thing == nil || read == nil || base == nil || baseRead == nil {
		t.Fatal("the test ground has no Thing, Read, Base or method Read on Base")
	}
	for _, pair := range []struct{ source, target *graph.Node }{
		{thing, read}, {base, baseRead},
	} {
		exists, err := Exists(pair.source, pair.target)
		if !exists || err != nil {
			t.Errorf("a structure to its method answers %v %v, want the connection to exist", exists, err)
		}
	}
	// A source the rule does not cover answers no rather than the third outcome.
	reader := nodeByName(nodes, "Reader")
	if reader == nil {
		t.Fatal("the test ground has no Reader node")
	}
	exists, err := Exists(reader, thing)
	if exists || err != nil {
		t.Errorf("a pair whose source is not a structure answers %v %v, want no connection", exists, err)
	}
}

func TestTheReferentRuleAnswersItsGroundPair(t *testing.T) {
	nodes := groundNodes(t)
	outer := nodeByName(nodes, "Outer")
	base := nodeByName(nodes, "Base")
	if outer == nil || base == nil {
		t.Fatal("the test ground has no Outer or Base node")
	}
	exists, err := Exists(outer, base)
	if !exists || err != nil {
		t.Errorf("Outer to Base answers %v %v, want the connection to exist", exists, err)
	}
}

// TestTheReferenceRuleAnswersItsGroundPairs checks the pairs the test ground answers for this
// relation, which are read from the references a source carries whatever kind the source is.
func TestTheReferenceRuleAnswersItsGroundPairs(t *testing.T) {
	nodes := groundNodes(t)
	pairs := []struct{ source, target string }{
		{"describe", "Reader"},
		{"NewStruct1", "Struct1"},
		{"zeroStruct", "Struct1"},
		{"main", "Five"},
	}
	for _, pair := range pairs {
		source, target := nodeByName(nodes, pair.source), nodeByName(nodes, pair.target)
		if source == nil || target == nil {
			t.Fatalf("the test ground has no %s or %s node", pair.source, pair.target)
		}
		exists, err := Exists(source, target)
		if !exists || err != nil {
			t.Errorf("%s to %s answers %v %v, want the connection to exist with no rule named", pair.source, pair.target, exists, err)
		}
	}
}

// TestTheReferenceRuleReadsOnlyUses checks that a source whose uses name no node answers nothing, and
// that the kinds recording a name rather than a referent keep answering nothing, so the exclusions of
// the node model still hold for this relation.
func TestTheReferenceRuleReadsOnlyUses(t *testing.T) {
	nodes := groundNodes(t)
	minFunc, struct1 := nodeByName(nodes, "min"), nodeByName(nodes, "Struct1")
	if minFunc == nil || struct1 == nil {
		t.Fatal("the test ground has no min or Struct1 node")
	}
	if use := attributeByName(minFunc, "a"); use == nil || use.Kind != graph.AttrUse || use.Ref.Name != "" {
		t.Fatalf("the use of a parameter is not recorded as a value with no reference: %v", minFunc.Attributes)
	}
	if yes, err := carriedReference.Answer(carriedReference.Name, minFunc, struct1); yes || err != nil {
		t.Errorf("a source whose uses name no node answers %v %v, want no connection with no rule named", yes, err)
	}
	// The body of describe uses Read through an interface value, so the name resolves to a method spec
	// and carries no reference, which keeps the pair with the package-level function of the same name
	// answered no.
	describe, read := nodeByName(nodes, "describe"), nodeByName(nodes, "Read")
	if describe == nil || read == nil {
		t.Fatal("the test ground has no describe or Read node")
	}
	exists, err := Exists(describe, read)
	if exists || err != nil {
		t.Errorf("describe to the package-level Read answers %v %v, want no connection with no rule named", exists, err)
	}
	// The ground records the alias of a declared type on the node it resolves to, so a pair needs
	// records built by hand to point an alias and a method attribute at a second member.
	pkg := graph.PackageID{Name: "stub"}
	target := stubNode(pkg, "Referent", graph.SemanticVariable)
	source := stubNode(pkg, "Carrier", graph.SemanticVariable)
	source.Attributes = []graph.Attribute{
		{Kind: graph.AttrAlias, Name: "Mirror", Ref: target.ID},
		{Kind: graph.AttrMethod, Name: "Read", Value: "func() error", Ref: target.ID},
	}
	if yes, err := carriedReference.Answer(carriedReference.Name, source, target); yes || err != nil {
		t.Errorf("an alias and a method attribute answer %v %v for this relation, want no connection", yes, err)
	}
}

// TestAMethodAnswersNothingForItsReceiver checks that the receiver is read from the receiver property
// rather than from a reference, so a method answers nothing for the node its receiver names while the
// reverse pair still carries the method.
func TestAMethodAnswersNothingForItsReceiver(t *testing.T) {
	nodes := groundNodes(t)
	base, read := nodeByName(nodes, "Base"), methodNode(nodes, "Base", "Read")
	if base == nil || read == nil {
		t.Fatal("the test ground has no Base node or method Read on Base")
	}
	if attributeByName(read, "Base") != nil {
		t.Fatalf("the method records its receiver as a use: %v", read.Attributes)
	}
	exists, err := Exists(read, base)
	if exists || err != nil {
		t.Errorf("the method to the type it is declared on answers %v %v, want no connection with no rule named", exists, err)
	}
	if exists, err := Exists(base, read); !exists || err != nil {
		t.Errorf("the reverse pair answers %v %v, want the method carried by the rule that reads the parent", exists, err)
	}
}

// TestTheReferenceRuleAnswersForAnUndecidedRecord checks that the rule answers for a pair whose member
// reports unresolved without naming itself, because it reads no semantic value, and that the rule named
// for an undecided pair is unchanged.
func TestTheReferenceRuleAnswersForAnUndecidedRecord(t *testing.T) {
	pkg := graph.PackageID{Name: "stub"}
	target := stubNode(pkg, "Referent", graph.SemanticVariable)
	source := stubNode(pkg, "Carrier", graph.SemanticUnresolved)
	source.Attributes = []graph.Attribute{{Kind: graph.AttrUse, Name: "referent", Ref: target.ID}}
	exists, err := consult(rules, source, target)
	if !exists || err != nil {
		t.Errorf("a pair with an unresolved member that carries a reference answers %v %v, want the connection to exist with no rule named", exists, err)
	}
	// The rule is never unable to answer, so the undecided outcome still names the last rule that
	// reads a semantic value.
	bare := stubNode(pkg, "Carrier", graph.SemanticUnresolved)
	undecided, err := consult(rules, bare, target)
	if undecided {
		t.Error("a pair with an unresolved member answered the connection as existing")
	}
	var u *UndecidedError
	if !errors.As(err, &u) {
		t.Fatalf("the pair with an unresolved member answers %v without naming a rule", err)
	}
	if u.Rule != satisfiedInterface.Name {
		t.Errorf("the answer names %q, want the last rule in the fixed order that could not answer", u.Rule)
	}
}

func TestTheSatisfactionRuleAnswersTheStructuresThatSatisfyTheInterface(t *testing.T) {
	nodes := groundNodes(t)
	reader, adder, base, struct1 := nodeByName(nodes, "Reader"), nodeByName(nodes, "Adder"), nodeByName(nodes, "Base"), nodeByName(nodes, "Struct1")
	if reader == nil || adder == nil || base == nil || struct1 == nil {
		t.Fatal("the test ground has no Reader, Adder, Base or Struct1 node")
	}
	for _, pair := range []struct{ source, target *graph.Node }{
		{reader, base},
		// Add5 is declared with a pointer receiver, so the method is read through the named type the
		// receiver is based on.
		{adder, struct1},
	} {
		exists, err := Exists(pair.source, pair.target)
		if !exists || err != nil {
			t.Errorf("%s to %s answers %v %v, want the connection to exist", pair.source.Name, pair.target.Name, exists, err)
		}
	}
	// Only the outgoing direction carries the connection, so the structure does not satisfy the
	// interface in the reverse pair.
	if exists, err := Exists(base, reader); exists || err != nil {
		t.Errorf("Base to Reader answers %v %v, want no connection", exists, err)
	}
}

func TestTheSatisfactionRuleAnswersNoForAnIncompleteMethodSet(t *testing.T) {
	nodes := groundNodes(t)
	reader, struct1, thing, outer, base := nodeByName(nodes, "Reader"), nodeByName(nodes, "Struct1"), nodeByName(nodes, "Thing"), nodeByName(nodes, "Outer"), nodeByName(nodes, "Base")
	if reader == nil || struct1 == nil || thing == nil || outer == nil || base == nil {
		t.Fatal("the test ground has no Reader, Struct1, Thing, Outer or Base node")
	}
	// Struct1 owns no method named Read, Thing owns a method named Read with a different signature,
	// and Outer reaches Read only through an embedding, which the model records as an embedding
	// rather than as a method of Outer.
	for _, target := range []*graph.Node{struct1, thing, outer} {
		exists, err := Exists(reader, target)
		if exists || err != nil {
			t.Errorf("Reader to %s answers %v %v, want no connection with no rule named", target.Name, exists, err)
		}
	}
	// An interface with no method spec states nothing to check, so the rule answers no rather than
	// the third outcome.
	empty := nodeByName(nodes, "EmptyInterface")
	if empty == nil {
		t.Fatal("the test ground has no EmptyInterface node")
	}
	if exists, err := Exists(empty, base); exists || err != nil {
		t.Errorf("EmptyInterface to Base answers %v %v, want no connection", exists, err)
	}
}

func TestMethodAttributesNameNoReferent(t *testing.T) {
	nodes := groundNodes(t)
	base := nodeByName(nodes, "Base")
	baseRead := methodNode(nodes, "Base", "Read")
	if base == nil || baseRead == nil {
		t.Fatal("the test ground has no Base node or method Read on Base")
	}
	// The method is a node, so it is also recorded as an attribute of the node that owns it. Only the
	// rule that reads a method by its parent answers this pair.
	if !recordsMethod(base, graph.Attribute{Kind: graph.AttrMethod, Name: "Read", Value: "func() error"}) {
		t.Fatalf("Base carries no method attribute for Read: %v", base.Attributes)
	}
	if yes, err := referentNode.Answer(referentNode.Name, base, baseRead); yes || err != nil {
		t.Errorf("the referent rule answers a method attribute %v %v, want no", yes, err)
	}
	if yes, err := parentMethod.Answer(parentMethod.Name, base, baseRead); !yes || err != nil {
		t.Errorf("the method rule answers Base to Read %v %v, want the connection to exist", yes, err)
	}
}

func TestTheSatisfactionRuleNamesItselfForAnUndecidedRecord(t *testing.T) {
	pkg := graph.PackageID{Name: "stub"}
	// The test ground produces no unresolved record, so the third outcome is asserted on records
	// built by hand, with the method set the rule reads.
	unresolved := stubInterface(pkg, "StubInterface", graph.SemanticUnresolved, stubMethodSpec("Read", "func() error"))
	satisfied := stubStructureMethods(pkg, "StubBase", stubMethodAttribute(pkg, "StubBase", "Read", "func() error"))
	exists, err := consult(rules, unresolved, satisfied)
	if exists {
		t.Error("the pair with an unresolved member answers the connection as existing")
	}
	var u *UndecidedError
	if !errors.As(err, &u) {
		t.Fatalf("the pair with an unresolved member answers %v without naming a rule", err)
	}
	if u.Rule != satisfiedInterface.Name {
		t.Errorf("the answer names %q, want the last rule in the fixed order that could not answer", u.Rule)
	}
}

func TestHandBuiltRecordsCarryTheMethodSetTheRuleReads(t *testing.T) {
	pkg := graph.PackageID{Name: "stub"}
	spec := stubMethodSpec("Describe", "func() string")
	reader := stubInterface(pkg, "StubReader", graph.SemanticInterface, spec)
	base := stubStructureMethods(pkg, "StubBase", stubMethodAttribute(pkg, "StubBase", "Describe", "func() string"))
	if len(reader.Attributes) != 1 || reader.Attributes[0].Kind != graph.AttrMethodSpec {
		t.Fatalf("the interface record carries %v, want a method spec", reader.Attributes)
	}
	if len(base.Attributes) != 1 || base.Attributes[0].Kind != graph.AttrMethod {
		t.Fatalf("the structure record carries %v, want a method attribute", base.Attributes)
	}
	if !recordsMethod(base, spec) {
		t.Errorf("the structure does not read as carrying the method the spec states: %v", base.Attributes)
	}
	if yes, err := satisfiedInterface.Answer(satisfiedInterface.Name, reader, base); !yes || err != nil {
		t.Errorf("the hand-built pair answers %v %v, want the connection to exist", yes, err)
	}
}

func TestTheAliasReceiverIsReadThroughTheNodeItResolvesTo(t *testing.T) {
	nodes := groundNodes(t)
	struct1 := nodeByName(nodes, "Struct1")
	alias := methodNode(nodes, "Struct1", "DescribeAlias")
	if struct1 == nil || alias == nil {
		t.Fatal("the test ground has no Struct1 node or method DescribeAlias on the alias receiver")
	}
	exists, err := Exists(struct1, alias)
	if !exists || err != nil {
		t.Errorf("Struct1 to DescribeAlias answers %v %v, want the connection to exist through the resolved node", exists, err)
	}
}

func TestAnUncoveredRelationAnswersNo(t *testing.T) {
	nodes := groundNodes(t)
	struct1 := nodeByName(nodes, "Struct1")
	base := nodeByName(nodes, "Base")
	if struct1 == nil || base == nil {
		t.Fatal("the test ground has no Struct1 or Base node")
	}
	// Base is not a method of Struct1 and is not a referent of one of its attributes.
	exists, err := Exists(struct1, base)
	if exists || err != nil {
		t.Errorf("a relation no rule covers answers %v %v, want no connection with no rule named", exists, err)
	}
}

func TestTheConsultationStopsAtTheFirstYes(t *testing.T) {
	// The first rule that answers yes decides the answer, and a later rule that could not answer is
	// not consulted.
	source := stubNode(graph.PackageID{Name: "stub"}, "a", graph.SemanticStructure)
	target := stubMethod(graph.PackageID{Name: "stub"}, "a", "m")
	rs := []Rule{yesRule("first"), undecidedRule("second")}
	exists, err := consult(rs, source, target)
	if !exists {
		t.Error("the consultation did not stop at the first yes")
	}
	if err != nil {
		t.Errorf("the consultation named %v, want no rule named with a yes", err)
	}
}

func TestTheOutcomeDoesNotDependOnTheOrderOfTheRules(t *testing.T) {
	source := stubNode(graph.PackageID{Name: "stub"}, "a", graph.SemanticStructure)
	target := stubMethod(graph.PackageID{Name: "stub"}, "a", "m")
	firstExists, firstErr := consult([]Rule{undecidedRule("u1"), noRule("n")}, source, target)
	secondExists, secondErr := consult([]Rule{noRule("n"), undecidedRule("u1")}, source, target)
	if firstExists != secondExists {
		// The outcome of the combination is the same for any order; only the reported rule changes.
		t.Errorf("the outcome changed with the order: %v and %v", firstExists, secondExists)
	}
	var e1, e2 *UndecidedError
	if !errors.As(firstErr, &e1) || !errors.As(secondErr, &e2) {
		t.Fatalf("the order changed the outcome to a missing rule: %v and %v", firstErr, secondErr)
	}
	if e1.Rule != e2.Rule || e1.Rule != "u1" {
		t.Errorf("the reported rule is %q and %q, want the only rule that could not answer", e1.Rule, e2.Rule)
	}
}

func TestTheAnsweredSetIsOrderedByThePair(t *testing.T) {
	pkg := graph.PackageID{Name: "stub"}
	reader := stubNode(pkg, "Reader", graph.SemanticInterface)
	base := stubNode(pkg, "Base", graph.SemanticStructure)
	// The same source with its attributes in a different order produces the same answered set.
	first := stubNode(pkg, "Struct", graph.SemanticStructure, reader.ID, base.ID)
	second := stubNode(pkg, "Struct", graph.SemanticStructure, base.ID, reader.ID)
	set1 := Pairs([]*graph.Node{first, reader, base})
	set2 := Pairs([]*graph.Node{second, reader, base})
	if len(set1) != len(set2) {
		t.Fatalf("a different attribute order changed the set size: %v and %v", set1, set2)
	}
	for i := range set1 {
		if set1[i].Source.String() != set2[i].Source.String() || set1[i].Target.String() != set2[i].Target.String() {
			t.Errorf("a different attribute order changed the answered set: %v and %v", set1, set2)
		}
	}
	for i := 1; i < len(set1); i++ {
		if set1[i-1].Source.String() > set1[i].Source.String() {
			t.Errorf("the answered set is not ordered by the source identifier: %v", set1)
		} else if set1[i-1].Source.String() == set1[i].Source.String() && set1[i-1].Target.String() > set1[i].Target.String() {
			t.Errorf("the answered set is not ordered by the target identifier: %v", set1)
		}
	}
}

func TestTwoRulesThatAnswerTheSamePairGiveOneFact(t *testing.T) {
	pkg := graph.PackageID{Name: "stub"}
	// The method records the structure as its parent, and the structure records the method as a
	// referent of one of its attributes, so both rules answer yes for the same pair.
	method := stubMethod(pkg, "Struct", "m")
	struct1 := stubNode(pkg, "Struct", graph.SemanticStructure, method.ID)
	set := Pairs([]*graph.Node{struct1, method})
	if len(set) != 1 {
		t.Fatalf("a pair answered yes by two rules produced %v facts, want one", set)
	}
	if set[0].Source.String() != struct1.ID.String() || set[0].Target.String() != method.ID.String() {
		t.Errorf("the fact is %v, want the outgoing direction of the pair", set[0])
	}
	// The reverse pair is answered separately as not existing, so only the outgoing direction appears in the answered set.
}

func TestAGrowthOfTheRuleSetProtectsOnlyTheRuleSet(t *testing.T) {
	nodes := groundNodes(t)
	struct1 := nodeByName(nodes, "Struct1")
	reader := nodeByName(nodes, "Reader")
	if struct1 == nil || reader == nil {
		t.Fatal("the test ground has no Struct1 or Reader node")
	}
	yes, err := consult(rules, struct1, reader)
	if !yes || err != nil {
		t.Fatalf("Struct1 to Reader answers %v %v, want the connection to exist", yes, err)
	}
	// A change to the source stops producing what the rule reads, so the outcome is read again from
	// the source rather than from the growth of the rule set.
	bare := stubNode(struct1.ID.Package, "Struct1", graph.SemanticStructure)
	no, err := consult(rules, bare, reader)
	if no || err != nil {
		t.Errorf("the changed source still carries the referent: %v %v, want no connection", no, err)
	}
}

func TestAReferentWithoutARecordProducesNoPair(t *testing.T) {
	pkg := graph.PackageID{Name: "stub"}
	base := stubNode(pkg, "Base", graph.SemanticStructure)
	// A referent with no record is not a member of any pair the rules read.
	referent := graph.ID{Package: pkg, Name: "Referent"}
	struct1 := stubNode(pkg, "Struct1", graph.SemanticStructure, base.ID, referent)
	set := Pairs([]*graph.Node{struct1, base})
	if len(set) != 1 {
		t.Fatalf("the collection with an external referent produced %v pairs, want one", set)
	}
	if set[0].Source.String() != struct1.ID.String() || set[0].Target.String() != base.ID.String() {
		t.Errorf("the answered pair is %v, want Struct1 to Base", set[0])
	}
	// A referent that has a record is a member of the collection, so its pair is answered with the
	// pairs already in the collection unchanged.
	grown := stubNode(pkg, "Referent", graph.SemanticVariable)
	grownSet := Pairs([]*graph.Node{struct1, base, grown})
	if len(grownSet) != 2 {
		t.Fatalf("the collection with the referent's record produced %v pairs, want two", grownSet)
	}
	for _, p := range grownSet {
		if p.Source.String() == struct1.ID.String() && p.Target.String() == referent.String() {
			continue
		}
		if p.Source.String() == struct1.ID.String() && p.Target.String() == base.ID.String() {
			continue
		}
		t.Errorf("the grown collection changed the answered set with %v", p)
	}
}

func TestTheAnsweredSetIsReproducibleFromTheRecordsAlone(t *testing.T) {
	nodes := groundNodes(t)
	first := Pairs(nodes)
	second := Pairs(nodes)
	if len(first) != len(second) {
		t.Fatalf("the same collection answered twice gave %v and %v", first, second)
	}
	for i := range first {
		if first[i].Source.String() != second[i].Source.String() || first[i].Target.String() != second[i].Target.String() {
			t.Fatalf("the same collection answered twice gave %v and %v", first, second)
		}
	}
}

// TestTheGroundAnsweredSetIsAssertedAsAWhole pins the answered set the test ground produces, so the
// growth of the rule that reads references is visible in the test rather than inferred rule by rule.
func TestTheGroundAnsweredSetIsAssertedAsAWhole(t *testing.T) {
	want := []string{
		"Adder -> Struct1",
		"Base -> Base.Describe",
		"Base -> Base.Read",
		"Box -> Box.PickVal",
		"NewStruct1 -> Struct1",
		"Option -> Struct1",
		"Outer -> Base",
		"Pick -> Box",
		"Reader -> Base",
		"Struct1 -> Reader",
		"Struct1 -> Struct1.Add5",
		"Struct1 -> Struct1.DescribeAlias",
		"Thing -> Thing.Read",
		"describe -> Reader",
		"main -> Base",
		"main -> Base.Read",
		"main -> Box",
		"main -> EmptyInterface",
		"main -> Five",
		"main -> LayoutMono",
		"main -> LayoutStereo",
		"main -> Name",
		"main -> NewStruct1",
		"main -> Option",
		"main -> Outer",
		"main -> Pick",
		"main -> StringMap",
		"main -> Struct1",
		"main -> Struct1.Add5",
		"main -> describe",
		"main -> first",
		"main -> min",
		"main -> namedFunc",
		"main -> second",
		"main -> two",
		"main -> zeroStruct",
		"zeroStruct -> Struct1",
	}
	got := []string{}
	for _, p := range Pairs(groundNodes(t)) {
		got = append(got, pairName(p))
	}
	if len(got) != len(want) {
		t.Fatalf("the test ground answered %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("the answered set is %v, want %v", got, want)
			break
		}
	}
}

func TestARecordIsNotAPairWithItself(t *testing.T) {
	nodes := groundNodes(t)
	struct1 := nodeByName(nodes, "Struct1")
	if struct1 == nil {
		t.Fatal("the test ground has no Struct1 node")
	}
	// The alias of a declared type is recorded as an attribute of the node it resolves to, so the
	// alias name names the node that owns it rather than a relation between two records.
	alias := attributeByName(struct1, "Mirror")
	if alias == nil || alias.Kind != graph.AttrAlias || alias.Ref != struct1.ID {
		t.Fatalf("alias Mirror is not an attribute of the node it resolves to: %v", struct1.Attributes)
	}
	for _, p := range Pairs(nodes) {
		if p.Source.String() == struct1.ID.String() && p.Target.String() == struct1.ID.String() {
			t.Errorf("the reflexive candidate produced the answered pair %v", p)
		}
	}
}

func TestTheGroundCoversEveryRule(t *testing.T) {
	nodes := groundNodes(t)
	if got := Coverage(rules, nodes); len(got) != 0 {
		t.Errorf("the test ground does not produce a pair for %v", got)
	}
	extra := append(rules, noRule("uncovered relation"))
	if got := Coverage(extra, nodes); len(got) != 1 || got[0] != "uncovered relation" {
		t.Errorf("a rule with no pair in the ground reported %v, want only itself", got)
	}
	if len(rules) != 4 {
		t.Errorf("adding a rule changed the rule set to %v", rules)
	}
}

func attributeByName(node *graph.Node, name string) *graph.Attribute {
	for _, a := range node.Attributes {
		if a.Name == name {
			return &a
		}
	}
	return nil
}

// pairName writes one answered pair with the names the documentation uses, which names a method by its
// receiver rather than by its name alone.
func pairName(p Pair) string {
	return pairPart(p.Source) + " -> " + pairPart(p.Target)
}

func pairPart(id graph.ID) string {
	if id.Receiver != "" {
		return id.Receiver + "." + id.Name
	}
	return id.Name
}
