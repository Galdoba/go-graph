package vault

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/Galdoba/go-graph/pkg/graph"
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
