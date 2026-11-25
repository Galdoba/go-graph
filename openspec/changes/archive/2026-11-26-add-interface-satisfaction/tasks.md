# Tasks

## 1. The method set on the record

- [x] 1.1 Add the attribute kind `method` to `pkg/graph/attribute.go`, including its `String`, and verify a test reads an attribute of that kind back.
- [x] 1.2 Record each method declared on a package-level type as an attribute of the node that owns it, carrying the name, the signature rendered without the receiver and the reference to the method node, and verify in `pkg/graph/build_test.go` that `Base` carries a method attribute for `Read` with the value `func() error` pointing at `Base.Read`.

## 2. The rule

- [x] 2.1 Add `type Adder interface { Add5() int }` to `testdata/main.go` and verify the test ground answers the pair `Adder` to `Struct1` as existing, which is the pointer receiver case read on a real declaration.
- [x] 2.2 Add `stubInterface` and a structure stub that carries method attributes to `pkg/link/testhelper_test.go`, used only for the outcome the test ground cannot produce, and verify a test can build an interface with method specs and a structure with method attributes.
- [x] 2.3 Add `answerSatisfiedInterface` and the `satisfiedInterface` rule in `pkg/link/link.go`, registered as the third entry in `rules`: an unresolved member names the rule, a source that is not an interface or a target that is not a structure answers no, and a pair answers yes only when every method spec of the interface matches a method attribute on the structure by name, receiver and signature. Verify that `Reader` to `Base` answers yes and `Reader` to `Thing` answers no against the test ground.
- [x] 2.4 Make `answerReferentNode` skip attributes of kind `method`, and verify that the pair from `Base` to `Base.Read` is answered by the method rule rather than by the referent rule, so the existing pair outcome does not change.
- [x] 2.5 Add tests for the rule in `pkg/link/link_test.go`, each answered on the test ground: satisfied (`Reader` to `Base`), missing method (`Reader` to `Struct1`), signature mismatch (`Reader` to `Thing`), pointer receiver (`Adder` to `Struct1`), reachable only through an embedding (`Reader` to `Outer` answers no), empty interface (`EmptyInterface` to `Base` answers no), and an undecided source that names the new rule on a hand-built record. Verify each test passes.

## 3. Integration check

- [x] 3.1 Verify that `Coverage` reports no uncovered rule against the test ground, so `satisfiedInterface` is answered yes by `Reader` to `Base`.
- [x] 3.2 Run all tests in `pkg/graph` and `pkg/link` and verify that adding the attribute kind and the rule changes no existing outcome.

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied.
- Verify the archived result.
