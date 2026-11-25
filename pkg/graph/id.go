package graph

import (
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// PackageID identifies the package a node belongs to. A package is a namespace that
// groups nodes rather than being a node, so it carries no name in the sense of a node name,
// but it records the name the package declares. That name is not always the last segment of
// the package path, because a directory can be named differently from the package it holds,
// so the record keeps both and the form reads them as two parts.
type PackageID struct {
	ModulePath    string `json:"module_path"`
	ModuleVersion string `json:"module_version"`
	PackagePath   string `json:"package_path"`
	Name          string `json:"name"`
}

// String prints the canonical form of the package part, which is the part of a note name
// that names the package. The path is written with tildes instead of slashes, because the
// identifier is used as a note name. The module is not repeated when the package path
// already names it, which is what a load of a package that sits on disk reports; the
// module path stays in the record for a consumer that groups nodes by module. The version
// is written after the module part, and it is left out when the module has none.
// String prints the canonical form of the package part, which is the part of a note name
// that names the package. The path is written with tildes instead of slashes, because the
// identifier is used as a note name. The module is not repeated when the package path
// already names it, which is what a load of a package that sits on disk reports; the
// module path stays in the record for a consumer that groups nodes by module. The version is
// written after the module part, and it is left out when the module has none. The name the
// package declares is written as its own part, and it is stored even when it matches the
// last segment of the path. A package
// that reports no module prints as the name it declares alone, because the printed form has no
// module part to carry the path; the path stays in the record. The package path is read as a path
// segment of the module, so a module whose name is a prefix of another module name is not read as
// the same module.
func (p PackageID) String() string {
	module := noteName(p.ModulePath)
	path := noteName(p.PackagePath)
	switch {
	case module == "":
		path = ""
	case moduleSegment(p.ModulePath, p.PackagePath):
		if p.ModuleVersion != "" {
			path = module + "@" + p.ModuleVersion + noteName(p.PackagePath[len(p.ModulePath):])
		}
	case p.ModuleVersion == "":
		path = module + "~" + path
	default:
		path = module + "@" + p.ModuleVersion + "~" + path
	}
	if p.Name == "" {
		return path
	}
	if path == "" {
		return p.Name
	}
	return path + "~" + p.Name
}

// noteName writes a path so that it can be used as a note name: the characters a note name
// cannot contain do not appear in it.
func noteName(name string) string {
	return strings.ReplaceAll(name, "/", "~")
}

// moduleSegment reports whether the package path begins with the module path at a path segment
// boundary, so a module named project1 does not read as the module of a package under a module
// named project12.
func moduleSegment(modulePath, packagePath string) bool {
	if !strings.HasPrefix(packagePath, modulePath) {
		return false
	}
	rest := packagePath[len(modulePath):]
	return rest == "" || strings.HasPrefix(rest, "/")
}

// ID identifies a node. It is built from the components of a declaration rather than
// from a printed form, so formatting or importers cannot change it. It is qualified by
// module path and module version, which keeps the same package path in two modules, or
// the same package at two versions, as two different sets of nodes. The receiver is set
// for function nodes only, which is what keeps a method named Read on Struct1 distinct
// from a method named Read on Struct2.
type ID struct {
	Package  PackageID `json:"package"`
	Receiver string    `json:"receiver"`
	Name     string    `json:"name"`
}

// String prints the canonical form of the identifier, which is the name of the note that
// holds the node. The package part is written with tildes and the node part with dots, so
// the two cannot be read for one another: a Go identifier contains neither, so the last
// segment is always the name. The receiver segment is left out for a node that has no
// receiver, and the package namespace has no name of its own, so its form is the package
// part alone rather than a form that ends in a dot.
func (id ID) String() string {
	s := id.Package.String()
	if id.Receiver != "" {
		s += "." + id.Receiver
	}
	if id.Name == "" {
		return s
	}
	return s + "." + id.Name
}

// NewID builds the identifier of an object inside a package. For a function node the
// receiver named type is part of the identifier; for the other kinds the qualified name
// is enough.
func NewID(pkg PackageID, obj types.Object) ID {
	id := ID{Package: pkg, Name: obj.Name()}
	if f, ok := obj.(*types.Func); ok {
		id.Receiver = receiverName(f.Signature().Recv())
	}
	return id
}

// receiverName reports the named type a receiver is based on. The pointer wrapper is a
// property of the node, not part of its identifier or of its parent.
func receiverName(recv *types.Var) string {
	if recv == nil {
		return ""
	}
	t := recv.Type()
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	// Resolve aliases to their target named type.
	if _, ok := t.(*types.Alias); ok {
		t = types.Unalias(t)
	}
	named, ok := t.(*types.Named)
	if !ok {
		return ""
	}
	return named.Obj().Name()
}

// baseNamed strips the pointer wrapper from a receiver type and resolves aliases
// to their target named type.
func baseNamed(t types.Type) (*types.Named, bool) {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	if _, ok := t.(*types.Alias); ok {
		t = types.Unalias(t)
	}
	named, ok := t.(*types.Named)
	return named, ok
}

// packageID reports the package a loaded package belongs to. Module information can be
// missing for the standard library, in which case the package path alone qualifies the
// identifier in the record, while the printed form prints the name the package declares, because
// the printed form has no module part to carry the path.
func packageID(p *packages.Package) PackageID {
	id := PackageID{PackagePath: p.PkgPath, Name: p.Name}
	if p.Module != nil {
		id.ModulePath = p.Module.Path
		id.ModuleVersion = p.Module.Version
	}
	return id
}
