package utils

import (
	"os"
	"strconv"
)

func GetEnvStringValue(key string) (string, bool) {
	if env := os.Getenv(key); env != "" {
		return env, true
	}
	return "", false
}

func GetEnvBoolValue(key string) (bool, bool) {
	if env := os.Getenv(key); env != "" {
		b, _ := strconv.ParseBool(env)
		return b, true
	}
	return false, false
}
