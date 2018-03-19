package main

import (
	"runtime"
)

// IsTargetDevice tells if we're running on real device
func IsTargetDevice() bool {

	if runtime.GOOS == "windows" {
		return false
	}

	return true
}
