//go:build !debug && !release

package logging

import (
	"log"
	"os"
)

// Configure keeps the default development-friendly logging behavior for
// builds that do not explicitly select a mode.
func Configure() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
