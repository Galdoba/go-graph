package link

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Galdoba/go-graph/pkg/graph"
	"golang.org/x/tools/go/packages"
)

// groundPackagePath is the test ground at the repository root, which is the same ground the
// node model runs against.
func groundPackage(t *testing.T) *packages.Package {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the test ground")
	}
	root := filepath.Join(filepath.Dir(file), "..", "..")
	pkgs, err := packages.Load(&packages.Config{Mode: graph.LoadMode, Dir: root}, "./testdata")
	if err != nil {
		t.Fatalf("load the test ground: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("the test ground: expected one package, got %d", len(pkgs))
	}
	p := pkgs[0]
	if len(p.Errors) > 0 {
		t.Fatalf("the test ground: %v", p.Errors)
	}
	return p
}

// groundNodes builds the test ground into records, which is the collection the rules read.
func groundNodes(t *testing.T) []*graph.Node {
	t.Helper()
	nodes, _, err := graph.Build(groundPackage(t))
	if err != nil {
		t.Fatalf("build the test ground: %v", err)
	}
	return nodes
}

// groundWithForeignRecords builds the test ground and the records its references name in packages
// the build reads from type information only, which is the collection a pair whose members come from
// two packages is answered from.
func groundWithForeignRecords(t *testing.T) []*graph.Node {
	t.Helper()
	nodes := groundNodes(t)
	foreign, err := graph.SecondPass(groundPackage(t))
	if err != nil {
		t.Fatalf("read the records the references of the test ground name: %v", err)
	}
	return append(nodes, foreign...)
}

func nodeByName(nodes []*graph.Node, name string) *graph.Node {
	for _, n := range nodes {
		if n.Name == name {
			return n
		}
	}
	return nil
}

// stubNode builds a structure record by hand, so a test can name what the rule reads rather than
// rely on a loaded package.
func stubNode(pkg graph.PackageID, name string, st graph.SemanticType, refs ...graph.ID) *graph.Node {
	attrs := make([]graph.Attribute, 0, len(refs))
	for _, ref := range refs {
		attrs = append(attrs, graph.Attribute{Kind: graph.AttrEmbeds, Name: ref.Name, Ref: ref})
	}
	id := graph.ID{Package: pkg, Name: name}
	return &graph.Node{
		ID: id, Kind: graph.KindType, Name: name, Package: pkg,
		Parent: graph.ID{Package: pkg}, SemanticType: st, Attributes: attrs,
	}
}

// stubMethod builds a method record by hand, including its receiver and the parent that owns it.
func stubMethod(pkg graph.PackageID, receiver, name string) *graph.Node {
	id := graph.ID{Package: pkg, Receiver: receiver, Name: name}
	return &graph.Node{
		ID: id, Kind: graph.KindFunc, Name: name, Package: pkg,
		Parent: graph.ID{Package: pkg, Name: receiver}, Receiver: receiver,
		SemanticType: graph.SemanticMethod,
	}
}

// stubMethodSpec builds a method spec attribute in the form an interface records it, which is the
// signature read without the receiver.
func stubMethodSpec(name, signature string) graph.Attribute {
	return graph.Attribute{Kind: graph.AttrMethodSpec, Name: name, Value: signature}
}

// stubMethodAttribute builds a method attribute in the form the build records it on the node that
// owns it, pointing at the method node such a declaration also produces.
func stubMethodAttribute(pkg graph.PackageID, receiver, name, signature string) graph.Attribute {
	return graph.Attribute{
		Kind: graph.AttrMethod, Name: name, Value: signature,
		Ref: graph.ID{Package: pkg, Receiver: receiver, Name: name},
	}
}

// stubInterface builds an interface record by hand carrying method specs, so a test can name what
// the rule reads rather than rely on a loaded package.
func stubInterface(pkg graph.PackageID, name string, st graph.SemanticType, specs ...graph.Attribute) *graph.Node {
	id := graph.ID{Package: pkg, Name: name}
	return &graph.Node{
		ID: id, Kind: graph.KindType, Name: name, Package: pkg,
		Parent: graph.ID{Package: pkg}, SemanticType: st, Attributes: specs,
	}
}

// stubStructureMethods builds a structure record by hand carrying method attributes. The test ground
// produces this shape from real declarations, so a hand-built record stands only for the outcome the
// ground cannot produce.
func stubStructureMethods(pkg graph.PackageID, name string, methods ...graph.Attribute) *graph.Node {
	id := graph.ID{Package: pkg, Name: name}
	return &graph.Node{
		ID: id, Kind: graph.KindType, Name: name, Package: pkg,
		Parent: graph.ID{Package: pkg}, SemanticType: graph.SemanticStructure, Attributes: methods,
	}
}

// stubRule builds a rule by hand whose answer is fixed, so a test can name the order and the
// outcome rather than the relation type itself.
func stubRule(name string, answer func(name string, source, target *graph.Node) (bool, error)) Rule {
	return Rule{Name: name, Answer: answer}
}

func yesRule(name string) Rule {
	return stubRule(name, func(_ string, _, _ *graph.Node) (bool, error) { return true, nil })
}

func noRule(name string) Rule {
	return stubRule(name, func(_ string, _, _ *graph.Node) (bool, error) { return false, nil })
}

func undecidedRule(name string) Rule {
	return stubRule(name, func(n string, _, _ *graph.Node) (bool, error) {
		// The third outcome carries only the rule name.
		return false, &UndecidedError{Rule: n}
	})
}
