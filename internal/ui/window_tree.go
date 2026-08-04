package ui

import (
	"fmt"
	"strings"

	"yamlviewer/internal/display"
	"yamlviewer/internal/yamlmodel"
)

type treeItem struct {
	node     *yamlmodel.Node
	label    string
	children []string
}

func (viewer *Viewer) childrenOf(id string) []string {
	item, ok := viewer.items[id]
	if !ok {
		return nil
	}
	return item.children
}

func (viewer *Viewer) refreshTree() {
	viewer.items = make(map[string]treeItem)
	root := treeItem{label: "YAML hierarchy"}
	viewer.items["tree-root"] = root

	if viewer.current != nil && viewer.current.Model != nil {
		for _, document := range viewer.current.Model.Documents {
			if document.Root == nil {
				id := fmt.Sprintf("document-%d", document.Number)
				viewer.items[id] = treeItem{label: fmt.Sprintf("Document %d (empty)", document.Number)}
				if viewer.includeDocument(document.Number, nil) {
					root.children = append(root.children, id)
				}
				continue
			}
			if !viewer.includeDocument(document.Number, document.Root) {
				continue
			}
			if len(viewer.current.Model.Documents) > 1 {
				id := fmt.Sprintf("document-%d", document.Number)
				viewer.items[id] = treeItem{label: fmt.Sprintf("Document %d", document.Number), children: []string{document.Root.ID}}
				root.children = append(root.children, id)
			} else {
				root.children = append(root.children, document.Root.ID)
			}
			viewer.addNode(document.Root)
		}
	}
	if len(root.children) == 0 {
		if viewer.current == nil || viewer.current.Model == nil {
			root.label = "YAML hierarchy — open a file"
		} else if viewer.current.Model.Empty || allDocumentsEmpty(viewer.current.Model) {
			root.label = "YAML hierarchy — Empty Document"
		} else {
			root.label = "YAML hierarchy — no matches"
		}
	}
	viewer.items["tree-root"] = root
	viewer.tree.Root = "tree-root"
	viewer.tree.Refresh()
	viewer.restoreBranches()
	viewer.updateStatus()
}

func (viewer *Viewer) addNode(node *yamlmodel.Node) {
	if node == nil || !viewer.visibleNode(node) {
		return
	}
	children := make([]string, 0, len(node.Children))
	for _, child := range node.Children {
		if viewer.visibleNode(child) {
			viewer.addNode(child)
			children = append(children, child.ID)
		}
	}
	viewer.items[node.ID] = treeItem{node: node, label: nodeLabel(node), children: children}
}

func (viewer *Viewer) visibleNode(node *yamlmodel.Node) bool {
	return strings.TrimSpace(viewer.searchEntry.Text) == "" || viewer.visible[node.ID]
}

func (viewer *Viewer) includeDocument(number int, root *yamlmodel.Node) bool {
	if strings.TrimSpace(viewer.searchEntry.Text) == "" {
		return true
	}
	if root != nil && viewer.visible[root.ID] {
		return true
	}
	for _, result := range viewer.results {
		if result.Node != nil && result.Node.Path != "" && documentNumber(result.Node.ID) == number {
			return true
		}
	}
	return false
}

func documentNumber(id string) int {
	var number int
	if _, err := fmt.Sscanf(id, "doc-%d-node-", &number); err == nil {
		return number
	}
	return 0
}

func nodeLabel(node *yamlmodel.Node) string {
	label := display.NodeLabel(node)
	if node.KeySet && display.HumanizeKey(node.Key) != node.Key && node.Key != "" {
		label += " [" + node.Key + "]"
	}
	if node.Duplicate {
		label += " (duplicate key)"
	}
	return label
}

func (viewer *Viewer) restoreBranches() {
	viewer.programmatic = true
	defer func() { viewer.programmatic = false }()
	viewer.tree.CloseAllBranches()
	for id, expanded := range viewer.state.Expanded {
		if expanded {
			viewer.tree.OpenBranch(id)
		}
	}
	if strings.TrimSpace(viewer.searchEntry.Text) != "" {
		for id, item := range viewer.items {
			if id != "tree-root" && len(item.children) > 0 && viewer.visible[id] {
				viewer.tree.OpenBranch(id)
			}
		}
		viewer.tree.OpenBranch("tree-root")
	}
	viewer.updateExpandCollapseButton()
}

func (viewer *Viewer) selectTreeItem(id string) {
	item, ok := viewer.items[id]
	if !ok || item.node == nil {
		return
	}
	if viewer.editingNode != item.node.ID {
		viewer.editingNode = ""
	}
	viewer.state.Selected = item.node
	viewer.updateInspector(item.node)
	viewer.updateCommands()
}

func (viewer *Viewer) selectedNode() *yamlmodel.Node {
	return viewer.state.Selected
}

func allDocumentsEmpty(file *yamlmodel.File) bool {
	if file == nil || len(file.Documents) == 0 {
		return true
	}
	for _, document := range file.Documents {
		if document.Root != nil {
			return false
		}
	}
	return true
}
