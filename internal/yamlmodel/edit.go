package yamlmodel

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

// FormatOptions controls how YAML is serialized when saving.
type FormatOptions struct {
	// Indent is the number of spaces per indentation level (2 or 4).
	Indent int
	// SortKeys causes mapping keys to be serialized in alphabetical order.
	SortKeys bool
}

// ScalarState is the editable portion of a scalar yaml.v3 node. Comments,
// anchors, identity, and parent relationships are intentionally not part of a
// value edit and remain attached to the original node.
type ScalarState struct {
	Value string
	Tag   string
	Style yaml.Style
}

// ScalarChange describes one committed scalar edit and is sufficient to
// replay or reverse it without copying the complete document.
type ScalarChange struct {
	NodeID string
	Before ScalarState
	After  ScalarState
}

// SetScalarLiteral parses and applies one YAML scalar literal. The returned
// change can be committed to an undo history or applied in reverse later.
func (file *File) SetScalarLiteral(nodeID, literal string) (ScalarChange, error) {
	node := file.findNode(nodeID)
	if node == nil {
		return ScalarChange{}, fmt.Errorf("node %q not found", nodeID)
	}
	if node.Kind != ScalarNode || node.source == nil {
		return ScalarChange{}, errors.New("only scalar nodes can be edited")
	}

	candidate, err := parseScalarLiteral(literal)
	if err != nil {
		return ScalarChange{}, err
	}
	before := scalarState(node.source)
	after := ScalarState{Value: candidate.Value, Tag: candidate.Tag, Style: candidate.Style}
	if before == after {
		return ScalarChange{NodeID: nodeID, Before: before, After: after}, nil
	}

	applyScalarState(node, after)
	return ScalarChange{NodeID: nodeID, Before: before, After: after}, nil
}

// ApplyScalarChange applies either the after or before side of a previously
// validated scalar change. It is used by document-level undo and redo.
func (file *File) ApplyScalarChange(change ScalarChange, forward bool) error {
	node := file.findNode(change.NodeID)
	if node == nil {
		return fmt.Errorf("node %q not found", change.NodeID)
	}
	if node.Kind != ScalarNode || node.source == nil {
		return errors.New("only scalar nodes can be edited")
	}
	if !forward {
		applyScalarState(node, change.Before)
		return nil
	}
	applyScalarState(node, change.After)
	return nil
}

// Marshal encodes the retained YAML documents, including edits made through
// this package, as a YAML stream.
func (file *File) Marshal() ([]byte, error) {
	return file.MarshalWithOptions(FormatOptions{Indent: 2})
}

// MarshalWithOptions encodes the retained YAML documents using the given
// format options for indentation and key sorting.
func (file *File) MarshalWithOptions(opts FormatOptions) ([]byte, error) {
	if file == nil || len(file.Documents) == 0 {
		return nil, nil
	}

	indent := opts.Indent
	if indent < 2 || indent > 4 {
		indent = 2
	}

	var output bytes.Buffer
	for index, document := range file.Documents {
		if document.Source == nil {
			continue
		}
		if index > 0 {
			output.WriteString("---\n")
		}
		if document.Empty || len(document.Source.Content) == 0 || isEmptyDocument(document.Source.Content) {
			if index == 0 {
				output.WriteString("---\n")
			}
			continue
		}
		if opts.SortKeys {
			sortMappingKeys(document.Source)
		}
		var encoded bytes.Buffer
		encoder := yaml.NewEncoder(&encoded)
		encoder.SetIndent(indent)
		if err := encoder.Encode(document.Source); err != nil {
			_ = encoder.Close()
			return nil, fmt.Errorf("encode document %d: %w", document.Number, err)
		}
		if err := encoder.Close(); err != nil {
			return nil, fmt.Errorf("finish YAML encoding: %w", err)
		}
		output.Write(encoded.Bytes())
	}
	return output.Bytes(), nil
}

// sortMappingKeys recursively sorts mapping node keys in-place.
func sortMappingKeys(node *yaml.Node) {
	if node == nil {
		return
	}
	if node.Kind == yaml.MappingNode {
		// Pair key/value nodes: Content[0]=key0, Content[1]=val0, Content[2]=key1, ...
		type pair struct {
			key   *yaml.Node
			value *yaml.Node
		}
		pairs := make([]pair, 0, len(node.Content)/2)
		for i := 0; i+1 < len(node.Content); i += 2 {
			pairs = append(pairs, pair{key: node.Content[i], value: node.Content[i+1]})
		}
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].key.Value < pairs[j].key.Value
		})
		sorted := make([]*yaml.Node, 0, len(node.Content))
		for _, p := range pairs {
			sorted = append(sorted, p.key, p.value)
		}
		node.Content = sorted
	}
	for _, child := range node.Content {
		sortMappingKeys(child)
	}
}

func parseScalarLiteral(literal string) (*yaml.Node, error) {
	decoder := yaml.NewDecoder(strings.NewReader(literal))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("parse scalar value: %w", err)
	}
	if len(document.Content) != 1 || document.Content[0] == nil {
		return nil, errors.New("value must contain one YAML scalar")
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, errors.New("value must contain one YAML scalar")
		}
		return nil, fmt.Errorf("parse scalar value: %w", err)
	}
	candidate := document.Content[0]
	if candidate.Kind != yaml.ScalarNode {
		return nil, errors.New("value must be a YAML scalar")
	}
	return candidate, nil
}

func scalarState(source *yaml.Node) ScalarState {
	return ScalarState{Value: source.Value, Tag: source.Tag, Style: source.Style}
}

func applyScalarState(node *Node, state ScalarState) {
	node.source.Value = state.Value
	node.source.Tag = state.Tag
	node.source.Style = state.Style
	node.Value = state.Value
	node.Tag = state.Tag
	node.Style = styleOf(state.Style)
	if encoded, err := yaml.Marshal(node.source); err == nil {
		node.YAML = strings.TrimSpace(string(encoded))
	}
}

func (file *File) findNode(nodeID string) *Node {
	if file == nil {
		return nil
	}
	for _, document := range file.Documents {
		if node := findNodeInTree(document.Root, nodeID); node != nil {
			return node
		}
	}
	return nil
}

func findNodeInTree(node *Node, nodeID string) *Node {
	if node == nil {
		return nil
	}
	if node.ID == nodeID {
		return node
	}
	for _, child := range node.Children {
		if found := findNodeInTree(child, nodeID); found != nil {
			return found
		}
	}
	return nil
}
