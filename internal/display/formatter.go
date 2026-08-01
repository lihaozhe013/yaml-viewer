package display

import (
	"fmt"
	"strings"

	"yamlviewer/internal/yamlmodel"
)

// NodeDisplayName is the human-readable field label. Sequence items and
// document roots intentionally have structural labels instead of fake keys.
func NodeDisplayName(node *yamlmodel.Node) string {
	if node == nil {
		return ""
	}
	if node.KeySet {
		return HumanizeKey(node.Key)
	}
	if node.Kind == yamlmodel.SequenceNode {
		return fmt.Sprintf("[%d]", node.Index)
	}
	if node.Path == "/" {
		return "Root"
	}
	return "Value"
}

// NodeLabel returns the concise text used in the hierarchy tree.
func NodeLabel(node *yamlmodel.Node) string {
	if node == nil {
		return ""
	}
	name := NodeDisplayName(node)
	if node.Kind == yamlmodel.AliasNode {
		if node.Alias == "" {
			return name + ": Alias"
		}
		return name + ": Alias *" + node.Alias
	}
	if node.Kind == yamlmodel.MappingNode {
		return fmt.Sprintf("%s (%d children)", name, len(node.Children))
	}
	if node.Kind == yamlmodel.SequenceNode {
		return fmt.Sprintf("%s (%d items)", name, len(node.Children))
	}
	if node.Value == "" {
		return name + ": " + scalarType(node.Tag)
	}
	return name + ": " + truncate(node.Value, 80)
}

func scalarType(tag string) string {
	if strings.HasPrefix(tag, "!!") {
		return strings.TrimPrefix(tag, "!!")
	}
	if tag == "" {
		return "scalar"
	}
	return tag
}

func truncate(value string, limit int) string {
	value = strings.ReplaceAll(strings.ReplaceAll(value, "\n", "\\n"), "\r", "\\r")
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit-1]) + "…"
}
