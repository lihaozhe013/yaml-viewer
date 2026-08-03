package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestPathUsesRequestedConfigLocation(t *testing.T) {
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".config", "yaml-viewer", "config.yaml")
	if path != want {
		t.Fatalf("Path() = %q, want %q", path, want)
	}
}

func TestConfigYAMLShape(t *testing.T) {
	data, err := yaml.Marshal(Config{LastOpenedFile: "/tmp/example.yaml"})
	if err != nil {
		t.Fatal(err)
	}
	var decoded Config
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.LastOpenedFile != "/tmp/example.yaml" {
		t.Fatalf("LastOpenedFile = %q", decoded.LastOpenedFile)
	}
	if decoded.SearchMode != SearchModeSmartFuzzy {
		t.Fatalf("SearchMode = %q, want %q", decoded.SearchMode, SearchModeSmartFuzzy)
	}
	if decoded.ThemeMode != ThemeModeLight {
		t.Fatalf("ThemeMode = %q, want %q", decoded.ThemeMode, ThemeModeLight)
	}
}

func TestLoadCreatesConfigFromEmbeddedTemplate(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	value, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if value.SearchMode != SearchModeSmartFuzzy {
		t.Fatalf("SearchMode = %q, want %q", value.SearchMode, SearchModeSmartFuzzy)
	}
	if value.ThemeMode != ThemeModeLight {
		t.Fatalf("ThemeMode = %q, want %q", value.ThemeMode, ThemeModeLight)
	}
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := yaml.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["search_mode"] != string(SearchModeSmartFuzzy) {
		t.Fatalf("saved search_mode = %#v", fields["search_mode"])
	}
}

func TestLoadFillsMissingFieldsAndPreservesUnknownFields(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("last_opened_file: /tmp/old.yaml\ncustom_setting: keep-me\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	value, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if value.LastOpenedFile != "/tmp/old.yaml" {
		t.Fatalf("LastOpenedFile = %q", value.LastOpenedFile)
	}
	if value.SearchMode != SearchModeSmartFuzzy {
		t.Fatalf("SearchMode = %q, want %q", value.SearchMode, SearchModeSmartFuzzy)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, expected := range []string{"custom_setting: keep-me", "search_mode: smart_fuzzy", "theme: light"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("merged config missing %q: %s", expected, text)
		}
	}
}

func TestLoadRepairsInvalidKnownFields(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("last_opened_file: 42\nsearch_mode: unknown\ntheme: system\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	value, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if value.LastOpenedFile != "" || value.SearchMode != SearchModeSmartFuzzy || value.ThemeMode != ThemeModeLight {
		t.Fatalf("repaired config = %#v", value)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, expected := range []string{"search_mode: smart_fuzzy", "theme: light"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("repaired config missing %q: %s", expected, data)
		}
	}
}

func TestLoadPreservesValidThemeValue(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("theme: dark\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	value, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if value.ThemeMode != ThemeModeDark {
		t.Fatalf("ThemeMode = %q, want %q", value.ThemeMode, ThemeModeDark)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "theme: dark") {
		t.Fatalf("saved config lost the selected theme: %s", data)
	}
}

func TestSavePreservesUnknownFields(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("last_opened_file: /tmp/old.yaml\ncustom_setting: keep-me\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	value, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	value.LastOpenedFile = "/tmp/new.yaml"
	if err := Save(value); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "last_opened_file: /tmp/new.yaml") || !strings.Contains(text, "custom_setting: keep-me") {
		t.Fatalf("saved config lost values: %s", text)
	}
}
