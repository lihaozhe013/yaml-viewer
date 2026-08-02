package fileio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAtomicReplacesFile(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "config.yaml")
	if err := os.WriteFile(path, []byte("before\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := WriteAtomic(path, []byte("after\n")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "after\n" {
		t.Fatalf("saved data = %q, want after", data)
	}
	if information, err := os.Stat(path); err != nil {
		t.Fatal(err)
	} else if information.Mode().Perm() != 0o640 {
		t.Fatalf("saved permissions = %o, want 640", information.Mode().Perm())
	}
}

func TestWriteAtomicCreatesNewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new.yaml")
	if err := WriteAtomic(path, []byte("value: true\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
