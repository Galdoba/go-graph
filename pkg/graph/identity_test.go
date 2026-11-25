package graph

import (
	"strings"
	"testing"
)

// TestIdentifierQualifiesModulesAndVersions checks that the identifier is built from the
// components of a declaration, so two modules that contain the same package path and the
// same entity name produce two different identifiers. The name alone cannot tell them
// apart, which is why the identifier is not derived from printed output.
func TestIdentifierQualifiesModulesAndVersions(t *testing.T) {
	source := map[string]string{"a.go": "package pkg1\n\nfunc Read() error { return nil }\n"}

	l := newSourceLoader(t)
	first := l.load("github.com/user/project1", "v1.0.0", "pkg1", source)
	second := l.load("github.com/user/project2", "v2.1.0", "pkg1", source)

	firstNodes, _ := mustBuild(t, first)
	secondNodes, _ := mustBuild(t, second)

	if len(firstNodes) != 1 || len(secondNodes) != 1 {
		t.Fatalf("expected one node per module, got %v and %v", nodeNames(firstNodes), nodeNames(secondNodes))
	}
	a, b := firstNodes[0], secondNodes[0]

	if a.ID == b.ID {
		t.Fatalf("two modules with the same package path produced the same identifier %v", a.ID)
	}
	if a.ID.Package.PackagePath != "pkg1" || a.Name != "Read" {
		t.Fatalf("identifier %v is not qualified by the package path and name", a.ID)
	}
	if got, want := a.ID.String(), "github.com~user~project1@v1.0.0~pkg1~pkg1.Read"; got != want {
		t.Errorf("identifier = %q, want %q", got, want)
	}
	if got, want := b.ID.String(), "github.com~user~project2@v2.1.0~pkg1~pkg1.Read"; got != want {
		t.Errorf("identifier = %q, want %q", got, want)
	}

	// The same package path at two module versions is two different sets of nodes.
	versioned := newSourceLoader(t)
	older := versioned.load("github.com/user/project1", "v1.0.0", "pkg1", source)
	newer := versioned.load("github.com/user/project1", "v2.0.0", "pkg1", source)
	olderNodes, _ := mustBuild(t, older)
	newerNodes, _ := mustBuild(t, newer)
	if olderNodes[0].ID == newerNodes[0].ID {
		t.Errorf("the same package at two versions produced the same identifier %v", olderNodes[0].ID)
	}

	// The name is what collides, so an identifier built from the name alone cannot
	// separate the two entities.
	if first.Types.Scope().Lookup("Read").Id() != second.Types.Scope().Lookup("Read").Id() {
		t.Error("the qualified names differ, so the collision this change guards against is not the one being demonstrated")
	}
}

// TestSameNamedMethodsStayDistinct checks that three methods with the same name on
// different receivers produce three nodes with three distinct identifiers.
func TestSameNamedMethodsStayDistinct(t *testing.T) {
	nodes, _ := mustBuild(t, loadFixture(t, "methods"))

	methods := make([]*Node, 0, 3)
	for _, n := range nodes {
		if n.Kind == KindFunc {
			methods = append(methods, n)
		}
	}
	if len(methods) != 3 {
		t.Fatalf("function nodes = %v, want three methods", nodeNames(nodes))
	}

	seen := make(map[string]bool)
	for _, n := range methods {
		key := n.ID.String()
		if seen[key] {
			t.Fatalf("two methods share the identifier %s: %v", key, nodeNames(nodes))
		}
		seen[key] = true
	}
	for _, want := range []string{
		"Struct1/Read",
		"Struct2/Read",
		"Struct2/ReadLoud",
	} {
		found := false
		for _, n := range methods {
			if n.ID.Receiver+"/"+n.ID.Name == want {
				found = true
			}
		}
		if !found {
			t.Errorf("no identifier ends with %s: %v", want, seen)
		}
	}
}

