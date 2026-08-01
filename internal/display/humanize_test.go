package display

import "testing"

func TestHumanizeKey(t *testing.T) {
	tests := map[string]string{
		"tick_rate":         "Tick Rate",
		"tick-rate":         "Tick Rate",
		"tickRate":          "Tick Rate",
		"TICK_RATE":         "TICK RATE",
		"HTTPServer":        "HTTP Server",
		"max2Dashes":        "Max 2 Dashes",
		"already Humanized": "Already Humanized",
		"服务器地址":             "服务器地址",
	}
	for input, expected := range tests {
		if actual := HumanizeKey(input); actual != expected {
			t.Errorf("HumanizeKey(%q) = %q, want %q", input, actual, expected)
		}
	}
}
