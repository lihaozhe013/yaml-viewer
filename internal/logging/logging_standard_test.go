//go:build !debug

package logging

import (
	"bytes"
	"log"
	"testing"
)

func TestDebugfDiscardsMessagesInNormalBuilds(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() {
		log.SetOutput(previous)
	})

	Debugf("config", "this message is discarded")

	if output.Len() != 0 {
		t.Fatalf("unexpected normal-build output: %q", output.String())
	}
}