// TestCanonicalFormIsUsableAsANoteName checks that the printed identifier avoids the
// characters a note name cannot contain, and that the package part and the node part stay
// apart. The package path is written with tildes and the receiver and the name with dots,
// so a package path that ends in a name cannot be read as a receiver.
func TestCanonicalFormIsUsableAsANoteName(t *testing.T) {
	nodes, _ := mustBuild(t, loadGround(t))

	forbidden := "#[]\\|^/"
	for _, n := range nodes {
		if strings.ContainsAny(n.ID.String(), forbidden) {
			t.Errorf("identifier %s cannot be used as a note name", n.ID)
		}
	}

	// The test ground is a directory named testdata that holds package main, so the name the
	// package declares is a part of its own rather than the last segment of the path.
	if got, want := nodes[0].ID.Package.String(), "github.com~Galdoba~go-graph~testdata~main"; got != want {
		t.Errorf("package part = %q, want %q", got, want)
	}

	// The module is named once, not twice, because the package path already names it.
	escapedModule := strings.ReplaceAll(nodes[0].ID.Package.ModulePath, "/", "~")
	for _, n := range nodes {
		if strings.Count(n.ID.Package.String(), escapedModule) != 1 {
			t.Errorf("package part %s names the module %s more than once", n.ID.Package, escapedModule)
		}
	}

	// The package namespace carries no name of its own, so its form is the package part
	// alone rather than a form that ends in a dot.
	namespace := ID{Package: nodes[0].ID.Package}
	if strings.HasSuffix(namespace.String(), ".") {
		t.Errorf("the package namespace reads as %s, which is not a note name", namespace)
	}

	l := newSourceLoader(t)
	methodNodes, _ := mustBuild(t, l.load("github.com/user/project1", "v1.0.0", "pkg1", map[string]string{
		"a.go": "package pkg1\n\ntype Struct1 struct{}\n\nfunc (s *Struct1) Read() error { return nil }\n\nfunc Read() error { return nil }\n",
	}))

	method := nodeByID(methodNodes, "github.com~user~project1@v1.0.0~pkg1~pkg1.Struct1.Read")
	if method == nil {
		t.Fatalf("the receiver is not separated from the package part: %v", nodeNames(methodNodes))
	}

	plain := nodeByID(methodNodes, "github.com~user~project1@v1.0.0~pkg1~pkg1.Read")
	if plain == nil || plain.Receiver != "" {
		t.Fatalf("a node with no receiver carries a receiver segment: %v", nodeNames(methodNodes))
	}

	other, _ := mustBuild(t, l.load("github.com/user/project1", "v1.0.0", "pkg1/Struct1", map[string]string{
		"a.go": "package pkg1\n\nfunc Read() error { return nil }\n",
	}))
	byPackagePath := nodeByID(other, "github.com~user~project1@v1.0.0~pkg1~Struct1~pkg1.Read")
	if byPackagePath == nil {
		t.Fatalf("the second package produced no node, got %v", nodeNames(other))
	}
	if byPackagePath.ID == method.ID {
		t.Errorf("the package path reads as the receiver part of %s", method.ID)
	}
}

// TestModuleIsReadAsAPathSegment checks that the package path is read as a segment of the
// module rather than as a plain prefix, so a module whose name is a prefix of another module
// name keeps its own part in the printed form.
func TestModuleIsReadAsAPathSegment(t *testing.T) {
	segment := PackageID{
		ModulePath:    "github.com/user/project1",
		ModuleVersion: "v1.0.0",
		PackagePath:   "github.com/user/project1/pkg2",
		Name:          "pkg2",
	}
	if got, want := segment.String(), "github.com~user~project1@v1.0.0~pkg2~pkg2"; got != want {
		t.Errorf("package part = %q, want %q", got, want)
	}

	prefix := PackageID{
		ModulePath:    "github.com/user/project1",
		ModuleVersion: "v1.0.0",
		PackagePath:   "github.com/user/project12/pkg2",
		Name:          "pkg2",
	}
	if got, want := prefix.String(), "github.com~user~project1@v1.0.0~github.com~user~project12~pkg2~pkg2"; got != want {
		t.Errorf("package part of a package under a different module = %q, want %q", got, want)
	}
}

// TestPromotedMethodIsExcluded checks the single declaration site invariant: a method
// reachable only through embedding is produced by a lookup rather than by a declaration,
// so it is reported as excluded and the outer type keeps the embedding as an attribute.
func TestPromotedMethodIsExcluded(t *testing.T) {
	nodes, excluded := mustBuild(t, loadFixture(t, "embed"))

	if got := nodeNames(nodes); len(got) != 4 {
		t.Fatalf("nodes = %v, want the four package-level declarations", got)
	}
	if !hasCandidate(excluded, "Read", ReasonProducedByLookup) {
		t.Fatalf("the promoted method is not reported as excluded: %v", excluded)
	}

	outer := mustNodeByName(t, nodes, "Outer")
	base := mustNodeByName(t, nodes, "Base")
	embeds := attributeByName(outer, "Base")
	if embeds == nil || embeds.Kind != AttrEmbeds {
		t.Fatalf("the outer type does not record the embedding: %v", outer.Attributes)
	}
	if embeds.Ref != base.ID {
		t.Errorf("embedding points at %v, want the embedded type node %v", embeds.Ref, base.ID)
	}

	declared := nodeByName(nodes, "Read")
	if declared == nil || declared.Receiver != "Base" {
		t.Fatalf("the declared method is not a node on Base: %v", nodeNames(nodes))
	}

	// The call site uses the method through the embedding, so it resolves to the node of
	// the type that declares it rather than to the type that is written in the code.
	use := mustNodeByName(t, nodes, "Use")
	call := attributeByName(use, "Read")
	if call == nil || call.Kind != AttrUse {
		t.Fatalf("the call through the embedding is not recorded on the use-site: %v", use.Attributes)
	}
	if call.Ref != declared.ID {
		t.Errorf("the call resolves to %v, want the method node declared on Base %v", call.Ref, declared.ID)
	}
}

