package graph

// Provenance states where the values of a record were read from, so a consumer can tell a value
// that is absent from a value that was never read. The set is closed.
type Provenance int

const (
	// ProvenanceUnstated marks a record that states no provenance. It is the zero value, so an
	// unset provenance is never mistaken for a decided one.
	ProvenanceUnstated Provenance = iota
	// ProvenanceFromSource states that the values were read from a declaration the build read as a
	// target, so the record states the full set its definition requires.
	ProvenanceFromSource
	// ProvenanceFromName states that the values were read from a name a reference carried, from a
	// package the build did not read as a target. A value the record does not state was not read
	// rather than found to be absent.
	ProvenanceFromName
)

func (p Provenance) String() string {
	switch p {
	case ProvenanceFromSource:
		return "read from a declaration the build read as a target"
	case ProvenanceFromName:
		return "read from a name the build did not read"
	default:
		return "unstated"
	}
}
