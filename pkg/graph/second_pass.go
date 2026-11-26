package graph

import (
	"fmt"
	"go/ast"
	"go/types"
	"sort"

	"golang.org/x/tools/go/packages"
)

// SecondPass produces the records for the entities that the references of the first pass name in
// packages the build did not read as a target. It reads the names those references already carry
// rather than enumerating a package, so a record is produced only for an entity a reference names,
// and only when the definition of a node covers it.
//
// A record produced here carries what a name and a type state: the identifier, the kind, the
// semantic type, the receiver, the parent, the exported property, the signature and the shape the
// type states. When the load reported the syntax that declares the entity, it also carries the
// region, the declaration and the comment above it. It carries no use attributes, because no
// declaration body of that package was read, and its provenance states that the values were read
// from a name, so a consumer reads an absent value as a value that was never read.
func SecondPass(p *packages.Package) ([]*Node, error) {
	if p == nil {
		return nil, fmt.Errorf("graph: no package to read the references of")
	}
	if p.Types == nil || p.TypesInfo == nil {
		return nil, fmt.Errorf("graph: package %s carries no type information to read the references from", p.ID)
	}

	// The demand set is the entities the references of the first pass name. It is read from the
	// objects, so the same entity named from several declarations is one record.
	demand := make(map[*types.Package][]types.Object)
	seen := make(map[types.Object]bool)
	for _, obj := range p.TypesInfo.Uses {
		if obj == nil || obj.Pkg() == nil || obj.Pkg() == p.Types {
			continue
		}
		referent := resolvedObject(obj)
		if referent == nil || seen[referent] || !definitionNamesNode(referent.Pkg(), referent) {
			continue
		}
		seen[referent] = true
		demand[referent.Pkg()] = append(demand[referent.Pkg()], referent)
	}

	nodes := make([]*Node, 0, len(seen))
	for pkg, objects := range demand {
		imported := p.Imports[pkg.Path()]
		decls, ranges := declarationIndex(imported)
		for _, obj := range objects {
			nodes = append(nodes, foreignRecord(p, imported, obj, decls, ranges))
		}
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID.String() < nodes[j].ID.String() })
	for _, n := range nodes {
		sort.Slice(n.Attributes, func(i, j int) bool { return attributeOrder(n.Attributes[i], n.Attributes[j]) })
	}
	return nodes, nil
}

// foreignRecord builds the record of one entity the build did not read as a target, from the load
// that read the reference naming it. The identifier is built by the rules that build the identifier
// of any other record, so the same entity named by a reference and recorded by a build of its own
// package is one node rather than two entries that have to be matched by name.
func foreignRecord(p *packages.Package, imported *packages.Package, obj types.Object,
	decls map[types.Object]ast.Decl, ranges map[types.Object]LineRange) *Node {
	pkg := referentPackage(p, obj.Pkg())
	id := NewID(pkg, obj)
	kind := Classify(obj.Pkg(), obj).Kind

	n := &Node{
		ID:           id,
		Kind:         kind,
		Name:         obj.Name(),
		Package:      pkg,
		Parent:       parentIn(pkg, obj),
		Receiver:     id.Receiver,
		SemanticType: semanticTypeOf(obj, kind, id.Receiver),
		Properties:   propertiesOf(pkg, obj),
		Attributes:   shapeOf(p, obj),
		Provenance:   ProvenanceFromName,
	}

	// The region, the declaration and the comment are stated only when the load reported the syntax
	// that declares the entity. A record that states none states no region, which is read as a value
	// that could not be stated rather than as a declaration with no region.
	if decl := decls[obj]; decl != nil {
		fset := p.Fset
		if imported != nil && imported.Fset != nil {
			fset = imported.Fset
		}
		n.Decl = decl
		n.Pos = position(fset, obj.Pos())
		n.Range = ranges[obj]
	}
	return n
}

// shapeOf records the shape a type states, which is all a record produced from a name can carry
// about an entity: the fields and the embedding it declares, the method specs an interface states,
// and the methods the package declares on that type, which leaves out a method reachable only
// through an embedding. The reference names the entity the type denotes, so it is built by the same
// rules as any other reference, and an entity the definition does not cover contributes no
// reference, leaving the value to carry it.
func shapeOf(p *packages.Package, obj types.Object) []Attribute {
	attrs := []Attribute{}
	tn, ok := obj.(*types.TypeName)
	if !ok {
		return attrs
	}
	named, ok := types.Unalias(tn.Type()).(*types.Named)
	if !ok {
		return attrs
	}

	switch t := named.Underlying().(type) {
	case *types.Struct:
		for i := 0; i < t.NumFields(); i++ {
			field := t.Field(i)
			kind := AttrField
			if field.Embedded() {
				kind = AttrEmbeds
			}
			attrs = append(attrs, Attribute{
				Kind:  kind,
				Name:  field.Name(),
				Value: types.TypeString(types.Unalias(field.Type()), pathQualifier),
				Ref:   shapeRef(p, field.Type()),
			})
		}
	case *types.Interface:
		for i := 0; i < t.NumExplicitMethods(); i++ {
			spec := t.ExplicitMethod(i)
			attrs = append(attrs, Attribute{
				Kind:  AttrMethodSpec,
				Name:  spec.Name(),
				Value: types.TypeString(spec.Signature(), pathQualifier),
			})
		}
	}

	for fn := range named.Methods() {
		plain := types.NewSignatureType(nil, nil, nil, fn.Signature().Params(), fn.Signature().Results(), fn.Signature().Variadic())
		attrs = append(attrs, Attribute{
			Kind:  AttrMethod,
			Name:  fn.Name(),
			Value: types.TypeString(plain, pathQualifier),
			Ref:   NewID(referentPackage(p, obj.Pkg()), fn),
		})
	}
	return attrs
}

// shapeRef reports the node a field type denotes. Only an entity the definition covers names a node,
// so a type parameter, a collection type and a predeclared identifier contribute no reference.
func shapeRef(p *packages.Package, expr types.Type) ID {
	named, ok := types.Unalias(expr).(*types.Named)
	if !ok {
		return ID{}
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil || !definitionNamesNode(obj.Pkg(), obj) {
		return ID{}
	}
	return NewID(referentPackage(p, obj.Pkg()), obj)
}

// resolvedObject reports the entity a name denotes, which is what a reference points at: an alias is
// a name for an existing entity rather than an entity of its own, so an alias name names the entity
// the package declares.
func resolvedObject(obj types.Object) types.Object {
	if tn, ok := obj.(*types.TypeName); ok && tn.IsAlias() {
		if named, ok := types.Unalias(tn.Type()).(*types.Named); ok {
			return named.Obj()
		}
		return nil
	}
	return obj
}

// referentPackage reports the package part of the identifier a reference builds, which is the module
// the load reported for the package that declares the entity and carries no version when that load
// reported no module.
func referentPackage(p *packages.Package, pkg *types.Package) PackageID {
	if pkg == p.Types {
		return packageID(p)
	}
	if imported := p.Imports[pkg.Path()]; imported != nil {
		return packageID(imported)
	}
	return PackageID{PackagePath: pkg.Path(), Name: pkg.Name()}
}

func pathQualifier(pkg *types.Package) string { return pkg.Path() }
