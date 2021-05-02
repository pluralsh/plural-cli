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