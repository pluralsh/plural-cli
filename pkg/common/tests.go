package common

import (
	"bytes"
	"io"
	"os"

	"github.com/urfave/cli"
)

func CaptureStdout(app *cli.App, arg []string) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := app.Run(arg)
	if err != nil {
		return "", err
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return "", err
	}
	return buf.String(), nil
}
