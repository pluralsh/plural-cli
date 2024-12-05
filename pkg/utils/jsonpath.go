package utils

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/client-go/util/jsonpath"
)

var jsonRegexp = regexp.MustCompile(`^\{\.?([^{}]+)\}$|^\.?([^{}]+)$`)

func ParseJSONPath(input string, data interface{}) error {
	after, ok := strings.CutPrefix(input, "jsonpath=")
	if !ok {
		return fmt.Errorf("invalid jsonpath format: %s", input)
	}
	field, err := RelaxedJSONPathExpression(after)
	if err != nil {
		return err
	}
	parser := jsonpath.New("parsing").AllowMissingKeys(true)
	err = parser.Parse(field)
	if err != nil {
		return fmt.Errorf("parsing error: %w", err)
	}
	buf := new(bytes.Buffer)
	if err := parser.Execute(buf, data); err != nil {
		return err
	}
	fmt.Print(buf.String())
	return nil
}

func RelaxedJSONPathExpression(pathExpression string) (string, error) {
	if len(pathExpression) == 0 {
		return pathExpression, nil
	}
	submatches := jsonRegexp.FindStringSubmatch(pathExpression)
	if submatches == nil {
		return "", fmt.Errorf("unexpected path string, expected a 'name1.name2' or '.name1.name2' or '{name1.name2}' or '{.name1.name2}'")
	}
	if len(submatches) != 3 {
		return "", fmt.Errorf("unexpected submatch list: %v", submatches)
	}
	var fieldSpec string
	if len(submatches[1]) != 0 {
		fieldSpec = submatches[1]
	} else {
		fieldSpec = submatches[2]
	}
	return fmt.Sprintf("{.%s}", fieldSpec), nil
}
