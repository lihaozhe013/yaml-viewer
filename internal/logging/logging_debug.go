//go:build debug

// Package logging configures process-wide logging for the application.
package logging

import (
	"fmt"
	"log"
	"os"
)

var debugLogger = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)

// Debugf writes a feature-prefixed diagnostic message in debug builds.
func Debugf(feature string, format string, args ...any) {
	message := fmt.Sprintf("[%s] %s", feature, fmt.Sprintf(format, args...))
	_ = debugLogger.Output(2, message)
}
