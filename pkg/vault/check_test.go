package vault

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/Galdoba/go-graph/pkg/graph"
)

// TestCheckConfirmsAFileThatMatchesItsRecord checks the case where the location, the name and the
// recorded region all agree with the record the file carries.
func TestCheckConfirmsAFileThatMatchesItsRecord(t *testing.T) {
	root := t.TempDir()
	second := loadGround(t, "./testdata/second")
	thing := nodeByName(second, "Thing")
	if thing == nil {
		t.Fatalf("the second ground produced no node named Thing: %v", nodeNames(second))
	}

	file, err := Render(root, thing, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	result := Check(root, file)
	if !result.Confirmed {
		t.Errorf("check of %s = %s, want it confirmed", file, result.Reason)
	}
}

// TestCheckRefutesAFileAtALocationItsRecordDoesNotMapTo checks that the folder is read against the
// package path the file itself claims, so a file that sits in a folder another package path maps to
// is refuted even though its name and its region still match.
func TestCheckRefutesAFileAtALocationItsRecordDoesNotMapTo(t *testing.T) {
	root := t.TempDir()
	second := loadGround(t, "./testdata/second")
	thing := nodeByName(second, "Thing")
	if thing == nil {
		t.Fatalf("the second ground produced no node named Thing: %v", nodeNames(second))
	}

	file, err := Render(root, thing, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	foreign := graph.PackageID{
		ModulePath:  thing.Package.ModulePath,
		PackagePath: "github.com/Galdoba/go-graph/random",
		Name:        "random",
	}
	if err := os.MkdirAll(Folder(root, foreign), 0o755); err != nil {
		t.Fatalf("create the folder the record does not map to: %v", err)
	}
	moved := path.Join(Folder(root, foreign), path.Base(file))
	if err := os.WriteFile(moved, readFile(t, file), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	result := Check(root, moved)
	if result.Confirmed || result.Reason != ReasonLocationMismatch {
		t.Errorf("check of %s = %s, want %s", moved, result.Reason, ReasonLocationMismatch)
	}
}

// TestCheckRefutesAFileWhoseNameIsNotTheNodePartOfItsRecord checks the name reason in both
// directions, read against the record the file carries rather than against the code: a record that
// states a method names an identifier whose node part carries a receiver, and a record that states a
// node the package declares on its own names one that carries none. The main ground declares Read at
// package level and a method named Read on the Thing it declares, so both records live in one folder
// and only the name tells them apart. The recorded region still declares the node, so the name is what
// refutes the file.
func TestCheckRefutesAFileWhoseNameIsNotTheNodePartOfItsRecord(t *testing.T) {
	root := t.TempDir()
	ground := loadGround(t, "./testdata")

	packageLevel := nodeByID(ground, "github.com~Galdoba~go-graph~testdata~main.Read")
	method := nodeByID(ground, "github.com~Galdoba~go-graph~testdata~main.Thing.Read")
	if packageLevel == nil || method == nil {
		t.Fatalf("the main ground produced both a package-level Read and a method named Read: %v", nodeNames(ground))
	}

	cases := []struct {
		record *graph.Node
		name   string
		why    string
	}{
		{packageLevel, "Thing.Read.md", "a record that states no method names a node with a receiver"},
		{method, "Read.md", "a record that states a method names a node with its receiver"},
		{packageLevel, "Thing..md", "a name with an empty part is not a node part"},
		{packageLevel, "Thing.Read.Read.md", "a name with more than one dot cannot come from two identifiers, so it is not a node part"},
	}
	for _, c := range cases {
		written, err := Render(root, c.record, nil)
		if err != nil {
			t.Fatalf("render the note of %s: %v", c.record.ID, err)
		}
		file := path.Join(Folder(root, c.record.Package), c.name)
		if err := os.WriteFile(file, readFile(t, written), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}

		result := Check(root, file)
		if result.Confirmed || result.Reason != ReasonNameMismatch {
			t.Errorf("check of %s = %s, want %s, because %s", file, result.Reason, ReasonNameMismatch, c.why)
		}
	}
}

// TestCheckRefutesAFileWhoseRegionNoLongerDeclaresItsNode checks a record whose coordinates point
// at a place in the file that no longer declares the node it records.
func TestCheckRefutesAFileWhoseRegionNoLongerDeclaresItsNode(t *testing.T) {
	root := t.TempDir()
	second := loadGround(t, "./testdata/second")
	thing := nodeByName(second, "Thing")
	if thing == nil {
		t.Fatalf("the second ground produced no node named Thing: %v", nodeNames(second))
	}

	moved := *thing
	moved.Pos.Line = thing.Pos.Line + 1
	file, err := Render(root, &moved, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	result := Check(root, file)
	if result.Confirmed || result.Reason != ReasonRegionNoLongerDeclares {
		t.Errorf("check of %s = %s, want %s", file, result.Reason, ReasonRegionNoLongerDeclares)
	}
}

// TestCheckReportsAFileItCannotRead checks a file that carries no frontmatter block, and that such a
// file survives the rebuild action because its ownership cannot be confirmed from it.
func TestCheckReportsAFileItCannotRead(t *testing.T) {
	root := t.TempDir()
	second := loadGround(t, "./testdata/second")
	pkg := second[0].Package

	if _, err := Rebuild(root, []graph.PackageID{pkg}, second); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	unreadable := path.Join(Folder(root, pkg), "notes_without_frontmatter.md")
	if err := os.WriteFile(unreadable, []byte("# a note with no frontmatter\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	result := Check(root, unreadable)
	if result.Confirmed || result.Reason != ReasonUnreadableFrontmatter {
		t.Errorf("check of %s = %s, want %s", unreadable, result.Reason, ReasonUnreadableFrontmatter)
	}

	report, err := Rebuild(root, []graph.PackageID{pkg}, second)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if _, err := os.Stat(unreadable); err != nil {
		t.Errorf("the file the check reports as unreadable did not survive: %v", err)
	}
	if len(report.Unreadable) != 1 || report.Unreadable[0] != unreadable {
		t.Errorf("reported unreadable files = %v, want %s", report.Unreadable, unreadable)
	}
	if len(report.Removed) != len(second) {
		t.Errorf("reported removals = %v, want the %d files the graph generated", report.Removed, len(second))
	}
}

// TestCheckReportsAValueOutsideItsClosedSetAsUnreadable checks the case the documented reason covers
// as well: the frontmatter reads as lines, but a value it states is not a member of the set that wrote
// that field, so the file cannot be judged from it. Every value below belongs to no closed set at all,
// so the assertion is about membership in the set that wrote the field and cannot be read as a value
// that belongs to the other set.
func TestCheckReportsAValueOutsideItsClosedSetAsUnreadable(t *testing.T) {
	root := t.TempDir()
	second := loadGround(t, "./testdata/second")
	thing := nodeByName(second, "Thing")
	if thing == nil {
		t.Fatalf("the second ground produced no node named Thing: %v", nodeNames(second))
	}

	file, err := Render(root, thing, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	written := string(readFile(t, file))

	cases := []struct {
		before string
		after  string
		claim  string
	}{
		{"kind: type", "kind: thing", "the kind field states a value that belongs to no closed set"},
		{"semantic_type: structure", "semantic_type: thing", "the semantic type field states a value that belongs to no closed set"},
	}
	for _, c := range cases {
		if !strings.Contains(written, c.before) {
			t.Fatalf("the file carries no %s line: %q", c.before, written)
		}
		if err := os.WriteFile(file, []byte(strings.Replace(written, c.before, c.after, 1)), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}

		result := Check(root, file)
		if result.Confirmed || result.Reason != ReasonUnreadableFrontmatter {
			t.Errorf("check of %s = %s, want %s, because %s", file, result.Reason, ReasonUnreadableFrontmatter, c.claim)
		}
	}
}

// TestCheckConfirmsAForeignNoteAgainstTheEntityItsPackageDeclares checks a note for a record that
// states no region. Nothing in the file names a place to read the node at, so the answer is given by
// the package the record names holding the entity, and the check confirms it without the build that
// produced the file.
func TestCheckConfirmsAForeignNoteAgainstTheEntityItsPackageDeclares(t *testing.T) {
	root := t.TempDir()
	_, second := loadForeignPairWithoutSyntax(t, "./pkg/graph/testdata/fixtures/use")
	hidden := nodeByID(second, "github.com~Galdoba~go-graph~pkg~graph~testdata~fixtures~shape~shape.hidden")
	if hidden == nil {
		t.Fatalf("the second pass produced no record for the entity the alias names: %v", nodeNames(second))
	}
	if hidden.Pos.File != "" {
		t.Fatalf("the record of %s states a region, want one that states none", hidden.ID)
	}

	file, err := Render(root, hidden, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	result := Check(root, file)
	if !result.Confirmed {
		t.Errorf("check of %s = %s, want the file answered against the entity the package declares", file, result.Reason)
	}
}

// TestCheckRefutesAForeignNoteWhosePackageNoLongerDeclaresTheEntity keeps the record of a note that
// states no region and asks for an entity its package does not declare, which is the answer the check
// gives when the code at that moment no longer holds the name the record states.
func TestCheckRefutesAForeignNoteWhosePackageNoLongerDeclaresTheEntity(t *testing.T) {
	root := t.TempDir()
	_, second := loadForeignPairWithoutSyntax(t, "./pkg/graph/testdata/fixtures/use")
	hidden := nodeByID(second, "github.com~Galdoba~go-graph~pkg~graph~testdata~fixtures~shape~shape.hidden")
	if hidden == nil {
		t.Fatalf("the second pass produced no record for the entity the alias names: %v", nodeNames(second))
	}
	if hidden.Pos.File != "" {
		t.Fatalf("the record of %s states a region, want one that states none", hidden.ID)
	}

	gone := *hidden
	gone.Name = "Gone"
	gone.ID = graph.ID{Package: hidden.Package, Name: "Gone"}

	file, err := Render(root, &gone, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	result := Check(root, file)
	if result.Confirmed || result.Reason != ReasonPackageNoLongerDeclares {
		t.Errorf("check of %s = %s, want %s", file, result.Reason, ReasonPackageNoLongerDeclares)
	}
}

// TestCheckRefutesAForeignNoteForAPackageItCannotRead checks the reason that stands on its own: a note
// whose record names a package no load can resolve is refuted for that reason rather than for a reason
// that needs a region the record does not carry.
func TestCheckRefutesAForeignNoteForAPackageItCannotRead(t *testing.T) {
	root := t.TempDir()
	unknown := graph.PackageID{ModulePath: "nowhere.test", PackagePath: "nowhere.test/shape", Name: "shape"}
	node := &graph.Node{ID: graph.ID{Package: unknown, Name: "Shape"}, Package: unknown, Name: "Shape",
		Kind: graph.KindType, SemanticType: graph.SemanticStructure}

	file, err := Render(root, node, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	result := Check(root, file)
	if result.Confirmed || result.Reason != ReasonPackageNotReadable {
		t.Errorf("check of %s = %s, want %s", file, result.Reason, ReasonPackageNotReadable)
	}
}

// TestCheckRefutesAForeignNoteInAFolderThatStatesAnotherVersion checks that the location is read
// against the package part the record states, version included. The same record in the folder that
// states its own version is confirmed, so the version is what refutes the file when the folder states
// another one.
func TestCheckRefutesAForeignNoteInAFolderThatStatesAnotherVersion(t *testing.T) {
	root := t.TempDir()
	_, second := loadForeignPairWithoutSyntax(t, "./pkg/graph/testdata/fixtures/use")
	hidden := nodeByID(second, "github.com~Galdoba~go-graph~pkg~graph~testdata~fixtures~shape~shape.hidden")
	if hidden == nil {
		t.Fatalf("the second pass produced no record for the entity the alias names: %v", nodeNames(second))
	}

	versioned := *hidden
	versioned.Package.ModuleVersion = "v1.2.3"
	versioned.ID.Package.ModuleVersion = "v1.2.3"

	file, err := Render(root, &versioned, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if path.Dir(file) != path.Join(root, hidden.Package.PackagePath, "v1.2.3") {
		t.Fatalf("the note was written in %s, want the folder that states the version the record states", path.Dir(file))
	}
	if result := Check(root, file); !result.Confirmed {
		t.Fatalf("check of %s = %s, want the file confirmed in the folder that states its version", file, result.Reason)
	}

	other := versioned.Package
	other.ModuleVersion = "v1.3.0"
	moved := path.Join(Folder(root, other), path.Base(file))
	if err := os.MkdirAll(Folder(root, other), 0o755); err != nil {
		t.Fatalf("create the folder that states another version: %v", err)
	}
	if err := os.WriteFile(moved, readFile(t, file), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	result := Check(root, moved)
	if result.Confirmed || result.Reason != ReasonLocationMismatch {
		t.Errorf("check of %s = %s, want %s", moved, result.Reason, ReasonLocationMismatch)
	}
}

// TestCheckRefutesAForeignNoteWhoseNameIsNotItsNodePart keeps a foreign record and writes its content
// under a name that is not the node part of the identifier the record states. The record states a
// method, so its name carries the type the method is declared on, and the check refutes the file for
// the name before it reads the package.
func TestCheckRefutesAForeignNoteWhoseNameIsNotItsNodePart(t *testing.T) {
	root := t.TempDir()
	_, _, second := loadForeignPair(t, "./pkg/graph/testdata/fixtures/use")
	method := nodeByID(second, "github.com~Galdoba~go-graph~pkg~graph~testdata~fixtures~shape~shape.Shape.Read")
	if method == nil {
		t.Fatalf("the second pass produced no record for the method a reference names: %v", nodeNames(second))
	}

	file, err := Render(root, method, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if path.Base(file) != "Shape.Read.md" {
		t.Fatalf("the note maps to %s, want the node part of the identifier the record states", path.Base(file))
	}

	moved := path.Join(Folder(root, method.Package), "Read.md")
	if err := os.WriteFile(moved, readFile(t, file), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	result := Check(root, moved)
	if result.Confirmed || result.Reason != ReasonNameMismatch {
		t.Errorf("check of %s = %s, want %s", moved, result.Reason, ReasonNameMismatch)
	}
}
