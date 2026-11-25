package vault

import (
	"os"
	"path"
	"slices"
	"testing"

	"github.com/Galdoba/go-graph/pkg/graph"
)

// TestRoundTripInFileForm runs the action over both test grounds, reads each file it writes back and
// confirms it against the code, which is the whole capability exercised in file form. The files present
// in each ground folder are exactly the names the mapping produces.
func TestRoundTripInFileForm(t *testing.T) {
	// Both ground packages declare a type named Thing, so two files share one name in two folders and
	// have to read back as two nodes.
	thingFiles := make(map[string]graph.ID)

	for _, pattern := range []string{"./testdata", "./testdata/second"} {
		nodes := loadGround(t, pattern)
		root := t.TempDir()

		if _, err := Rebuild(root, nodes[0].Package, nodes); err != nil {
			t.Fatalf("rebuild %s: %v", pattern, err)
		}

		entries, err := os.ReadDir(Folder(root, nodes[0].Package))
		if err != nil {
			t.Fatalf("read folder: %v", err)
		}
		present := make([]string, len(entries))
		for i, entry := range entries {
			present[i] = entry.Name()
		}
		want := make([]string, len(nodes))
		for i, n := range nodes {
			want[i] = FileName(n)
		}
		slices.Sort(present)
		slices.Sort(want)
		if !slices.Equal(present, want) {
			t.Errorf("the folder of %s holds %v, want the names the mapping produces: %v", pattern, present, want)
		}

		for _, n := range nodes {
			file := NotePath(root, n)
			record, err := Read(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}

			if !record.Generated {
				t.Errorf("%s does not claim graph ownership when read back", file)
			}
			if record.Module != n.Package.ModulePath || record.Version != n.Package.ModuleVersion ||
				record.PackagePath != n.Package.PackagePath || record.Package != n.Package.Name {
				t.Errorf("identifier read from %s does not match the record of %s: %+v", file, n.ID, record)
			}

			parts, err := ParseFileName(path.Base(file))
			if err != nil {
				t.Fatalf("read the name %s back: %v", path.Base(file), err)
			}
			if parts.Receiver != n.Receiver || parts.Name != n.Name {
				t.Errorf("name %s reads as %+v, want the node part of %s", path.Base(file), parts, n.ID)
			}
			if !slices.Equal(record.Properties, propertyNames(n.Properties)) {
				t.Errorf("properties read from %s = %v, want %v", file, record.Properties, propertyNames(n.Properties))
			}

			result := Check(root, file)
			if !result.Confirmed {
				t.Errorf("check of %s = %s, want the file confirmed", file, result.Reason)
			}

			if path.Base(file) == "Thing.md" {
				thingFiles[pattern] = graph.ID{
					Package: graph.PackageID{
						ModulePath:    record.Module,
						ModuleVersion: record.Version,
						PackagePath:   record.PackagePath,
						Name:          record.Package,
					},
					Name: parts.Name,
				}
			}
		}
	}

	if len(thingFiles) != 2 {
		t.Fatalf("no file named Thing.md read back from both grounds: %v", thingFiles)
	}
	if thingFiles["./testdata"] == thingFiles["./testdata/second"] {
		t.Errorf("the two files named Thing.md both read as %s, want the same name to read as two nodes",
			thingFiles["./testdata"].String())
	}
}

// TestRoundTripWritesNoFileForAPackageTheGraphNeverBuilt checks a recorded reference that resolves
// outside the node set. The ground reference from main to strings.ToLower names a package the
// graph never loaded as a target, so the file count stays the number of nodes the build produced.
func TestRoundTripWritesNoFileForAPackageTheGraphNeverBuilt(t *testing.T) {
	nodes := loadGround(t, "./testdata")

	var referent graph.ID
	for _, n := range nodes {
		for _, a := range n.Attributes {
			if a.Ref.Package.PackagePath == "strings" {
				referent = a.Ref
			}
		}
	}
	if referent.Name == "" {
		t.Fatalf("the ground records no reference to a package the graph never built: %v", nodeNames(nodes))
	}

	root := t.TempDir()
	for _, n := range nodes {
		if _, err := Render(root, n, nil); err != nil {
			t.Fatalf("render %s: %v", n.ID, err)
		}
	}

	entries, err := os.ReadDir(Folder(root, nodes[0].Package))
	if err != nil {
		t.Fatalf("read folder: %v", err)
	}
	if len(entries) != len(nodes) {
		t.Errorf("the folder holds %d files, want one for each of the %d nodes", len(entries), len(nodes))
	}
	if _, err := os.Stat(Folder(root, referent.Package)); err == nil {
		t.Errorf("a folder was written for package %s, which the graph never built", referent.Package.PackagePath)
	}
}
