# Spec Delta

## ADDED Requirements

### Requirement: The domain holds a pair with a record produced from a reference
A pair SHALL be answered when one of its members is a record the second pass produced, because that member is
a record the rules can read, and a pair whose members come from two packages SHALL be answered by the same
rules as a pair whose members come from one package. The exported filter SHALL reach only records produced from
a reference, so a pair inside the package the build read as a target is answered whatever the exported value of
its target is.

#### Scenario: An unexported source reaches an exported node of another package
- **WHEN** a node of the package the build read as a target is not exported and carries a use whose
  reference names a record the second pass produced
- **THEN** the pair is answered as existing, because the filter reaches only records produced from a
  reference

#### Scenario: An unexported target inside the package
- **WHEN** a source of the package the build read as a target carries a use that names an unexported node of
  that same package
- **THEN** the pair is answered as it is answered today, because no record of that package was produced from
  a reference

#### Scenario: A member comes from a package the second pass produced no record for
- **WHEN** a reference names a node for which no record was produced
- **THEN** no pair is answered for it, because a member of a pair must be a record

#### Scenario: A pair that needs two records from the same foreign package
- **WHEN** the second pass produced a record for a method declared on a type of another package, and the code
  that produced that record named no record for the type the method is declared on
- **THEN** the relation from that type to its method is not answered, because the pair would need a record for
  the type as its other member

### Requirement: A rule cannot answer for a value a record does not carry
A rule that needs a value a record produced from a reference does not carry SHALL report that it cannot
answer for that pair, and SHALL NOT answer no, because the record does not claim that the value is absent. A
rule that reads only what a record produced from a reference can carry SHALL answer for that pair as it does
for any other record.

#### Scenario: The rule that reads uses meets a record that states no uses
- **WHEN** a pair whose source is a record produced from a reference is answered by the rule that reads the
  uses the source carries
- **THEN** that rule reports that it cannot answer, and the outcome for the pair names it when no rule answered
  yes

#### Scenario: A rule that reads shape answers for a record produced from a reference
- **WHEN** a record produced from a reference carries the method specs of the interface it states, and a
  structure of the package the build read as a target records a method for each of them
- **THEN** the pair is answered as existing, because the rule reads only what the two members carry

## MODIFIED Requirements

### Requirement: Every rule is exercised by the test ground
The test ground SHALL produce a pair for every rule in the rule set, so that each rule is answered on real
code. A rule answered yes by no entity of the test ground SHALL be reported as not covered. The test ground
SHALL include a declaration that references a node of a package the build reads from type information only,
so that a rule is answered on a pair whose members come from two packages.

#### Scenario: A rule has no pair in the test ground
- **WHEN** a rule in the rule set is answered yes by no pair the test ground produces
- **THEN** the run against the test ground reports that rule as not covered

#### Scenario: A pair whose members come from two packages
- **WHEN** the test ground declares a package that calls a method declared on a type of another package the
  build reads from type information only
- **THEN** the rule that reads references is answered yes for that pair, and the run reports it as covered
