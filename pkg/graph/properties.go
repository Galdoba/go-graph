package graph

import "go/types"

// NodeProperties holds properties of a node that never change its semantic type.
// Each property is false when it does not apply to the node.
type NodeProperties struct {
	Exported   bool `json:"exported"`
	Generic    bool `json:"generic"`
	EntryPoint bool `json:"entry_point"`
}

// propertiesOf reports the properties of an entity from the entity itself, which is the same
// reading for a record the build produced and for a record produced from a name.
func propertiesOf(pkg PackageID, obj types.Object) NodeProperties {
	var props NodeProperties
	if obj.Exported() {
		props.Exported = true
	}
	if tn, ok := obj.(*types.TypeName); ok {
		if named, ok := tn.Type().(*types.Named); ok && named.TypeParams() != nil && named.TypeParams().Len() > 0 {
			props.Generic = true
		}
	}
	if f, ok := obj.(*types.Func); ok {
		if f.Signature().TypeParams() != nil && f.Signature().TypeParams().Len() > 0 {
			props.Generic = true
		}
		if pkg.Name == "main" && (obj.Name() == "main" || obj.Name() == "init") {
			props.EntryPoint = true
		}
	}
	return props
}
