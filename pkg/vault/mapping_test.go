package vault

import (
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/Galdoba/go-graph/pkg/graph"
)

// TestMappingMirrorsTheRecordedPackagePath checks the folder a record maps to, including the case
// where the directory a package sits in is not the name it declares, which is why the declared name
// has to be read from the file.
func TestMappingMirrorsTheRecordedPackagePath(t *testing.T) {
	root := t.TempDir()

	second := loadGround(t, "./testdata/second")
	thing := nodeByName(second, "Thing")
	if thing == nil {
		t.Fatalf("the second ground produced no node named Thing: %v", nodeNames(second))
	}
	want := path.Join(root, "github.com/Galdoba/go-graph/testdata/second", "Thing.md")
	if got := NotePath(root, thing); got != want {
		t.Errorf("note path = %s, want %s", got, want)
	}

	ground := loadGround(t, "./testdata")
	box := nodeByName(ground, "Box")
	if box == nil {
		t.Fatalf("the test ground produced no node named Box: %v", nodeNames(ground))
	}

	want = path.Join(root, "github.com/Galdoba/go-graph/testdata", "Box.md")
	file, err := Render(root, box, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if file != want {
		t.Errorf("rendered at %s, want %s", file, want)
	}

	record, err := Read(file)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if record.PackagePath != "github.com/Galdoba/go-graph/testdata" {
		t.Errorf("recorded package path = %s, want github.com/Galdoba/go-graph/testdata", record.PackagePath)
	}
	if record.Package != "main" {
		t.Errorf("declared name read from the file = %s, want main", record.Package)
	}
}

// TestFileNameIsTheNodePartOfTheIdentifier checks that a name carries the receiver and the name the
// package declares and states no classification. The main ground declares Read at package level and a
// method named Read on Thing, so a name made of the node name alone would ask for one file for two
// nodes in the same folder. Both ground packages declare a type named Thing, so the same name appears
// in two folders. The two nodes named Read are resolved by the canonical form, because a lookup by
// name alone returns the first match and cannot tell a package-level node from a method.
func TestFileNameIsTheNodePartOfTheIdentifier(t *testing.T) {
	second := loadGround(t, "./testdata/second")
	ground := loadGround(t, "./testdata")

	secondThing := nodeByName(second, "Thing")
	groundThing := nodeByName(ground, "Thing")
	if secondThing == nil || groundThing == nil {
		t.Fatalf("both grounds must produce a node named Thing: %v, %v", nodeNames(second), nodeNames(ground))
	}
	if got := FileName(secondThing); got != "Thing.md" {
		t.Errorf("name for the type node of the second ground = %s, want Thing.md", got)
	}
	if FileName(groundThing) != FileName(secondThing) {
		t.Errorf("the same node name in two packages gives %s and %s, want one name read as two files in two folders",
			FileName(groundThing), FileName(secondThing))
	}

	function := nodeByName(second, "New")
	if function == nil {
		t.Fatalf("the second ground produced no function node: %v", nodeNames(second))
	}
	if got := FileName(function); got != "New.md" {
		t.Errorf("name for the function node = %s, want New.md", got)
	}
	secondMethod := nodeByName(second, "Read")
	if secondMethod == nil {
		t.Fatalf("the second ground produced no method named Read: %v", nodeNames(second))
	}
	if got := FileName(secondMethod); got != "Thing.Read.md" {
		t.Errorf("name for the method Read of the second ground = %s, want Thing.Read.md", got)
	}

	packageLevel := nodeByID(ground, "github.com~Galdoba~go-graph~testdata~main.Read")
	method := nodeByID(ground, "github.com~Galdoba~go-graph~testdata~main.Thing.Read")
	if packageLevel == nil || method == nil {
		t.Fatalf("the main ground produced both a package-level Read and a method named Read: %v", nodeNames(ground))
	}
	if got := FileName(packageLevel); got != "Read.md" {
		t.Errorf("name for the package-level node = %s, want Read.md", got)
	}
	if got := FileName(method); got != "Thing.Read.md" {
		t.Errorf("name for the method named Read = %s, want Thing.Read.md", got)
	}
	if FileName(method) == FileName(packageLevel) {
		t.Errorf("both nodes named Read map to %s, want two distinct names in one folder", FileName(method))
	}

	root := t.TempDir()
	packageFile, err := Render(root, packageLevel, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	methodFile, err := Render(root, method, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if path.Dir(packageFile) != path.Dir(methodFile) {
		t.Fatalf("the two nodes named Read were written in %s and %s, want both in the folder of the main ground",
			path.Dir(packageFile), path.Dir(methodFile))
	}
	entries, err := os.ReadDir(Folder(root, packageLevel.Package))
	if err != nil {
		t.Fatalf("read folder: %v", err)
	}
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	want := []string{"Read.md", "Thing.Read.md"}
	slices.Sort(names)
	if !slices.Equal(names, want) {
		t.Errorf("the folder of the main ground holds %v, want both %s and %s", names, want[0], want[1])
	}

	shape := &graph.Node{Name: "Shape", SemanticType: graph.SemanticUnresolved, Package: second[0].Package}
	if got := FileName(shape); got != "Shape.md" {
		t.Errorf("name for a node no rule classifies = %s, want Shape.md, because a name carries no classification", got)
	}
}

// TestParsingAFileNameKeepsTheNodeNameIntact checks that reading a name back gives the receiver and
// the node name without matching any part against a closed set, so a node name that contains an
// underscore survives, and that the parts read back to the node part of the identifier they came from.
// A name that two identifiers cannot produce is not read at all.
func TestParsingAFileNameKeepsTheNodeNameIntact(t *testing.T) {
	parts, err := ParseFileName("Foo_Bar.md")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parts.Name != "Foo_Bar" || parts.Receiver != "" {
		t.Errorf("parts = %+v, want node name Foo_Bar with no receiver", parts)
	}

	parts, err = ParseFileName("Thing.Read.md")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parts.Receiver != "Thing" || parts.Name != "Read" {
		t.Errorf("parts = %+v, want receiver Thing and node name Read", parts)
	}

	second := loadGround(t, "./testdata/second")
	for _, n := range second {
		parts, err := ParseFileName(FileName(n))
		if err != nil {
			t.Fatalf("read the name %s back: %v", FileName(n), err)
		}
		if parts.Receiver != n.Receiver || parts.Name != n.Name {
			t.Errorf("the name %s reads as %+v, want the node part of %s", FileName(n), parts, n.ID)
		}
	}

	if _, err := ParseFileName("Thing..md"); err == nil {
		t.Error("a name with an empty part was read as the node part of an identifier")
	}
	if _, err := ParseFileName("Thing.Read.Read.md"); err == nil {
		t.Error("a name with more than one dot was read as the node part of an identifier, which two identifiers cannot produce")
	}
}

// TestTwoVersionsOfOnePackagePathMapToTwoFolders checks the part of the location a record states that
// the folder has to carry. Two records that differ only by version cannot share one file, because a
// file that holds two identifiers cannot be read back to one node, while a record that states no
// version keeps the folder the package path alone gives.
func TestTwoVersionsOfOnePackagePathMapToTwoFolders(t *testing.T) {
	root := t.TempDir()

	unversioned := graph.PackageID{ModulePath: "mod.test", PackagePath: "mod.test/shape", Name: "shape"}
	if got := Folder(root, unversioned); got != path.Join(root, "mod.test/shape") {
		t.Errorf("folder for a record that states no version = %s, want %s", got, path.Join(root, "mod.test/shape"))
	}

	first := unversioned
	first.ModuleVersion = "v1.2.3"
	second := unversioned
	second.ModuleVersion = "v1.3.0"

	if got := Folder(root, first); got != path.Join(root, "mod.test/shape", "v1.2.3") {
		t.Errorf("folder for the record that states v1.2.3 = %s, want %s", got, path.Join(root, "mod.test/shape", "v1.2.3"))
	}
	if Folder(root, first) == Folder(root, second) {
		t.Errorf("both versions map to %s, want one folder per version", Folder(root, first))
	}

	older := &graph.Node{ID: graph.ID{Package: first, Name: "Shape"}, Package: first, Name: "Shape",
		Kind: graph.KindType, SemanticType: graph.SemanticStructure}
	newer := &graph.Node{ID: graph.ID{Package: second, Name: "Shape"}, Package: second, Name: "Shape",
		Kind: graph.KindType, SemanticType: graph.SemanticStructure}
	if NotePath(root, older) == NotePath(root, newer) {
		t.Errorf("one entity at two versions maps to one file %s, want two files", NotePath(root, older))
	}
}

// TestAForeignNoteReadsBackToTheIdentifierItsRecordStates renders a note for a record produced from a
// reference and reads it back from its location and its name alone. The location states the package
// the record names rather than the package that referenced it, so the note is read as a node of the
// package that declares the entity.
func TestAForeignNoteReadsBackToTheIdentifierItsRecordStates(t *testing.T) {
	root := t.TempDir()
	_, _, foreign := loadForeignPair(t, "./pkg/graph/testdata/fixtures/use")
	record := nodeByID(foreign, "github.com~Galdoba~go-graph~pkg~graph~testdata~fixtures~shape~shape.Shape")
	if record == nil {
		t.Fatalf("the second pass produced no record for the shape a reference names: %v", nodeNames(foreign))
	}

	file, err := Render(root, record, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	location := strings.TrimPrefix(path.Dir(file), root+"/")
	if location != record.Package.PackagePath {
		t.Errorf("the location reads as %s, want the package path the record states, %s", location, record.Package.PackagePath)
	}

	read, err := Read(file)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	parts, err := ParseFileName(path.Base(file))
	if err != nil {
		t.Fatalf("read the name back: %v", err)
	}

	assembled := graph.ID{
		Package: graph.PackageID{
			ModulePath:    read.Module,
			ModuleVersion: read.Version,
			PackagePath:   location,
			Name:          read.Package,
		},
		Receiver: parts.Receiver,
		Name:     parts.Name,
	}
	if assembled != record.ID {
		t.Errorf("the location and the name read as %s, want the identifier the record states, %s", assembled, record.ID)
	}
}
