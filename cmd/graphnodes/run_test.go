package main

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/Galdoba/go-graph/pkg/graph"
)

// moduleRoot is the module root, which is the directory the loader has to run in to resolve a package of
// the test ground by import path.
func moduleRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// TestACollectOverTwoTargetPackagesCountsOneEntryPerIdentifier checks the counts the command states for a
// run that reads two packages as targets at one time: the collection is one entry per identifier, so an
// entity one pass built and the other named from a reference is counted once, and the collection is read
// in the order the identifiers give whatever order the load returned the packages in.
func TestACollectOverTwoTargetPackagesCountsOneEntryPerIdentifier(t *testing.T) {
	found, err := packages.Load(&packages.Config{Mode: graph.LoadMode, Dir: moduleRoot()},
		"./pkg/graph/testdata/fixtures/use", "./pkg/graph/testdata/fixtures/shape")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("load: expected two packages, got %d", len(found))
	}

	targets, collection, counts, err := collect(found)
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("the run counted %d packages as targets, want two", len(targets))
	}
	if len(counts) != 2 {
		t.Fatalf("the run states counts for %d packages, want two", len(counts))
	}

	// The count the run states is the count of the identifiers it read, because the merge keeps one entry
	// for an entity both passes read.
	read := map[string]bool{}
	stated := 0
	for i, p := range found {
		built, _, err := graph.Build(p)
		if err != nil {
			t.Fatalf("build %s: %v", p.PkgPath, err)
		}
		fromNames, err := graph.SecondPass(p)
		if err != nil {
			t.Fatalf("second pass over %s: %v", p.PkgPath, err)
		}
		if counts[i].pkg != p.PkgPath || counts[i].built != len(built) || counts[i].fromNames != len(fromNames) {
			t.Errorf("the run states %d nodes and %d records for %s, want %d and %d", counts[i].built, counts[i].fromNames, p.PkgPath, len(built), len(fromNames))
		}
		stated += counts[i].built + counts[i].fromNames
		for _, n := range append(built, fromNames...) {
			read[n.ID.String()] = true
		}
	}
	if len(collection) != len(read) {
		t.Errorf("the run counted %d records, want one entry for each of the %d identifiers it read", len(collection), len(read))
	}
	if stated <= len(collection) {
		t.Errorf("the run stated %d records for the %d identifiers it read, want the merge to count an entity both passes read once", stated, len(collection))
	}

	seen := map[string]bool{}
	for i, n := range collection {
		if seen[n.ID.String()] {
			t.Errorf("the collection counts %s twice", n.ID)
		}
		seen[n.ID.String()] = true
		if i > 0 && collection[i-1].ID.String() > n.ID.String() {
			t.Errorf("the collection is not ordered by identifier at %d: %s then %s", i, collection[i-1].ID, n.ID)
		}
	}
}
