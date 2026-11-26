package vault

import (
	"os"
	"path"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/Galdoba/go-graph/pkg/graph"
	"github.com/Galdoba/go-graph/pkg/link"
)

// TestARunOverTwoPackagesOfTheRepositoryAnswersTheConnectionBetweenThemFromTheRun writes the notes one
// run produces for two packages of the repository read as targets at one time. The collection of the run
// holds the records of both, so the connection a reference from `pkg/link` carries to a node `pkg/graph`
// declares is answered, which the records of the package that declares the node could not show on its
// own, and the note written for a shared entity is the record read from source. One note is written per
// record of the collection, in the folder that record's package maps to, and no file for a name the run
// produced no record for. The vault root is a constant, so the run is reproduced in a temporary folder
// rather than in the vault the command writes.
func TestARunOverTwoPackagesOfTheRepositoryAnswersTheConnectionBetweenThemFromTheRun(t *testing.T) {
	root := t.TempDir()
	found, err := packages.Load(&packages.Config{Mode: graph.LoadMode, Dir: projectRoot()}, "./pkg/graph", "./pkg/link")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("load: expected two packages, got %d", len(found))
	}

	// The pair between the two packages can only go from the package that reads the other, because a
	// record produced from a name states no uses to answer the rule that reads references from.
	graphPackage := found[0]
	linkPackage := found[1]
	if graphPackage.PkgPath != "github.com/Galdoba/go-graph/pkg/graph" || linkPackage.PkgPath != "github.com/Galdoba/go-graph/pkg/link" {
		t.Fatalf("the load returned %s and %s, want the graph package and the link package", graphPackage.PkgPath, linkPackage.PkgPath)
	}
	if linkPackage.Imports[graphPackage.PkgPath] == nil {
		t.Fatalf("pkg/link does not import pkg/graph, so the run reads no reference between the two")
	}

	var own [2][]*graph.Node
	var collection []*graph.Node
	for i, p := range []*packages.Package{graphPackage, linkPackage} {
		if len(p.Errors) > 0 {
			t.Fatalf("load %s: %v", p.PkgPath, p.Errors)
		}
		built, _, err := graph.Build(p)
		if err != nil {
			t.Fatalf("build %s: %v", p.PkgPath, err)
		}
		fromNames, err := graph.SecondPass(p)
		if err != nil {
			t.Fatalf("second pass over %s: %v", p.PkgPath, err)
		}
		own[i] = append(append([]*graph.Node{}, built...), fromNames...)
		collection = append(collection, own[i]...)
	}
	collection = graph.Combine(collection)

	seen := map[string]bool{}
	for _, record := range collection {
		if seen[record.ID.String()] {
			t.Errorf("the collection holds two entries for %s", record.ID)
		}
		seen[record.ID.String()] = true
	}

	nodeRecord := nodeByID(collection, "github.com~Galdoba~go-graph~pkg~graph~graph.Node")
	if nodeRecord == nil {
		t.Fatalf("the collection holds no record for the node pkg/link references: %v", nodeNames(collection))
	}
	if nodeRecord.Provenance != graph.ProvenanceFromSource {
		t.Fatalf("the record of %s was read from %v, want the record the run read as a target", nodeRecord.ID, nodeRecord.Provenance)
	}

	pairFound := false
	for _, pair := range link.Pairs(collection) {
		if pair.Target != nodeRecord.ID {
			continue
		}
		if nodeByID(collection, pair.Source.String()) == nil {
			t.Fatalf("the pair to %s names %s, which is not in the collection", nodeRecord.ID, pair.Source)
		}
		pairFound = true
	}
	if !pairFound {
		t.Fatalf("the collection of the run answers no pair that names %s", nodeRecord.ID)
	}
	// The records of the package that declares the node name no record of the other package, so this pair
	// is answered only from the collection the run holds for both of them.
	for _, pair := range link.Pairs(own[0]) {
		if pair.Source.Package.PackagePath == linkPackage.PkgPath || pair.Target.Package.PackagePath == linkPackage.PkgPath {
			t.Errorf("the records of the package that declares %s answer the pair %v, want no pair that names a record of %s", nodeRecord.ID, pair, linkPackage.PkgPath)
		}
	}

	targets := []graph.PackageID{graph.PackageIDOf(graphPackage), graph.PackageIDOf(linkPackage)}
	if _, err := Rebuild(root, targets, collection); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	// One note per record of the collection, in the folder that record's package maps to.
	for _, record := range collection {
		file := NotePath(root, record)
		if _, err := os.Stat(file); err != nil {
			t.Errorf("no note written for %s at %s: %v", record.ID, file, err)
		}
	}

	// A reference that names an entity the definition of a node does not cover gets no file. The references
	// of pkg/link read graph.Node through its field Ref, which is not a node, so the folder the graph
	// package maps to holds the notes of the records of this run and nothing for that field.
	graphFolder := Folder(root, nodeRecord.Package)
	entries, err := os.ReadDir(graphFolder)
	if err != nil {
		t.Fatalf("read folder %s: %v", graphFolder, err)
	}
	produced := map[string]bool{}
	for _, record := range collection {
		if record.Package.PackagePath == graphPackage.PkgPath {
			produced[FileName(record)] = true
		}
	}
	if len(entries) != len(produced) {
		t.Errorf("folder %s holds %d files, want the %d notes this run produced for %s", graphFolder, len(entries), len(produced), graphPackage.PkgPath)
	}
	for _, entry := range entries {
		if !produced[entry.Name()] {
			t.Errorf("folder %s holds a file for a name no record exists for: %s", graphFolder, entry.Name())
		}
	}
	if _, err := os.Stat(path.Join(graphFolder, "Ref.md")); err == nil {
		t.Errorf("a file was written for the field Ref of %s that pkg/link reads, which no record exists for: %s", nodeRecord.ID, "Ref.md")
	}

	// The note written for the node the other package references is the record read from source, and the
	// note of the record that carries the reference lists the connection it points at, which is the pair
	// the run answered across the two packages.
	carrier := nodeByID(collection, "github.com~Galdoba~go-graph~pkg~link~link.Exists")
	if carrier == nil {
		t.Fatalf("the collection holds no record for the function that references the node: %v", nodeNames(collection))
	}
	want := "- [[" + path.Join(Folder("", nodeRecord.Package), FileName(nodeRecord)) + "|" + nodeRecord.Name + "]]"
	content := string(readFile(t, NotePath(root, carrier)))
	section := content[strings.Index(content, "## Related"):]
	if !strings.Contains(section, want) {
		t.Errorf("the Related section of the note that carries the reference does not list %s:\n%s", want, section)
	}

	record, err := Read(NotePath(root, nodeRecord))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if record.Kind != nodeRecord.Kind.String() || record.SemanticType != nodeRecord.SemanticType.String() {
		t.Errorf("the note read back states %s and %s, want the record the run read as a target: %s and %s",
			record.Kind, record.SemanticType, nodeRecord.Kind, nodeRecord.SemanticType)
	}
	if written := Content(nodeRecord, collection); string(readFile(t, NotePath(root, nodeRecord))) != written {
		t.Errorf("the note of %s is not the content the record read from source renders", nodeRecord.ID)
	}
}
