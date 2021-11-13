package utils

import (
	"fmt"
	"regexp"
)

func ValidateRegex(val, regex, message string) error {
	reg, err := regexp.Compile(fmt.Sprintf("^%s$", regex))
	if err != nil {
		return err
	}

	if reg.MatchString(val) {
		return nil
	}

	return ErrorWrap(fmt.Errorf(message), "Validation Failure")
}