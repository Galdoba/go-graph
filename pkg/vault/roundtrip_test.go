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

		if _, err := Rebuild(root, []graph.PackageID{nodes[0].Package}, nodes); err != nil {
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

// TestRoundTripForAForeignNoteReadsBackToTheRecordItClaims runs the round trip for a note written for a
// record produced from a reference. The recorded reference from main to a name of a package the build never
// read as a target now produces a record, so a file exists for it in the folder that package path maps to,
// and the location and the name together read back to the identifier that record states. Two records for one
// package path at two module versions give two files, because one folder cannot hold two identifiers that
// read back as one node, and the check reads each folder against the version its own record states.
func TestRoundTripForAForeignNoteReadsBackToTheRecordItClaims(t *testing.T) {
	root := t.TempDir()
	_, built, records := loadForeignPair(t, "./testdata")
	if len(records) == 0 {
		t.Fatalf("the second pass produced no record for a package the build never read as a target: %v", nodeNames(built))
	}

	for _, n := range records {
		file, err := Render(root, n, append(built, records...))
		if err != nil {
			t.Fatalf("render %s: %v", n.ID, err)
		}

		location := path.Dir(file)
		if location != Folder(root, n.Package) {
			t.Errorf("the note of %s sits in %s, want the folder the package its record names maps to", n.ID, location)
		}

		read, err := Read(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		parts, err := ParseFileName(path.Base(file))
		if err != nil {
			t.Fatalf("read the name %s back: %v", path.Base(file), err)
		}
		pkg := graph.PackageID{
			ModulePath:    read.Module,
			ModuleVersion: read.Version,
			PackagePath:   read.PackagePath,
			Name:          read.Package,
		}
		if pkg != n.Package {
			t.Errorf("the package parts read from %s do not match the record of %s: %+v", file, n.ID, read)
		}
		assembled := graph.ID{Package: pkg, Receiver: parts.Receiver, Name: parts.Name}
		if assembled != n.ID {
			t.Errorf("the location and the name read as %s, want the identifier the record states, %s", assembled, n.ID)
		}
		if result := Check(root, file); !result.Confirmed {
			t.Errorf("check of %s = %s, want the file confirmed against the package its record names", file, result.Reason)
		}
	}

	// The same entity at two module versions is two records, so it is two files.
	_, _, foreign := loadForeignPair(t, "./pkg/graph/testdata/fixtures/use")
	shape := nodeByID(foreign, "github.com~Galdoba~go-graph~pkg~graph~testdata~fixtures~shape~shape.Shape")
	if shape == nil {
		t.Fatalf("the second pass produced no record for the shape a reference names: %v", nodeNames(foreign))
	}

	older := *shape
	older.Package.ModuleVersion = "v1.2.3"
	older.ID.Package = older.Package
	newer := *shape
	newer.Package.ModuleVersion = "v1.3.0"
	newer.ID.Package = newer.Package

	olderFile, err := Render(root, &older, nil)
	if err != nil {
		t.Fatalf("render %s: %v", older.ID, err)
	}
	newerFile, err := Render(root, &newer, nil)
	if err != nil {
		t.Fatalf("render %s: %v", newer.ID, err)
	}
	if olderFile == newerFile {
		t.Fatalf("one entity at two versions maps to one file %s, want two files", olderFile)
	}

	for _, want := range []struct {
		file string
		id   graph.ID
	}{{olderFile, older.ID}, {newerFile, newer.ID}} {
		read, err := Read(want.file)
		if err != nil {
			t.Fatalf("read %s: %v", want.file, err)
		}
		parts, err := ParseFileName(path.Base(want.file))
		if err != nil {
			t.Fatalf("read the name %s back: %v", path.Base(want.file), err)
		}
		pkg := graph.PackageID{
			ModulePath:    read.Module,
			ModuleVersion: read.Version,
			PackagePath:   read.PackagePath,
			Name:          read.Package,
		}
		if path.Dir(want.file) != Folder(root, pkg) {
			t.Errorf("the note read from %s sits in a folder that does not state the version its record states", want.file)
		}
		assembled := graph.ID{Package: pkg, Receiver: parts.Receiver, Name: parts.Name}
		if assembled != want.id {
			t.Errorf("%s reads as %s, want the identifier its record states, %s", want.file, assembled, want.id)
		}
		if assembled != want.id {
			t.Errorf("%s reads as %s, want the identifier its record states, %s", want.file, assembled, want.id)
		}
		if result := Check(root, want.file); !result.Confirmed {
			t.Errorf("check of %s = %s, want the file confirmed in the folder that states the version its record states",
				want.file, result.Reason)
		}
	}

	// A file read in a folder that states another version is refuted, because the location is read
	// against the version the record states rather than against the folder it happens to sit in.
	misplaced := path.Join(Folder(root, newer.Package), FileName(&older))
	data, err := os.ReadFile(olderFile)
	if err != nil {
		t.Fatalf("read %s: %v", olderFile, err)
	}
	if err := os.WriteFile(misplaced, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", misplaced, err)
	}
	if result := Check(root, misplaced); result.Confirmed || result.Reason != ReasonLocationMismatch {
		t.Errorf("check of %s = %s, want the file refuted as a location that states another version", misplaced, result.Reason)
	}
}
