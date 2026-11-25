# Tasks

## 1. Update note body rendering

- [x] 1.1 Refactor `pkg/vault/note.go`: change `Content(n *graph.Node)` to `Content(n *graph.Node, nodes []*graph.Node)`, add `body(n *graph.Node, nodes []*graph.Node)`, and implement the new structured body format with `# Name`, `## Code`, and `## Related` sections (always present — "None detected" when no incoming links) — verify by running `go test ./pkg/vault/`
- [x] 1.2 Add helper `codeSnippet(n *graph.Node) string` that reads the source file and returns the lines in `[StartLine-1:EndLine]` — verify with a unit test in `note_test.go`
- [x] 1.3 Add helper `incomingLinks(n *graph.Node, nodes []*graph.Node) []string` that computes incoming connections and formats them as Obsidian wikilinks — verify with a unit test in `note_test.go`
- [x] 1.4 Integrate doc comment extraction: prepend `n.Decl.Doc.Text()` to the code snippet when present — verify by checking the rendered body for a node with a doc comment

## 2. Wire the node collection through rebuild

- [x] 2.1 Update `pkg/vault/rebuild.go`: pass `nodes` to `Content(n, nodes)` — verify by running `go test ./pkg/vault/`

## 3. Remove the separate appendLinks pass

- [x] 3.1 Remove `appendLinks()` function and its call from `cmd/graphnodes/main.go` — verify by running `go build ./cmd/graphnodes/` and confirming the command still produces correct output

## 4. Integration verification

- [x] 4.1 Run `go build ./...` and `go test ./...` — verify all tests pass
- [x] 4.2 Run `graphnodes` against the testground package and inspect a generated note file — verify it matches the `examples/Thing.md` format
