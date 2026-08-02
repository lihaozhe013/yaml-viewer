package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"yamlviewer/internal/yamlmodel"
)

func TestPaletteForLightAndDarkThemes(t *testing.T) {
	light := paletteFor(theme.VariantLight)
	dark := paletteFor(theme.VariantDark)
	for name, palette := range map[string]detailPalette{"light": light, "dark": dark} {
		for section, value := range map[detailSection]struct {
			surface interface{}
			accent  interface{}
		}{
			sectionIdentity: {surface: palette.identitySurface, accent: palette.identityAccent},
			sectionValue:    {surface: palette.valueSurface, accent: palette.valueAccent},
			sectionMetadata: {surface: palette.metadataSurface, accent: palette.metadataAccent},
			sectionComments: {surface: palette.commentSurface, accent: palette.commentAccent},
			sectionWarning:  {surface: palette.warningSurface, accent: palette.warningAccent},
		} {
			if value.surface == nil || value.accent == nil {
				t.Errorf("%s palette %s has an empty color", name, section)
			}
		}
	}
	if light.codeBackground == dark.codeBackground {
		t.Error("light and dark code blocks should use different backgrounds")
	}
}

func TestValueForDisplay(t *testing.T) {
	tests := []struct {
		name string
		node *yamlmodel.Node
		want string
	}{
		{name: "scalar", node: &yamlmodel.Node{Kind: yamlmodel.ScalarNode, Value: "30"}, want: "30"},
		{name: "null", node: &yamlmodel.Node{Kind: yamlmodel.ScalarNode, YAML: "null"}, want: "null"},
		{name: "mapping", node: &yamlmodel.Node{Kind: yamlmodel.MappingNode, YAML: "{}"}, want: "{}"},
		{name: "sequence", node: &yamlmodel.Node{Kind: yamlmodel.SequenceNode, YAML: "[]"}, want: "[]"},
		{name: "alias", node: &yamlmodel.Node{Kind: yamlmodel.AliasNode, Alias: "base"}, want: "*base"},
		{name: "empty", node: &yamlmodel.Node{Kind: yamlmodel.ScalarNode}, want: "null / empty"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := valueForDisplay(test.node); got != test.want {
				t.Fatalf("valueForDisplay() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestHasComments(t *testing.T) {
	if hasComments(&yamlmodel.Node{}) {
		t.Error("empty comments should not render a comments card")
	}
	if !hasComments(&yamlmodel.Node{Comments: yamlmodel.Comments{Line: "# rate"}}) {
		t.Error("line comments should render a comments card")
	}
}

func TestUpdateInspectorBuildsCardLayout(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	application := test.NewApp()
	defer application.Quit()
	viewer := New(application)
	viewer.updateInspector(&yamlmodel.Node{
		Kind:   yamlmodel.ScalarNode,
		Key:    "tick_rate",
		KeySet: true,
		Value:  "30",
		Path:   "/settings/tick_rate",
		Tag:    "!!int",
	})
	if len(viewer.inspector.Objects) < 5 {
		t.Fatalf("inspector contains %d objects, want a card-based layout", len(viewer.inspector.Objects))
	}
}
