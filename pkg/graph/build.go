package graph

import (
	"fmt"
	"go/ast"
	"go/types"
	"sort"

	"golang.org/x/tools/go/packages"
)

// LoadMode is what a package must be loaded with for the enumeration to be complete:
// type information for identity, syntax for the declaration record, and module
// information for the identifier.
const LoadMode = packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps |
	packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedModule

// Build produces the nodes of one package. The enumeration is the definition: the
// package scope gives the package-level declarations, the methods declared on
// package-level named types give the method nodes, and every other entity is reported as
// an excluded candidate instead of recorded, so there is no filter step to keep in sync
// with the definition.
//
// Nodes are ordered by identifier and candidates by name, so the same source produces
// the same result twice. Entities that are not nodes survive as attributes of the node
// that owns them.
func Build(p *packages.Package) ([]*Node, []Candidate, error) {
	if p == nil {
		return nil, nil, fmt.Errorf("graph: no package to build")
	}
	if p.Types == nil || p.TypesInfo == nil || p.Syntax == nil {
		return nil, nil, fmt.Errorf("graph: package %s is incomplete, load it with graph.LoadMode", p.ID)
	}
	b := &builder{
		pkg:      p,
		id:       packageID(p),
		nodes:    make(map[ID]*Node),
		seen:     make(map[types.Object]bool),
		reported: make(map[string]bool),
		decls:    make(map[types.Object]ast.Decl),
		ranges:   make(map[types.Object]LineRange),
	}

	b.indexDeclarations()
	b.enumerateScope()
	b.enumerateMethods()
	b.enumerateDefinitions()
	b.recordAttributes()

	nodes := make([]*Node, 0, len(b.nodes))
	for _, n := range b.nodes {
		sort.Slice(n.Attributes, func(i, j int) bool { return attributeOrder(n.Attributes[i], n.Attributes[j]) })
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID.String() < nodes[j].ID.String() })
	sort.Slice(b.excluded, func(i, j int) bool {
		if b.excluded[i].Name != b.excluded[j].Name {
			return b.excluded[i].Name < b.excluded[j].Name
		}
		return b.excluded[i].Reason < b.excluded[j].Reason
	})
	return nodes, b.excluded, nil
}

// builder holds the state of one enumeration. It produces the nodes of a single package
// and is not a graph, so nothing here depends on nodes outside the package.
type builder struct {
	pkg      *packages.Package
	id       PackageID
	nodes    map[ID]*Node
	excluded []Candidate
	seen     map[types.Object]bool
	reported map[string]bool
	decls    map[types.Object]ast.Decl
	ranges   map[types.Object]LineRange
}

// indexDeclarations records, for each object, the declaration that created it and the
// range of its declaration spec. The range is kept per spec so that two variables inside
// one group declaration keep different ranges.
func (b *builder) indexDeclarations() {
	for _, file := range b.pkg.Syntax {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				b.recordDecl(b.pkg.TypesInfo.Defs[d.Name], decl, d)
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						b.recordDecl(b.pkg.TypesInfo.Defs[s.Name], decl, s)
					case *ast.ValueSpec:
						for _, name := range s.Names {
							b.recordDecl(b.pkg.TypesInfo.Defs[name], decl, s)
						}
					}
				}
			}
		}
	}
}

func (b *builder) recordDecl(obj types.Object, decl ast.Decl, spec ast.Node) {
	if obj == nil {
		return
	}
	b.decls[obj] = decl
	b.ranges[obj] = lineRange(b.pkg.Fset, spec)
}

// enumerateScope records the objects the package scope holds, which are exactly the
// entities created by package-level declarations.
func (b *builder) enumerateScope() {
	scope := b.pkg.Types.Scope()
	for _, name := range scope.Names() {
		b.add(scope.Lookup(name))
	}
}

