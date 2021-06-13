package utils

import (
	"bufio"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"
	"fmt"
	"os"
	"syscall"
	"strings"
)

func ReadLine(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	color.New(color.Bold).Printf(prompt)
	res, err := reader.ReadString('\n')
	return strings.TrimSpace(string(res)), err
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
	pwd, err := terminal.ReadPassword(int(syscall.Stdin))
	return strings.TrimSpace(string(pwd)), err
}

func Warn(line string, args... interface{}) {
	color.New(color.FgYellow, color.Bold).Printf(line, args...)
}

func Success(line string, args... interface{}) {
	color.New(color.FgGreen, color.Bold).Printf(line, args...)
}

func Highlight(line string, args... interface{}) {
	color.New(color.Bold).Printf(line, args...)
}