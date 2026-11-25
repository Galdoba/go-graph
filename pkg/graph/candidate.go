package graph

import "go/types"

// Reasons a candidate is excluded from the node set. The exclusion rules are part of
// the definition, so a rejected entity is reported with its reason instead of being
// dropped silently.
const (
	ReasonNoDeclarationSite        = "no single declaration site"
	ReasonBlankIdentifier          = "blank identifier names no entity"
	ReasonOutsideClosedSet         = "kind outside the closed set"
	ReasonPredeclaredIdentifier    = "predeclared identifier, never created by a package-level declaration"
	ReasonCreatedInsideDeclaration = "created inside a declaration body"
	ReasonTypeParameter            = "generic type parameter, not a package-level declaration"
	ReasonPartOfDeclarationSpec    = "part of a declaration spec rather than a declaration"
	ReasonProducedByLookup         = "produced by a lookup rather than by a declaration"
	ReasonPackageIsNamespace       = "the package groups nodes and is not a node"
	ReasonForeignDeclaration       = "declared by another package and owned by that package graph"
	ReasonAliasIsName              = "alias is a name, not an entity"
)

// Candidate is an entity the enumeration looked at. Recorded is true only when the
// entity satisfies the definition of a node, otherwise Reason says why it is not one.
type Candidate struct {
	Name     string `json:"name"`
	Kind     Kind   `json:"kind"`
	Recorded bool   `json:"recorded"`
	Reason   string `json:"reason"`
}

// Classify reports whether an object found while enumerating a package is a node. The
// test is the definition itself: a node is a named entity created by a package-level
// declaration. Anything inside a declaration body, any alias, any predeclared
// identifier and anything outside the closed set of kinds is reported as excluded.
func Classify(pkg *types.Package, obj types.Object) Candidate {
	if obj == nil {
		return Candidate{Reason: ReasonNoDeclarationSite}
	}
	c := Candidate{Name: obj.Name(), Kind: kindOf(obj)}
	if obj.Name() == "_" {
		c.Reason = ReasonBlankIdentifier
		return c
	}
	if _, ok := obj.(*types.PkgName); ok {
		c.Reason = ReasonPackageIsNamespace
		return c
	}
	// Universe objects have no package and no declaration position, so they cannot
	// satisfy the node record. A package that declares its own entity named after a
	// predeclared identifier is a node, and references resolve to that node.
	if obj.Pkg() == nil {
		c.Reason = ReasonPredeclaredIdentifier
		return c
	}
	if obj.Pkg() != pkg {
		c.Reason = ReasonForeignDeclaration
		return c
	}
	if c.Kind == KindUnknown {
		c.Reason = ReasonOutsideClosedSet
		return c
	}
	if tn, ok := obj.(*types.TypeName); ok {
		if tn.IsAlias() {
			c.Reason = ReasonAliasIsName
			return c
		}
		if _, isParam := tn.Type().(*types.TypeParam); isParam {
			c.Reason = ReasonTypeParameter
			return c
		}
	}
	if v, ok := obj.(*types.Var); ok {
		switch v.Kind() {
		case types.PackageVar:
			c.Recorded = true
		case types.FieldVar:
			c.Reason = ReasonPartOfDeclarationSpec
		default:
			c.Reason = ReasonCreatedInsideDeclaration
		}
		return c
	}
	c.Recorded = true
	return c
}
