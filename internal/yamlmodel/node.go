// Package yamlmodel contains the format-independent tree used by the viewer.
package yamlmodel

// NodeKind identifies the YAML construct represented by a Node.
type NodeKind string

const (
	MappingNode  NodeKind = "mapping"
	SequenceNode NodeKind = "sequence"
	ScalarNode   NodeKind = "scalar"
	AliasNode    NodeKind = "alias"
)

// Comments preserves all comments attached to a YAML node.
type Comments struct {
	Head string
	Line string
	Foot string
}

// Node is a generic YAML tree node. Key is populated for a mapping value and
// Index is populated for a sequence item. The original value and metadata are
// intentionally kept separate from display formatting.
type Node struct {
	ID       string
	Kind     NodeKind
	Key      string
	KeySet   bool
	Index    int
	Value    string
	YAML     string
	Path     string
	Tag      string
	Style    string
	Anchor   string
	Alias    string
	Line     int
	Column   int
	Comments Comments
	Children []*Node

	// KeyYAML contains the source-like representation used when a mapping key
	// is not a string. It is empty for sequence items and document roots.
	KeyYAML string
	// KeyTag keeps typed mapping keys distinct when their rendered text is the
	// same, for example the integer key 1 and the string key "1".
	KeyTag string
	// Duplicate marks a mapping value whose key occurs more than once in its
	// parent mapping. Duplicate entries remain separate nodes.
	Duplicate bool
	// Parent is useful to UI consumers and is not used as part of identity.
	Parent *Node
}

// IsContainer reports whether the node can have children.
func (n *Node) IsContainer() bool {
	return n != nil && (n.Kind == MappingNode || n.Kind == SequenceNode)
}

// ChildCount returns the number of direct children without exposing nil checks
// to presentation code.
func (n *Node) ChildCount() int {
	if n == nil {
		return 0
	}
	return len(n.Children)
}
