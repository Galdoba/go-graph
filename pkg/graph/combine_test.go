package graph

import (
	"testing"
)

// TestARunOverTwoTargetPackagesHoldsOneRecordForAnEntityBothPassesRead checks the collection a run holds
// when it reads as a target a package that a reference of another package of the same run names: the entity
// has one entry, the record the first pass produced, because that record states the uses its declaration
// carries while the record read from a name cannot.
func TestARunOverTwoTargetPackagesHoldsOneRecordForAnEntityBothPassesRead(t *testing.T) {
	use, shape := loadForeignPair(t, singleNameUseSource)

	useRecords, _ := mustBuild(t, use)
	shapeRecords, _ := mustBuild(t, shape)
	nameRecords, err := SecondPass(use)
	if err != nil {
		t.Fatalf("second pass over the package that references: %v", err)
	}
	if len(nameRecords) != 1 {
		t.Fatalf("the second pass produced %d records for %v, want the record the reference names",
			len(nameRecords), nodeNames(nameRecords))
	}

	combined := Combine(append(append(useRecords, shapeRecords...), nameRecords...))
	if len(combined) != len(useRecords)+len(shapeRecords) {
		t.Fatalf("the collection holds %d entries for %v, want one entry per identifier",
			len(combined), nodeNames(combined))
	}

	declared := mustNodeByName(t, shapeRecords, "Shape")
	entry := mustNodeByID(t, combined, declared.ID.String())
	if entry != declared {
		t.Errorf("the entry for %v is the record read from %v, want the record read from source",
			declared.ID, entry.Provenance)
	}
	for _, n := range combined {
		if n.Provenance == ProvenanceFromName {
			t.Errorf("the collection holds the record %v read from a name, want the record read from source", n.ID)
		}
	}
}

// TestTheCollectionOfARunIsTheSameWhateverOrderThePackagesArriveIn checks that a run that read the same
// packages produces the same collection when they arrive in a different order. The records that can differ
// between two loads of the same package are the ones that state a region, so the collection is read from the
// record that states a region rather than from whichever reference reached the entity first.
func TestTheCollectionOfARunIsTheSameWhateverOrderThePackagesArriveIn(t *testing.T) {
	readingLoad := newSourceLoader(t)
	readingLoad.load("mod.test", "", "mod.test/shape", map[string]string{"a.go": foreignShapeSource})
	reading := readingLoad.load("mod.test", "", "mod.test/use", map[string]string{"a.go": singleNameUseSource})

	readingWithoutSyntax := newSourceLoader(t)
	shapeWithoutSyntax := readingWithoutSyntax.load("mod.test", "", "mod.test/shape", map[string]string{"a.go": foreignShapeSource})
	shapeWithoutSyntax.Syntax = nil
	readingFromTypeInformation := readingWithoutSyntax.load("mod.test", "", "mod.test/use", map[string]string{"a.go": singleNameUseSource})

	statesRegion, err := SecondPass(reading)
	if err != nil {
		t.Fatalf("second pass over %s: %v", reading.PkgPath, err)
	}
	statesNothing, err := SecondPass(readingFromTypeInformation)
	if err != nil {
		t.Fatalf("second pass over %s: %v", readingFromTypeInformation.PkgPath, err)
	}
	if len(statesRegion) != 1 || len(statesNothing) != 1 {
		t.Fatalf("the two loads produced %v and %v, want one record each", nodeNames(statesRegion), nodeNames(statesNothing))
	}
	if statesRegion[0].ID != statesNothing[0].ID {
		t.Fatalf("the two loads produced %v and %v for the same entity, want one identifier",
			statesRegion[0].ID, statesNothing[0].ID)
	}
	if statesRegion[0].Pos.File == "" || statesNothing[0].Pos.File != "" {
		t.Fatalf("the two loads state %q and %q for the same entity, want a region and no region",
			statesRegion[0].Pos.File, statesNothing[0].Pos.File)
	}

	first := Combine(append([]*Node{}, append(statesRegion, statesNothing...)...))
	second := Combine(append([]*Node{}, append(statesNothing, statesRegion...)...))
	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("the run holds %v and %v for one entity, want one entry each", nodeNames(first), nodeNames(second))
	}
	if first[0] != statesRegion[0] || second[0] != statesRegion[0] {
		t.Errorf("the run keeps %v and %v, want the record that states the region the load reported",
			first[0].ID, second[0].ID)
	}
}
