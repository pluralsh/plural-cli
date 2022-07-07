package errors

import (
	"fmt"

	"github.com/fatih/color"
)

func ErrorWrap(err error, explanation string) error {
	if err == nil {
		return err
	}

	return fmt.Errorf("%s: %w", color.New(color.FgRed, color.Bold).Sprint(explanation), err)
}
