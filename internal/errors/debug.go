package errors

import (
	"os"
	"sync"
)

var (
	debugMode     bool
	debugModeMux  sync.RWMutex
	debugModeOnce sync.Once
)

// initDebugMode initializes debug mode from environment
func initDebugMode() {
	debugModeOnce.Do(func() {
		// Check GRIMOIRE_DEBUG environment variable
		if os.Getenv("GRIMOIRE_DEBUG") != "" {
			debugMode = true
		}
	})
}

// SetDebugMode sets the debug mode
func SetDebugMode(enabled bool) {
	debugModeMux.Lock()
	defer debugModeMux.Unlock()
	debugMode = enabled
}

// getDebugMode returns the current debug mode
func getDebugMode() bool {
	initDebugMode()
	debugModeMux.RLock()
	defer debugModeMux.RUnlock()
	return debugMode
}

// EnableDebugMode enables debug mode
func EnableDebugMode() {
	SetDebugMode(true)
}

// DisableDebugMode disables debug mode
func DisableDebugMode() {
	SetDebugMode(false)
}