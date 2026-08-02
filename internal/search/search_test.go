package search

import (
	"testing"

	"yamlviewer/internal/yamlmodel"
)

func TestNormalizeEquivalentKeyStyles(t *testing.T) {
	want := Normalize("tick_rate")
	for _, query := range []string{"Tick Rate", "tick-rate", "tickRate", "TICK RATE"} {
		got := Normalize(query)
		if got.Compact != want.Compact {
			t.Errorf("Normalize(%q) = %#v, want %#v", query, got, want)
		}
	}
}

func TestSearchMatchesEquivalentFormsAndKeepsAncestors(t *testing.T) {
	file, err := yamlmodel.Decode([]byte("settings:\n  tick_rate: 30\n  tick-rate-copy: 31\n"))
	if err != nil {
		t.Fatal(err)
	}
	index := NewIndex(file)
	for _, query := range []string{"tick_rate", "Tick Rate", "tick-rate", "tickRate", "tickrate"} {
		results := index.Search(query, ModeSmartFuzzy)
		if len(results) == 0 || results[0].Node.Key != "tick_rate" {
			t.Errorf("Search(%q) = %#v", query, results)
		}
		visible := VisibleIDs(results)
		if !visible["doc-1-node-1"] {
			t.Errorf("Search(%q) did not retain the root ancestor: %#v", query, visible)
		}
	}
}

func TestSearchValueAndMetadata(t *testing.T) {
	file, err := yamlmodel.Decode([]byte("settings: !custom\n  value: &anchor hello-world # searchable\ncopy: *anchor\n"))
	if err != nil {
		t.Fatal(err)
	}
	index := NewIndex(file)
	for _, query := range []string{"hello", "custom", "anchor", "searchable"} {
		if len(index.Search(query, ModeSmartFuzzy)) == 0 {
			t.Errorf("Search(%q) returned no results", query)
		}
	}
}

func TestKeywordSearchIgnoresOrderAndRequiresWholeKeywords(t *testing.T) {
	file, err := yamlmodel.Decode([]byte("player:\n  attack_speed: 30\n  attack: 10\nplaying: true\n"))
	if err != nil {
		t.Fatal(err)
	}
	index := NewIndex(file)
	for _, query := range []string{"player speed attack", "player attack speed"} {
		results := index.Search(query, ModeKeyword)
		if len(results) == 0 || results[0].Node.Key != "attack_speed" {
			t.Errorf("Keyword search %q = %#v, want attack_speed first", query, results)
		}
	}
	if results := index.Search("pla speed attack", ModeKeyword); len(results) != 0 {
		t.Fatalf("Keyword search accepted partial keyword: %#v", results)
	}
}

func BenchmarkSearchNested(b *testing.B) {
	file, err := yamlmodel.Decode([]byte("root:\n  section:\n    tick_rate: 30\n    other: value\n"))
	if err != nil {
		b.Fatal(err)
	}
	index := NewIndex(file)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = index.Search("tick rate", ModeSmartFuzzy)
	}
}
