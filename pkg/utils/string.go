package utils

import (
	"fmt"
)

func Pluralize(one, many string, count int) string {
	if count == 1 {
		return one
	}
	return many
}

func ToString(val interface{}) string {
	return fmt.Sprintf("%v", val)
}

func TruncString(s string, c int) string {
	if c < 0 && len(s)+c > 0 {
		return s[len(s)+c:]
	}
	if c >= 0 && len(s) > c {
		return s[:c]
	}
	return s
}
