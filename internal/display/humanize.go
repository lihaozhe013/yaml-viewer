// Package display contains presentation-only formatting helpers.
package display

import (
	"strings"
	"unicode"
)

// HumanizeKey converts common machine-oriented key styles into a readable
// label. It deliberately has no application-specific acronym dictionary.
func HumanizeKey(raw string) string {
	words := splitWords(raw)
	for index, word := range words {
		words[index] = formatWord(word)
	}
	return strings.Join(words, " ")
}

// Words exposes the same boundary rules used by HumanizeKey to other
// presentation-independent features such as search normalization.
func Words(raw string) []string {
	return splitWords(raw)
}

func splitWords(raw string) []string {
	runes := []rune(raw)
	words := make([]string, 0, 4)
	start := -1
	flush := func(end int) {
		if start >= 0 && end > start {
			words = append(words, string(runes[start:end]))
		}
		start = -1
	}

	for index, current := range runes {
		if !unicode.IsLetter(current) && !unicode.IsDigit(current) {
			flush(index)
			continue
		}
		if start < 0 {
			start = index
			continue
		}
		previous := runes[index-1]
		var next rune
		if index+1 < len(runes) {
			next = runes[index+1]
		}
		if boundary(previous, current, next) {
			flush(index)
			start = index
		}
	}
	flush(len(runes))
	return words
}

func boundary(previous, current, next rune) bool {
	if unicode.IsLower(previous) && unicode.IsUpper(current) {
		return true
	}
	if unicode.IsLetter(previous) && unicode.IsDigit(current) {
		return true
	}
	if unicode.IsDigit(previous) && unicode.IsLetter(current) {
		return true
	}
	// HTTPServer -> HTTP Server: split before the final capital in an acronym.
	return unicode.IsUpper(previous) && unicode.IsUpper(current) && unicode.IsLower(next)
}

func formatWord(word string) string {
	if word == "" {
		return word
	}
	allUpper := true
	for _, character := range word {
		if unicode.IsLetter(character) && !unicode.IsUpper(character) {
			allUpper = false
			break
		}
	}
	if allUpper {
		return word
	}
	runes := []rune(strings.ToLower(word))
	for index, character := range runes {
		if unicode.IsLetter(character) {
			runes[index] = unicode.ToUpper(character)
			break
		}
	}
	return string(runes)
}
