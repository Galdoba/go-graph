package link

import (
	"testing"

	"github.com/Galdoba/go-graph/pkg/graph"
)

const (
	graphPackagePath = "github.com/Galdoba/go-graph/pkg/graph"
	linkPackagePath  = "github.com/Galdoba/go-graph/pkg/link"
)

// TestAPairBetweenTwoPackagesOfTheSameRunIsAnsweredFromTheCollectionOfTheRun checks that the rules read the
// records of the run rather than the records of one package, so a pair whose two members are records of two
// packages the run read as targets is answered, and that the same reference names no pair when the collection
// holds only the package that carries it.
func TestAPairBetweenTwoPackagesOfTheSameRunIsAnsweredFromTheCollectionOfTheRun(t *testing.T) {
	pkgs := runPackages(t, "./pkg/graph", "./pkg/link")

	records := make(map[string][]*graph.Node, len(pkgs))
	for _, p := range pkgs {
		nodes, _, err := graph.Build(p)
		if err != nil {
			t.Fatalf("build %s: %v", p.PkgPath, err)
		}
		records[p.PkgPath] = nodes
	}

	onePackage := records[linkPackagePath]
	run := graph.Combine(append(append([]*graph.Node{}, records[graphPackagePath]...), onePackage...))

	carrying := nodeByName(onePackage, "Pair")
	if carrying == nil {
		t.Fatalf("the records of %s hold no structure that refers to a node of %s: %v",
			linkPackagePath, graphPackagePath, nodeNames(onePackage))
	}
	var named *graph.ID
	for _, attribute := range carrying.Attributes {
		if attribute.Ref.Package.PackagePath == graphPackagePath {
			named = &attribute.Ref
			break
		}
	}
	if named == nil {
		t.Fatalf("the record %v carries no reference to a node of %s: %v",
			carrying.ID, graphPackagePath, carrying.Attributes)
	}

	answered := false
	for _, pair := range Pairs(run) {
		if pair.Source == carrying.ID && pair.Target == *named {
			answered = true
			break
		}
	}
	if !answered {
		t.Errorf("the collection of the run answered %v, want the pair from %v to the record the reference names",
			Pairs(run), carrying.ID)
	}

	for _, pair := range Pairs(onePackage) {
		if pair.Source == *named || pair.Target == *named {
			t.Errorf("a pair was answered for %v, want no pair for a member the collection holds no record for", pair)
		}
	}
}
