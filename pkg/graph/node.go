package graph

import (
	"go/ast"
	"go/token"
)

// Position is where a node is declared. It is provenance of the build that produced the
// record rather than part of node identity, because any edit above a declaration moves
// it without changing the node.
type Position struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// LineRange is the range of the declaration spec that created a node. The range is per
// spec rather than per group declaration, because one group declaration can create
// several nodes and a shared range makes verification ambiguous. It covers the
// declaration of the node, including parts of it that are not nodes, such as a struct
// body, and it is recomputed on rebuild rather than matched between builds. The file is
// not part of the range: a declaration spec and the identifier it declares are in the same
// file, which the record carries once, in the position.
type LineRange struct {
	StartLine int `json:"start_line"`
	EndLine   int `json:"end_line"`
}

func position(fset *token.FileSet, pos token.Pos) Position {
	p := fset.Position(pos)
	return Position{File: p.Filename, Line: p.Line, Column: p.Column}
}

func lineRange(fset *token.FileSet, node ast.Node) LineRange {
	start, end := fset.Position(node.Pos()), fset.Position(node.End())
	return LineRange{StartLine: start.Line, EndLine: end.Line}
}

// Node is the record of one entity the enumeration accepted. The record carries enough
// information to locate its declaration from the node alone.
type Node struct {
	ID           ID             `json:"id"`
	Kind         Kind           `json:"kind"`
	Name         string         `json:"name"`
	Package      PackageID      `json:"package"`
	Parent       ID             `json:"parent"`   // the node that owns it; the package namespace when Parent.Name is empty
	Receiver     string         `json:"receiver"` // the named type a function node is based on, empty for other kinds
	Pos          Position       `json:"pos"`      // the position of the declared identifier
	Range        LineRange      `json:"range"`
	Decl         ast.Decl       `json:"decl"` // the declaration that created the node
	SemanticType SemanticType   `json:"semantic_type"`
	Properties   NodeProperties `json:"properties"`
	Attributes   []Attribute    `json:"attributes"`
	Provenance   Provenance     `json:"provenance"` // where the values of the record were read from
}