// TestPositionIsProvenanceNotIdentity checks that an edit above a declaration moves the
// recorded position without changing the identifier, so matching two builds uses identity.
func TestPositionIsProvenanceNotIdentity(t *testing.T) {
	source := map[string]string{"a.go": "package pkg1\n\nfunc Read() error { return nil }\n"}
	edited := map[string]string{"a.go": "package pkg1\n\n// A line added above the declaration.\n\nfunc Read() error { return nil }\n"}

	before := newSourceLoader(t).load("github.com/user/project1", "v1.0.0", "pkg1", source)
	after := newSourceLoader(t).load("github.com/user/project1", "v1.0.0", "pkg1", edited)

	beforeNodes, _ := mustBuild(t, before)
	afterNodes, _ := mustBuild(t, after)

	a, b := beforeNodes[0], afterNodes[0]
	if a.ID != b.ID {
		t.Fatalf("identifier changed with the position: %v and %v", a.ID, b.ID)
	}
	if a.Pos.Line == b.Pos.Line {
		t.Fatalf("position did not move, so the test does not show provenance: %v and %v", a.Pos, b.Pos)
	}
}

// TestRebuildChangesSemanticTypeWithoutChangingIdentity checks that a rebuild recomputes the
// semantic type from the declaration while the identifier that names the note stays the same,
// so comparing two builds has to look at the semantic type as well as at identity.
func TestRebuildChangesSemanticTypeWithoutChangingIdentity(t *testing.T) {
	source := map[string]string{"a.go": "package pkg1\n\ntype Shape string\n"}
	rewritten := map[string]string{"a.go": "package pkg1\n\ntype Shape interface{}\n"}

	before := newSourceLoader(t).load("github.com/user/project1", "v1.0.0", "pkg1", source)
	after := newSourceLoader(t).load("github.com/user/project1", "v1.0.0", "pkg1", rewritten)

	beforeNodes, _ := mustBuild(t, before)
	afterNodes, _ := mustBuild(t, after)

	a, b := beforeNodes[0], afterNodes[0]
	if a.ID != b.ID {
		t.Fatalf("identifier changed with the semantic type: %v and %v", a.ID, b.ID)
	}
	if a.SemanticType != SemanticStructure {
		t.Fatalf("semantic type before the rebuild = %v, want %v", a.SemanticType, SemanticStructure)
	}
	if b.SemanticType != SemanticInterface {
		t.Errorf("semantic type after the rebuild = %v, want %v", b.SemanticType, SemanticInterface)
	}
}

// TestNodeResolvesBackToItsDeclaration checks that a recorded node can be traced back to
// its declaration from the record alone, and that nodes created by one group declaration
// keep the range of their own declaration spec.
func TestNodeResolvesBackToItsDeclaration(t *testing.T) {
	p := loadFixture(t, "ranges")
	nodes, _ := mustBuild(t, p)

	if len(nodes) != 3 {
		t.Fatalf("nodes = %v, want one per declaration", nodeNames(nodes))
	}

	// The second load is a different parse of the same source, so resolving the record
	// against it shows that the coordinates alone are enough to find the declaration.
	fresh := loadFixture(t, "ranges")
	for _, n := range nodes {
		if n.Decl == nil {
			t.Errorf("node %s has no reference to its declaration", n.ID)
			continue
		}
		if err := Verify(fresh, n); err != nil {
			t.Errorf("%s: %v", n.ID, err)
		}
	}

	first := nodeByName(nodes, "first")
	second := nodeByName(nodes, "second")
	if first == nil || second == nil {
		t.Fatalf("the group declaration did not produce both nodes, got %v", nodeNames(nodes))
	}
	if first.Range == second.Range {
		t.Errorf("two variables inside one var block share the range %v", first.Range)
	}
	counter := mustNodeByName(t, nodes, "Counter")
	if counter.Range.StartLine >= counter.Range.EndLine {
		t.Errorf("range of a type declaration covers only its name: %v", counter.Range)
	}

	// The same coordinates read as a different name are rejected, which is the invariant
	// seen from the other side: the record may point at one declaration only.
	shifted := *counter
	shifted.Name = "Value"
	if err := Verify(fresh, &shifted); err == nil {
		t.Error("a record pointing at a field was accepted as a declaration site")
	}
}
