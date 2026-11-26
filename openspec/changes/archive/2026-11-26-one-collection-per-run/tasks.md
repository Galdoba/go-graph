# Tasks

## 1. One collection for the run

- [x] 1.1 Add the merge that returns one entry per identifier for the records a run produced, keeping the record read from source when both passes produced one, ordered by the identifier, and verify with a test in `pkg/graph` that a run over two fixtures that name each other yields one record for the shared entity and that it is the one read from source
- [x] 1.2 Add a test that the same records assembled in a different order produce the same collection, and document the rule in `docs/node-model.md` next to the second pass section

## 2. The rules read the collection of the run

- [x] 2.1 Add a test in `pkg/link` that a pair whose two members are records of two packages the run read as targets is answered from one collection, and document in `docs/node-links.md` that the collection the rules read is the collection of the run rather than of one package

## 3. Ownership stated for the run

- [x] 3.1 Change the rebuild action so it owns the folders of the packages the run read as targets, keeping the rule that a folder only records produced from reference map to loses only the files this run wrote, and verify with a test that a run over two packages removes the marked files in both folders and nothing in a folder it does not own
- [x] 3.2 Document the ownership for the whole run in `docs/vault.md`

## 4. The Related section answered from the run

- [x] 4.1 Pin the direction with tests: a record read from source lists the pairs it is the source of, and a record read from a name lists the pairs that name it as their target, verified by reading the `## Related` section of the rendered file
- [x] 4.2 Rewrite the run test in `pkg/vault` so it runs over two packages of the repository at once and the note of the record that carries the reference shows the connection it points at, verified by reading the file back
- [x] 4.3 Document in `docs/vault.md` that the `## Related` section is answered from the records of the run, that its direction depends on how the record was read, and that a folder two runs write shows the last writer's view

## 5. Wire the command to the run

- [x] 5.1 Make `cmd/graphnodes` collect the records of every package it read as a target, merge them once, and call the rebuild once for the run, verified by `go build ./...` and by the counts the command prints for a run that reads two packages
- [x] 5.2 Run `go build ./...`, `go vet ./...` and `go test ./...` and confirm the whole suite is green with the new tests in place

## Workflow follow-up

- Sync the three delta specs into the main specs when the implementation is verified.
- Archive the change after the review requirements are satisfied.
