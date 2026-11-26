package vault

import (
	"os"
	"path"
	"sort"
	"strings"

	"github.com/Galdoba/go-graph/pkg/graph"
)

// Report is what the rebuild action states about the files it touched.
type Report struct {
	Removed    []string
	Unreadable []string
}

// Rebuild runs the action for one package. It removes the files the graph owns, those whose marker
// reads as true, states which files it removed, and then writes the files the mapping produces for the
// records of this run. A file whose frontmatter cannot be read is reported rather than removed, because
// ownership cannot be confirmed from it, and a file without a marker is left untouched.
//
// For the package the run built, removal covers every marked file in the folder that package path maps
// to, including a name the mapping no longer produces for a node that still exists, because the mapping
// is a function of the identifier alone and nothing is carried from the build that wrote the file. For a
// record produced from a reference, the folder it maps to can hold files another package's run wrote, so
// the action removes only the marked files the records of this run map to there and leaves a marked file
// it did not produce. The folders and the entries in them are visited in sorted order, so the removals
// the action states are the same for the same records.
func Rebuild(root string, pkg graph.PackageID, nodes []*graph.Node) (Report, error) {
	own := Folder(root, pkg)

	owned := make(map[string]bool, len(nodes))
	folders := []string{own}
	seen := map[string]bool{own: true}
	for _, n := range nodes {
		file := NotePath(root, n)
		owned[file] = true
		if folder := Folder(root, n.Package); !seen[folder] {
			seen[folder] = true
			folders = append(folders, folder)
		}
	}
	sort.Strings(folders)

	var report Report
	for _, folder := range folders {
		entries, err := os.ReadDir(folder)
		if err != nil && !os.IsNotExist(err) {
			return report, err
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), Extension) {
				continue
			}
			file := path.Join(folder, entry.Name())
			if folder != own && !owned[file] {
				continue
			}
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
	}

	for _, n := range nodes {
		if _, err := Render(root, n, nodes); err != nil {
			return report, err
		}
	}
	return report, nil
}
