// Command graphnodes finds the nodes of one package, prints them to stderr, reads
// every record back through a second, independent load, and writes the note file each record
// maps to in the vault, which is the round trip that shows a record carries enough information
// to locate its declaration on its own and to name the file it belongs to.
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
	for _, p := range reloaded {
		resolved[p.PkgPath] = p
	}

	failed := 0
	for _, p := range found {
		report(p)

		nodes, _, err := graph.Build(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%s: %d nodes\n", p.PkgPath, len(nodes))
		for _, n := range nodes {
			fmt.Fprintf(os.Stderr, "  %s kind=%v semantic=%v receiver=%q parent=%v exported=%v generic=%v entry_point=%v file=%s line=%d column=%d range=%d-%d attributes=%d\n",
				n.ID, n.Kind, n.SemanticType, n.Receiver, n.Parent, n.Properties.Exported, n.Properties.Generic, n.Properties.EntryPoint,
				n.Pos.File, n.Pos.Line, n.Pos.Column, n.Range.StartLine, n.Range.EndLine, len(n.Attributes))
		}

		for _, n := range nodes {
			other := resolved[n.Package.PackagePath]
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

		notes, err := writeNotes(nodes)
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
	}

	if failed > 0 {
		os.Exit(1)
	}
}

// writeNotes runs the action over one package, states what it removed, and lists the files it
// wrote so each can be read back and confirmed against the code.
func writeNotes(nodes []*graph.Node) ([]string, error) {
	if len(nodes) == 0 {
		return nil, nil
	}

	report, err := vault.Rebuild(vault.Root, nodes[0].Package, nodes)
	if err != nil {
		return nil, err
	}
	for _, file := range report.Removed {
		fmt.Fprintf(os.Stderr, "  removed %s\n", file)
	}
	for _, file := range report.Unreadable {
		fmt.Fprintf(os.Stderr, "  unreadable %s\n", file)
	}

	files := make([]string, 0, len(nodes))
	for _, n := range nodes {
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
