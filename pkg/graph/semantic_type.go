package graph

import "go/types"

// SemanticType is the semantic type of a node. The set is closed, so an entity that
// is not one of these types is never a node. unresolved is the zero value, so an
// unset semantic type is never mistaken for a decided one.
type SemanticType int

const (
	// SemanticUnresolved marks a node that no rule classifies. It is the zero value,
	// so an unset semantic type is never mistaken for a decided one.
	SemanticUnresolved SemanticType = iota
	SemanticStructure
	SemanticInterface
	SemanticFunction
	SemanticMethod
	SemanticConstant
	SemanticVariable
)

func (s SemanticType) String() string {
	switch s {
	case SemanticStructure:
		return "structure"
	case SemanticInterface:
		return "interface"
	case SemanticFunction:
		return "function"
	case SemanticMethod:
		return "method"
	case SemanticConstant:
		return "constant"
	case SemanticVariable:
		return "variable"
	default:
		return "unresolved"
	}
}

// ParseSemanticType reports the semantic type a written value names, which is how a record read
// from a file is compared against a build. Membership is decided by the closed set itself, so a
// value added to the set extends the range it is read from.
func ParseSemanticType(name string) (SemanticType, bool) {
	for s := SemanticUnresolved; s <= SemanticVariable; s++ {
		if s.String() == name {
			return s, true
		}
	}
	return SemanticUnresolved, false
}

// semanticTypeOf reports the semantic type of a language object. The value is
// derived from the node's own declaration and from what the language fixes for
// every package.
func semanticTypeOf(obj types.Object, kind Kind, receiver string) SemanticType {
	switch kind {
	case KindType:
		tn, ok := obj.(*types.TypeName)
		if !ok {
			return SemanticUnresolved
		}
		named, ok := tn.Type().(*types.Named)
		if !ok {
			return SemanticUnresolved
		}
		_, isInterface := named.Underlying().(*types.Interface)
		if isInterface {
			return SemanticInterface
		}
		return SemanticStructure
	case KindFunc:
		if receiver != "" {
			return SemanticMethod
		}
		return SemanticFunction
	case KindVar:
		return SemanticVariable
	case KindConst:
		return SemanticConstant
	default:
		return SemanticUnresolved
	}
}
