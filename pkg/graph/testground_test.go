package graph

import (
	"path"
	"testing"
)

// TestTestgroundCoversEveryNodeKind checks that the test ground declares one entity of every
// node kind, so a run against it exercises the whole definition at once.
func TestTestgroundCoversEveryNodeKind(t *testing.T) {
	nodes, excluded := mustBuild(t, loadGround(t))

	kinds := map[Kind][]string{}
	for _, n := range nodes {
		kinds[n.Kind] = append(kinds[n.Kind], n.Name)
	}
	for _, kind := range []Kind{KindType, KindFunc, KindVar, KindConst} {
		if len(kinds[kind]) == 0 {
			t.Errorf("the test ground produces no %v node: %v", kind, nodeNames(nodes))
		}
	}

	for _, want := range []struct{ name, reason string }{
		{"Field1", ReasonPartOfDeclarationSpec},
		{"Read", ReasonPartOfDeclarationSpec},
		{"T", ReasonTypeParameter},
		{"s1", ReasonCreatedInsideDeclaration},
		{"local", ReasonCreatedInsideDeclaration},
		{"anon", ReasonCreatedInsideDeclaration},
		{"_", ReasonBlankIdentifier},
		{"str", ReasonPackageIsNamespace},
		{"Read", ReasonProducedByLookup},
	} {
		if !hasCandidate(excluded, want.name, want.reason) {
			t.Errorf("the test ground does not exercise %s (%s): %v", want.name, want.reason, excluded)
		}
	}
}

// TestTestgroundCoversEverySemanticType checks that the test ground declares one entity of
// every decided semantic type, so a run against it exercises the whole semantic set at once.
func TestTestgroundCoversEverySemanticType(t *testing.T) {
	nodes, _ := mustBuild(t, loadGround(t))

	types := map[SemanticType][]string{}
	for _, n := range nodes {
		types[n.SemanticType] = append(types[n.SemanticType], n.Name)
	}
	for _, st := range []SemanticType{
		SemanticStructure,
		SemanticInterface,
		SemanticFunction,
		SemanticMethod,
		SemanticConstant,
		SemanticVariable,
	} {
		if len(types[st]) == 0 {
			t.Errorf("the test ground produces no %v node: %v", st, nodeNames(nodes))
		}
	}
}

// TestTestgroundProducesNoUnresolvedSemanticType checks that no node in the test ground
// has an unresolved semantic type, because every entity is classified by a rule already
// decided.
func TestTestgroundProducesNoUnresolvedSemanticType(t *testing.T) {
	nodes, _ := mustBuild(t, loadGround(t))
	for _, n := range nodes {
		if n.SemanticType == SemanticUnresolved {
			t.Errorf("node %s has unresolved semantic type", n.ID)
		}
	}
}

// TestTestgroundAliasProducesNoNode checks that a declared type alias produces no node,
// and that the alias name is recorded as an attribute of the resolved node.
func TestTestgroundAliasProducesNoNode(t *testing.T) {
	nodes, excluded := mustBuild(t, loadGround(t))

	if nodeByName(nodes, "Mirror") != nil {
		t.Error("alias Mirror produces a node, want no node")
	}
	if nodeByName(nodes, "NameAlias") != nil {
		t.Error("alias NameAlias produces a node, want no node")
	}

	mirrorNode := nodeByName(nodes, "Struct1")
	if mirrorNode == nil {
		t.Fatal("no node for Struct1")
	}
	alias := attributeByName(mirrorNode, "Mirror")
	if alias == nil || alias.Kind != AttrAlias {
		t.Errorf("alias Mirror is not an attribute of Struct1: %v", mirrorNode.Attributes)
	}

	nameAliasNode := nodeByName(nodes, "Name")
	if nameAliasNode == nil {
		t.Fatal("no node for Name")
	}
	nameAlias := attributeByName(nameAliasNode, "NameAlias")
	if nameAlias == nil || nameAlias.Kind != AttrAlias {
		t.Errorf("alias NameAlias is not an attribute of Name: %v", nameAliasNode.Attributes)
	}

	// An alias of a predeclared identifier names an entity that is not a node, so the build has
	// no node to place an alias attribute on: the name stays an excluded candidate, and a use of
	// it carries the resolved type as a value with no reference.
	if nodeByName(nodes, "Counter") != nil {
		t.Error("alias of a predeclared identifier produces a node, want no node")
	}
	if !hasCandidate(excluded, "Counter", ReasonAliasIsName) {
		t.Errorf("alias Counter is not reported as excluded: %v", excluded)
	}
	owner := nodeByName(nodes, "main")
	if owner == nil {
		t.Fatal("no node for main")
	}
	use := attributeByName(owner, "Counter")
	if use == nil {
		t.Fatalf("the use of Counter is not recorded on the node that owns the declaration: %v", owner.Attributes)
	}
	if use.Kind != AttrUse || use.Value != "int" || use.Ref.Name != "" {
		t.Errorf("use of Counter = %v, want the value int with no reference to a node", use)
	}

	// A use of an alias name references the node the alias resolves to, because an alias is not a
	// node the reference can point at.
	for _, pair := range []struct{ alias, resolved string }{{"Mirror", "Struct1"}, {"NameAlias", "Name"}} {
		resolved := nodeByName(nodes, pair.resolved)
		if resolved == nil {
			t.Fatalf("no node for %s", pair.resolved)
		}
		ref := attributeByName(owner, pair.alias)
		if ref == nil || ref.Kind != AttrUse {
			t.Errorf("the use of %s is not recorded on main: %v", pair.alias, owner.Attributes)
			continue
		}
		if ref.Ref != resolved.ID {
			t.Errorf("the use of %s resolves to %v, want the node %s", pair.alias, ref.Ref, resolved.ID)
		}
	}
}

