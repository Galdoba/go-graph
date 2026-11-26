package vault

import (
	"errors"
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/Galdoba/go-graph/pkg/graph"
	"github.com/Galdoba/go-graph/pkg/link"
)

// TestFrontmatterWritesTheFixedFieldOrder checks the order the fields are written in and that a
// field the record does not carry is not written as an empty value.
func TestFrontmatterWritesTheFixedFieldOrder(t *testing.T) {
	root := t.TempDir()
	second := loadGround(t, "./testdata/second")

	thing := nodeByName(second, "Thing")
	if thing == nil {
		t.Fatalf("the second ground produced no node named Thing: %v", nodeNames(second))
	}

	// The ground node has no module version, so the version field is a field the record does not
	// carry, and the properties are cleared to show the same rule for a list field.
	record := *thing
	record.Properties = graph.NodeProperties{}

	file, err := Render(root, &record, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	lines := strings.Split(string(readFile(t, file)), "\n")
	if lines[0] != "---" {
		t.Fatalf("first line = %q, want a frontmatter fence", lines[0])
	}
	if lines[1] != "generated: true" {
		t.Errorf("first field = %q, want generated: true", lines[1])
	}

	var written []string
	for _, line := range lines[2:] {
		if line == "---" {
			break
		}
		if line == "" {
			continue
		}
		field, _, ok := strings.Cut(line, ":")
		if !ok {
			t.Fatalf("frontmatter line without a value: %q", line)
		}
		written = append(written, field)
	}

	want := []string{"module", "package_path", "package", "kind", "semantic_type", "file", "line", "column", "start_line", "end_line"}
	if !slices.Equal(written, want) {
		t.Errorf("fields written = %v, want %v", written, want)
	}
}

// TestNoteStatesWhatItsNodeIs checks that the file names the kind, the semantic type and the
// properties without the build that produced it, because the file name states the semantic type
// only.
func TestNoteStatesWhatItsNodeIs(t *testing.T) {
	root := t.TempDir()
	ground := loadGround(t, "./testdata")

	box := nodeByName(ground, "Box")
	if box == nil {
		t.Fatalf("the test ground produced no node named Box: %v", nodeNames(ground))
	}

	file, err := Render(root, box, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	record, err := Read(file)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if record.Kind != "type" {
		t.Errorf("kind read from the file = %s, want type", record.Kind)
	}
	if record.SemanticType != "structure" {
		t.Errorf("semantic type read from the file = %s, want structure", record.SemanticType)
	}
	if !slices.Contains(record.Properties, "exported") || !slices.Contains(record.Properties, "generic") {
		t.Errorf("properties read from the file = %v, want exported and generic", record.Properties)
	}

	if content := string(readFile(t, file)); !strings.Contains(content, "properties: [exported, generic]") {
		t.Errorf("the note does not state both properties:\n%s", content)
	}
}

// TestNoteStatesBothPropertiesForAGenericEntryPointNode checks a node that reports generic and
// entry_point, which no node of the test ground reports together, so the record is assembled with
// both properties and the file is read back without the build.
func TestNoteStatesBothPropertiesForAGenericEntryPointNode(t *testing.T) {
	root := t.TempDir()
	second := loadGround(t, "./testdata/second")

	node := nodeByName(second, "New")
	if node == nil {
		t.Fatalf("the second ground produced no node named New: %v", nodeNames(second))
	}

	record := *node
	record.Properties = graph.NodeProperties{Generic: true, EntryPoint: true}

	file, err := Render(root, &record, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	read, err := Read(file)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !slices.Contains(read.Properties, "generic") || !slices.Contains(read.Properties, "entry_point") {
		t.Errorf("properties read from the file = %v, want generic and entry_point", read.Properties)
	}
	if content := string(readFile(t, file)); !strings.Contains(content, "properties: [generic, entry_point]") {
		t.Errorf("the note does not state both properties:\n%s", content)
	}
}

// TestGenerationIsDeterministic checks that the same record produces the same location, the same
// name and the same bytes.
func TestGenerationIsDeterministic(t *testing.T) {
	root := t.TempDir()
	second := loadGround(t, "./testdata/second")

	thing := nodeByName(second, "Thing")
	if thing == nil {
		t.Fatalf("the second ground produced no node named Thing: %v", nodeNames(second))
	}

	first, err := Render(root, thing, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	again, err := Render(root, thing, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if first != again {
		t.Errorf("second rendering at %s, want %s", again, first)
	}
	if string(readFile(t, first)) != string(readFile(t, again)) {
		t.Errorf("second rendering differs from the first for %s", first)
	}
}

func readFile(t *testing.T, file string) []byte {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	return data
}

// TestAForeignNoteHoldsTheValuesItsRecordStates renders a note for a record produced from a
// reference. The load reported the syntax that declares the entity, so the note states the region and
// the comment above the declaration, and its `## Related` section lists the connection that points at
// it, which is the reference that named it, because a record produced from a reference carries nothing
// it could point at.
func TestAForeignNoteHoldsTheValuesItsRecordStates(t *testing.T) {
	root := t.TempDir()
	_, first, second := loadForeignPair(t, "./pkg/graph/testdata/fixtures/use")
	shape := nodeByID(second, "github.com~Galdoba~go-graph~pkg~graph~testdata~fixtures~shape~shape.Shape")
	if shape == nil {
		t.Fatalf("the second pass produced no record for the shape a reference names: %v", nodeNames(second))
	}

	file, err := Render(root, shape, append(first, second...))
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if file != path.Join(Folder(root, shape.Package), "Shape.md") {
		t.Errorf("the note was written at %s, want the folder the referenced package path maps to", file)
	}

	record, err := Read(file)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if record.File != shape.Pos.File || record.Line != shape.Pos.Line || record.Column != shape.Pos.Column ||
		record.StartLine != shape.Range.StartLine || record.EndLine != shape.Range.EndLine {
		t.Errorf("the region read from the file does not match the region the record states: %+v", record)
	}

	content := string(readFile(t, file))
	if !strings.Contains(content, "// Shape is the type a reference from another package names.") {
		t.Errorf("the note does not state the comment above the declaration:\n%s", content)
	}
	if !strings.Contains(content, "type Shape struct {") {
		t.Errorf("the note does not state the code snippet of the declaration:\n%s", content)
	}

	want := "- [[github.com/Galdoba/go-graph/pkg/graph/testdata/fixtures/use/Use.md|Use]]"
	if !strings.Contains(content, want) {
		t.Errorf("the Related section does not list the link that points at the record, want %s:\n%s", want, content)
	}
}

// TestAForeignNoteWritesNoFieldItsRecordDoesNotState renders a record produced from a reference for an
// entity whose package the load reported no syntax for. The record states no region, so the note writes
// no region fields and no code snippet, and the rule that reads the uses a record carries is not
// answered against it, because the absence of uses is a value that was never read rather than a claim
// that the entity refers to nothing.
func TestAForeignNoteWritesNoFieldItsRecordDoesNotState(t *testing.T) {
	first, second := loadForeignPairWithoutSyntax(t, "./pkg/graph/testdata/fixtures/use")
	hidden := nodeByID(second, "github.com~Galdoba~go-graph~pkg~graph~testdata~fixtures~shape~shape.hidden")
	if hidden == nil {
		t.Fatalf("the second pass produced no record for the entity the alias names: %v", nodeNames(second))
	}
	if hidden.Pos.File != "" || hidden.Decl != nil {
		t.Fatalf("the record of %s states a region, want one that states none", hidden.ID)
	}

	root := t.TempDir()
	file, err := Render(root, hidden, append(first, second...))
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	content := string(readFile(t, file))
	lines := strings.Split(content, "\n")
	if lines[0] != "---" {
		t.Fatalf("first line = %q, want a frontmatter fence", lines[0])
	}
	var written []string
	for _, line := range lines[1:] {
		if line == "---" {
			break
		}
		if line == "" {
			continue
		}
		field, _, ok := strings.Cut(line, ":")
		if !ok {
			t.Fatalf("frontmatter line without a value: %q", line)
		}
		written = append(written, field)
	}

	wantFields := []string{"generated", "module", "package_path", "package", "kind", "semantic_type"}
	if !slices.Equal(written, wantFields) {
		t.Errorf("fields written = %v, want only the parts the record states: %v", written, wantFields)
	}
	if strings.Contains(content, "```go") {
		t.Errorf("the note wrote a code snippet for a record that states no region:\n%s", content)
	}

	yes, err := link.Exists(hidden, nodeByName(first, "Use"))
	var undecided *link.UndecidedError
	if err == nil || !errors.As(err, &undecided) || undecided.Rule != "node to a node a reference it carries names" {
		t.Errorf("the reference rule answered %v for a record that states no uses, want it unable to answer", err)
	}
	if yes {
		t.Errorf("the reference rule answered yes for a record that states no uses")
	}

	want := "- [[github.com/Galdoba/go-graph/pkg/graph/testdata/fixtures/use/Use.md|Use]]"
	if !strings.Contains(content, want) {
		t.Errorf("the Related section does not list the link that points at the record, want %s:\n%s", want, content)
	}
}

// TestANoteForANodeThatIsAMemberOfNoPairStatesNoneDetected checks the section for a record that no pair
// names, which is what a record whose fields name no node gives: nothing it carries points at a record and
// no other record names it, so the section is present and states that nothing was detected.
func TestANoteForANodeThatIsAMemberOfNoPairStatesNoneDetected(t *testing.T) {
	root := t.TempDir()
	nodes := loadGround(t, "./pkg/graph/testdata/fixtures/basic")

	counter := nodeByName(nodes, "GlobalCounter")
	if counter == nil {
		t.Fatalf("the ground produced no node named GlobalCounter: %v", nodeNames(nodes))
	}
	for _, pair := range link.Pairs(nodes) {
		if pair.Source == counter.ID || pair.Target == counter.ID {
			t.Fatalf("the record %s is a member of the pair %v, want a record that is a member of no pair",
				counter.ID, pair)
		}
	}

	file, err := Render(root, counter, nodes)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	content := string(readFile(t, file))
	start := strings.Index(content, "## Related")
	if start < 0 {
		t.Fatalf("the file carries no Related section:\n%s", content)
	}
	if !strings.Contains(content[start:], "None detected") {
		t.Errorf("the Related section does not state that nothing was detected:\n%s", content[start:])
	}
}

// TestARecordReadFromSourceListsTheConnectionsItPointsAtInTheCollectionOfTheRun pins the direction the
// section answers for a record the build produced: it lists the pairs that have that record as their
// source, because such a record states the uses its definition carries. The run reads both fixtures as
// targets, so the record of the shape is read from source, and the interface in the same package that
// answers to it is a pair that names it as its target, which the section does not list.
func TestARecordReadFromSourceListsTheConnectionsItPointsAtInTheCollectionOfTheRun(t *testing.T) {
	root := t.TempDir()
	_, records := loadRun(t, "./pkg/graph/testdata/fixtures/use", "./pkg/graph/testdata/fixtures/shape")

	shape := nodeByID(records, "github.com~Galdoba~go-graph~pkg~graph~testdata~fixtures~shape~shape.Shape")
	if shape == nil {
		t.Fatalf("the collection holds no record for the shape: %v", nodeNames(records))
	}
	if shape.Provenance != graph.ProvenanceFromSource {
		t.Fatalf("the record of %s was read from %v, want a record the run read as a target", shape.ID, shape.Provenance)
	}

	file, err := Render(root, shape, records)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	content := string(readFile(t, file))
	section := content[strings.Index(content, "## Related"):]

	want := "- [[github.com/Galdoba/go-graph/pkg/graph/testdata/fixtures/shape/Shape.Read.md|Shape.Read]]"
	if !strings.Contains(section, want) {
		t.Errorf("the Related section does not list the pair the record is the source of, want %s:\n%s", want, section)
	}
	if strings.Contains(section, "Reader.md") {
		t.Errorf("the Related section of a record read from source lists a pair that points at it:\n%s", section)
	}
}

// TestARecordReadFromANameListsTheConnectionsThatPointAtItInTheCollectionOfTheRun pins the direction the
// section answers for a record produced from a reference, which is the record of a package the run did
// not read as a target. Such a record states no uses, so the only pair it can be a member of is the one
// the reference names it as the target of, and the pair the shape would be the source of is not listed.
func TestARecordReadFromANameListsTheConnectionsThatPointAtItInTheCollectionOfTheRun(t *testing.T) {
	root := t.TempDir()
	_, first, second := loadForeignPair(t, "./pkg/graph/testdata/fixtures/use")
	records := graph.Combine(append(first, second...))

	shape := nodeByID(records, "github.com~Galdoba~go-graph~pkg~graph~testdata~fixtures~shape~shape.Shape")
	if shape == nil {
		t.Fatalf("the collection holds no record for the shape a reference names: %v", nodeNames(records))
	}
	if shape.Provenance != graph.ProvenanceFromName {
		t.Fatalf("the record of %s was read from %v, want a record produced from a name", shape.ID, shape.Provenance)
	}

	file, err := Render(root, shape, records)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	content := string(readFile(t, file))
	section := content[strings.Index(content, "## Related"):]

	want := "- [[github.com/Galdoba/go-graph/pkg/graph/testdata/fixtures/use/Use.md|Use]]"
	if !strings.Contains(section, want) {
		t.Errorf("the Related section does not list the pair that names the record as its target, want %s:\n%s", want, section)
	}
	if strings.Contains(section, "Shape.Read.md") {
		t.Errorf("the Related section of a record produced from a name lists a pair it is the source of:\n%s", section)
	}
}
