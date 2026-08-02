package yamlmodel

import "go.yaml.in/yaml/v3"

// Document is one YAML document from a stream. Root is nil for an empty
// document, such as a stream containing only "---".
type Document struct {
	Number int
	Root   *Node
	Empty  bool

	// Source is the original yaml.v3 document node. It is kept private to the
	// model package's editing and encoding operations while Root remains the
	// presentation tree consumed by the UI.
	Source *yaml.Node
}

// File is the decoded representation of one source file.
type File struct {
	Documents []Document
	Empty     bool
}

// NodeCount returns the number of concrete nodes across all documents.
func (f *File) NodeCount() int {
	if f == nil {
		return 0
	}
	count := 0
	for _, document := range f.Documents {
		count += countNodes(document.Root)
	}
	return count
}

func countNodes(node *Node) int {
	if node == nil {
		return 0
	}
	count := 1
	for _, child := range node.Children {
		count += countNodes(child)
	}
	return count
}
