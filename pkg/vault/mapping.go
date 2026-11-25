package vault

import (
	"fmt"
	"path"
	"strings"

	"github.com/Galdoba/go-graph/pkg/graph"
)

// Root is the vault root. It is a constant for now and moves to configuration later.
const Root = "/home/galdoba/Documents/Obsidian/GoRepository"

// Extension is the suffix a note file carries.
const Extension = ".md"

// Folder reports the folder a recorded package path maps to under the vault root. The folder is
// built from the path rather than from the printed form, because the printed form has no module
// part to carry the path for a package that reports no module.
func Folder(root string, pkg graph.PackageID) string {
	return path.Join(root, pkg.PackagePath)
}

// FileName reports the file name a record maps to: the node part of the identifier, the receiver
// when the node has one and then the name the package declares. No part of the name states a
// classification, so nothing a rule decides about the node changes the name. The receiver is part of
// the name because a package can declare an entity and a method with the same name.
func FileName(n *graph.Node) string {
	name := ""
	if n.Receiver != "" {
		name += n.Receiver + "."
	}
	return name + n.Name + Extension
}

// NotePath reports the file a record maps to.
func NotePath(root string, n *graph.Node) string {
	return path.Join(Folder(root, n.Package), FileName(n))
}

// NameParts are the parts a file name carries. A file name holds no package information and states
// no classification, so the parts it carries are the receiver when it carries one and the node name.
type NameParts struct {
	Receiver string
	Name     string
}

// ParseFileName reads a file name back into the node part of the identifier. A Go identifier cannot
// contain a dot, so a name the mapping produces carries at most one dot and that dot separates the
// receiver from the node name, while a name without a dot names an entity the package declares on its
// own. No part of a name is matched against a closed set, so a node name that contains an underscore
// stays intact. A name that carries more than one dot cannot come from two identifiers, so it is not a
// name the mapping produces and is not read.
func ParseFileName(name string) (NameParts, error) {
	base := strings.TrimSuffix(name, Extension)
	if base == "" {
		return NameParts{}, fmt.Errorf("vault: name %s carries no node name", name)
	}

	parts := NameParts{}
	receiver, node, ok := strings.Cut(base, ".")
	if !ok {
		parts.Name = base
		return parts, nil
	}
	if strings.Contains(node, ".") {
		return NameParts{}, fmt.Errorf("vault: name %s carries more than the dot that separates the receiver from the node name", name)
	}
	if receiver == "" || node == "" {
		return NameParts{}, fmt.Errorf("vault: name %s does not state a receiver and a node name", name)
	}
	parts.Receiver = receiver
	parts.Name = node
	return parts, nil
}
