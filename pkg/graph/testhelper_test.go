package graph

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"testing"

	"golang.org/x/tools/go/packages"
)

// loadFixture loads a fixture package through go/packages, which is the path the
// enumeration is designed for.
func loadFixture(t *testing.T, dir string) *packages.Package {
	t.Helper()
	pkgs, err := packages.Load(&packages.Config{Mode: LoadMode}, "./testdata/fixtures/"+dir)
	if err != nil {
		t.Fatalf("load fixture %s: %v", dir, err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("fixture %s: expected one package, got %d", dir, len(pkgs))
	}
	p := pkgs[0]
	if len(p.Errors) > 0 {
		t.Fatalf("fixture %s: %v", dir, p.Errors)
	}
	return p
}

// sourceLoader type-checks source text into a package, which lets a test choose the
// module path and version that qualify an identifier.
type sourceLoader struct {
	t      *testing.T
	fset   *token.FileSet
	loaded map[string]*packages.Package
}

func newSourceLoader(t *testing.T) *sourceLoader {
	return &sourceLoader{t: t, fset: token.NewFileSet(), loaded: make(map[string]*packages.Package)}
}

func (l *sourceLoader) load(modulePath, version, pkgPath string, sources map[string]string) *packages.Package {
	l.t.Helper()
	names := make([]string, 0, len(sources))
	for name := range sources {
		names = append(names, name)
	}
	sort.Strings(names)

	files := make([]*ast.File, 0, len(names))
	for _, name := range names {
		file, err := parser.ParseFile(l.fset, modulePath+"/"+pkgPath+"/"+name, sources[name], parser.ParseComments)
		if err != nil {
			l.t.Fatalf("parse %s: %v", name, err)
		}
		files = append(files, file)
	}

	info := &types.Info{
		Types:  make(map[ast.Expr]types.TypeAndValue),
		Defs:   make(map[*ast.Ident]types.Object),
		Uses:   make(map[*ast.Ident]types.Object),
		Scopes: make(map[ast.Node]*types.Scope),
	}
	tp, err := (&types.Config{Importer: l}).Check(pkgPath, l.fset, files, info)
	if err != nil {
		l.t.Fatalf("check %s: %v", pkgPath, err)
	}

	p := &packages.Package{
		ID:        pkgPath,
		Name:      tp.Name(),
		PkgPath:   pkgPath,
		Module:    &packages.Module{Path: modulePath, Version: version},
		Types:     tp,
		Fset:      l.fset,
		Syntax:    files,
		TypesInfo: info,
		Imports:   make(map[string]*packages.Package),
	}
	for _, file := range files {
		for _, imp := range file.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				l.t.Fatalf("import %s: %v", imp.Path.Value, err)
			}
			if imported, ok := l.loaded[path]; ok {
				p.Imports[path] = imported
			}
		}
	}
	l.loaded[pkgPath] = p
	return p
}

func (l *sourceLoader) Import(path string) (*types.Package, error) {
	if p, ok := l.loaded[path]; ok {
		return p.Types, nil
	}
	return nil, fmt.Errorf("graph test: package %s is not loaded", path)
}

func mustBuild(t *testing.T, p *packages.Package) ([]*Node, []Candidate) {
	t.Helper()
	nodes, excluded, err := Build(p)
	if err != nil {
		t.Fatalf("build %s: %v", p.ID, err)
	}
	return nodes, excluded
}

func nodeByName(nodes []*Node, name string) *Node {
	for _, n := range nodes {
		if n.Name == name {
			return n
		}
	}
	return nil
}

func hasCandidate(candidates []Candidate, name, reason string) bool {
	for _, c := range candidates {
		if c.Name == name && c.Reason == reason {
			return true
		}
	}
	return false
}

func nodeNames(nodes []*Node) []string {
	names := make([]string, 0, len(nodes))
	for _, n := range nodes {
		names = append(names, n.Name)
	}
	return names
}

func nodeByID(nodes []*Node, id string) *Node {
	for _, n := range nodes {
		if n.ID.String() == id {
			return n
		}
	}
	return nil
}

// loadGround loads the test ground at the repository root, which declares one entity of every
// node kind and one entity of every kind the model excludes.
func loadGround(t *testing.T) *packages.Package {
	t.Helper()
	return loadGroundPackage(t, "./testdata")
}

// loadGroundPackage loads one package of the test ground by pattern, so a second ground
// package can be checked on its own without changing what the first one asserts.
func loadGroundPackage(t *testing.T, pattern string) *packages.Package {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the test ground")
	}
	root := filepath.Join(filepath.Dir(file), "..", "..")
	pkgs, err := packages.Load(&packages.Config{Mode: LoadMode, Dir: root}, pattern)
	if err != nil {
		t.Fatalf("load the test ground: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("the test ground %s: expected one package, got %d", pattern, len(pkgs))
	}
	p := pkgs[0]
	if len(p.Errors) > 0 {
		t.Fatalf("the test ground %s: %v", pattern, p.Errors)
	}
	return p
}
