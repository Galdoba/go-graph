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

	if _, err := Rebuild(root, []graph.PackageID{pkg}, second); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	own := "---\ntitle: my note\n---\n\nwritten by hand\n"
	human := path.Join(folder, "own note.md")
	if err := os.WriteFile(human, []byte(own), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	report, err := Rebuild(root, []graph.PackageID{pkg}, second)
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

	if _, err := Rebuild(root, []graph.PackageID{pkg}, second); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	own := "---\ngenerated: false\ntitle: my note\n---\n\nwritten by hand\n"
	human := path.Join(folder, "own note.md")
	if err := os.WriteFile(human, []byte(own), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	report, err := Rebuild(root, []graph.PackageID{pkg}, second)
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

	if _, err := Rebuild(root, []graph.PackageID{pkg}, second); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	gone := second[len(second)-1]
	report, err := Rebuild(root, []graph.PackageID{pkg}, second[:len(second)-1])
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

	if _, err := Rebuild(root, []graph.PackageID{thing.Package}, before); err != nil {
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

	report, err := Rebuild(root, []graph.PackageID{reclassified.Package}, after)
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
	if _, err := Rebuild(root, []graph.PackageID{oldNode.Package}, before); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	old := NotePath(root, oldNode)

	writeFile(t, dir, "ground.go", "package ground\n\nfunc NewThing() int { return 0 }\n")
	after := loadDir(t, dir)
	renamed := nodeByName(after, "NewThing")
	if renamed == nil {
		t.Fatalf("the rewritten ground produced no node named NewThing: %v", nodeNames(after))
	}

	report, err := Rebuild(root, []graph.PackageID{renamed.Package}, after)
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

	report, err := Rebuild(root, []graph.PackageID{pkg}, second)
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

// TestRebuildRemovesOnlyTheFilesItsOwnRecordsMapToInAForeignFolder checks that a run writing in a
// folder of a package it did not build removes only the files its own records map to there, so a file
// another run wrote in the same folder survives. The two runs build the same package and each carry a
// different record of the same foreign package, which is the pair the action has to tell apart.
func TestRebuildRemovesOnlyTheFilesItsOwnRecordsMapToInAForeignFolder(t *testing.T) {
	root := t.TempDir()
	_, built, records := loadForeignPair(t, "./pkg/graph/testdata/fixtures/use")
	pkg := built[0].Package

	own := make([]string, len(built))
	for i, n := range built {
		own[i] = NotePath(root, n)
	}
	slices.Sort(own)

	kept := nodeByID(records, "github.com~Galdoba~go-graph~pkg~graph~testdata~fixtures~shape~shape.Shape")
	if kept == nil {
		t.Fatalf("the second pass produced no record for the type a reference names: %v", nodeNames(records))
	}
	other := nodeByID(records, "github.com~Galdoba~go-graph~pkg~graph~testdata~fixtures~shape~shape.Inner.Count")
	if other == nil {
		t.Fatalf("the second pass produced no record for a method of a type a reference names: %v", nodeNames(records))
	}

	// The first run writes the file of the record it carries, which is the file the second run has to
	// leave in place.
	if _, err := Rebuild(root, []graph.PackageID{pkg}, append(append([]*graph.Node{}, built...), kept)); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	// In the foreign folder the second run has no file of its own yet, so it judges nothing there.
	secondRun, err := Rebuild(root, []graph.PackageID{pkg}, append(append([]*graph.Node{}, built...), other))
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if !slices.Equal(secondRun.Removed, own) {
		t.Errorf("reported removals = %v, want only the files of the package this run built: %v", secondRun.Removed, own)
	}
	if len(secondRun.Unreadable) != 0 {
		t.Errorf("reported unreadable files = %v, want none, because the action judged no file it did not map to", secondRun.Unreadable)
	}
	if _, err := os.Stat(NotePath(root, kept)); err != nil {
		t.Errorf("the file the other run wrote in the foreign folder did not survive: %v", err)
	}

	// The second run now owns the file it wrote there, so it is the only foreign file it removes.
	want := append([]string{NotePath(root, other)}, own...)
	slices.Sort(want)
	secondRun, err = Rebuild(root, []graph.PackageID{pkg}, append(append([]*graph.Node{}, built...), other))
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if !slices.Equal(secondRun.Removed, want) {
		t.Errorf("reported removals = %v, want the files this run mapped to: %v", secondRun.Removed, want)
	}
	if len(secondRun.Unreadable) != 0 {
		t.Errorf("reported unreadable files = %v, want none, because the action judged no file it did not map to", secondRun.Unreadable)
	}
	if _, err := os.Stat(NotePath(root, kept)); err != nil {
		t.Errorf("the file the other run wrote in the foreign folder did not survive: %v", err)
	}

	// The same assertion mirrored: the run that produced the first file keeps the file the other run
	// wrote, so ownership in a foreign folder follows the records of the run rather than the folder.
	firstRun, err := Rebuild(root, []graph.PackageID{pkg}, append(append([]*graph.Node{}, built...), kept))
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	want = append([]string{NotePath(root, kept)}, own...)
	slices.Sort(want)
	if !slices.Equal(firstRun.Removed, want) {
		t.Errorf("reported removals = %v, want the files this run mapped to: %v", firstRun.Removed, want)
	}
	if len(firstRun.Unreadable) != 0 {
		t.Errorf("reported unreadable files = %v, want none, because the action judged no file it did not map to", firstRun.Unreadable)
	}
	if _, err := os.Stat(NotePath(root, other)); err != nil {
		t.Errorf("the file the other run wrote in the foreign folder did not survive: %v", err)
	}
}

// TestRebuildOwnsEveryMarkedFileInTheFoldersOfThePackagesTheRunRead checks a run that read two
// packages as targets: it removes every marked file in the folder of each of those packages,
// including a file no record of the run maps to, while in a folder that only records produced from a
// reference map to it removes only the files this run wrote there. The two fixtures name each other
// only through a reference, so the collection of the run holds the records of both packages and one
// record of a package the run never read.
func TestRebuildOwnsEveryMarkedFileInTheFoldersOfThePackagesTheRunRead(t *testing.T) {
	root := t.TempDir()
	targets, records := loadRun(t, "./pkg/graph/testdata/fixtures/use", "./pkg/graph/testdata/fixtures/shape")
	if len(targets) != 2 {
		t.Fatalf("the run read %d packages as targets, want two", len(targets))
	}

	owned := make(map[string]bool, len(targets))
	for _, pkg := range targets {
		owned[Folder(root, pkg)] = true
	}

	var foreign *graph.Node
	for _, record := range records {
		if !owned[Folder(root, record.Package)] {
			foreign = record
			break
		}
	}
	if foreign == nil {
		t.Fatalf("the collection maps to no folder the run read as a target: %v", nodeNames(records))
	}

	if _, err := Rebuild(root, targets, records); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	marker := "---\ngenerated: true\n---\n\nwritten by another run\n"
	inOwned := path.Join(Folder(root, targets[0]), "gone record.md")
	inForeign := path.Join(Folder(root, foreign.Package), "other record.md")
	for _, file := range []string{inOwned, inForeign} {
		if err := os.MkdirAll(path.Dir(file), 0o755); err != nil {
			t.Fatalf("create folder for %s: %v", file, err)
		}
		if err := os.WriteFile(file, []byte(marker), 0o644); err != nil {
			t.Fatalf("write %s: %v", file, err)
		}
	}

	report, err := Rebuild(root, targets, records)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	want := make([]string, 0, len(records)+1)
	for _, record := range records {
		want = append(want, NotePath(root, record))
	}
	want = append(want, inOwned)
	slices.Sort(want)
	got := append([]string{}, report.Removed...)
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("reported removals = %v, want the files of the records of the run and the marked file in the folder the run owns in full: %v", got, want)
	}

	if _, err := os.ReadFile(inForeign); err != nil {
		t.Errorf("the marked file in a folder the run does not own in full did not survive: %v", err)
	}
	if slices.Contains(report.Removed, inForeign) {
		t.Errorf("a marked file the run did not write was treated as owned: %v", report.Removed)
	}
	for _, record := range records {
		if _, err := os.Stat(NotePath(root, record)); err != nil {
			t.Errorf("no file written for %s: %v", record.ID, err)
		}
	}
}
