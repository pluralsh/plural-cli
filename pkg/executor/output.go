package executor

import (
	"io"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/utils"
)

type OutputWriter struct {
	Delegate    io.WriteCloser
	UseDelegate bool
	lines       []string
}

func (out *OutputWriter) Write(line []byte) (int, error) {
	if out.UseDelegate {
		return out.Delegate.Write(line)
	}

	out.lines = append(out.lines, string(line))
	utils.LogInfo().Println(string(line))
	if !utils.EnableDebug {
		_, err := out.Delegate.Write([]byte("."))
		if err != nil {
			return 0, err
		}
	}

	return len(line), nil
}

func (out *OutputWriter) Close() error {
	return out.Delegate.Close()
}

func (out *OutputWriter) Format() string {
	return strings.Join(out.lines, "")
}
