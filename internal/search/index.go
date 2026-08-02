package search

import (
	"sort"
	"strings"

	"yamlviewer/internal/display"
	"yamlviewer/internal/yamlmodel"
)

// Entry is the searchable projection of a model node. The original values are
// kept for result inspection while normalized values remain implementation
// details of the matcher.
type Entry struct {
	Node         *yamlmodel.Node
	Document     int
	DocumentName string
	RawKey       string
	DisplayName  string
	Path         string
	Value        string
	Tag          string
	Anchor       string
	Alias        string
	Comment      string

	name     Normalized
	path     Normalized
	value    Normalized
	tag      Normalized
	anchor   Normalized
	alias    Normalized
	comment  Normalized
	document Normalized
	display  Normalized
	order    int
}

// Result is a matching node and its stable ranking score.
type Result struct {
	Node  *yamlmodel.Node
	Score int
}

// Mode controls how a query is matched against indexed nodes.
type Mode string

const (
	ModeSmartFuzzy Mode = "smart_fuzzy"
	ModeKeyword    Mode = "keyword"
)

// NormalizeMode returns a supported mode and falls back to smart fuzzy
// matching for missing or unknown values.
func NormalizeMode(mode Mode) Mode {
	if mode == ModeKeyword {
		return ModeKeyword
	}
	return ModeSmartFuzzy
}

// Index contains entries in YAML source order.
type Index struct {
	Entries []*Entry
}

// NewIndex builds an index for every concrete node, preserving document and
// child order.
func NewIndex(file *yamlmodel.File) *Index {
	index := &Index{}
	if file == nil {
		return index
	}
	order := 0
	for _, document := range file.Documents {
		if document.Root == nil {
			continue
		}
		walk(document.Root, document.Number, "Document "+itoa(document.Number), &order, index)
	}
	return index
}

func walk(node *yamlmodel.Node, document int, documentName string, order *int, index *Index) {
	if node == nil {
		return
	}
	entry := &Entry{
		Node:         node,
		Document:     document,
		DocumentName: documentName,
		RawKey:       node.Key,
		Path:         node.Path,
		Value:        node.Value,
		Tag:          node.Tag,
		Anchor:       node.Anchor,
		Alias:        node.Alias,
		Comment:      strings.Join([]string{node.Comments.Head, node.Comments.Line, node.Comments.Foot}, "\n"),
		order:        *order,
	}
	if node.Key == "" && node.Path == "/" {
		entry.DisplayName = documentName
	} else {
		entry.DisplayName = display.HumanizeKey(node.Key)
	}
	entry.name = Normalize(entry.RawKey)
	entry.display = Normalize(entry.DisplayName)
	entry.path = Normalize(entry.Path)
	entry.value = Normalize(entry.Value)
	entry.tag = Normalize(entry.Tag)
	entry.anchor = Normalize(entry.Anchor)
	entry.alias = Normalize(entry.Alias)
	entry.comment = Normalize(entry.Comment)
	entry.document = Normalize(entry.DocumentName)
	index.Entries = append(index.Entries, entry)
	*order++
	for _, child := range node.Children {
		walk(child, document, documentName, order, index)
	}
}

// Search returns matching nodes ordered by score, with ties in original YAML
// order. An empty query returns all indexed nodes in source order.
func (index *Index) Search(query string, mode Mode) []Result {
	if index == nil {
		return nil
	}
	mode = NormalizeMode(mode)
	normalizedQuery := Normalize(query)
	if strings.TrimSpace(normalizedQuery.Compact) == "" {
		results := make([]Result, 0, len(index.Entries))
		for _, entry := range index.Entries {
			results = append(results, Result{Node: entry.Node})
		}
		return results
	}
	results := make([]Result, 0)
	for _, entry := range index.Entries {
		var score int
		var ok bool
		if mode == ModeKeyword {
			score, ok = scoreKeywordEntry(entry, normalizedQuery)
		} else {
			score, ok = scoreFuzzyEntry(entry, normalizedQuery)
		}
		if ok {
			results = append(results, Result{Node: entry.Node, Score: score})
		}
	}
	sort.SliceStable(results, func(left, right int) bool {
		return results[left].Score > results[right].Score
	})
	return results
}

// VisibleIDs returns matching nodes and every ancestor needed to render their
// paths in a filtered tree.
func VisibleIDs(results []Result) map[string]bool {
	visible := make(map[string]bool)
	for _, result := range results {
		for node := result.Node; node != nil; node = node.Parent {
			visible[node.ID] = true
		}
	}
	return visible
}

func searchableFields(entry *Entry) []searchField {
	return []searchField{
		{value: entry.name, base: 1000, kind: "name"},
		{value: entry.display, base: 950, kind: "name"},
		{value: entry.path, base: 850, kind: "path"},
		{value: entry.value, base: 500, kind: "value"},
		{value: entry.tag, base: 250, kind: "metadata"},
		{value: entry.anchor, base: 250, kind: "metadata"},
		{value: entry.alias, base: 250, kind: "metadata"},
		{value: entry.comment, base: 150, kind: "metadata"},
		{value: entry.document, base: 400, kind: "document"},
	}
}

func scoreFuzzyEntry(entry *Entry, query Normalized) (int, bool) {
	fields := searchableFields(entry)

	bestTotal := 0
	for _, token := range query.Tokens {
		token = normalizeToken(token)
		if token == "" {
			continue
		}
		best, found := 0, false
		for _, field := range fields {
			if rank := match(field.value, token, query); rank > 0 {
				candidate := field.base + rank
				if field.kind == "name" && exactWhole(field.value, query) {
					candidate += 500
				}
				if candidate > best {
					best, found = candidate, true
				}
			}
		}
		if !found {
			return 0, false
		}
		bestTotal += best
	}
	return bestTotal, true
}

func scoreKeywordEntry(entry *Entry, query Normalized) (int, bool) {
	fields := searchableFields(entry)
	seen := make(map[string]bool, len(query.Tokens))
	bestTotal := 0
	for _, token := range query.Tokens {
		token = normalizeToken(token)
		if token == "" || seen[token] {
			continue
		}
		seen[token] = true
		best, found := 0, false
		for _, field := range fields {
			if !containsToken(field.value, token) {
				continue
			}
			candidate := field.base + 100
			if candidate > best {
				best, found = candidate, true
			}
		}
		if !found {
			return 0, false
		}
		bestTotal += best
	}
	return bestTotal, true
}

type searchField struct {
	value Normalized
	base  int
	kind  string
}

func containsToken(field Normalized, token string) bool {
	for _, fieldToken := range field.Tokens {
		if normalizeToken(fieldToken) == token {
			return true
		}
	}
	return false
}

func exactWhole(field, query Normalized) bool {
	return field.Compact == query.Compact || field.Segmented == query.Segmented
}

func match(field Normalized, token string, query Normalized) int {
	if field.Compact == "" {
		return 0
	}
	if exactWhole(field, query) || field.Compact == token {
		return 100
	}
	if strings.HasPrefix(field.Compact, token) {
		return 80
	}
	if strings.Contains(field.Compact, token) {
		return 60
	}
	if isSubsequence(token, field.Compact) {
		return 40
	}
	return 0
}

func isSubsequence(needle, haystack string) bool {
	needleRunes, haystackRunes := []rune(needle), []rune(haystack)
	needleIndex := 0
	for _, character := range haystackRunes {
		if needleIndex < len(needleRunes) && character == needleRunes[needleIndex] {
			needleIndex++
		}
	}
	return needleIndex == len(needleRunes)
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	index := len(digits)
	for value > 0 {
		index--
		digits[index] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[index:])
}
