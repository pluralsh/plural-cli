package diff

import (
	"io"
	"os"
)

type TeeWriter struct {
	File io.Writer
}

func (tee *TeeWriter) Write(p []byte) (int, error) {
	os.Stdout.Write(p)
	return tee.File.Write(p)
}
