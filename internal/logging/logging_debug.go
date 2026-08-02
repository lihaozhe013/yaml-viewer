//go:build debug

// Package logging configures process-wide logging for the application.
package logging

import (
	"log"
	"os"
)

// Configure enables diagnostic logging for development builds.
func Configure() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
