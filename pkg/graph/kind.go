package graph

import "go/types"

// Kind is the kind of a node. The set is closed, so an entity that is not one of
// these kinds is never a node.
type Kind int

const (
	// KindUnknown marks a candidate outside the closed set. It is the zero value, so
	// an unset kind is never mistaken for a node.
	KindUnknown Kind = iota
	KindType
	KindFunc
	KindVar
	KindConst
)

func (k Kind) String() string {
	switch k {
	case KindType:
		return "type"
	case KindFunc:
		return "function"
	case KindVar:
		return "variable"
	case KindConst:
		return "constant"
	default:
		return "unknown"
	}
}

// ParseKind reports the kind a written value names, which is how a record read from a file is
// compared against a build. Membership is decided by the closed set itself, so a value added
// to the set extends the range it is read from.
func ParseKind(name string) (Kind, bool) {
	for k := KindUnknown; k <= KindConst; k++ {
		if k.String() == name {
			return k, true
		}
	}
	return KindUnknown, false
}

// kindOf reports the node kind of a language object. Objects outside the closed set,
// such as packages, labels and predeclared functions, report KindUnknown.
func kindOf(obj types.Object) Kind {
	switch obj.(type) {
	case *types.TypeName:
		return KindType
	case *types.Func:
		return KindFunc
	case *types.Var:
		return KindVar
	case *types.Const:
		return KindConst
	default:
		return KindUnknown
	}
}
