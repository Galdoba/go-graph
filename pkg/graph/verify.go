package graph

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"
)

// Verify resolves a recorded node back to its declaration from the record alone. It opens
// the recorded file path in a loaded package, finds the declaration the recorded position
// points at, and compares the name, the kind and the line range against the record. Only
// the coordinates are used, so the check works on a load that did not produce the record,
// which is how the coordinates are meant to be read: as provenance of a build rather than
// as part of node identity. A record read back from a file carries no reference to its
// declaration, so a recorded reference is checked only when the record has one.
func Verify(p *packages.Package, n *Node) error {
	if p == nil {
		return fmt.Errorf("graph: no package to resolve the node against")
	}
	if p.Syntax == nil || p.TypesInfo == nil || p.Fset == nil {
		return fmt.Errorf("graph: package %s is incomplete, load it with graph.LoadMode", p.ID)
	}
	if n.Decl != nil && !declaresName(n.Decl, n.Name) {
		return fmt.Errorf("graph: node %s references a declaration that does not name %s", n.ID, n.Name)
	}
	for _, file := range p.Syntax {
		if p.Fset.Position(file.Pos()).Filename != n.Pos.File {
			continue
		}
		for _, decl := range file.Decls {
			for _, ident := range declaredNames(decl) {
				if ident.Name != n.Name {
					continue
				}
				pos := p.Fset.Position(ident.Pos())
				if pos.Line != n.Pos.Line || pos.Column != n.Pos.Column {
					continue
				}
				if pos.Line < n.Range.StartLine || pos.Line > n.Range.EndLine {
					continue
				}
				obj := p.TypesInfo.Defs[ident]
				if obj == nil || obj.Name() != n.Name || kindOf(obj) != n.Kind {
					continue
				}
				return nil
			}
		}
	}
	return fmt.Errorf("graph: node %s does not resolve to a declaration of %s %s at %s:%d",
		n.ID, n.Kind, n.Name, n.Pos.File, n.Range.StartLine)
}

// declaresName reports whether a declaration declares an entity with this name, which is
// how the recorded reference is checked against the record it belongs to.
func declaresName(decl ast.Decl, name string) bool {
	for _, ident := range declaredNames(decl) {
		if ident.Name == name {
			return true
		}
	}
	return false
}

// declaredNames returns the identifiers a declaration defines, which are the names a node
// can have. Blank names are skipped because they name no entity.
func declaredNames(decl ast.Decl) []*ast.Ident {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		if d.Name != nil && d.Name.Name != "_" {
			return []*ast.Ident{d.Name}
		}
	case *ast.GenDecl:
		var names []*ast.Ident
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				if s.Name != nil && s.Name.Name != "_" {
					names = append(names, s.Name)
				}
			case *ast.ValueSpec:
				for _, name := range s.Names {
					if name.Name != "_" {
						names = append(names, name)
					}
				}
			}
		}
		return names
	}
	return nil
}
