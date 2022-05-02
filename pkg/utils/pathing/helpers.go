package pathing

import (
	"runtime"
	"strings"
)

// returns a string (filepath) with appropriate slashes, dependent on OS
func GenOSFilepathString(filepath string) string {
	os := runtime.GOOS
	switch os {
	case "windows":
		return strings.Replace(filepath, "\\", "/", -1)
	default:
		return filepath
	}
}

// returns a string (filepath) with appropriate slashes, dependent on OS
func GenOSFilepathStringWithError(filepath string, err error) (string, error) {
	os := runtime.GOOS
	switch os {
	case "windows":
		return strings.Replace(filepath, "\\", "/", -1), err
	default:
		return filepath, err
	}
}
