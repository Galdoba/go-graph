# Tasks

## 1. Receiver normalization through one decision point

- [ ] 1.1 Start from a committed green baseline and record `go build ./... && go vet ./... && go test ./...` as the result every later step is compared against.
- [ ] 1.2 Resolve the receiver name through the shared named-type helper instead of repeating the pointer strip and alias resolution, preferring a gopls inline or extract over a hand edit, and verify the suite is green with the identity tests printing the same canonical forms.
- [ ] 1.3 Check that both paths the single helper now serves are exercised by existing tests, a pointer receiver and a receiver written with an alias name, and verify `go test ./pkg/graph/... -cover` reports no coverage drop for the file.

## 2. Enumeration helper that never reads an unused argument

- [ ] 2.1 List the call sites of the helper that classifies one entity with gopls references, confirm the reason argument is empty at every call site, and verify the count matches the three enumeration paths.
- [ ] 2.2 Drop the reason argument and the branch that only runs when a caller passes one, and verify the suite is green and the excluded candidate set the test ground reports is unchanged.

## 3. Test helpers in one place

- [ ] 3.1 Move the attribute lookup and the fixture line lookup helpers into the shared test helper file, and verify the suite is green with each helper declared once in the package.
- [ ] 3.2 Add a shared helper that returns the node a test looks for or fails, convert the repeated guards that already fail on a missing node to it, and verify the suite is green and one run shows the failure message still names the expected node.

## 4. Record types and the functions that fill them

- [ ] 4.1 Move the position and line range builders next to the record types they fill, and verify the suite is green with no change to any recorded coordinate.
- [ ] 4.2 Run `go test ./...` to confirm `pkg/vault`, `pkg/link` and `cmd/graphnodes` still resolve nodes and print the same note names, which is the check that nothing in this change reached outside the package.

## Workflow follow-up

- Commit each group on its own as a structural commit, never mixed with a behavior change, and commit the moment a step goes green.
- If a step goes red, revert to the last green commit rather than patching forward.
- Archive the change after the project's review requirements are satisfied; the deferred work (splitting the enumeration state by responsibility, the shared exclusion rule, the record field that holds a declaration reference) is a separate change and is not part of this one.
