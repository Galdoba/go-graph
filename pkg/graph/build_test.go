package graph

import (
	"testing"
)

// TestPackageScopeProducesOneNodePerDeclaration checks that enumerating the package
// scope produces exactly one node per package-level declaration, and that entities
// created inside a declaration are reported as excluded rather than recorded.
func TestPackageScopeProducesOneNodePerDeclaration(t *testing.T) {
	p := loadFixture(t, "basic")
	nodes, excluded := mustBuild(t, p)

	if len(nodes) != 4 {
		t.Fatalf("nodes = %v, want one node per declaration", nodeNames(nodes))
	}

	want := map[string]Kind{
		"Struct1":       KindType,
		"Read":          KindFunc,
		"GlobalCounter": KindVar,
		"MaxRetries":    KindConst,
	}
	for name, kind := range want {
		n := nodeByName(nodes, name)
		if n == nil {
			t.Fatalf("no node for declaration %s, got %v", name, nodeNames(nodes))
		}
		if n.Kind != kind {
			t.Errorf("%s: kind = %v, want %v", name, n.Kind, kind)
		}
		if n.Package != packageID(p) {
			t.Errorf("%s: package = %v, want %v", name, n.Package, packageID(p))
		}
	}

	for _, name := range []string{"Field", "local", "_"} {
		if nodeByName(nodes, name) != nil {
			t.Errorf("entity created inside a declaration was recorded as a node: %s", name)
		}
	}
	for _, c := range []struct{ name, reason string }{
		{"Field", ReasonPartOfDeclarationSpec},
		{"local", ReasonCreatedInsideDeclaration},
		{"_", ReasonBlankIdentifier},
	} {
		if !hasCandidate(excluded, c.name, c.reason) {
			t.Errorf("candidate %s is not reported as excluded with reason %q: %v", c.name, c.reason, excluded)
		}
	}
}

// TestMethodNodeCarriesReceiverAndParent checks that a method is a node whose receiver
// names the type it is declared on, and whose parent is the node of that type.
func TestMethodNodeCarriesReceiverAndParent(t *testing.T) {
	p := loadFixture(t, "methods")
	nodes, _ := mustBuild(t, p)

	owner := nodeByName(nodes, "Struct1")
	if owner == nil {
		t.Fatalf("no node for Struct1, got %v", nodeNames(nodes))
	}
	method := nodeByName(nodes, "Read")
	if method == nil {
		t.Fatalf("no node for the method, got %v", nodeNames(nodes))
	}
	if method.Receiver != "Struct1" {
		t.Errorf("receiver = %q, want Struct1", method.Receiver)
	}
	if method.Parent != owner.ID {
		t.Errorf("parent = %v, want the type node %v", method.Parent, owner.ID)
	}

	other := nodeByName(nodes, "ReadLoud")
	if other == nil {
		t.Fatalf("no node for ReadLoud, got %v", nodeNames(nodes))
	}
	if other.Receiver != "Struct2" {
		t.Errorf("receiver = %q, want Struct2", other.Receiver)
	}
}

// TestFieldsAndMethodSpecsAreAttributes checks that entities which are not declarations
// survive as attributes of the node that owns them and never as nodes.
func TestFieldsAndMethodSpecsAreAttributes(t *testing.T) {
	p := loadFixture(t, "attrs")
	nodes, _ := mustBuild(t, p)

	if got := nodeNames(nodes); len(got) != 4 {
		t.Fatalf("nodes = %v, want only the package-level declarations", got)
	}

	counter := nodeByName(nodes, "Counter")
	if counter == nil {
		t.Fatalf("no node for Counter, got %v", nodeNames(nodes))
	}
	value := attributeByName(counter, "Value")
	if value == nil || value.Kind != AttrField {
		t.Fatalf("field Value is not an attribute of Counter: %v", counter.Attributes)
	}
	if value.Ref.Name != "" || value.Value != "int" {
		t.Errorf("field Value = %v, want the predeclared type int as an attribute value", value)
	}
	named := attributeByName(counter, "Named")
	if named == nil || named.Kind != AttrField {
		t.Fatalf("field Named is not an attribute of Counter: %v", counter.Attributes)
	}
	target := nodeByName(nodes, "Struct1")
	if named.Ref != target.ID {
		t.Errorf("field Named points at %v, want the node it refers to %v", named.Ref, target.ID)
	}

	reader := nodeByName(nodes, "Reader")
	if reader == nil {
		t.Fatalf("no node for Reader, got %v", nodeNames(nodes))
	}
	spec := attributeByName(reader, "Read")
	if spec == nil || spec.Kind != AttrMethodSpec {
		t.Fatalf("interface method spec is not an attribute of Reader: %v", reader.Attributes)
	}
	if nodeByName(nodes, "Read") != nil {
		t.Error("the interface method spec was recorded as a node")
	}
}

