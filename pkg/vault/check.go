package vault

import (
	"fmt"
	"path"

	"golang.org/x/tools/go/packages"

	"github.com/Galdoba/go-graph/pkg/graph"
)

// Reasons a check refutes a file. The set is closed so a test can assert on it.
const (
	ReasonUnreadableFrontmatter   = "frontmatter cannot be read"
	ReasonLocationMismatch        = "location is not the folder the recorded package path maps to"
	ReasonNameMismatch            = "file name does not match the record the file carries"
	ReasonRegionNoLongerDeclares  = "the recorded region no longer declares the node"
	ReasonPackageNotReadable      = "the package the record names cannot be read"
	ReasonPackageNoLongerDeclares = "the package the record names no longer declares the node"
)

// Result is the answer of a check for one file. Reason is empty when the file is confirmed.
type Result struct {
	Confirmed bool
	Reason    string
}

// Check confirms or refutes one file from the file and the code at the moment it is invoked,
// without the build that produced the file. The recorded package path is the claim the folder is
// compared against, so a file that sits in a folder its own record does not map to is refuted. A
// record that states a region is answered at that region, and a record that states none is answered
// against the entity the package declares, because nothing in the file names a place to read it at.
// A package that cannot be read is its own reason rather than a reason about a region.
func Check(root string, file string) Result {
	record, err := Read(file)
	if err != nil {
		return Result{Reason: ReasonUnreadableFrontmatter}
	}

	parts, err := ParseFileName(path.Base(file))
	if err != nil {
		return Result{Reason: ReasonNameMismatch}
	}

	semantic, ok := graph.ParseSemanticType(record.SemanticType)
	if !ok {
		return Result{Reason: ReasonUnreadableFrontmatter}
	}
	kind, ok := graph.ParseKind(record.Kind)
	if !ok {
		return Result{Reason: ReasonUnreadableFrontmatter}
	}

	pkg := graph.PackageID{
		ModulePath:    record.Module,
		ModuleVersion: record.Version,
		PackagePath:   record.PackagePath,
		Name:          record.Package,
	}
	node := &graph.Node{
		ID:           graph.ID{Package: pkg, Receiver: parts.Receiver, Name: parts.Name},
		Package:      pkg,
		Kind:         kind,
		Name:         parts.Name,
		Receiver:     parts.Receiver,
		SemanticType: semantic,
		Properties:   propertiesFrom(record.Properties),
		Pos:          graph.Position{File: record.File, Line: record.Line, Column: record.Column},
		Range:        graph.LineRange{StartLine: record.StartLine, EndLine: record.EndLine},
	}

	if folderOf(file) != Folder(root, pkg) {
		return Result{Reason: ReasonLocationMismatch}
	}
	// The name is read as the node part of the identifier the record states. The semantic type the
	// record states says whether that part carries a receiver, because a method is named with the type
	// it is declared on and every other node is named by the package alone, so a name that carries a
	// receiver where the record states none, or omits one where the record states a method, is not the
	// node part of the identifier the record states.
	if (semantic == graph.SemanticMethod) != (parts.Receiver != "") {
		return Result{Reason: ReasonNameMismatch}
	}

	if record.File == "" {
		// A record that states no region names no place to read the node at, so the answer is given by
		// the entity the package declares, which is the whole a record produced from a name can state.
		found, err := loadForPackage(record.PackagePath)
		if err != nil {
			return Result{Reason: ReasonPackageNotReadable}
		}
		if !graph.DeclaresEntity(found, node) {
			return Result{Reason: ReasonPackageNoLongerDeclares}
		}
		return Result{Confirmed: true}
	}

	found, err := loadForCode(record.File)
	if err != nil {
		return Result{Reason: ReasonPackageNotReadable}
	}
	if err := graph.Verify(found, node); err != nil {
		return Result{Reason: ReasonRegionNoLongerDeclares}
	}
	return Result{Confirmed: true}
}

// propertiesFrom reads the properties the frontmatter states back into the set that wrote them, so the
// exported value a record states can be compared against the entity the package declares.
func propertiesFrom(names []string) graph.NodeProperties {
	var p graph.NodeProperties
	for _, name := range names {
		switch name {
		case "exported":
			p.Exported = true
		case "generic":
			p.Generic = true
		case "entry_point":
			p.EntryPoint = true
		}
	}
	return p
}

// loadForCode loads the package the recorded region belongs to from the directory the recorded
// file names, which is how a check reads the code without the build that produced the record.
func loadForCode(file string) (*packages.Package, error) {
	found, err := packages.Load(&packages.Config{Mode: graph.LoadMode, Dir: path.Dir(file)}, ".")
	if err != nil {
		return nil, err
	}
	if len(found) != 1 {
		return nil, fmt.Errorf("vault: %s does not resolve to one package", file)
	}
	if len(found[0].Errors) > 0 {
		return nil, fmt.Errorf("vault: %s cannot be read: %v", file, found[0].Errors[0])
	}
	return found[0], nil
}

// loadForPackage loads the package a record names from the directory the check runs in, which is how a
// check answers a record that states no region: the package is read on its own, without the build that
// produced the record.
func loadForPackage(packagePath string) (*packages.Package, error) {
	found, err := packages.Load(&packages.Config{Mode: graph.LoadMode}, packagePath)
	if err != nil {
		return nil, err
	}
	if len(found) != 1 {
		return nil, fmt.Errorf("vault: %s does not resolve to one package", packagePath)
	}
	if len(found[0].Errors) > 0 {
		return nil, fmt.Errorf("vault: %s cannot be read: %v", packagePath, found[0].Errors[0])
	}
	return found[0], nil
}
