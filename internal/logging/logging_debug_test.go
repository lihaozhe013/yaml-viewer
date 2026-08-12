//go:build debug

package logging

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

func TestDebugfWritesFeaturePrefixedMessage(t *testing.T) {
	var output bytes.Buffer
	previous := debugLogger
	previousInitialized := debugLoggerInitialized
	previousFile := debugLogFile
	debugLogger = log.New(&output, "", 0)
	debugLoggerInitialized = true
	debugLogFile = nil
	t.Cleanup(func() {
		debugLogger = previous
		debugLoggerInitialized = previousInitialized
		debugLogFile = previousFile
	})

	Debugf("config", "load failed: %s", "invalid YAML")

	message := output.String()
	if !strings.Contains(message, "[config] load failed: invalid YAML") {
		t.Fatalf("unexpected debug output: %q", message)
	}
}

func TestDebugfWritesToDebugLogByDefault(t *testing.T) {
	t.Chdir(t.TempDir())

	previousLogger := debugLogger
	previousInitialized := debugLoggerInitialized
	previousFile := debugLogFile
	debugLogger = log.New(io.Discard, "", 0)
	debugLoggerInitialized = false
	debugLogFile = nil
	t.Cleanup(func() {
		if debugLogFile != nil {
			_ = debugLogFile.Close()
		}
		debugLogger = previousLogger
		debugLoggerInitialized = previousInitialized
		debugLogFile = previousFile
	})

	Debugf("search", "query failed")

	contents, err := os.ReadFile(debugLogPath)
	if err != nil {
		t.Fatalf("read debug log: %v", err)
	}
	if !strings.Contains(string(contents), "[search] query failed") {
		t.Fatalf("unexpected debug log contents: %q", contents)
	}
}
