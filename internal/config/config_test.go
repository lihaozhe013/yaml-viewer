package config

import (
	"os"
	"path/filepath"
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
}