// enumerateMethods records the methods declared on package-level named types. The
// language exposes no scope for a method, so its owner is the named type its receiver is
// based on.
func (b *builder) enumerateMethods() {
	scope := b.pkg.Types.Scope()
	for _, name := range scope.Names() {
		tn, ok := scope.Lookup(name).(*types.TypeName)
		if !ok {
			continue
		}
		named, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}
		for fn := range named.Methods() {
			b.add(fn)
		}
		// A method set also holds methods that are reachable only through embedding.
		// Such a method is produced by a lookup rather than by a declaration, so it is
		// reported as excluded and the outer type keeps the embedding as an attribute.
		// The set is taken through a pointer so that methods declared with a pointer
		// receiver are part of it, as the language does for a value the caller can take
		// the address of.
		set := types.NewMethodSet(types.NewPointer(named))
		for sel := range set.Methods() {
			fn, ok := sel.Obj().(*types.Func)
			if !ok {
				continue
			}
			receiver, ok := baseNamed(fn.Signature().Recv().Type())
			if ok && receiver.Obj() == tn {
				continue
			}
			b.report(Candidate{Name: fn.Name(), Kind: KindFunc, Reason: ReasonProducedByLookup})
		}
	}
}

// enumerateDefinitions classifies every entity the type checker recorded. This is how
// entities created inside a declaration body, struct fields, interface method specs and
// import aliases are reported as excluded rather than silently dropped.
func (b *builder) enumerateDefinitions() {
	type entry struct {
		ident *ast.Ident
		obj   types.Object
	}
	entries := make([]entry, 0, len(b.pkg.TypesInfo.Defs))
	for ident, obj := range b.pkg.TypesInfo.Defs {
		entries = append(entries, entry{ident, obj})
	}
	sort.Slice(entries, func(i, j int) bool {
		pi, pj := b.pkg.Fset.Position(entries[i].ident.Pos()), b.pkg.Fset.Position(entries[j].ident.Pos())
		if pi.Filename != pj.Filename {
			return pi.Filename < pj.Filename
		}
		if pi.Line != pj.Line {
			return pi.Line < pj.Line
		}
		return entries[i].ident.Name < entries[j].ident.Name
	})
	for _, e := range entries {
		if e.obj == nil {
			if e.ident.Name == "_" {
				b.report(Candidate{Name: e.ident.Name, Reason: ReasonBlankIdentifier})
			}
			continue
		}
		b.add(e.obj)
	}
}

// recordAttributes records the entities that are not nodes as attributes of the node that
// owns them: struct fields, interface method specs, embedded fields, import aliases and
// the nodes a declaration refers to. The use-site is the node that owns the declaration
// the reference appears in, and the attribute names the node the reference resolves to.
func (b *builder) recordAttributes() {
	files := append([]*ast.File(nil), b.pkg.Syntax...)
	sort.Slice(files, func(i, j int) bool {
		return b.pkg.Fset.Position(files[i].Pos()).Filename < b.pkg.Fset.Position(files[j].Pos()).Filename
	})
	for _, file := range files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				obj := b.pkg.TypesInfo.Defs[d.Name]
				if node := b.nodeOf(obj); node != nil {
					if fn, ok := obj.(*types.Func); ok {
						b.recordMethod(node, fn)
					}
					b.recordUses(node, decl)
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						tn, ok := b.pkg.TypesInfo.Defs[s.Name].(*types.TypeName)
						if !ok {
							break
						}
						if tn.IsAlias() {
							// An alias resolves to an existing entity. Record it as an
							// attribute of the node the alias resolves to, when that
							// target is a node of the same package.
							resolved := types.Unalias(tn.Type())
							if named, ok := resolved.(*types.Named); ok {
								resolvedObj := named.Obj()
								if resolvedObj.Pkg() == b.pkg.Types {
									resolvedID := NewID(b.id, resolvedObj)
									if resolvedNode := b.nodes[resolvedID]; resolvedNode != nil {
										b.addAttribute(resolvedNode, Attribute{
											Kind:  AttrAlias,
											Name:  s.Name.Name,
											Value: resolvedID.String(),
											Ref:   resolvedID,
										})
									}
								}
							}
							break
						}
						node := b.nodeOf(b.pkg.TypesInfo.Defs[s.Name])
						if node == nil {
							continue
						}
						b.recordTypeAttributes(node, s)
						b.recordUses(node, s)
					case *ast.ValueSpec:
						for _, name := range s.Names {
							if node := b.nodeOf(b.pkg.TypesInfo.Defs[name]); node != nil {
								b.recordUses(node, s)
							}
						}
					}
				}
			}
		}
	}
}

