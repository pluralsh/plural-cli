package format

import (
	"encoding/csv"
	"os"
)

type csvFormatter struct {
	writer *csv.Writer
}

func NewCsvFormatter() *csvFormatter {
	return &csvFormatter{writer: csv.NewWriter(os.Stdout)}
}

func (csv *csvFormatter) Write(line []string) error {
	return csv.writer.Write(line)
}

func (csv *csvFormatter) Dump(lines [][]string) error {
	return csv.writer.WriteAll(lines)
}

func (csv *csvFormatter) Flush() error {
	csv.writer.Flush()
	return nil
}

func (csv *csvFormatter) Header(line []string) {
	if err := csv.writer.Write(line); err != nil {
		return
	}
}
