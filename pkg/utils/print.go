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
	"sigs.k8s.io/yaml"
)

func ReadLine(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	_, _ = color.New(color.Bold).Printf(prompt) //nolint:govet
	res, err := reader.ReadString('\n')
	return strings.TrimSpace(res), err
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
	_, _ = color.New(color.Bold).Printf(prompt) //nolint:govet
	pwd, err := term.ReadPassword(syscall.Stdin)
	return strings.TrimSpace(string(pwd)), err
}

func Warn(line string, args ...interface{}) {
	_, _ = color.New(color.FgYellow, color.Bold).Fprintf(color.Error, line, args...)
}

func Success(line string, args ...interface{}) {
	_, _ = color.New(color.FgGreen, color.Bold).Printf(line, args...)
}

func Error(line string, args ...interface{}) {
	_, _ = color.New(color.FgRed, color.Bold).Fprintf(color.Error, line, args...)
}

func Highlight(line string, args ...interface{}) {
	_, _ = color.New(color.Bold).Printf(line, args...)
}

func HighlightError(err error) error {
	if err != nil {
		err = fmt.Errorf(color.New(color.FgRed, color.Bold).Sprint(err.Error())) //nolint:all
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

type Printer interface {
	PrettyPrint()
}

type jsonPrinter struct {
	i interface{}
}

func (this *jsonPrinter) PrettyPrint() {
	s, _ := json.MarshalIndent(this.i, "", "  ")
	fmt.Println(string(s))
}

type yamlPrinter struct {
	i interface{}
}

func (this *yamlPrinter) PrettyPrint() {
	s, _ := yaml.Marshal(this.i)
	fmt.Println(string(s))
}

func NewJsonPrinter(i interface{}) Printer {
	return &jsonPrinter{i: i}
}

func NewYAMLPrinter(i interface{}) Printer {
	return &yamlPrinter{i: i}
}
