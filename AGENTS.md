# AGENTS.md

## Working

- Use the gopls MCP tools wherever they can answer the task: workspace and package API summaries, symbol search, symbol references, call hierarchy, diagnostics, and semantic rename. Fall back to grep or file reading only when the MCP cannot do it.
- Prefer smaller edits. Edit files with one change at the time. It is better to Edit file 2 times adding new functions with 1 functions each, than try to Edit once with 2 functions at once.
- Avoid rewriting whole file!

## Language

- All documentation is written in English: specs, design documents, README, godoc comments, commit messages, and PR text. Code identifiers follow Go conventions.

## Subagents

- Subagents are not allowed. Do not dispatch or delegate work to subagents; do all investigation and editing in the current session.