// TestDeclaredMethodsAreAttributesOfTheOwner checks that a method declared on a package-level
// type is recorded as an attribute of the node that owns it, in the same form a method spec is
// recorded, so that a relation can read the method set from the record of the type alone.
func TestDeclaredMethodsAreAttributesOfTheOwner(t *testing.T) {
	nodes, _ := mustBuild(t, loadGround(t))

	base := nodeByName(nodes, "Base")
	if base == nil {
		t.Fatalf("no node for Base, got %v", nodeNames(nodes))
	}
	var method *Node
	for _, n := range nodes {
		if n.Receiver == "Base" && n.Name == "Read" {
			method = n
			break
		}
	}
	if method == nil {
		t.Fatalf("no method node for Base.Read, got %v", nodeNames(nodes))
	}

	declared := attributeByName(base, "Read")
	if declared == nil || declared.Kind != AttrMethod {
		t.Fatalf("method Read is not an attribute of Base: %v", base.Attributes)
	}
	if declared.Kind.String() != "method" {
		t.Errorf("the kind reads back as %q, want method", declared.Kind.String())
	}
	if declared.Value != "func() error" {
		t.Errorf("method Read value = %q, want the signature without the receiver", declared.Value)
	}
	if declared.Ref != method.ID {
		t.Errorf("method Read points at %v, want the method node %v", declared.Ref, method.ID)
	}

	reader := nodeByName(nodes, "Reader")
	if reader == nil {
		t.Fatalf("no node for Reader, got %v", nodeNames(nodes))
	}
	spec := attributeByName(reader, "Read")
	if spec == nil || spec.Kind != AttrMethodSpec {
		t.Fatalf("Read is not a method spec attribute of Reader: %v", reader.Attributes)
	}
	// A relation can compare the two sides only because both are recorded in the same form, with the
	// signature read without the receiver.
	if declared.Value != spec.Value {
		t.Errorf("the method value %q is not the form a method spec records %q", declared.Value, spec.Value)
	}
}

// TestAliasesAreAttributesAndImportsCreateNoNodes checks that an alias is recorded as an
// attribute of the node it resolves to, and that neither the imported package nor a
// predeclared identifier becomes a node.
func TestAliasesAreAttributesAndImportsCreateNoNodes(t *testing.T) {
	p := loadFixture(t, "alias")
	nodes, excluded := mustBuild(t, p)

	if got := nodeNames(nodes); len(got) != 1 || got[0] != "Use" {
		t.Fatalf("nodes = %v, want only the package-level declaration", got)
	}
	for _, name := range []string{"fx", "fmt", "foreign", "error"} {
		if nodeByName(nodes, name) != nil {
			t.Errorf("no node may be created for %s", name)
		}
	}

	use := nodes[0]
	alias := attributeByName(use, "fx")
	if alias == nil || alias.Kind != AttrAlias {
		t.Fatalf("alias fx is not an attribute of the use-site: %v", use.Attributes)
	}

	foreignNodes, _ := mustBuild(t, loadFixture(t, "foreign"))
	target := nodeByName(foreignNodes, "Export")
	if target == nil {
		t.Fatalf("no node for the foreign declaration, got %v", nodeNames(foreignNodes))
	}
	if alias.Ref != target.ID {
		t.Errorf("alias resolves to %v, want the foreign node %v", alias.Ref, target.ID)
	}
	if alias.Value != p.Imports[target.Package.PackagePath].PkgPath {
		t.Errorf("alias value = %q, want the package it resolves to", alias.Value)
	}

	predeclared := attributeByName(use, "error")
	if predeclared == nil || predeclared.Kind != AttrUse {
		t.Fatalf("the predeclared identifier is not recorded as an attribute value: %v", use.Attributes)
	}
	if predeclared.Value != "error" || predeclared.Ref.Name != "" {
		t.Errorf("predeclared reference = %v, want the value error with no referent node", predeclared)
	}

	if !hasCandidate(excluded, "fx", ReasonPackageIsNamespace) {
		t.Errorf("import alias is not reported as excluded: %v", excluded)
	}
}

