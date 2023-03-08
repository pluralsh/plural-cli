package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
)

func ReadLine(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	color.New(color.Bold).Printf(prompt)
	res, err := reader.ReadString('\n')
	return strings.TrimSpace(string(res)), err
}

func ReadAlphaNum(prompt string) (string, error) {
	val, err := ReadLine(prompt)
	if err != nil {
		return val, err
	}

	return val, ValidateRegex(val, "[a-z][0-9\\-a-z]+", "String can only contain alphanumeric characters or hyphens")
}

func ReadAlphaNumDefault(prompt string, def string) (string, error) {
	result, err := ReadAlphaNum(fmt.Sprintf("%s [%s]: ", prompt, def))
	if result == "" {
		return def, nil
	}

	return result, err
}

func ReadLineDefault(prompt string, def string) (string, error) {
	result, err := ReadLine(fmt.Sprintf("%s [%s]: ", prompt, def))
	if result == "" {
		return def, nil
	}

	return result, err
}

func ReadPwd(prompt string) (string, error) {
	color.New(color.Bold).Printf(prompt)
	pwd, err := term.ReadPassword(int(syscall.Stdin))
	return strings.TrimSpace(string(pwd)), err
}

func Warn(line string, args ...interface{}) {
	color.New(color.FgYellow, color.Bold).Fprintf(os.Stderr, line, args...)
}

func Success(line string, args ...interface{}) {
	color.New(color.FgGreen, color.Bold).Printf(line, args...)
}

func Error(line string, args ...interface{}) {
	color.New(color.FgRed, color.Bold).Printf(line, args...)
}

func Highlight(line string, args ...interface{}) {
	color.New(color.Bold).Printf(line, args...)
}

func Note(line string, args ...interface{}) {
	Warn("**NOTE** :: ")
	Highlight(line, args...)
}

func HighlightError(err error) error {
	if err != nil {
		err = fmt.Errorf(color.New(color.FgRed, color.Bold).Sprint(err.Error()))
	}
	return err
}

func PrintTable[T any](list []T, headers []string, rowFun func(T) ([]string, error)) error {
	length := len(headers)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	for _, v := range list {
		row, err := rowFun(v)
		if err != nil {
			return err
		}
		if len(row) != length {
			return fmt.Errorf("row lengths don't align")
		}
		table.Append(row)
	}
	table.Render()
	return nil
}

func PrintAttributes(attrs map[string]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Attribute", "Value"})
	for k, v := range attrs {
		table.Append([]string{k, v})
	}
	table.Render()
}
