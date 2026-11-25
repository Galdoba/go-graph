# Design

## Context

The vault package generates markdown note files from graph nodes. The current `body()` function produces a prose paragraph. A separate `appendLinks()` pass in `cmd/graphnodes/main.go` appends a `## links` section after the file is written.

## Goals / Non-Goals

**Goals:**
- Embed code snippet, doc comment, and related links directly in the note body on first write.
- Remove the separate `appendLinks()` pass.

**Non-Goals:**
- Change the frontmatter format or fields.
- Add new link types or change the link rules.
- Support outgoing links (only incoming connections are listed).

## Decisions

1. **`Content(n, nodes)` accepts the full node collection.**
   - Rationale: `Rebuild()` already has `nodes []*graph.Node`; passing it through avoids a second load or a global registry.
   - Alternative: pass nodes only when needed. Rejected — the body always needs them for incoming-link computation.

2. **Incoming links are computed by inverting `link.Pairs`.**
   - For each node, iterate all pairs and collect targets where this node is the source.
   - Rationale: `link.Pairs()` already iterates all pairs; inverting it is O(n²) but the node count per package is small.
   - Alternative: build an incoming adjacency map. Rejected — adds complexity for marginal gain.

3. **Code snippet is extracted from the source file using `Pos.File` and `Range`.**
   - Read the file, split by lines, slice `[StartLine-1 : EndLine]`, join with newlines.
   - Rationale: preserves exact source formatting; `n.Decl` AST rendering would not match the original whitespace.

4. **Doc comment is taken from `n.Decl.Doc.Text()` when present.**
   - `n.Decl` is `ast.Decl`; `*ast.GenDecl` and `*ast.FuncDecl` both have `Doc *ast.CommentGroup`.
   - Only included when non-empty, keeping the section clean for nodes without docs.

5. **`## Related` is always present.** When there are no incoming connections it renders "None detected"; otherwise it lists wikilinks.
   - Mirrors the existing `appendLinks()` format for consistency.

## Risks / Trade-offs

- [File read latency] Reading the source file for every node adds I/O. Mitigation: the node count per package is small (typically <50), and the file is likely cached.
- [Stale code if source changes between build and vault write] The source is read at render time, not build time. If the file changes between `graph.Build()` and `vault.Render()`, the snippet may not match the build. Mitigation: in practice, build and render happen in the same `main()` invocation with no intervening edits.

## Migration Plan

No migration needed. The change is internal to the tool: `graphnodes` produces different body content on the next run. Existing note files are removed by `Rebuild()` before new ones are written.

## Open Questions

None.
