package vault

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/Galdoba/go-graph/pkg/graph"
)

// projectRoot is the module root, which is the directory the loader has to run in to resolve the
// test ground by import path.
func projectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// loadGround builds one package of the test ground. The ground is loaded as a target, so its nodes
// are the nodes the vault writes files for.
func loadGround(t *testing.T, pattern string) []*graph.Node {
	t.Helper()

	found, err := packages.Load(&packages.Config{Mode: graph.LoadMode, Dir: projectRoot()}, pattern)
	return loadPackage(t, found, pattern, err)
}

// groundModule writes a package of its own in a module of its own, so a test can rewrite a
// declaration body and load the package again without changing the test ground the other tests
// assert against.
func groundModule(t *testing.T, source string) string {
	t.Helper()

	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module vaultground.test\n\ngo 1.27.4\n")
	writeFile(t, dir, "ground.go", source)
	return dir
}

// loadForeignPair builds one package of the graph fixtures as a target and the records the second
// pass produces for the names its references carry. The load is returned with both sets, so a test
// can take away the syntax the load reported and see what a record that states no region writes.
func loadForeignPair(t *testing.T, pattern string) (*packages.Package, []*graph.Node, []*graph.Node) {
	t.Helper()

	found, err := packages.Load(&packages.Config{Mode: graph.LoadMode, Dir: projectRoot()}, pattern)
	if err != nil {
		t.Fatalf("load %s: %v", pattern, err)
	}
	if len(found) != 1 {
		t.Fatalf("load %s: expected one package, got %d", pattern, len(found))
	}
	if len(found[0].Errors) > 0 {
		t.Fatalf("load %s: %v", pattern, found[0].Errors)
	}

	first, _, err := graph.Build(found[0])
	if err != nil {
		t.Fatalf("build %s: %v", pattern, err)
	}
	second, err := graph.SecondPass(found[0])
	if err != nil {
		t.Fatalf("second pass over %s: %v", pattern, err)
	}
	return found[0], first, second
}

// loadForeignPairWithoutSyntax reads the same pair with the syntax of the imported package taken away,
// which is the load that produces a record stating no region. The check then answers such a record
// against the entity the package declares, so the helper keeps the record and the package that declares
// the entity apart.
func loadForeignPairWithoutSyntax(t *testing.T, pattern string) (first, second []*graph.Node) {
	t.Helper()

	found, first, _ := loadForeignPair(t, pattern)
	imported := found.Imports["github.com/Galdoba/go-graph/pkg/graph/testdata/fixtures/shape"]
	if imported == nil || len(imported.Syntax) == 0 {
		t.Fatalf("the load reported no syntax for the package a reference names, so there is nothing to take away")
	}
	imported.Syntax = nil

	second, err := graph.SecondPass(found)
	if err != nil {
		t.Fatalf("second pass: %v", err)
	}
	return first, second
}

// loadDir builds the package that sits in the given directory.
func loadDir(t *testing.T, dir string) []*graph.Node {
	t.Helper()

	found, err := packages.Load(&packages.Config{Mode: graph.LoadMode, Dir: dir}, ".")
	return loadPackage(t, found, dir, err)
}

func loadPackage(t *testing.T, found []*packages.Package, what string, err error) []*graph.Node {
	t.Helper()

	if err != nil {
		t.Fatalf("load %s: %v", what, err)
	}
	if len(found) != 1 {
		t.Fatalf("load %s: expected one package, got %d", what, len(found))
	}
	if len(found[0].Errors) > 0 {
		t.Fatalf("load %s: %v", what, found[0].Errors)
	}

	nodes, _, err := graph.Build(found[0])
	if err != nil {
		t.Fatalf("build %s: %v", what, err)
	}
	return nodes
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func nodeByName(nodes []*graph.Node, name string) *graph.Node {
	for _, n := range nodes {
		if n.Name == name {
			return n
		}
	}
	return nil
}

// nodeByID resolves a node by the canonical form of its identifier, which is what a test has to use
// for a name shared by a package-level node and a method in the same package, because a lookup by
// name alone returns the first match.
func nodeByID(nodes []*graph.Node, id string) *graph.Node {
	for _, n := range nodes {
		if n.ID.String() == id {
			return n
		}
	}
	return nil
}

func nodeNames(nodes []*graph.Node) []string {
	var names []string
	for _, n := range nodes {
		names = append(names, n.ID.String())
	}
	return names
}
