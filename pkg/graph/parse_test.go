package graph

import "testing"

// TestAWrittenValueReadsBackIntoItsClosedSet checks the inverse of `String`, which is the step a
// note is navigated through to reach the node it records: membership is decided by the set that wrote
// the value, so every value of the set reads back to itself and a value outside it is refused. A value
// that belongs to the other set is refused as well, because the two sets share no members, and so is a
// value that belongs to neither set.
func TestAWrittenValueReadsBackIntoItsClosedSet(t *testing.T) {
	for k := KindUnknown; k <= KindConst; k++ {
		got, ok := ParseKind(k.String())
		if !ok || got != k {
			t.Errorf("kind %q reads back as %v, want %v", k.String(), got, k)
		}
	}
	if _, ok := ParseKind("method"); ok {
		t.Error("ParseKind accepted method, which is a value of the semantic set and not of the kind set")
	}
	if _, ok := ParseKind("thing"); ok {
		t.Error("ParseKind accepted thing, which is a value of neither set")
	}

	for s := SemanticUnresolved; s <= SemanticVariable; s++ {
		got, ok := ParseSemanticType(s.String())
		if !ok || got != s {
			t.Errorf("semantic type %q reads back as %v, want %v", s.String(), got, s)
		}
	}
	if _, ok := ParseSemanticType("type"); ok {
		t.Error("ParseSemanticType accepted type, which is a value of the kind set and not of the semantic set")
	}
	if _, ok := ParseSemanticType("thing"); ok {
		t.Error("ParseSemanticType accepted thing, which is a value of neither set")
	}
}
