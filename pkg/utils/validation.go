package utils

import (
	"fmt"
	"regexp"
)

const (
	dnsRegex = "(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])"
)

type fetcher func() (string, error)
type validator func(string) error

func UntilValid(fetch func() error) {
	for {
		if err := fetch(); err != nil {
			fmt.Printf("%s\n", HighlightError(err))
			continue
		}
		break
	}
}

func UntilInputValid(fetch fetcher, valid validator) string {
	for {
		res, err := fetch()
		if err != nil {
			fmt.Printf("%s\n", HighlightError(err))
			continue
		}

		if err := valid(res); err != nil {
			fmt.Printf("%s\n", HighlightError(err))
			continue
		}

		return res
	}
}

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

func ValidateDns(val string) error {
	return ValidateRegex(val, dnsRegex, "String must be a dns compliant hostname")
}