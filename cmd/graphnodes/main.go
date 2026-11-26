// Command graphnodes finds the nodes of the packages it reads as targets, prints them to stderr, reads
// every record back through a second, independent load, and writes the note file each record maps to in
// the vault once for the run, which is the round trip that shows a record carries enough information to
// locate its declaration on its own and to name the file it belongs to.
package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/Galdoba/go-graph/pkg/graph"
	"github.com/Galdoba/go-graph/pkg/vault"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [--check] [--recursive] <package>\n", os.Args[0])
		os.Exit(2)
	}

	checkFlag := false
	recursiveFlag := false
	var target string
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--check":
			checkFlag = true
		case "--recursive":
			recursiveFlag = true
		default:
			target = arg
		}
	}
	if target == "" {
		fmt.Fprintf(os.Stderr, "usage: %s [--check] [--recursive] <package>\n", os.Args[0])
		os.Exit(2)
	}

	found, err := load(target, recursiveFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	reloaded, err := load(target, recursiveFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	resolved := make(map[string]*packages.Package, len(reloaded))
	imported := make(map[string]*packages.Package, len(reloaded))
	for _, p := range reloaded {
		resolved[p.PkgPath] = p
		for path, other := range p.Imports {
			if imported[path] == nil {
				imported[path] = other
			}
		}
	}

	for _, p := range found {
		report(p)
	}

	targets, collection, counts, err := collect(found)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	for _, c := range counts {
		fmt.Fprintf(os.Stderr, "%s: %d nodes, %d records produced from a reference\n", c.pkg, c.built, c.fromNames)
	}

	// The run holds one collection for every package it read as a target, so an entity both passes read
	// is one entry and the collection is read in the same order every time.
	fmt.Fprintf(os.Stderr, "run: %d packages, %d records\n", len(targets), len(collection))

	failed := 0
	for _, n := range collection {
		fmt.Fprintf(os.Stderr, "  %s kind=%v semantic=%v receiver=%q parent=%v exported=%v generic=%v entry_point=%v file=%s line=%d column=%d range=%d-%d attributes=%d provenance=%v\n",
			n.ID, n.Kind, n.SemanticType, n.Receiver, n.Parent, n.Properties.Exported, n.Properties.Generic, n.Properties.EntryPoint,
			n.Pos.File, n.Pos.Line, n.Pos.Column, n.Range.StartLine, n.Range.EndLine, len(n.Attributes), n.Provenance)

		other := resolved[n.Package.PackagePath]
		if other == nil {
			// A record produced from a name belongs to a package this run did not load as a target, so it
			// is read back against the second load as it holds the import.
			other = imported[n.Package.PackagePath]
		}
		if other == nil {
			fmt.Fprintf(os.Stderr, "  round trip %s: package %s was not loaded again\n", n.ID, n.Package.PackagePath)
			failed++
			continue
		}
		if err := graph.Verify(other, n); err != nil {
			fmt.Fprintf(os.Stderr, "  round trip %s: %v\n", n.ID, err)
			failed++
			continue
		}
		fmt.Fprintf(os.Stderr, "  round trip %s: ok\n", n.ID)
	}

	notes, err := writeNotes(targets, collection)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if checkFlag {
		for _, file := range notes {
			result := vault.Check(vault.Root, file)
			if !result.Confirmed {
				fmt.Fprintf(os.Stderr, "  check %s: %s\n", file, result.Reason)
				failed++
				continue
			}
			fmt.Fprintf(os.Stderr, "  check %s: ok\n", file)
		}
	}

	if failed > 0 {
		os.Exit(1)
	}
}

// writeNotes runs the action for the records of the run, states what it removed, and lists the files it
// wrote so each can be read back and confirmed against the code. The packages the run read as targets are
// passed with the records, because the run owns the folders they map to, while a record produced from a
// reference maps to a folder another run may write in.
func writeNotes(targets []graph.PackageID, records []*graph.Node) ([]string, error) {
	if len(records) == 0 {
		return nil, nil
	}

	report, err := vault.Rebuild(vault.Root, targets, records)
	if err != nil {
		return nil, err
	}
	for _, file := range report.Removed {
		fmt.Fprintf(os.Stderr, "  removed %s\n", file)
	}
	for _, file := range report.Unreadable {
		fmt.Fprintf(os.Stderr, "  unreadable %s\n", file)
	}

	files := make([]string, 0, len(records))
	for _, n := range records {
		file := vault.NotePath(vault.Root, n)
		fmt.Fprintf(os.Stderr, "  wrote %s\n", file)
		files = append(files, file)
	}
	return files, nil
}

// load resolves the target from the module that owns it. The loader reads a pattern against
// the main module of its own directory, so a path to a package of another repository has to
// be loaded from that path rather than from the working directory. When recursive is true,
// the loader also resolves all subpackages beneath the target.
func load(target string, recursive bool) ([]*packages.Package, error) {
	cfg := &packages.Config{Mode: graph.LoadMode}
	if dir, ok := packageDir(target); ok {
		cfg.Dir = dir
		if recursive {
			return packages.Load(cfg, "./...")
		}
		return packages.Load(cfg, ".")
	}
	if recursive {
		return packages.Load(cfg, target+"/...")
	}
	return packages.Load(cfg, target)
}

// packageCounts states what one package of the run contributed to the collection, so the counts the run
// states can be read without the command printing them.
type packageCounts struct {
	pkg       string
	built     int
	fromNames int
}

// collect assembles the collection the run holds for the packages it read as targets: the records each
// package produced together with the records its references name, merged into one entry per identifier
// and ordered by it. The targets are returned in the order the load gave them, because the folders the
// run owns in full are the folders those packages map to.
func collect(found []*packages.Package) (targets []graph.PackageID, records []*graph.Node, counts []packageCounts, err error) {
	for _, p := range found {
		built, _, err := graph.Build(p)
		if err != nil {
			return nil, nil, nil, err
		}
		fromNames, err := graph.SecondPass(p)
		if err != nil {
			return nil, nil, nil, err
		}
		counts = append(counts, packageCounts{pkg: p.PkgPath, built: len(built), fromNames: len(fromNames)})

		targets = append(targets, graph.PackageIDOf(p))
		records = append(records, built...)
		records = append(records, fromNames...)
	}
	return targets, graph.Combine(records), counts, nil
}

// packageDir reports the directory a filesystem target names, which is the directory the
// loader has to run in. An import path is left to the loader, because only a module can
// resolve one.
func packageDir(target string) (string, bool) {
	if !strings.HasPrefix(target, "/") && !strings.HasPrefix(target, ".") {
		return "", false
	}
	dir := strings.TrimSuffix(target, "/...")
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir, true
	}
	return "", false
}

// report prints what the load found for a package, including the errors a package can
// carry while still being partly usable.
func report(p *packages.Package) {
	for _, err := range p.Errors {
		fmt.Fprintf(os.Stderr, "%s: %v\n", p.PkgPath, err)
	}
}
