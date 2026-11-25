package vault

import (
	"path"
	"testing"

	"github.com/Galdoba/go-graph/pkg/graph"
)

// TestSameNodeNameInTwoPackagesReadsAsTwoNodes covers the two packages of the test ground that both
// declare a type named Thing, which gives two notes with the same file name in two folders. The records
// are the ones the build produced for real declarations, so each file is read back and confirmed against
// the declaration in its own package.
func TestSameNodeNameInTwoPackagesReadsAsTwoNodes(t *testing.T) {
	root := t.TempDir()
	patterns := []string{"./testdata", "./testdata/second"}

	files := make([]string, len(patterns))
	identifiers := make([]graph.ID, len(patterns))
	for i, pattern := range patterns {
		nodes := loadGround(t, pattern)
		thing := nodeByName(nodes, "Thing")
		if thing == nil {
			t.Fatalf("%s produced no node named Thing: %v", pattern, nodeNames(nodes))
		}
		if got := FileName(thing); got != "Thing.md" {
			t.Fatalf("%s maps to %s, want the same file name for the same node name", pattern, got)
		}

		file, err := Render(root, thing, nil)
		if err != nil {
			t.Fatalf("render the note of %s: %v", thing.ID, err)
		}
		files[i] = file

		record, err := Read(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if record.Module != thing.Package.ModulePath || record.PackagePath != thing.Package.PackagePath ||
			record.Package != thing.Package.Name {
			t.Errorf("identifier read from %s does not match the record of %s: %+v", file, thing.ID, record)
		}

		parts, err := ParseFileName(path.Base(file))
		if err != nil {
			t.Fatalf("read the name %s back: %v", path.Base(file), err)
		}
		if parts.Name != thing.Name || parts.Receiver != thing.Receiver {
			t.Errorf("the name %s reads as %+v, want the node part of %s", path.Base(file), parts, thing.ID)
		}

		identifiers[i] = graph.ID{
			Package: graph.PackageID{
				ModulePath:    record.Module,
				ModuleVersion: record.Version,
				PackagePath:   record.PackagePath,
				Name:          record.Package,
			},
			Name: parts.Name,
		}

		result := Check(root, file)
		if !result.Confirmed {
			t.Errorf("check of %s = %s, want the file confirmed against the declaration in its own package",
				file, result.Reason)
		}
	}

	if path.Dir(files[0]) == path.Dir(files[1]) {
		t.Fatalf("both notes were written in %s, want one file in the folder of each package", path.Dir(files[0]))
	}
	if identifiers[0] == identifiers[1] {
		t.Errorf("both files read as the node %s, want the same name to read as two nodes", identifiers[0])
	}
	if identifiers[0].String() == identifiers[1].String() {
		t.Errorf("both files read as the note %s, want two notes", identifiers[0].String())
	}
}
