package main

import (
	"os"
	"runtime"
)

// IsTargetDevice tells if we're running on real device
func IsTargetDevice() bool {

	if runtime.GOOS == "windows" {
		return false
	}

	return true
}

// IsDebugMode check if we have debug enabled
func IsDebugMode() bool {
	return os.Getenv("DEBUG") != ""
}

// IsTraceMode check if we have trace enabled
func IsTraceMode() bool {
	return os.Getenv("TRACE") != ""
}
