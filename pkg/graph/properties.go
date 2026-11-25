package graph

// NodeProperties holds properties of a node that never change its semantic type.
// Each property is false when it does not apply to the node.
type NodeProperties struct {
	Exported   bool `json:"exported"`
	Generic    bool `json:"generic"`
	EntryPoint bool `json:"entry_point"`
}