func TestPackageDeclarationShadowsPredeclaredIdentifier(t *testing.T) {
	source := map[string]string{"a.go": "package pkg1\n\nfunc min(a, b int) int {\n\tif a < b {\n\t\treturn a\n\t}\n\treturn b\n}\n\nvar chosen = min(1, 2)\n"}
	p := newSourceLoader(t).load("github.com/user/project1", "v1.0.0", "pkg1", source)
	nodes, excluded := mustBuild(t, p)

	declared := nodeByName(nodes, "min")
	if declared == nil {
		t.Fatalf("the package-level declaration named after a predeclared identifier is not a node: %v", nodeNames(nodes))
	}
	for _, name := range []string{"a", "b"} {
		if nodeByName(nodes, name) != nil {
			t.Errorf("a parameter was recorded as a node: %s", name)
		}
		if !hasCandidate(excluded, name, ReasonCreatedInsideDeclaration) {
			t.Errorf("parameter %s is not reported as excluded: %v", name, excluded)
		}
	}

	use := nodeByName(nodes, "chosen")
	if use == nil {
		t.Fatalf("no node for the variable, got %v", nodeNames(nodes))
	}
	ref := attributeByName(use, "min")
	if ref == nil || ref.Kind != AttrUse {
		t.Fatalf("the reference to min is not recorded on the use-site: %v", use.Attributes)
	}
	if ref.Ref != declared.ID {
		t.Errorf("the reference resolves to %v, want the package node %v", ref.Ref, declared.ID)
	}
}

// TestEntryPointNameOutsidePackageMainIsAnOrdinaryFunction checks that the entry point
// property follows the package rather than the name, so a function named main in a package
// that is not main is an ordinary function with no entry point.
func TestEntryPointNameOutsidePackageMainIsAnOrdinaryFunction(t *testing.T) {
	l := newSourceLoader(t)

	ordinary, _ := mustBuild(t, l.load("github.com/user/project1", "v1.0.0", "pkg1", map[string]string{
		"a.go": "package pkg1\n\nfunc main() {\n}\n",
	}))

	fn := nodeByName(ordinary, "main")
	if fn == nil {
		t.Fatalf("no node for main, got %v", nodeNames(ordinary))
	}
	if fn.Properties.EntryPoint {
		t.Error("main in a package that is not main reports entry_point=true, want false")
	}
	if fn.SemanticType != SemanticFunction {
		t.Errorf("main in a package that is not main has semantic type %v, want %v", fn.SemanticType, SemanticFunction)
	}

	// The same name inside package main is the entry point the language assigns.
	entryPackage, _ := mustBuild(t, l.load("github.com/user/project1", "v1.0.0", "main", map[string]string{
		"a.go": "package main\n\nfunc main() {\n}\n",
	}))
	entry := nodeByName(entryPackage, "main")
	if entry == nil {
		t.Fatalf("no node for main in package main, got %v", nodeNames(entryPackage))
	}
	if !entry.Properties.EntryPoint {
		t.Error("main in package main reports entry_point=false, want true")
	}
}

func attributeByName(node *Node, name string) *Attribute {
	for _, a := range node.Attributes {
		if a.Name == name {
			return &a
		}
	}
	return nil
}
