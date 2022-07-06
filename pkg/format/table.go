package format

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

type tableFormatter struct {
	writer *tablewriter.Table
}

func NewTableFormatter() *tableFormatter {
	return &tableFormatter{writer: tablewriter.NewWriter(os.Stdout)}
}

func (f *tableFormatter) Write(line []string) error {
	f.writer.Append(line)
	return nil
}

func (f *tableFormatter) Dump(lines [][]string) error {
	for _, line := range lines {
		f.writer.Append(line)
	}

	return nil
}

func (f *tableFormatter) Flush() error {
	f.writer.Render()
	return nil
}

func (f *tableFormatter) Header(line []string) {
	f.writer.SetHeader(line)
}
