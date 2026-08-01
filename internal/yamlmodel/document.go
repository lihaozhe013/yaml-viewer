package yamlmodel

// Document is one YAML document from a stream. Root is nil for an empty
// document, such as a stream containing only "---".
type Document struct {
	Number int
	Root   *Node
	Empty  bool
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
