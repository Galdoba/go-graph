package vault

import (
	"os"
	"path"
	"slices"
	"testing"

	"github.com/Galdoba/go-graph/pkg/graph"
)

// TestRebuildRemovesTheFilesItOwns checks that a marked file is replaced while a file the graph did
// not generate in the same folder survives.
func TestRebuildRemovesTheFilesItOwns(t *testing.T) {
	root := t.TempDir()
	second := loadGround(t, "./testdata/second")
	pkg := second[0].Package
	folder := Folder(root, pkg)

	if _, err := Rebuild(root, pkg, second); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	own := "---\ntitle: my note\n---\n\nwritten by hand\n"
	human := path.Join(folder, "own note.md")
	if err := os.WriteFile(human, []byte(own), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	report, err := Rebuild(root, pkg, second)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if len(report.Removed) != len(second) {
		t.Errorf("reported removals = %v, want the %d files the graph generated", report.Removed, len(second))
	}
	if len(report.Unreadable) != 0 {
		t.Errorf("reported unreadable files = %v, want none", report.Unreadable)
	}

	data, err := os.ReadFile(human)
	if err != nil {
		t.Fatalf("the file without the marker did not survive: %v", err)
	}
	if string(data) != own {
		t.Errorf("the file without the marker changed: %q", string(data))
	}
	if slices.Contains(report.Removed, human) {
		t.Errorf("the file without the marker was treated as owned: %v", report.Removed)
	}
	for _, n := range second {
		if _, err := os.Stat(NotePath(root, n)); err != nil {
			t.Errorf("no file written for %s: %v", n.ID, err)
		}
	}
}

// TestRebuildLeavesAMarkerThatDoesNotClaimOwnership checks a file that carries the marker as false.
// Its frontmatter reads, so the file is judged and the marker answers that the graph does not own
// it, which is the case the action leaves in place.
func TestRebuildLeavesAMarkerThatDoesNotClaimOwnership(t *testing.T) {
	root := t.TempDir()
	second := loadGround(t, "./testdata/second")
	pkg := second[0].Package
	folder := Folder(root, pkg)

	if _, err := Rebuild(root, pkg, second); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	own := "---\ngenerated: false\ntitle: my note\n---\n\nwritten by hand\n"
	human := path.Join(folder, "own note.md")
	if err := os.WriteFile(human, []byte(own), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	report, err := Rebuild(root, pkg, second)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if _, err := os.ReadFile(human); err != nil {
		t.Errorf("the file whose marker does not claim ownership did not survive: %v", err)
	}
	if slices.Contains(report.Removed, human) {
		t.Errorf("the file whose marker does not claim ownership was treated as owned: %v", report.Removed)
	}
	if len(report.Unreadable) != 0 {
		t.Errorf("reported unreadable files = %v, want none, because its marker reads as a boolean", report.Unreadable)
	}
}

// TestRebuildStatesWhatItRemoved checks that the removals the action reports are the files it
// removed, including the file of a node that is no longer in the node set.
func TestRebuildStatesWhatItRemoved(t *testing.T) {
	root := t.TempDir()
	second := loadGround(t, "./testdata/second")
	pkg := second[0].Package

	if _, err := Rebuild(root, pkg, second); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	gone := second[len(second)-1]
	report, err := Rebuild(root, pkg, second[:len(second)-1])
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	want := make([]string, len(second))
	for i, n := range second {
		want[i] = NotePath(root, n)
	}
	slices.Sort(want)
	if !slices.Equal(report.Removed, want) {
		t.Errorf("reported removals = %v, want the files the action removed: %v", report.Removed, want)
	}
	if _, err := os.Stat(NotePath(root, gone)); err == nil {
		t.Errorf("the file of the node that no longer exists survived at %s", NotePath(root, gone))
	}
	for _, n := range second[:len(second)-1] {
		if _, err := os.Stat(NotePath(root, n)); err != nil {
			t.Errorf("the file of a node still in the node set was not written back: %v", err)
		}
	}
}

// TestRebuildWritesTheSameFileWhenOnlyTheClassificationChanges checks that a node whose declaration
// body changes from a struct to an interface is written to the same file, because the name is built
// from the node part of the identifier and from nothing a rule decides. The declaration is rewritten
// between runs, so the semantic type is recomputed by the build rather than assumed by the test. The
// action removes the marked file before writing it, so the assertion is about the mapping producing
// the same path, not about a name carried from the earlier build.
func TestRebuildWritesTheSameFileWhenOnlyTheClassificationChanges(t *testing.T) {
	root := t.TempDir()
	dir := groundModule(t, "package ground\n\ntype Thing struct {\n\tValue int\n}\n")

	before := loadDir(t, dir)
	thing := nodeByName(before, "Thing")
	if thing == nil {
		t.Fatalf("the ground produced no node named Thing: %v", nodeNames(before))
	}
	if thing.SemanticType != graph.SemanticStructure {
		t.Fatalf("Thing reads as %s, want the semantic type a struct body gives", thing.SemanticType)
	}

	if _, err := Rebuild(root, thing.Package, before); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	file := NotePath(root, thing)

	writeFile(t, dir, "ground.go", "package ground\n\ntype Thing interface {\n\tRead() int\n}\n")
	after := loadDir(t, dir)
	reclassified := nodeByName(after, "Thing")
	if reclassified == nil {
		t.Fatalf("the rewritten ground produced no node named Thing: %v", nodeNames(after))
	}
	if reclassified.SemanticType == thing.SemanticType {
		t.Fatalf("the rewritten ground still reads Thing as %s, want a build that classifies it differently",
			reclassified.SemanticType)
	}

	report, err := Rebuild(root, reclassified.Package, after)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	if NotePath(root, reclassified) != file {
		t.Errorf("the rewritten node maps to %s, want the file it mapped to before: %s", NotePath(root, reclassified), file)
	}
	if !slices.Contains(report.Removed, file) {
		t.Errorf("reported removals = %v, want the file the action removed before writing it again: %s", report.Removed, file)
	}
	if _, err := os.Stat(file); err != nil {
		t.Errorf("no file written for the rewritten node: %v", err)
	}

	entries, err := os.ReadDir(Folder(root, reclassified.Package))
	if err != nil {
		t.Fatalf("read folder: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("the folder holds %d files, want one for the node: %v", len(entries), nodeNames(after))
	}
}

// TestRebuildWritesTheNameTheCurrentSourceProduces checks that a change to the node part of the
// identifier gives a different name, and that the marked file holding the node under the earlier name
// is removed rather than left behind, because no name is carried from the build that wrote it.
func TestRebuildWritesTheNameTheCurrentSourceProduces(t *testing.T) {
	root := t.TempDir()
	dir := groundModule(t, "package ground\n\nfunc New() int { return 0 }\n")

	before := loadDir(t, dir)
	oldNode := nodeByName(before, "New")
	if oldNode == nil {
		t.Fatalf("the ground produced no node named New: %v", nodeNames(before))
	}
	if _, err := Rebuild(root, oldNode.Package, before); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	old := NotePath(root, oldNode)

	writeFile(t, dir, "ground.go", "package ground\n\nfunc NewThing() int { return 0 }\n")
	after := loadDir(t, dir)
	renamed := nodeByName(after, "NewThing")
	if renamed == nil {
		t.Fatalf("the rewritten ground produced no node named NewThing: %v", nodeNames(after))
	}

	report, err := Rebuild(root, renamed.Package, after)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	current := NotePath(root, renamed)
	if current == old {
		t.Fatalf("the renamed node maps to %s, want a name built from the node part the source now has", current)
	}
	if !slices.Contains(report.Removed, old) {
		t.Errorf("reported removals = %v, want the file that held the node under the earlier name: %s", report.Removed, old)
	}
	if _, err := os.Stat(old); err == nil {
		t.Errorf("the file under the earlier name survived at %s", old)
	}
	if _, err := os.Stat(current); err != nil {
		t.Errorf("no file written under the name this build produces: %v", err)
	}

	entries, err := os.ReadDir(Folder(root, renamed.Package))
	if err != nil {
		t.Fatalf("read folder: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("the folder holds %d files, want one for the node: %v", len(entries), nodeNames(after))
	}
}

// TestRebuildReplacesAFileNamedInTheEarlierForm checks that a marked file whose name the mapping no
// longer produces is removed and written again under the node part of the identifier, so the change of
// form happens through the action and a name this build does not produce is replaced rather than left
// behind in the folder.
func TestRebuildReplacesAFileNamedInTheEarlierForm(t *testing.T) {
	root := t.TempDir()
	second := loadGround(t, "./testdata/second")
	pkg := second[0].Package
	folder := Folder(root, pkg)

	thing := nodeByName(second, "Thing")
	if thing == nil {
		t.Fatalf("the second ground produced no node named Thing: %v", nodeNames(second))
	}
	written, err := Render(root, thing, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	content := readFile(t, written)
	if err := os.Remove(written); err != nil {
		t.Fatalf("remove: %v", err)
	}

	earlier := path.Join(folder, "structure_Thing.md")
	if err := os.WriteFile(earlier, content, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	report, err := Rebuild(root, pkg, second)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if !slices.Contains(report.Removed, earlier) {
		t.Errorf("reported removals = %v, want the file in the earlier form: %s", report.Removed, earlier)
	}
	if _, err := os.Stat(earlier); err == nil {
		t.Errorf("the file in the earlier form survived at %s", earlier)
	}

	current := NotePath(root, thing)
	result := Check(root, current)
	if !result.Confirmed {
		t.Errorf("check of %s = %s, want the file written in the current form to be confirmed", current, result.Reason)
	}

	entries, err := os.ReadDir(folder)
	if err != nil {
		t.Fatalf("read folder: %v", err)
	}
	if len(entries) != len(second) {
		t.Errorf("the folder holds %d files, want one for each of the %d nodes: %v", len(entries), len(second), nodeNames(second))
	}
}
