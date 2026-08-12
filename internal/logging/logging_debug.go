//go:build debug

// Package logging configures process-wide logging for the application.
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

const debugLogPath = "debug.log"

var (
	debugLogger            = log.New(io.Discard, "", log.LstdFlags|log.Lshortfile)
	debugLoggerMu          sync.Mutex
	debugLoggerInitialized bool
	debugLogFile           *os.File
)

// Debugf writes a feature-prefixed diagnostic message in debug builds.
func Debugf(feature string, format string, args ...any) {
	message := fmt.Sprintf("[%s] %s", feature, fmt.Sprintf(format, args...))
	debugLoggerMu.Lock()
	defer debugLoggerMu.Unlock()

	if !debugLoggerInitialized {
		debugLoggerInitialized = true
		if file, err := os.OpenFile(debugLogPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600); err == nil {
			debugLogFile = file
			debugLogger.SetOutput(file)
		}
	}

	_ = debugLogger.Output(2, message)
}
