//go:build !debug

package logging

// Debugf discards diagnostic messages in normal builds.
func Debugf(_ string, _ string, _ ...any) {}
