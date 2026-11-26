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

// Rebuild runs the action for the records of one run, which covers every package the run read as a
// target. It removes the marked files the records of this run map to before writing them and states which
// files it removed. A file whose frontmatter cannot be read is reported rather than removed, because
// ownership cannot be confirmed from it, and a file without a marker is left untouched.
//
// The folders the run owns in full are the folders the packages it read as targets map to, including a name
// the mapping no longer produces for a record that still exists, because the mapping is a function of the
// identifier alone and nothing is carried from the build that wrote the file. A folder that only records
// produced from a reference map to can hold files another run wrote, so there the action removes only the
// marked files the records of this run map to. The folders and the entries in them are visited in sorted
// order, so the removals the action states are the same for the same records.
func Rebuild(root string, targets []graph.PackageID, records []*graph.Node) (Report, error) {
	ownedFolders := make(map[string]bool, len(targets))
	for _, pkg := range targets {
		ownedFolders[Folder(root, pkg)] = true
	}

	ownFiles := make(map[string]bool, len(records))
	visited := make(map[string]bool, len(ownedFolders)+len(records))
	folders := make([]string, 0, len(ownedFolders)+len(records))
	for folder := range ownedFolders {
		visited[folder] = true
		folders = append(folders, folder)
	}
	for _, record := range records {
		file := NotePath(root, record)
		ownFiles[file] = true
		folder := Folder(root, record.Package)
		if !visited[folder] {
			visited[folder] = true
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
			if !ownedFolders[folder] && !ownFiles[file] {
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

	for _, record := range records {
		if _, err := Render(root, record, records); err != nil {
			return report, err
		}
	}
	return report, nil
}
