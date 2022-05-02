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
		return strings.Replace(filepath, "\\", "/", -1)
	default:
		return filepath
	}
}

// returns a string (filepath) with appropriate slashes, dependent on OS if error is empty. Optional parameters and method overloading unavailable in Go.
func SanitizeFilePathWithError(filepath string, err error) (string, error) {
	os := runtime.GOOS
	switch os {
	case "windows":
		return strings.Replace(filepath, "\\", "/", -1), err
	default:
		return filepath, err
	}
}
