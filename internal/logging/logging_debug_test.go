//go:build debug

package logging

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestDebugfWritesFeaturePrefixedMessage(t *testing.T) {
	var output bytes.Buffer
	previous := debugLogger
	debugLogger = log.New(&output, "", 0)
	t.Cleanup(func() {
		debugLogger = previous
	})

	Debugf("config", "load failed: %s", "invalid YAML")

	message := output.String()
	if !strings.Contains(message, "[config] load failed: invalid YAML") {
		t.Fatalf("unexpected debug output: %q", message)
	}
}
