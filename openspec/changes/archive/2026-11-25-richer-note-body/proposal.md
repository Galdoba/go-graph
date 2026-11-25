# Proposal

## Why

The current note body is a prose paragraph that describes the node in natural language. This format is hard to scan, provides no inline code reference, and requires a second pass (`appendLinks`) to add connections after the file is already written.

## What Changes

- Replace the prose body with a structured format: `# Name` heading, `## Code` section with the source code snippet and optional doc comment, `## Related` section with Obsidian wikilinks to incoming connections (or "None detected" when there are none).
- Compute incoming links during note generation so the body is complete on first write — no separate `appendLinks` pass needed.
- Extract the code snippet from the source file using the node's `Pos.File` and `Range` fields; prepend the declaration's doc comment (`ast.CommentGroup.Text()`) when present.

## Capabilities

### Modified Capabilities

- `vault`: requirement "The note states what its node is" — the body format changes from prose to structured sections with code and related links.

## Impact

- `pkg/vault/note.go` — `Content()` and `body()` accept `nodes` parameter; new helpers for code extraction and incoming-link computation.
- `pkg/vault/rebuild.go` — passes `nodes` to `Content()`.
- `cmd/graphnodes/main.go` — removes `appendLinks()` and its call.
