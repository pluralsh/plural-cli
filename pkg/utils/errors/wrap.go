package errors

import (
	"github.com/fatih/color"
	"fmt"
)

func ErrorWrap(err error, explanation string) error {
	if err == nil { return err }

	return fmt.Errorf("%s: %s", color.New(color.FgRed, color.Bold).Sprint(explanation), err.Error())
}