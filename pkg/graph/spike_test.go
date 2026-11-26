package graph

import (
	"go/types"
	"testing"
)

// TestSpikeWhatANonTargetPackageCarries records what a load that does not read a package as a
// target preserves for a declared entity, which is what a record produced from a reference can
// carry. The answers are written into design.md.
// TestSpikeWhatANonTargetPackageCarries records what a load that does not read a package as a
// target preserves for a declared entity, which is what a record produced from a reference can
// carry. The answers are written into design.md.
func TestSpikeWhatANonTargetPackageCarries(t *testing.T) {
	p := loadFixture(t, "use")

	// The load reports the syntax and the type information of an imported package as well, so a
	// record read from a name can state the region and the comment of the declaration it names.
	shape := p.Imports[shapeFixturePath]
	if shape == nil {
		t.Fatal("the fixture imports no package of the repository")
	}
	if len(shape.Syntax) != 1 || shape.TypesInfo == nil {
		t.Errorf("the load reports syntax=%d and type info=%v for a package of the repository, want one file and type info",
			len(shape.Syntax), shape.TypesInfo != nil)
	}
	if len(shape.CompiledGoFiles) != 0 {
		t.Errorf("an imported package reports %d compiled files, which cannot tell a foreign record from a target record",
			len(shape.CompiledGoFiles))
	}
	standard := p.Imports["strings"]
	if standard == nil || len(standard.Syntax) == 0 || standard.TypesInfo == nil {
		t.Errorf("the load reports syntax=%d and type info=%v for a standard library package, want both",
			len(standard.Syntax), standard.TypesInfo != nil)
	}

	// A package-level object reports a usable position in the file of the package that declares
	// it, which is the region a record read from a name states when the syntax is reported.
	shapeObject := shape.Types.Scope().Lookup("Shape")
	if shapeObject == nil {
		t.Fatal("the imported package declares no Shape")
	}
	region := p.Fset.Position(shapeObject.Pos())
	if !region.IsValid() || region.Filename != p.Fset.Position(shape.Syntax[0].Pos()).Filename {
		t.Errorf("a package-level entity of an imported package reports %v, want a position in the file of that package", region)
	}
	if kindOf(shapeObject) != KindType || !shapeObject.Exported() {
		t.Errorf("the entity reports kind %v and exported=%v", kindOf(shapeObject), shapeObject.Exported())
	}

	// A generic named type reports its type parameters.
	box, _ := shape.Types.Scope().Lookup("Box").(*types.TypeName)
	if box == nil {
		t.Fatal("the imported package declares no generic type")
	}
	namedBox, _ := types.Unalias(box.Type()).(*types.Named)
	if namedBox == nil || namedBox.TypeParams() == nil || namedBox.TypeParams().Len() != 1 {
		t.Errorf("the generic type reports no type parameters, want one")
	}

	// An exported alias of an unexported named type reports that it is an alias and resolves to
	// the entity the package declares.
	shown, _ := shape.Types.Scope().Lookup("Shown").(*types.TypeName)
	if shown == nil || !shown.IsAlias() {
		t.Fatalf("the exported name that is an alias reports it is not an alias: %v", shown)
	}
	target, _ := types.Unalias(shown.Type()).(*types.Named)
	if target == nil || target.Obj().Name() != "hidden" || target.Obj().Exported() {
		t.Errorf("the alias resolves to %v, want the unexported entity the package declares", target)
	}

	// A method object identifies the type that declares it through its receiver, and a method is
	// not in the package scope, so node-ness for a foreign method follows the definition.
	namedShape, _ := types.Unalias(shapeObject.Type()).(*types.Named)
	if namedShape == nil || namedShape.NumMethods() == 0 {
		t.Fatal("the imported package declares no method on the named type")
	}
	var method *types.Func
	for fn := range namedShape.Methods() {
		method = fn
	}
	if method == nil {
		t.Fatal("no method object is carried for an imported package")
	}
	namedReceiver, ok := baseNamed(method.Signature().Recv().Type())
	if !ok || namedReceiver.Obj() != shapeObject {
		t.Errorf("the method object names no receiver, want the type that declares it")
	}
	if methodPos := p.Fset.Position(method.Pos()); !methodPos.IsValid() {
		t.Errorf("the method object reports no usable position")
	}
	if shape.Types.Scope().Lookup(method.Name()) != nil {
		t.Errorf("the method appears in the package scope, so the definition is not what decides its node-ness")
	}

	// The load reports a module for an imported package only when it resolved one, so the version
	// belongs to a location only for a record that was named from a module.
	if shape.Module == nil || shape.Module.Path != "github.com/Galdoba/go-graph" || shape.Module.Version != "" {
		t.Errorf("the load reports module %v for a package of the repository, want its path with no version", shape.Module)
	}
	if standard.Module != nil {
		t.Errorf("the load reports a module for a standard library package, want none")
	}
}
