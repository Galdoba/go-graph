package graph

// AttributeKind names what an attribute records. Attributes hold the information about
// entities that are not nodes, so that a relation can later be built from them.
type AttributeKind int

const (
	// AttrUnknown is the sentinel for an attribute whose kind is not known.
	AttrUnknown    AttributeKind = iota
	AttrField                    // a struct field of a type node
	AttrMethodSpec               // a method spec of an interface type node
	AttrEmbeds                   // an embedded field, which points at the type node that owns the method
	AttrAlias                    // a name used at a use-site that is not the declaration name
	AttrUse                      // an entity a declaration refers to
)

func (a AttributeKind) String() string {
	switch a {
	case AttrField:
		return "field"
	case AttrMethodSpec:
		return "method spec"
	case AttrEmbeds:
		return "embeds"
	case AttrAlias:
		return "alias"
	case AttrUse:
		return "use"
	default:
		return "unknown"
	}
}

// Attribute records information about an entity that is not a node on the node that
// owns it. Ref is the identifier of the node an attribute points at, and it is empty
// when the referent is not a node, in which case Value carries it, as the predeclared
// type int does for a field.
type Attribute struct {
	Kind  AttributeKind `json:"kind"`
	Name  string        `json:"name"`
	Value string        `json:"value"`
	Ref   ID            `json:"ref"`
}

func (a Attribute) String() string {
	if a.Ref.Name == "" {
		return a.Kind.String() + " " + a.Name + " = " + a.Value
	}
	return a.Kind.String() + " " + a.Name + " -> " + a.Ref.String()
}

func attributeOrder(a, b Attribute) bool {
	if a.Kind != b.Kind {
		return a.Kind < b.Kind
	}
	if a.Name != b.Name {
		return a.Name < b.Name
	}
	return a.Ref.String() < b.Ref.String()
}