// recordMethod records a method declared on a package-level type as an attribute of the node
// that owns it, so that a relation can read the method set from the record of the type alone.
// The signature is written without the receiver, in the form a method spec is written, so that
// the two sides of a satisfaction check are comparable as recorded.
func (b *builder) recordMethod(node *Node, fn *types.Func) {
	signature := fn.Signature()
	recv := signature.Recv()
	if recv == nil {
		return
	}
	receiver, ok := baseNamed(recv.Type())
	if !ok {
		return
	}
	owner := b.nodeOf(receiver.Obj())
	if owner == nil || owner.ID == node.ID {
		return
	}
	plain := types.NewSignatureType(nil, nil, nil, signature.Params(), signature.Results(), signature.Variadic())
	b.addAttribute(owner, Attribute{
		Kind:  AttrMethod,
		Name:  fn.Name(),
		Value: types.TypeString(plain, b.qualifier),
		Ref:   node.ID,
	})
}

func (b *builder) recordTypeAttributes(node *Node, spec *ast.TypeSpec) {
	switch t := spec.Type.(type) {
	case *ast.StructType:
		for _, field := range t.Fields.List {
			if len(field.Names) == 0 {
				b.addAttribute(node, Attribute{
					Kind:  AttrEmbeds,
					Name:  b.exprName(field.Type),
					Value: b.valueText(field.Type),
					Ref:   b.referentExpr(field.Type),
				})
				continue
			}
			for _, name := range field.Names {
				b.addAttribute(node, Attribute{
					Kind:  AttrField,
					Name:  name.Name,
					Value: b.valueText(field.Type),
					Ref:   b.referentExpr(field.Type),
				})
			}
		}
	case *ast.InterfaceType:
		for _, field := range t.Methods.List {
			info, ok := b.pkg.TypesInfo.Types[field.Type]
			if !ok {
				continue
			}
			signature, ok := info.Type.(*types.Signature)
			if !ok {
				continue
			}
			name := ""
			if len(field.Names) > 0 {
				name = field.Names[0].Name
			}
			b.addAttribute(node, Attribute{
				Kind:  AttrMethodSpec,
				Name:  name,
				Value: types.TypeString(signature, b.qualifier),
			})
			b.report(Candidate{Name: name, Kind: KindFunc, Reason: ReasonPartOfDeclarationSpec})
		}
	}
}

func (b *builder) recordUses(node *Node, decl ast.Node) {
	// The name written as the receiver names no reference, because the receiver is recorded as the
	// receiver property of the node.
	var receiver ast.Node
	if f, ok := decl.(*ast.FuncDecl); ok {
		receiver = f.Recv
	}
	ast.Inspect(decl, func(n ast.Node) bool {
		if receiver != nil && n == receiver {
			return false
		}
		switch x := n.(type) {
		case *ast.SelectorExpr:
			qualifier, ok := x.X.(*ast.Ident)
			if !ok {
				return true
			}
			pn, ok := b.pkg.TypesInfo.Uses[qualifier].(*types.PkgName)
			if !ok {
				sel := b.pkg.TypesInfo.Selections[x]
				if sel == nil || sel.Obj() == nil || sel.Obj().Name() == "_" {
					return true
				}
				b.addAttribute(node, Attribute{
					Kind:  AttrUse,
					Name:  sel.Obj().Name(),
					Value: types.TypeString(sel.Obj().Type(), b.qualifier),
					Ref:   b.referentObj(sel.Obj()),
				})
				return true
			}
			if pn.Name() != pn.Imported().Name() {
				b.addAttribute(node, Attribute{
					Kind:  AttrAlias,
					Name:  pn.Name(),
					Value: pn.Imported().Path(),
					Ref:   b.referentObj(b.pkg.TypesInfo.Uses[x.Sel]),
				})
			}
			return true
		case *ast.Ident:
			if b.pkg.TypesInfo.Defs[x] != nil {
				return true
			}
			obj := b.pkg.TypesInfo.Uses[x]
			if obj == nil || obj.Name() == "_" {
				return true
			}
			if _, ok := obj.(*types.PkgName); ok {
				return true
			}
			b.addAttribute(node, Attribute{
				Kind:  AttrUse,
				Name:  obj.Name(),
				Value: types.TypeString(b.resolvedType(obj.Type()), b.qualifier),
				Ref:   b.referentObj(obj),
			})
		}
		return true
	})
}

