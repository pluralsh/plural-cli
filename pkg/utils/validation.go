package utils

import (
	"fmt"
	"os"
	"regexp"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/utils/errors"
)

const (
	dnsRegex = "(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])"
)

func ValidateSingleLevelDeep(val, subdomain string) error {
	reg, err := regexp.Compile(fmt.Sprintf("\\w+.%s", subdomain))
	if err != nil {
		return err
	}
	submatch := reg.FindStringSubmatch(val)

	if len(submatch) == 1 {
		if submatch[0] == val {
			return nil
		}
	}
	return fmt.Errorf("use single level deep domain: <domainName>.%s", subdomain)
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
	prompt := &survey.Confirm{Message: msg, Default: true}
	if err := survey.AskOne(prompt, &res, survey.WithValidator(survey.Required)); err != nil {
		return false
	}
	return res
}

func FileExists(val interface{}) error {
	path, ok := val.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", val)
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("provided path points to a directory, not a file: %s", path)
	}

	return nil
}
