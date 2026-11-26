# Design

## Context

Notes for nodes are rendered in `pkg/vault/note.go` by `body`, which appends the `## Code` heading, then any doc comment, then the ```go fenced code snippet. The doc comment currently renders as `// ...` lines before the fence. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Render the doc comment inside the ```go fence, immediately before the code snippet lines, so the comment is visually grouped with the code it documents.

**Non-Goals:**
- Change the format of the code snippet, the `# Name` heading, or the `## Related` section.
- Change which records get a doc comment or when the code block is written.

## Decisions

- **Move the doc-comment lines inside the fence.** In `body`, the doc comment is generated before `lines = append(lines, "```go")`. The change moves that generation to after the ```go marker and before the loop that appends the code snippet lines. This is the minimal structural change that places the comment inside the block.
- **No change to `docComment` or `codeSnippet`.** Both helpers return the same values; only their ordering in the output changes.
- **Tests need no behavioral changes.** `note_test.go` asserts doc-comment presence with `strings.Contains`, which is position-independent, so the existing scenarios pass unchanged.

## Risks / Trade-offs

- [Markdown rendering] The doc comment now appears inside a code fence, so it renders as code rather than prose. This is intentional and matches the source-file grouping.

## Migration Plan

The change is a note-format adjustment only. No external migration is required; notes regenerate on the next rebuild.

## Open Questions

None.
