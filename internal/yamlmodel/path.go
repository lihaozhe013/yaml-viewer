package yamlmodel

import "strings"

// AppendPath appends a mapping key or sequence index to a JSON-pointer-like
// path. Escaping makes keys containing '/' unambiguous while keeping paths
// readable in the inspector.
func AppendPath(base, segment string) string {
	segment = strings.NewReplacer("~", "~0", "/", "~1").Replace(segment)
	if base == "" || base == "/" {
		return "/" + segment
	}
	return strings.TrimRight(base, "/") + "/" + segment
}
