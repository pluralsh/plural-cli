package pathing

import (
	"runtime"
	"strings"
)

// returns a string (filepath) with appropriate slashes, dependent on OS. Optional parameters and method overloading unavailable in Go.
func SanitizeFilepath(filepath string) string {
	os := runtime.GOOS
	switch os {
	case "windows":
		return strings.ReplaceAll(filepath, "\\", "/")
	default:
		return filepath
	}
}
