package yamlmodel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Decode parses a YAML stream without converting it to a fixed application
// struct. All documents and node metadata are retained in the returned model.
func Decode(data []byte) (*File, error) {
	return DecodeReader(bytes.NewReader(data))
}

// DecodeReader is the reader variant of Decode.
func DecodeReader(reader io.Reader) (*File, error) {
	decoder := yaml.NewDecoder(reader)
	file := &File{}
	for documentNumber := 1; ; documentNumber++ {
		var document yaml.Node
		err := decoder.Decode(&document)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		decoded := Document{Number: documentNumber}
		decoded.Source = &document
		if len(document.Content) == 0 || isEmptyDocument(document.Content) {
			decoded.Empty = true
		} else {
			builder := nodeBuilder{documentNumber: documentNumber, nextID: 1}
			decoded.Root = builder.build(document.Content[0], "/", "", -1, nil)
		}
		file.Documents = append(file.Documents, decoded)
	}
	file.Empty = len(file.Documents) == 0
	return file, nil
}

func isEmptyDocument(documentContent []*yaml.Node) bool {
	if len(documentContent) != 1 || documentContent[0] == nil {
		return false
	}
	node := documentContent[0]
	return node.Kind == yaml.ScalarNode && node.Tag == "!!null" && node.Value == ""
}

type nodeBuilder struct {
	documentNumber int
	nextID         int
}

func (b *nodeBuilder) build(source *yaml.Node, path, key string, index int, parent *Node) *Node {
	if source == nil {
		return nil
	}
	node := &Node{
		ID:       fmt.Sprintf("doc-%d-node-%d", b.documentNumber, b.nextID),
		Kind:     kindOf(source),
		Key:      key,
		Index:    index,
		Value:    source.Value,
		Path:     path,
		Tag:      source.Tag,
		Style:    styleOf(source.Style),
		Anchor:   source.Anchor,
		Line:     source.Line,
		Column:   source.Column,
		Comments: Comments{Head: source.HeadComment, Line: source.LineComment, Foot: source.FootComment},
		Parent:   parent,
		source:   source,
	}
	if encoded, err := yaml.Marshal(source); err == nil {
		node.YAML = strings.TrimSpace(string(encoded))
	}
	b.nextID++
	if source.Kind == yaml.AliasNode {
		node.Alias = source.Value
		node.Value = ""
		return node
	}

	switch source.Kind {
	case yaml.MappingNode:
		seen := make(map[string]int)
		for pair := 0; pair+1 < len(source.Content); pair += 2 {
			keyNode, valueNode := source.Content[pair], source.Content[pair+1]
			rawKey := mappingKey(keyNode)
			keyYAML := mappingKeyYAML(keyNode)
			identity := keyNode.Tag + "\x00" + keyYAML
			seen[identity]++
			child := b.build(valueNode, AppendPath(path, rawKey), rawKey, -1, node)
			child.KeySet = true
			child.KeyYAML = keyYAML
			child.KeyTag = keyNode.Tag
			child.Duplicate = seen[identity] > 1
			if seen[identity] > 1 {
				for _, existing := range node.Children {
					if existing.KeyYAML == keyYAML && existing.KeyTag == keyNode.Tag {
						existing.Duplicate = true
					}
				}
			}
			node.Children = append(node.Children, child)
		}
	case yaml.SequenceNode:
		for childIndex, valueNode := range source.Content {
			node.Children = append(node.Children, b.build(valueNode, AppendPath(path, strconv.Itoa(childIndex)), "", childIndex, node))
		}
	}
	return node
}

func kindOf(source *yaml.Node) NodeKind {
	switch source.Kind {
	case yaml.MappingNode:
		return MappingNode
	case yaml.SequenceNode:
		return SequenceNode
	case yaml.AliasNode:
		return AliasNode
	default:
		return ScalarNode
	}
}

func styleOf(style yaml.Style) string {
	styles := make([]string, 0, 2)
	if style&yaml.TaggedStyle != 0 {
		styles = append(styles, "tagged")
	}
	if style&yaml.DoubleQuotedStyle != 0 {
		styles = append(styles, "double-quoted")
	}
	if style&yaml.SingleQuotedStyle != 0 {
		styles = append(styles, "single-quoted")
	}
	if style&yaml.LiteralStyle != 0 {
		styles = append(styles, "literal")
	}
	if style&yaml.FoldedStyle != 0 {
		styles = append(styles, "folded")
	}
	if style&yaml.FlowStyle != 0 {
		styles = append(styles, "flow")
	}
	return strings.Join(styles, ", ")
}

func mappingKey(keyNode *yaml.Node) string {
	if keyNode.Kind == yaml.ScalarNode {
		if keyNode.Value == "" {
			if keyNode.Tag == "!!null" {
				return "null"
			}
			if keyNode.Tag == "!!str" {
				return "\"\""
			}
		}
		return keyNode.Value
	}
	return mappingKeyYAML(keyNode)
}

func mappingKeyYAML(keyNode *yaml.Node) string {
	if keyNode.Kind == yaml.ScalarNode {
		return keyNode.Value
	}
	encoded, err := yaml.Marshal(keyNode)
	if err == nil {
		return strings.TrimSpace(string(encoded))
	}
	// A complex key should still have a deterministic searchable/path value if
	// serialization fails for an unusual node.
	encoded, _ = json.Marshal(keyNode.Value)
	return strings.TrimSpace(string(encoded))
}
