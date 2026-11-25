package vault

import (
	"os"
	"path"
	"strings"

	"github.com/Galdoba/go-graph/pkg/graph"
)

// Report is what the rebuild action states about the files it touched.
type Report struct {
	Removed    []string
	Unreadable []string
}

// Rebuild runs the action for one package. It removes the files the graph owns, those whose marker
// reads as true, in the folder the recorded package path maps to, states which files it removed,
// and then writes the files the mapping produces for the nodes the build produced. A file whose
// frontmatter cannot be read is reported rather than removed, because ownership cannot be
// confirmed from it, and a file without a marker is left untouched.
//
// Removal covers every marked file in the folder, including a name the mapping no longer produces for
// a node that still exists, because the mapping is a function of the identifier alone and nothing is
// carried from the build that wrote the file.
func Rebuild(root string, pkg graph.PackageID, nodes []*graph.Node) (Report, error) {
	folder := Folder(root, pkg)

	var report Report
	entries, err := os.ReadDir(folder)
	if err != nil && !os.IsNotExist(err) {
		return report, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), Extension) {
			continue
		}
		file := path.Join(folder, entry.Name())
		record, err := Read(file)
		if err != nil {
			report.Unreadable = append(report.Unreadable, file)
			continue
		}
		if !record.Generated {
			continue
		}
		if err := os.Remove(file); err != nil {
			return report, err
		}
		report.Removed = append(report.Removed, file)
	}

	for _, n := range nodes {
		if _, err := Render(root, n, nodes); err != nil {
			return report, err
		}
	}
	return report, nil
}