// TestTestgroundGenericProperties checks that generic types and methods report
// generic as a property, and that entry points report entry_point as a property.
func TestTestgroundGenericProperties(t *testing.T) {
	nodes, _ := mustBuild(t, loadGround(t))

	box := nodeByName(nodes, "Box")
	if box == nil {
		t.Fatal("no node for Box")
	}
	if !box.Properties.Generic {
		t.Error("Box has generic=false, want true")
	}
	if box.SemanticType != SemanticStructure {
		t.Errorf("Box semantic type = %v, want %v", box.SemanticType, SemanticStructure)
	}

	pick := nodeByName(nodes, "Pick")
	if pick == nil {
		t.Fatal("no node for Pick")
	}
	if pick.Receiver != "" {
		t.Fatalf("Pick receiver = %q, want empty for package-level function", pick.Receiver)
	}
	if !pick.Properties.Generic {
		t.Error("Pick has generic=false, want true")
	}
	if pick.SemanticType != SemanticFunction {
		t.Errorf("Pick semantic type = %v, want %v", pick.SemanticType, SemanticFunction)
	}

	mainFn := nodeByName(nodes, "main")
	if mainFn == nil {
		t.Fatal("no node for main")
	}
	if !mainFn.Properties.EntryPoint {
		t.Error("main has entry_point=false, want true")
	}
	if mainFn.SemanticType != SemanticFunction {
		t.Errorf("main semantic type = %v, want %v", mainFn.SemanticType, SemanticFunction)
	}

	initFn := nodeByName(nodes, "init")
	if initFn == nil {
		t.Fatal("no node for init")
	}
	if !initFn.Properties.EntryPoint {
		t.Error("init has entry_point=false, want true")
	}
	if initFn.SemanticType != SemanticFunction {
		t.Errorf("init semantic type = %v, want %v", initFn.SemanticType, SemanticFunction)
	}
}

// TestTestgroundMethodWithTypeParams checks that a method declaring its own type
// parameters reports generic as a property with semantic type method.
func TestTestgroundMethodWithTypeParams(t *testing.T) {
	nodes, _ := mustBuild(t, loadGround(t))

	// The package-level Pick function has type parameters.
	pick := nodeByName(nodes, "Pick")
	if pick == nil {
		t.Fatal("no node for Pick")
	}
	if pick.Receiver != "" {
		t.Fatalf("Pick receiver = %q, want empty for package-level function", pick.Receiver)
	}
	if !pick.Properties.Generic {
		t.Error("Pick has generic=false, want true")
	}
	if pick.SemanticType != SemanticFunction {
		t.Errorf("Pick semantic type = %v, want %v", pick.SemanticType, SemanticFunction)
	}

	// The method PickVal on Box has its own type parameters.
	pickVal := nodeByID(nodes, "github.com~Galdoba~go-graph~testdata~main.Box.PickVal")
	if pickVal == nil {
		t.Fatalf("no node for Box/PickVal, got %v", nodeNames(nodes))
	}
	if !pickVal.Properties.Generic {
		t.Error("Box/PickVal has generic=false, want true")
	}
	if pickVal.SemanticType != SemanticMethod {
		t.Errorf("Box/PickVal semantic type = %v, want %v", pickVal.SemanticType, SemanticMethod)
	}
}

// TestTestgroundMethodOnAlias checks that a method written with an alias name
// as its receiver is attributed to the node the alias resolves to.
func TestTestgroundMethodOnAlias(t *testing.T) {
	nodes, _ := mustBuild(t, loadGround(t))

	// The method DescribeAlias should be owned by Struct1, not Mirror.
	struct1 := nodeByName(nodes, "Struct1")
	if struct1 == nil {
		t.Fatal("no node for Struct1")
	}
	aliasMethod := nodeByID(nodes, "github.com~Galdoba~go-graph~testdata~main.Struct1.DescribeAlias")
	if aliasMethod == nil {
		t.Fatalf("no node for Struct1/DescribeAlias, got %v", nodeNames(nodes))
	}
	if aliasMethod.Receiver != "Struct1" {
		t.Errorf("DescribeAlias receiver = %q, want Struct1", aliasMethod.Receiver)
	}
}

// TestSecondGroundPackageStatesNameAndLastSegment checks the second part of the test ground,
// a package whose declared name is the same as the last segment of its package path, so the
// record states both and the printed form states both.
func TestSecondGroundPackageStatesNameAndLastSegment(t *testing.T) {
	p := loadGroundPackage(t, "./testdata/second")
	nodes, _ := mustBuild(t, p)

	if p.Name != "second" {
		t.Fatalf("declared name = %q, want second", p.Name)
	}
	if last := path.Base(p.PkgPath); last != "second" {
		t.Fatalf("last segment of the package path = %q, want second", last)
	}
	if len(nodes) == 0 {
		t.Fatal("the second ground package produced no node")
	}

	read := nodeByName(nodes, "Read")
	if read == nil {
		t.Fatalf("no node for Read, got %v", nodeNames(nodes))
	}
	if read.ID.Package.Name != "second" {
		t.Errorf("record name = %q, want second", read.ID.Package.Name)
	}

	const wantPackage = "github.com~Galdoba~go-graph~testdata~second~second"
	if got := read.ID.Package.String(); got != wantPackage {
		t.Errorf("package part = %s, want %s", got, wantPackage)
	}
	if got := read.ID.String(); got != wantPackage+".Thing.Read" {
		t.Errorf("canonical form = %s, want %s", got, wantPackage+".Read")
	}
}
