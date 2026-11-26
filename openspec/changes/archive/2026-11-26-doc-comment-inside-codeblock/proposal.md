# Proposal

## Why

Notes generated for nodes include a declaration doc comment that currently renders as `// ...` lines outside the ```go fence, separating it visually from the code it documents. Placing the doc comment inside the code block keeps the comment with the code it describes, matching how the source file itself groups the comment with the declaration.

## What Changes

- The `## Code` section of a generated note renders the doc comment inside the ```go fence, before the code snippet, instead of as separate `// ...` lines above the fence.
- The vault spec scenarios that state "the doc comment above the code snippet" and "the comment above the declaration" are updated to state the doc comment is inside the code block, before the code snippet.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `vault`: the `## Code` section includes the doc comment inside the code block, before the code snippet, and the scenarios "Struct node with doc comment and related method" and "A foreign note is read by these rules" are updated accordingly.

## Impact

- `pkg/vault/note.go`: the `body` function moves the doc-comment lines to render after the ```go marker and before the code snippet lines.
- `openspec/specs/vault/spec.md`: the two scenarios above are updated.
- Tests in `pkg/vault/note_test.go`: no behavioral changes required, since they assert on presence via `strings.Contains`.
