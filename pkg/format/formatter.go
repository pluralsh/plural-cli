package format

type Formatter interface {
	Header(line []string)
	Write(line []string) error
	Dump(lines [][]string) error
	Flush() error
}

type FormatType string

const (
	CsvFormat   FormatType = "csv"
	TableFormat FormatType = "table"
)

func New(format FormatType) Formatter {
	switch format {
	case CsvFormat:
		return NewCsvFormatter()
	default:
		return NewTableFormatter()
	}
}