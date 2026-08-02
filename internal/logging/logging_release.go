//go:build release

package logging

import (
	"io"
	"log"
)

// Configure disables standard-library logging for production builds.
func Configure() {
	log.SetOutput(io.Discard)
	log.SetFlags(0)
}
