package graph

import (
	"go/ast"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// The two sources below let a test control what a reference names, so the record set of the
// second pass is read as the demand set rather than as the shape of a fixture on disk.
const foreignShapeSource = `package shape

// Shape is the entity a reference from another package names.
type Shape struct {
	Value int
}

func Read() error { return nil }

func (Shape) Read() error { return nil }

type Thing struct{}

func (Thing) Read() error { return nil }

// Reader states the method spec a record read from a name can carry.
type Reader interface {
	Read() error
}

// Wrapper reaches Count only through the type it embeds.
type Wrapper struct {
	Inner
}

type Inner struct{}

func (Inner) Count() int { return 0 }

type Unused struct{}
`

const foreignUseSource = `package use

import shp "mod.test/shape"

func UseShape() error {
	var s shp.Shape
	return s.Read()
}

func FirstReader() error {
	var r shp.Reader
	return r.Read()
}

func SecondReader(r shp.Reader) error {
	return r.Read()
}

func UseThing() error {
	var t shp.Thing
	return t.Read()
}

func UseRead() error { return shp.Read() }

func UseWrapper() int {
	var w shp.Wrapper
	return w.Count()
}
`

const singleNameUseSource = `package use

import shp "mod.test/shape"

func UseShape() error {
	var s shp.Shape
	_ = s
	return nil
}
`

// loadForeignPair type-checks a package that references an imported package, which is the
// situation the second pass reads: the referencing package is the target and the imported
// package is not.
func loadForeignPair(t *testing.T, useSource string) (*packages.Package, *packages.Package) {
	t.Helper()
	loader := newSourceLoader(t)
	shape := loader.load("mod.test", "", "mod.test/shape", map[string]string{"a.go": foreignShapeSource})
	use := loader.load("mod.test", "", "mod.test/use", map[string]string{"a.go": useSource})
	return use, shape
}

// TestTheSecondPassProducesOnlyTheRecordAReferenceNames checks that the second pass reads the
// names its own references carry rather than enumerating the imported package.
func TestTheSecondPassProducesOnlyTheRecordAReferenceNames(t *testing.T) {
	use, shape := loadForeignPair(t, singleNameUseSource)

	records, err := SecondPass(use)
	if err != nil {
		t.Fatalf("second pass: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("second pass produced %d records for %v, want one record for the name the reference names",
			len(records), nodeNames(records))
	}
	record := records[0]
	if record.Name != "Shape" {
		t.Errorf("record names %q, want the entity the reference names", record.Name)
	}
	if record.Provenance != ProvenanceFromName {
		t.Errorf("provenance = %v, want values read from a name", record.Provenance)
	}

	// The same entity read as a target is the same node, so the two passes share one
	// identifier space and a consumer does not have to match them by name.
	target, _ := mustBuild(t, shape)
	targetRecord := mustNodeByName(t, target, "Shape")
	if record.ID != targetRecord.ID {
		t.Errorf("record identifier %v, want the identifier the first pass produces for the same entity: %v",
			record.ID, targetRecord.ID)
	}
	if targetRecord.Provenance != ProvenanceFromSource {
		t.Errorf("the record of the package read as a target states provenance %v, want values read from a declaration",
			targetRecord.Provenance)
	}
}

// TestTwoReferencesFromTwoDeclarationsNameOneForeignRecord checks that the second pass keeps no
// identifier space of its own, so the same entity named from two declarations is one record.
func TestTwoReferencesFromTwoDeclarationsNameOneForeignRecord(t *testing.T) {
	use, _ := loadForeignPair(t, foreignUseSource)

	records, err := SecondPass(use)
	if err != nil {
		t.Fatalf("second pass: %v", err)
	}
	if record := nodeByID(records, "mod.test~shape~shape.Reader"); record == nil {
		t.Fatalf("no record for the entity two declarations name, got %v", nodeNames(records))
	}

	// A package-level name and a method of the same name are two entities, so the receiver
	// keeps them apart.
	packageLevel := nodeByID(records, "mod.test~shape~shape.Read")
	if packageLevel == nil {
		t.Fatalf("no record for the package-level name, got %v", nodeNames(records))
	}
	method := nodeByID(records, "mod.test~shape~shape.Thing.Read")
	if method == nil {
		t.Fatalf("no record for the method of the same name, got %v", nodeNames(records))
	}
	if packageLevel.ID == method.ID {
		t.Errorf("the package-level name and the method share %v, want two records", packageLevel.ID)
	}

	// A method record keeps the parent that names the type declaring it, which the shape of the
	// record reads from the entity itself.
	thing := mustNodeByID(t, records, "mod.test~shape~shape.Thing")
	if method.Parent != thing.ID {
		t.Errorf("method record has parent %v, want the record for the type that declares it: %v", method.Parent, thing.ID)
	}
	if method.Receiver != "Thing" {
		t.Errorf("method record has receiver %q, want the type that declares it", method.Receiver)
	}
	if packageLevel.Parent.Name != "" {
		t.Errorf("package-level record has parent %v, want the package namespace", packageLevel.Parent)
	}
}

// TestAForeignRecordCarriesTheRegionAndCommentTheSyntaxShows checks that the region and the
// comment of a record read from a name come from the syntax the load reports for the package
// that declares the entity.
func TestAForeignRecordCarriesTheRegionAndCommentTheSyntaxShows(t *testing.T) {
	use, _ := loadForeignPair(t, foreignUseSource)

	records, err := SecondPass(use)
	if err != nil {
		t.Fatalf("second pass: %v", err)
	}
	record := mustNodeByID(t, records, "mod.test~shape~shape.Shape")

	if record.Decl == nil {
		t.Fatal("the record carries no reference to the declaration the syntax shows")
	}
	decl, ok := record.Decl.(*ast.GenDecl)
	if !ok {
		t.Fatalf("the record carries %T, want the declaration that created it", record.Decl)
	}
	if decl.Doc == nil || decl.Doc.Text() != "Shape is the entity a reference from another package names.\n" {
		t.Errorf("the record carries no comment above the declaration: %v", decl.Doc)
	}

	if !strings.HasSuffix(record.Pos.File, "mod.test/shape/a.go") || record.Pos.Line == 0 || record.Pos.Column == 0 {
		t.Errorf("record states region %v, want the region the reported syntax shows in the file of the package that declares the entity", record.Pos)
	}
	if record.Range.StartLine == 0 || record.Range.EndLine < record.Range.StartLine {
		t.Errorf("record states range %v, want the range of the declaration spec", record.Range)
	}

	// A method record carries the receiver that names the type it is declared on.
	method := mustNodeByID(t, records, "mod.test~shape~shape.Shape.Read")
	if method.Receiver != "Shape" || method.SemanticType != SemanticMethod {
		t.Errorf("method record states receiver %q and semantic type %v", method.Receiver, method.SemanticType)
	}
	if method.Provenance != ProvenanceFromName {
		t.Errorf("provenance = %v, want values read from a name", method.Provenance)
	}
}

// TestAForeignRecordThatStatesNoRegionCarriesOnlyWhatANameStates checks that a load which reports
// no syntax for the package that declares the entity leaves the region and the comment unstated,
// so a consumer reads the absence as a value that could not be stated.
func TestAForeignRecordThatStatesNoRegionCarriesOnlyWhatANameStates(t *testing.T) {
	use, shape := loadForeignPair(t, singleNameUseSource)
	shape.Syntax = nil

	records, err := SecondPass(use)
	if err != nil {
		t.Fatalf("second pass: %v", err)
	}
	record := mustNodeByID(t, records, "mod.test~shape~shape.Shape")

	if record.Decl != nil || record.Pos != (Position{}) || record.Range != (LineRange{}) {
		t.Errorf("record states a region it could not have read: %v %v", record.Pos, record.Range)
	}
	if record.Provenance != ProvenanceFromName {
		t.Errorf("provenance = %v, want the record to state that its values were read from a name", record.Provenance)
	}
	if record.Kind != KindType || record.SemanticType != SemanticStructure {
		t.Errorf("record states kind %v and semantic type %v, want what a name and a type state",
			record.Kind, record.SemanticType)
	}
}

// TestAForeignRecordCarriesTheShapeItStatesAndNoUses checks that a record read from a name carries
// the method specs an interface states and only the methods the package declares on a type, while
// it carries no uses, because no declaration body of that package was read.
func TestAForeignRecordCarriesTheShapeItStatesAndNoUses(t *testing.T) {
	use, _ := loadForeignPair(t, foreignUseSource)

	records, err := SecondPass(use)
	if err != nil {
		t.Fatalf("second pass: %v", err)
	}

	for _, record := range records {
		for _, attribute := range record.Attributes {
			if attribute.Kind == AttrUse {
				t.Errorf("record %v carries a use, which no body read could have stated", record.ID)
			}
		}
	}

	reader := mustNodeByID(t, records, "mod.test~shape~shape.Reader")
	spec := attributeByName(reader, "Read")
	if spec == nil || spec.Kind != AttrMethodSpec {
		t.Fatalf("the interface record carries no method spec: %v", reader.Attributes)
	}
	if spec.Value != "func() error" {
		t.Errorf("method spec value = %q, want the signature without the receiver", spec.Value)
	}

	wrapper := mustNodeByID(t, records, "mod.test~shape~shape.Wrapper")
	embedded := attributeByName(wrapper, "Inner")
	if embedded == nil || embedded.Kind != AttrEmbeds {
		t.Fatalf("the record for the type that embeds carries no embedding: %v", wrapper.Attributes)
	}
	if want := (ID{Package: PackageID{ModulePath: "mod.test", PackagePath: "mod.test/shape", Name: "shape"}, Name: "Inner"}); embedded.Ref != want {
		t.Errorf("the embedding names %v, want the entity the type embeds", embedded.Ref)
	}
	if len(wrapper.Attributes) != 1 {
		t.Errorf("the record for a type that reaches a method only through an embedding carries %v, want the embedding alone",
			wrapper.Attributes)
	}

	// A method reached through an embedding names the declaration the embedded type owns, so the
	// record is produced for that method even though its parent has no record of its own.
	count := mustNodeByID(t, records, "mod.test~shape~shape.Inner.Count")
	if count.Receiver != "Inner" {
		t.Errorf("method record has receiver %q, want the type that declares it", count.Receiver)
	}
	if nodeByID(records, "mod.test~shape~shape.Inner") != nil {
		t.Errorf("the second pass produced a record for a name no reference named: %v", nodeNames(records))
	}
}

// TestAnExportedAliasProducesTheRecordOfTheEntityItResolvesTo checks that the demand set holds the entity
// a name resolves to rather than the alias, and that the record states the exported value that entity has,
// which is not the value the exported name the import wrote.
func TestAnExportedAliasProducesTheRecordOfTheEntityItResolvesTo(t *testing.T) {
	loader := newSourceLoader(t)
	shape := loader.load("mod.test", "", "mod.test/shape", map[string]string{"a.go": `package shape

// hidden is the entity the package does not export.
type hidden struct {
	Value int
}

// Shown is an exported name for that entity.
type Shown = hidden
`})
	use := loader.load("mod.test", "", "mod.test/use", map[string]string{"a.go": `package use

import shp "mod.test/shape"

func Use() error {
	var shown shp.Shown
	_ = shown
	return nil
}
`})

	records, err := SecondPass(use)
	if err != nil {
		t.Fatalf("second pass: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("second pass produced %d records, want one for the entity the alias resolves to: %v",
			len(records), nodeNames(records))
	}
	record := records[0]
	if record.Name != "hidden" {
		t.Errorf("the record names %q, want the entity the alias resolves to", record.Name)
	}
	if record.Properties.Exported {
		t.Errorf("the record states the exported value %v, want the value the entity it names has",
			record.Properties.Exported)
	}

	target, _ := mustBuild(t, shape)
	targetRecord := mustNodeByName(t, target, "hidden")
	if record.ID != targetRecord.ID {
		t.Errorf("record identifier %v, want the identifier the first pass produces for the same entity: %v",
			record.ID, targetRecord.ID)
	}
	if record.Properties.Exported != targetRecord.Properties.Exported {
		t.Errorf("the record read from a name states exported %v, want the value the entity has: %v",
			record.Properties.Exported, targetRecord.Properties.Exported)
	}
}
