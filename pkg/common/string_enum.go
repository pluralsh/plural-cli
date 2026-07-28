package common

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

type StringEnum struct {
	name    string
	value   string
	allowed []string
}

func NewStringEnum(name, defaultValue string, allowed ...string) *StringEnum {
	enum := &StringEnum{name: name, allowed: allowed}
	if err := enum.Set(defaultValue); err != nil {
		panic(err)
	}

	return enum
}

func StringEnumFlag(name, usage, defaultValue string, allowed ...string) cli.GenericFlag {
	return cli.GenericFlag{
		Name:  name,
		Usage: fmt.Sprintf("%s (%s)", usage, strings.Join(allowed, " or ")),
		Value: NewStringEnum(primaryFlagName(name), defaultValue, allowed...),
	}
}

func ValidateStringEnum(name, value string, allowed ...string) error {
	return validateStringEnumValue(name, value, allowed)
}

func (e *StringEnum) Set(value string) error {
	if err := validateStringEnumValue(e.name, value, e.allowed); err != nil {
		return err
	}

	e.value = value
	return nil
}

func (e *StringEnum) String() string {
	if e == nil {
		return ""
	}

	return e.value
}

func validateStringEnumValue(name, value string, allowed []string) error {
	if len(allowed) == 0 {
		return fmt.Errorf("no allowed values configured for %s", name)
	}

	for _, option := range allowed {
		if value == option {
			return nil
		}
	}

	return fmt.Errorf("unsupported %s %q (must be one of: %s)", name, value, strings.Join(allowed, ", "))
}

func primaryFlagName(name string) string {
	primary, _, _ := strings.Cut(name, ",")
	return strings.TrimSpace(primary)
}
