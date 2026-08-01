package search

import (
	"strings"
	"unicode"

	"yamlviewer/internal/display"
)

// Normalized stores both forms needed to treat snake_case, camelCase and
// whitespace-separated queries as equivalent.
type Normalized struct {
	Segmented string
	Compact   string
	Tokens    []string
}

// Normalize applies Unicode-aware case normalization and common word-boundary
// rules. Punctuation is treated as a separator for search purposes.
func Normalize(raw string) Normalized {
	words := display.Words(raw)
	for index, word := range words {
		words[index] = strings.ToLower(word)
	}
	segmented := strings.Join(words, " ")
	return Normalized{Segmented: segmented, Compact: strings.Join(words, ""), Tokens: words}
}

func normalizeToken(raw string) string {
	var builder strings.Builder
	for _, character := range strings.ToLower(raw) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			builder.WriteRune(character)
		}
	}
	return builder.String()
}
