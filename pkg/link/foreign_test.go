package link

import (
	"errors"
	"testing"

	"github.com/Galdoba/go-graph/pkg/graph"
)

// TestTheGroundAnswersAPairWhoseMembersComeFromTwoPackages checks that the rule that reads references
// is covered by a pair whose members come from the package the build read as a target and from a
// package the build reads from type information only.
func TestTheGroundAnswersAPairWhoseMembersComeFromTwoPackages(t *testing.T) {
	nodes := groundWithForeignRecords(t)

	if !answeredFor(nodes, "main", "ToLower") {
		t.Fatalf("the test ground answered %v, want the pair from main to the entity the imported package declares",
			Pairs(nodes))
	}
	if got := Coverage(rules, nodes); len(got) != 0 {
		t.Errorf("the test ground does not produce a pair for %v", got)
	}
}

// TestAnUnexportedSourceAnswersAPairWithARecordProducedFromAReference checks that the exported filter
// reaches only records produced from a reference, so a pair inside the package the build read as a
// target is answered whatever the exported value of its target is.
func TestAnUnexportedSourceAnswersAPairWithARecordProducedFromAReference(t *testing.T) {
	nodes := groundWithForeignRecords(t)

	source := nodeByName(nodes, "main")
	target := nodeByName(nodes, "ToLower")
	if source == nil || target == nil {
		t.Fatalf("the collection holds no pair to answer, got %v", nodeNames(nodes))
	}
	if source.Properties.Exported {
		t.Fatalf("the source %v is exported, so the test does not read the filter", source.ID)
	}
	if !target.Properties.Exported {
		t.Errorf("the record produced from a reference is not exported, which no reference from the package can name")
	}
	yes, err := Exists(source, target)
	if !yes || err != nil {
		t.Errorf("the pair %v -> %v answered %v (%v), want it answered as existing", source.ID, target.ID, yes, err)
	}
}

// TestARuleCannotAnswerForARecordThatStatesNoUses checks that the rule that reads uses reports the
// third outcome rather than no, because a record read from a name never read a declaration body.
func TestARuleCannotAnswerForARecordThatStatesNoUses(t *testing.T) {
	nodes := groundWithForeignRecords(t)

	foreign := nodeByName(nodes, "ToLower")
	other := nodeByName(nodes, "Struct1")
	if foreign == nil || other == nil {
		t.Fatalf("the collection holds no pair to answer, got %v", nodeNames(nodes))
	}
	if foreign.Provenance != graph.ProvenanceFromName {
		t.Fatalf("the record %v does not state that its values were read from a name", foreign.ID)
	}

	yes, err := Exists(foreign, other)
	if yes {
		t.Fatal("a rule answered for a record that states no uses")
	}
	var undecided *UndecidedError
	if !errors.As(err, &undecided) {
		t.Fatalf("the outcome for the pair is %v, want the rule that cannot answer", err)
	}
	if undecided.Rule != carriedReference.Name {
		t.Errorf("the outcome names rule %q, want the rule that reads uses", undecided.Rule)
	}
}

// TestARuleThatReadsShapeAnswersForARecordProducedFromAReference checks that a rule reading only what
// such a record can carry answers for it as it does for any other record. The records are built by
// hand, because the test ground cannot produce a foreign interface that a structure of the built
// package satisfies.
func TestARuleThatReadsShapeAnswersForARecordProducedFromAReference(t *testing.T) {
	foreign := graph.PackageID{ModulePath: "mod.test", PackagePath: "mod.test/shape", Name: "shape"}
	built := graph.PackageID{ModulePath: "mod.test", PackagePath: "mod.test/use", Name: "use"}

	record := stubInterface(foreign, "Reader", graph.SemanticInterface, stubMethodSpec("Read", "func() error"))
	record.Provenance = graph.ProvenanceFromName
	structure := stubStructureMethods(built, "Local", stubMethodAttribute(built, "Local", "Read", "func() error"))

	yes, err := Exists(record, structure)
	if !yes || err != nil {
		t.Errorf("the pair %v -> %v answered %v (%v), want it answered as existing", record.ID, structure.ID, yes, err)
	}
}

// TestAPairNeedsARecordForBothOfItsMembers checks that a relation is not answered for a member the
// second pass produced no record for, which is the case for a foreign method whose receiver type no
// reference named.
func TestAPairNeedsARecordForBothOfItsMembers(t *testing.T) {
	foreign := graph.PackageID{ModulePath: "mod.test", PackagePath: "mod.test/shape", Name: "shape"}
	method := stubMethod(foreign, "Thing", "Read")
	method.Provenance = graph.ProvenanceFromName

	nodes := []*graph.Node{method, stubNode(foreign, "Other", graph.SemanticStructure)}
	if got := Pairs(nodes); len(got) != 0 {
		t.Errorf("the answered set holds %v, want no pair for a member with no record of its own", got)
	}
	if method.Parent.Name != "Thing" {
		t.Fatalf("the method record names no parent, so the test does not read the missing member: %v", method.Parent)
	}
}

func answeredFor(nodes []*graph.Node, sourceName, targetName string) bool {
	source, target := nodeByName(nodes, sourceName), nodeByName(nodes, targetName)
	if source == nil || target == nil {
		return false
	}
	yes, err := Exists(source, target)
	return yes && err == nil
}

func nodeNames(nodes []*graph.Node) []string {
	names := make([]string, 0, len(nodes))
	for _, n := range nodes {
		names = append(names, n.Name)
	}
	return names
}
