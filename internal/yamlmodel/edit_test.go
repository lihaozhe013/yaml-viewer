package yamlmodel

import (
	"strings"
	"testing"
)

func TestSetScalarLiteralPreservesDocumentMetadata(t *testing.T) {
	file, err := Decode([]byte("# config\nname: old # keep this\nnumber: 30\nbase: &base one\ncopy: *base\n---\nsecond: true\n"))
	if err != nil {
		t.Fatal(err)
	}
	number := file.Documents[0].Root.Children[1]
	change, err := file.SetScalarLiteral(number.ID, "42")
	if err != nil {
		t.Fatal(err)
	}
	if change.Before.Value != "30" || change.After.Value != "42" {
		t.Fatalf("unexpected change: %#v", change)
	}
	if number.Value != "42" || number.Tag != "!!int" {
		t.Fatalf("display node was not synchronized: %#v", number)
	}

	encoded, err := file.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	for _, expected := range []string{"name: old # keep this", "number: 42", "base: &base one", "copy: *base", "second: true"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("encoded YAML missing %q:\n%s", expected, text)
		}
	}

	if err := file.ApplyScalarChange(change, false); err != nil {
		t.Fatal(err)
	}
	if number.Value != "30" {
		t.Fatalf("undo value = %q, want 30", number.Value)
	}
	if err := file.ApplyScalarChange(change, true); err != nil {
		t.Fatal(err)
	}
	if number.Value != "42" {
		t.Fatalf("redo value = %q, want 42", number.Value)
	}
}

func TestSetScalarLiteralAllowsTypeChanges(t *testing.T) {
	file, err := Decode([]byte("value: 30\n"))
	if err != nil {
		t.Fatal(err)
	}
	node := file.Documents[0].Root.Children[0]
	if _, err := file.SetScalarLiteral(node.ID, `"30"`); err != nil {
		t.Fatal(err)
	}
	if node.Value != "30" || node.Tag != "!!str" {
		t.Fatalf("type change = value %q tag %q", node.Value, node.Tag)
	}
}

func TestSetScalarLiteralRejectsNonScalarsAndMultipleDocuments(t *testing.T) {
	file, err := Decode([]byte("value: 30\nitems: []\n"))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		node  *Node
		input string
	}{
		{name: "mapping input", node: file.Documents[0].Root.Children[0], input: "key: value"},
		{name: "sequence node", node: file.Documents[0].Root.Children[1], input: "1"},
		{name: "multiple documents", node: file.Documents[0].Root.Children[0], input: "1\n---\n2"},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := test.node.Value
			if _, err := file.SetScalarLiteral(test.node.ID, test.input); err == nil {
				t.Fatal("SetScalarLiteral() unexpectedly succeeded")
			}
			if test.node.Value != before {
				t.Fatalf("failed edit changed value to %q", test.node.Value)
			}
		})
	}
}

func TestMarshalPreservesEmptyDocuments(t *testing.T) {
	file, err := Decode([]byte("---\n---\n"))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := file.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("encoded empty documents are invalid: %v\n%s", err, encoded)
	}
	if len(decoded.Documents) != 2 || !decoded.Documents[0].Empty || !decoded.Documents[1].Empty {
		t.Fatalf("decoded empty documents = %#v", decoded.Documents)
	}
}

func TestMarshalWithOptionsSortsKeys(t *testing.T) {
	file, err := Decode([]byte("zebra: 1\nalpha: 2\nbeta: 3\n"))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := file.MarshalWithOptions(FormatOptions{Indent: 2, SortKeys: true})
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	alphaIdx := strings.Index(text, "alpha:")
	betaIdx := strings.Index(text, "beta:")
	zebraIdx := strings.Index(text, "zebra:")
	if alphaIdx < 0 || betaIdx < 0 || zebraIdx < 0 {
		t.Fatalf("missing keys in output:\n%s", text)
	}
	if alphaIdx > betaIdx || betaIdx > zebraIdx {
		t.Fatalf("keys not sorted: alpha@%d beta@%d zebra@%d\n%s", alphaIdx, betaIdx, zebraIdx, text)
	}
}

func TestMarshalWithOptionsPreservesOriginalOrder(t *testing.T) {
	file, err := Decode([]byte("zebra: 1\nalpha: 2\nbeta: 3\n"))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := file.MarshalWithOptions(FormatOptions{Indent: 2, SortKeys: false})
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	zebraIdx := strings.Index(text, "zebra:")
	alphaIdx := strings.Index(text, "alpha:")
	betaIdx := strings.Index(text, "beta:")
	if zebraIdx < 0 || alphaIdx < 0 || betaIdx < 0 {
		t.Fatalf("missing keys in output:\n%s", text)
	}
	if zebraIdx > alphaIdx || alphaIdx > betaIdx {
		t.Fatalf("original order not preserved: zebra@%d alpha@%d beta@%d\n%s", zebraIdx, alphaIdx, betaIdx, text)
	}
}

func TestMarshalWithOptionsUsesIndent(t *testing.T) {
	file, err := Decode([]byte("parent:\n  child: value\n"))
	if err != nil {
		t.Fatal(err)
	}
	encoded2, err := file.MarshalWithOptions(FormatOptions{Indent: 2, SortKeys: false})
	if err != nil {
		t.Fatal(err)
	}
	encoded4, err := file.MarshalWithOptions(FormatOptions{Indent: 4, SortKeys: false})
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded2) == string(encoded4) {
		t.Fatalf("2-space and 4-space encoding should differ:\n2: %s\n4: %s", encoded2, encoded4)
	}
	if !strings.Contains(string(encoded4), "    child:") {
		t.Fatalf("4-space indent not found:\n%s", encoded4)
	}
}