// add classifies one entity and either records it as a node or reports it as an excluded
// candidate. An entity seen more than once is recorded once, so an entity referenced from
// several files is one node with one declaration site.
func (b *builder) add(obj types.Object) Candidate {
	if obj == nil || b.seen[obj] {
		return Candidate{}
	}
	c := Classify(b.pkg.Types, obj)
	if c.Reason == "" && b.decls[obj] == nil {
		c.Reason = b.reasonForInnerEntity(obj)
	}
	// A reason is the whole point of a candidate: an entity that needs one is not a node.
	if c.Reason != "" {
		c.Recorded = false
	}
	b.seen[obj] = true
	if !c.Recorded {
		b.report(c)
		return c
	}
	id := NewID(b.id, obj)
	st := semanticTypeOf(obj, c.Kind, id.Receiver)
	var props NodeProperties
	if obj.Exported() {
		props.Exported = true
	}
	if tn, ok := obj.(*types.TypeName); ok {
		if named, ok := tn.Type().(*types.Named); ok && named.TypeParams() != nil && named.TypeParams().Len() > 0 {
			props.Generic = true
		}
	}
	if f, ok := obj.(*types.Func); ok {
		if f.Signature().TypeParams() != nil && f.Signature().TypeParams().Len() > 0 {
			props.Generic = true
		}
		if b.pkg.Name == "main" && (obj.Name() == "main" || obj.Name() == "init") {
			props.EntryPoint = true
		}
	}
	b.nodes[id] = &Node{
		ID:           id,
		Kind:         c.Kind,
		Name:         obj.Name(),
		Package:      b.id,
		Parent:       b.parent(obj),
		Receiver:     id.Receiver,
		Pos:          position(b.pkg.Fset, obj.Pos()),
		Range:        b.ranges[obj],
		Decl:         b.decls[obj],
		SemanticType: st,
		Properties:   props,
	}
	return c
}

// reasonForInnerEntity explains why an entity whose kind is in the closed set is still not
// a node: it was created inside a declaration rather than by one. The declaration index
// holds only entities created by package-level declarations.
func (b *builder) reasonForInnerEntity(obj types.Object) string {
	switch o := obj.(type) {
	case *types.Var:
		if o.Kind() == types.FieldVar {
			return ReasonPartOfDeclarationSpec
		}
		return ReasonCreatedInsideDeclaration
	case *types.TypeName:
		if _, isParam := o.Type().(*types.TypeParam); isParam {
			return ReasonTypeParameter
		}
		return ReasonCreatedInsideDeclaration
	case *types.Func:
		return ReasonPartOfDeclarationSpec
	default:
		return ReasonCreatedInsideDeclaration
	}
}

// parent reports the node that owns an object. A method is owned by the node of the named
// type behind its receiver. A package-level declaration is owned by the package namespace,
// which groups nodes rather than being a node, so its parent carries no name.
func (b *builder) parent(obj types.Object) ID {
	if f, ok := obj.(*types.Func); ok {
		recv := f.Signature().Recv()
		if recv == nil {
			return ID{Package: b.id}
		}
		if named, ok := baseNamed(recv.Type()); ok {
			return ID{Package: b.id, Name: named.Obj().Name()}
		}
	}
	return ID{Package: b.id}
}

