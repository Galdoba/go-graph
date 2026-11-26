package vault

import (
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/Galdoba/go-graph/pkg/graph"
	"github.com/Galdoba/go-graph/pkg/link"
)

// TestARunOverAPackageThatImportsTheRepositoryWritesOneNotePerReferencedName checks the vault a run
// produces for one package that imports a package of the repository and one standard library package:
// one note for each name its references name, written in the folder of the package that declares the
// entity, and nothing for a name of that package that no reference names. The vault root is a constant,
// so the run is reproduced in a temporary folder rather than in the vault the command writes.
func TestARunOverAPackageThatImportsTheRepositoryWritesOneNotePerReferencedName(t *testing.T) {
	root := t.TempDir()
	found, err := packages.Load(&packages.Config{Mode: graph.LoadMode, Dir: projectRoot()}, "./pkg/link")
	if err != nil {
		t.Fatalf("load pkg/link: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("load pkg/link: expected one package, got %d", len(found))
	}
	if len(found[0].Errors) > 0 {
		t.Fatalf("load pkg/link: %v", found[0].Errors)
	}
	for _, importPath := range []string{"github.com/Galdoba/go-graph/pkg/graph", "sort"} {
		if found[0].Imports[importPath] == nil {
			t.Fatalf("pkg/link does not import %s, so the run does not read a package it never built", importPath)
		}
	}

	built, _, err := graph.Build(found[0])
	if err != nil {
		t.Fatalf("build pkg/link: %v", err)
	}
	records, err := graph.SecondPass(found[0])
	if err != nil {
		t.Fatalf("second pass over pkg/link: %v", err)
	}
	if len(records) == 0 {
		t.Fatalf("the run produced no record from a reference: %v", nodeNames(built))
	}
	named := map[string]bool{}
	for _, n := range records {
		named[n.Package.PackagePath] = true
	}
	for _, importPath := range []string{"github.com/Galdoba/go-graph/pkg/graph", "sort"} {
		if !named[importPath] {
			t.Fatalf("the run produced no record for a name of %s: %v", importPath, nodeNames(records))
		}
	}

	combined := append(append([]*graph.Node{}, built...), records...)
	if _, err := Rebuild(root, graph.PackageIDOf(found[0]), combined); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	// One note per referenced name, in the folder of the package that declares the entity.
	notes := map[string][]string{}
	for _, n := range records {
		file := NotePath(root, n)
		if _, err := os.Stat(file); err != nil {
			t.Errorf("no note written for the name %s: %v", n.ID, err)
			continue
		}
		notes[Folder(root, n.Package)] = append(notes[Folder(root, n.Package)], file)
	}
	for folder, files := range notes {
		entries, err := os.ReadDir(folder)
		if err != nil {
			t.Fatalf("read folder %s: %v", folder, err)
		}
		if len(entries) != len(files) {
			t.Errorf("folder %s holds %d files, want the %d notes this run produced: %v", folder, len(entries), len(files), files)
			continue
		}
		for _, entry := range entries {
			if !slices.Contains(files, path.Join(folder, entry.Name())) {
				t.Errorf("folder %s holds a note for a name no reference names: %s", folder, entry.Name())
			}
		}
	}

	// A name of a referenced package that no reference names gets no file. The references of pkg/link
	// name graph.Node but never graph.Verify, so the folder the graph package maps to holds no note for
	// a name the run did not produce.
	if _, err := os.Stat(path.Join(Folder(root, graph.PackageIDOf(found[0].Imports["github.com/Galdoba/go-graph/pkg/graph"])), "Verify.md")); err == nil {
		t.Errorf("a note was written for a name no reference names: %s", "Verify.md")
	}

	// The unexported node of the built package lists the links its references answer for.
	unexported := nodeByName(built, "recordsMethod")
	if unexported == nil {
		t.Fatalf("pkg/link produced no node named recordsMethod: %v", nodeNames(built))
	}
	if unexported.Properties.Exported {
		t.Fatalf("recordsMethod is exported, want an unexported node of the package the run built")
	}
	var links []string
	foreignLinks := 0
	for _, pair := range link.Pairs(combined) {
		if pair.Source != unexported.ID {
			continue
		}
		target := nodeByID(combined, pair.Target.String())
		if target == nil {
			t.Fatalf("the pair from %s names %s, which is not in the collection", unexported.ID, pair.Target)
		}
		if target.Provenance == graph.ProvenanceFromName {
			foreignLinks++
		}
		label := target.Name
		if target.Receiver != "" {
			label = target.Receiver + "." + target.Name
		}
		links = append(links, "- [["+path.Join(Folder("", target.Package), FileName(target))+"|"+label+"]]")
	}
	if len(links) == 0 {
		t.Fatalf("no connection is answered from the unexported node %s", unexported.ID)
	}
	if foreignLinks == 0 {
		t.Fatalf("the connections from %s name no record produced from a reference: %v", unexported.ID, links)
	}

	content := string(readFile(t, NotePath(root, unexported)))
	section := content[strings.Index(content, "## Related"):]
	for _, link := range links {
		if !strings.Contains(section, link) {
			t.Errorf("the Related section does not list %s:\n%s", link, section)
		}
	}
}
