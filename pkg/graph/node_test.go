package graph

import (
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"strings"
	"testing"
)

// TestNodeRecordReportsEveryField checks that a node built from a package-level
// declaration carries the whole record: identifier, kind, name, package, parent, receiver,
// the properties part and the declaration position.
func TestNodeRecordReportsEveryField(t *testing.T) {
	p := loadFixture(t, "basic")
	nodes, _ := mustBuild(t, p)

	n := nodeByName(nodes, "GlobalCounter")
	if n == nil {
		t.Fatalf("no node for the package-level declaration, got %v", nodeNames(nodes))
	}

	packageID := PackageID{ModulePath: p.Module.Path, ModuleVersion: p.Module.Version, PackagePath: p.PkgPath, Name: p.Name}
	if n.Kind != KindVar {
		t.Errorf("kind = %v, want %v", n.Kind, KindVar)
	}
	if n.Name != "GlobalCounter" {
		t.Errorf("name = %q, want GlobalCounter", n.Name)
	}
	if n.Package != packageID {
		t.Errorf("package = %v, want %v", n.Package, packageID)
	}
	if n.ID.Name != "GlobalCounter" || n.ID.Package != packageID {
		t.Errorf("id = %v, want %v", n.ID, ID{Package: packageID, Name: "GlobalCounter"})
	}
	if n.Parent.Name != "" || n.Parent.Package != packageID {
		t.Errorf("parent = %v, want the package namespace %v", n.Parent, ID{Package: packageID})
	}
	if n.Receiver != "" {
		t.Errorf("receiver = %q, want empty for a variable", n.Receiver)
	}
	if !n.Properties.Exported {
		t.Error("exported = false, want true")
	}
	if n.Pos.File == "" || n.Pos.Line == 0 || n.Pos.Column == 0 {
		t.Errorf("position = %v, want a declared position", n.Pos)
	}

	line := declarationLine(t, p.GoFiles[0], "var GlobalCounter int")
	if n.Pos.File != p.GoFiles[0] {
		t.Errorf("position file = %q, want the file the record was read from %q", n.Pos.File, p.GoFiles[0])
	}
	if n.Pos.Line != line {
		t.Errorf("position line = %d, want the declaration line %d", n.Pos.Line, line)
	}
	if n.Range.StartLine != line || n.Range.EndLine != line {
		t.Errorf("range = %v, want the range of its own declaration spec at line %d", n.Range, line)
	}
	if _, ok := n.Decl.(*ast.GenDecl); !ok {
		t.Errorf("declaration = %T, want a GenDecl", n.Decl)
	}
}

// TestClosedSetOfNodeKinds checks that the four kinds are accepted and that a candidate
// of any other kind is reported as excluded rather than recorded.
func TestClosedSetOfNodeKinds(t *testing.T) {
	pkg := types.NewPackage("github.com/user/project1/pkg1", "pkg1")

	cases := []struct {
		what     string
		obj      types.Object
		kind     Kind
		recorded bool
		reason   string
	}{
		{"type", types.NewTypeName(token.NoPos, pkg, "Struct1", nil), KindType, true, ""},
		{"function", types.NewFunc(token.NoPos, pkg, "Read", types.NewSignatureType(nil, nil, nil, nil, nil, false)), KindFunc, true, ""},
		{"variable", types.NewVar(token.NoPos, pkg, "GlobalCounter", types.Typ[types.Int]), KindVar, true, ""},
		{"constant", types.NewConst(token.NoPos, pkg, "MaxRetries", types.Typ[types.Int], nil), KindConst, true, ""},
		{"label", types.NewLabel(token.NoPos, pkg, "loop"), KindUnknown, false, ReasonOutsideClosedSet},
		{"import alias", types.NewPkgName(token.NoPos, pkg, "fx", types.NewPackage("fmt", "fmt")), KindUnknown, false, ReasonPackageIsNamespace},
		{"predeclared function", types.Universe.Lookup("len"), KindUnknown, false, ReasonPredeclaredIdentifier},
		{"predeclared type", types.Universe.Lookup("error"), KindType, false, ReasonPredeclaredIdentifier},
		{"no declaration site", nil, KindUnknown, false, ReasonNoDeclarationSite},
	}

	for _, c := range cases {
		got := Classify(pkg, c.obj)
		if got.Recorded != c.recorded {
			t.Errorf("%s: recorded = %v, want %v", c.what, got.Recorded, c.recorded)
		}
		if got.Kind != c.kind {
			t.Errorf("%s: kind = %v, want %v", c.what, got.Kind, c.kind)
		}
		if got.Reason != c.reason {
			t.Errorf("%s: reason = %q, want %q", c.what, got.Reason, c.reason)
		}
		if c.recorded && got.Name != c.obj.Name() {
			t.Errorf("%s: name = %q, want %q", c.what, got.Name, c.obj.Name())
		}
	}
}

func TestSemanticSetHasExactlyTheDocumentedValues(t *testing.T) {
	documented := map[SemanticType]string{
		SemanticStructure:  "structure",
		SemanticInterface:  "interface",
		SemanticFunction:   "function",
		SemanticMethod:     "method",
		SemanticConstant:   "constant",
		SemanticVariable:   "variable",
		SemanticUnresolved: "unresolved",
	}

	for st, want := range documented {
		if got := st.String(); got != want {
			t.Errorf("semantic type %d reads as %q, want %q", int(st), got, want)
		}
	}

	// The set is closed by construction, so every value it holds is documented.
	for st := SemanticUnresolved; st <= SemanticVariable; st++ {
		if _, ok := documented[st]; !ok {
			t.Errorf("semantic type %d is a value the documentation does not state", int(st))
		}
	}

	// A value past the last documented one is not a decided case, so it reads as the
	// reserved value rather than as a name a consumer could mistake for a decided one.
	if got := (SemanticVariable + 1).String(); got != "unresolved" {
		t.Errorf("a value outside the set reads as %q, want the reserved value", got)
	}
}

// TestSemanticSetReportsUnresolvedForAnUnclassifiedObject checks the reserved value on the
// path the test ground cannot show, because a compiling entity is always classified by a
// rule already decided.
func TestSemanticSetReportsUnresolvedForAnUnclassifiedObject(t *testing.T) {
	pkg := types.NewPackage("github.com/user/project1/pkg1", "pkg1")
	label := types.NewLabel(token.NoPos, pkg, "loop")

	if kindOf(label) != KindUnknown {
		t.Fatalf("kind of a label = %v, want %v", kindOf(label), KindUnknown)
	}
	if got := semanticTypeOf(label, kindOf(label), ""); got != SemanticUnresolved {
		t.Errorf("semantic type = %v, want %v", got, SemanticUnresolved)
	}
}

func declarationLine(t *testing.T, path, needle string) int {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	for i, line := range strings.Split(string(content), "\n") {
		if line == needle {
			return i + 1
		}
	}
	t.Fatalf("fixture %s has no line %q", path, needle)
	return 0
}
