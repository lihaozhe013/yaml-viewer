package yamlmodel

import "testing"

func TestDecodePreservesStructureAndMetadata(t *testing.T) {
	file, err := Decode([]byte("# head\nsettings:\n  tick_rate: 30 # rate\n  items: []\n  enabled: null\n---\nvalue: &base one\ncopy: *base\n"))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if len(file.Documents) != 2 {
		t.Fatalf("got %d documents, want 2", len(file.Documents))
	}
	root := file.Documents[0].Root
	if root == nil || root.Kind != MappingNode || root.Children[0].Path != "/settings" {
		t.Fatalf("unexpected root: %#v", root)
	}
	settings := root.Children[0]
	if settings.Children[0].Key != "tick_rate" || settings.Children[0].KeySet == false {
		t.Fatalf("raw key was not preserved: %#v", settings.Children[0])
	}
	if settings.Children[0].Comments.Line != "# rate" {
		t.Errorf("line comment = %q, want %q", settings.Children[0].Comments.Line, "# rate")
	}
	if settings.Children[1].Kind != SequenceNode || len(settings.Children[1].Children) != 0 {
		t.Errorf("empty sequence was not retained: %#v", settings.Children[1])
	}
	if settings.Children[2].Tag != "!!null" {
		t.Errorf("null tag = %q", settings.Children[2].Tag)
	}
	if file.Documents[1].Root.Children[1].Kind != AliasNode || file.Documents[1].Root.Children[1].Alias != "base" {
		t.Errorf("alias was not retained: %#v", file.Documents[1].Root.Children[1])
	}
}

func TestDecodeEmptyFile(t *testing.T) {
	file, err := Decode([]byte("\n # only a comment\n"))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if !file.Empty || len(file.Documents) != 0 {
		t.Fatalf("empty file = %#v", file)
	}
}

func TestDecodeDuplicateKeys(t *testing.T) {
	file, err := Decode([]byte("name: first\nname: second\n"))
	if err != nil {
		t.Fatalf("Decode() duplicate key error = %v", err)
	}
	children := file.Documents[0].Root.Children
	if len(children) != 2 || !children[0].Duplicate || !children[1].Duplicate {
		t.Fatalf("duplicate nodes = %#v", children)
	}
}
