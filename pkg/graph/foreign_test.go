package graph

import (
	"testing"
)

const shapeFixturePath = "github.com/Galdoba/go-graph/pkg/graph/testdata/fixtures/shape"

// TestAReferenceNamesAMethodDeclaredByAnotherPackage checks that node-ness for a referent from another
// package is decided by the definition of a node rather than by that package's scope, which holds
// package-level declarations only and so cannot hold a method.
func TestAReferenceNamesAMethodDeclaredByAnotherPackage(t *testing.T) {
	p := loadFixture(t, "use")
	nodes, _ := mustBuild(t, p)

	shape := packageID(p.Imports[shapeFixturePath])
	use := mustNodeByName(t, nodes, "Use")

	read := attributeByName(use, "Read")
	if read == nil || read.Kind != AttrUse {
		t.Fatalf("the call of a method of another package is not recorded as a use: %v", use.Attributes)
	}
	if want := (ID{Package: shape, Receiver: "Shape", Name: "Read"}); read.Ref != want {
		t.Errorf("the reference names %v, want the method the imported package declares", want)
	}

	// A method reached through an embedding names the declaration the embedded type owns, which is the
	// entity the definition covers. The reachable position names no node, because a lookup produced it.
	count := attributeByName(use, "Count")
	if count == nil || count.Kind != AttrUse {
		t.Fatalf("a call reached through an embedding is not recorded as a use: %v", use.Attributes)
	}
	if want := (ID{Package: shape, Receiver: "Inner", Name: "Count"}); count.Ref != want {
		t.Errorf("the reference names %v, want the method the embedded type declares", want)
	}
	ref := attributeByName(use, "Wrapper")
	if ref == nil || ref.Kind != AttrUse {
		t.Fatalf("the type that carries the embedding is not recorded as a use: %v", use.Attributes)
	}
	if want := (ID{Package: shape, Name: "Wrapper"}); ref.Ref != want {
		t.Errorf("the reference names %v, want the type the imported package declares", want)
	}
}

// TestAQualifiedReferenceNamesTheEntityItDenotes checks that a reference through an import is recorded as
// a use of the entity the imported package declares whatever name the import wrote, and that the name the
// import wrote keeps its alias attribute beside it.
func TestAQualifiedReferenceNamesTheEntityItDenotes(t *testing.T) {
	p := loadFixture(t, "use")
	nodes, _ := mustBuild(t, p)

	use := mustNodeByName(t, nodes, "Use")
	shape := packageID(p.Imports[shapeFixturePath])

	lower := attributeByName(use, "ToLower")
	if lower == nil || lower.Kind != AttrUse {
		t.Fatalf("a reference through the name the package declares is not recorded as a use: %v", use.Attributes)
	}
	if want := (ID{Package: packageID(p.Imports["strings"]), Name: "ToLower"}); lower.Ref != want {
		t.Errorf("the reference names %v, want the entity the imported package declares", want)
	}

	alias := attributeByName(use, "shp")
	if alias == nil || alias.Kind != AttrAlias {
		t.Fatalf("the name the import wrote is not recorded as an alias attribute: %v", use.Attributes)
	}
	if alias.Value != shapeFixturePath {
		t.Errorf("alias value = %q, want the package it resolves to", alias.Value)
	}

	shown := attributeByName(use, "Shown")
	if shown == nil || shown.Kind != AttrUse {
		t.Fatalf("the entity the reference names is not recorded as a use: %v", use.Attributes)
	}
	if want := (ID{Package: shape, Name: "hidden"}); shown.Ref != want {
		t.Errorf("the reference names %v, want the entity the alias resolves to", want)
	}
}
