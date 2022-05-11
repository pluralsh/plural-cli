package utils

import (
	"fmt"
	"regexp"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/utils/errors"
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

	return errors.ErrorWrap(fmt.Errorf(message), "Validation Failure")
}

func RegexValidator(regex, message string) survey.Validator {
	return func(val interface{}) error {
		str, ok := val.(string)
		if !ok {
			return fmt.Errorf("Result is not a string")
		}

		return ValidateRegex(str, regex, message)
	}
}

var ValidateAlphaNumeric = survey.ComposeValidators(
	survey.Required,
	RegexValidator("[a-z][0-9\\-a-z]+", "Must be an alphanumeric string"),
)

var ValidateAlphaNumExtended = survey.ComposeValidators(
	survey.Required,
	RegexValidator("[a-zA-Z][0-9\\-_a-zA-Z]+", "Must be an alphanumeric string"),
)

func ValidateDns(val string) error {
	return ValidateRegex(val, dnsRegex, "String must be a dns compliant hostname")
}

func Confirm(msg string) bool {
	res := true
	prompt := &survey.Confirm{Message: msg}
	survey.AskOne(prompt, &res, survey.WithValidator(survey.Required))
	return res
}