// nodeOf returns the recorded node an object produced, or nil when the entity is not a
// node.
func (b *builder) nodeOf(obj types.Object) *Node {
	if obj == nil || !Classify(b.pkg.Types, obj).Recorded {
		return nil
	}
	return b.nodes[NewID(b.id, obj)]
}

// namesNode reports whether an object is an entity the model records as a node in the graph that
// owns it, which is what a reference may point at. An entity created inside a declaration, a field,
// a type parameter or a method spec names no node, so its attribute carries the referent as a value.
func (b *builder) namesNode(obj types.Object) bool {
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	if obj.Pkg() == b.pkg.Types {
		return b.nodeOf(obj) != nil
	}
	// A referent from another package belongs to that package graph, so its node-ness is read from
	// the scope that declares it rather than from the nodes this build produced.
	return obj.Pkg().Scope().Lookup(obj.Name()) == obj && Classify(obj.Pkg(), obj).Recorded
}

// referentObj builds the identifier of the node an attribute points at. An alias name names an
// existing entity rather than creating one, so the reference points at the node the name resolves
// to, and an alias that resolves to something that is not a node contributes no reference. A
// referent from another package belongs to that package graph, so its module comes from the
// imports of the loaded package. A predeclared identifier has no package and is never a node, so
// it contributes no identifier and is carried as an attribute value instead.
func (b *builder) referentObj(obj types.Object) ID {
	if obj == nil || obj.Pkg() == nil {
		return ID{}
	}
	if tn, ok := obj.(*types.TypeName); ok && tn.IsAlias() {
		named, ok := types.Unalias(tn.Type()).(*types.Named)
		if !ok {
			return ID{}
		}
		obj = named.Obj()
	}
	if !b.namesNode(obj) {
		return ID{}
	}
	pkg := PackageID{PackagePath: obj.Pkg().Path(), Name: obj.Pkg().Name()}
	if imported := b.pkg.Imports[obj.Pkg().Path()]; imported != nil && imported.Module != nil {
		pkg.ModulePath = imported.Module.Path
		pkg.ModuleVersion = imported.Module.Version
	} else if obj.Pkg() == b.pkg.Types {
		pkg = b.id
	}
	return NewID(pkg, obj)
}

// referentExpr reports the node a type expression names. Only an identifier that denotes a
// package-level declaration names a node, so any other expression is carried as a value.
func (b *builder) referentExpr(expr ast.Expr) ID {
	obj, ok := expr.(*ast.Ident)
	if !ok {
		return ID{}
	}
	return b.referentObj(b.objectOf(obj))
}

func (b *builder) objectOf(ident *ast.Ident) types.Object {
	if obj := b.pkg.TypesInfo.Uses[ident]; obj != nil {
		return obj
	}
	return b.pkg.TypesInfo.Defs[ident]
}

func (b *builder) exprName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return b.valueText(expr)
}

func (b *builder) valueText(expr ast.Expr) string {
	info, ok := b.pkg.TypesInfo.Types[expr]
	if !ok {
		return ""
	}
	return types.TypeString(b.resolvedType(info.Type), b.qualifier)
}

// resolvedType reports the type a name or an expression denotes, with an alias read as the type it
// resolves to, so an attribute value names the entity the model records rather than a name that is
// not a node.
func (b *builder) resolvedType(t types.Type) types.Type {
	return types.Unalias(t)
}

func (b *builder) qualifier(pkg *types.Package) string { return pkg.Path() }

func (b *builder) addAttribute(node *Node, a Attribute) {
	if a.Name == "" {
		return
	}
	node.Attributes = append(node.Attributes, a)
}

func (b *builder) report(c Candidate) {
	key := c.Name + "|" + c.Reason
	if b.reported[key] {
		return
	}
	b.reported[key] = true
	b.excluded = append(b.excluded, c)
}
