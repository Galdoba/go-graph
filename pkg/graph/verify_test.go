package graph

import "testing"

// TestVerifyResolvesARecordWithoutItsDeclaration checks that a record carrying only its
// coordinates resolves against a load that did not produce it. A record read back from a note
// file has no reference to the declaration that created the node, so the coordinates are the
// only thing a check can read from it.
func TestVerifyResolvesARecordWithoutItsDeclaration(t *testing.T) {
	nodes, _ := mustBuild(t, loadFixture(t, "basic"))
	original := nodeByName(nodes, "Struct1")
	if original == nil {
		t.Fatalf("fixture basic produced no node named Struct1: %v", nodeNames(nodes))
	}

	record := *original
	record.Decl = nil

	if err := Verify(loadFixture(t, "basic"), &record); err != nil {
		t.Errorf("resolve the record without its declaration: %v", err)
	}
	if record.Decl != nil {
		t.Errorf("Verify attached a declaration reference to the record: %v", record.Decl)
	}
}
