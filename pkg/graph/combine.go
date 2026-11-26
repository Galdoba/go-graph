package graph

import "sort"

// Combine returns the collection of a run: one entry for each identifier the records state, ordered by
// that identifier. An entity the run both built and named from a reference has one entry, the record read
// from source, because it states the full set its definition requires while a record read from a name states
// only what a name and a type show. Between two records read from a name for the same entity, the entry is
// the one that states a declaration region, so the collection does not depend on which package reached the
// entity first.
func Combine(records []*Node) []*Node {
	entries := make(map[string]*Node, len(records))
	for _, n := range records {
		id := n.ID.String()
		kept, ok := entries[id]
		if !ok || prefers(kept, n) {
			entries[id] = n
		}
	}

	combined := make([]*Node, 0, len(entries))
	for _, n := range entries {
		combined = append(combined, n)
	}
	sort.Slice(combined, func(i, j int) bool { return combined[i].ID.String() < combined[j].ID.String() })
	return combined
}

// prefers reports the record that stands for an entity when two records state the same identifier. A record
// read from source takes precedence over one read from a name, and between two records read from a name the
// one that states a region wins, because a region is a value only some of the loads can state.
func prefers(kept, candidate *Node) bool {
	if kept.Provenance != candidate.Provenance {
		return candidate.Provenance == ProvenanceFromSource
	}
	return kept.Pos.File == "" && candidate.Pos.File != ""
}
