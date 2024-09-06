package seed

import (
	"fmt"
	"log/slog"
	"strings"
)

type ParserRow struct {
	value   string
	context map[string]string
}

func NewParserRow(value string, context map[string]string) *ParserRow {
	return &ParserRow{
		value:   value,
		context: context,
	}
}

type Parser interface {
	Parse(url string, data []byte) ([]ParserRow, error)
}

type CSVParser struct {
	Parser

	skipHeader    bool
	delimiter     string
	desiredColumn int
}

func NewCSVParser(header bool, delimiter string, desiredColumn int) *CSVParser {
	return &CSVParser{
		skipHeader:    header,
		delimiter:     delimiter,
		desiredColumn: desiredColumn,
	}
}

func (p *CSVParser) Parse(url string, data []byte) ([]ParserRow, error) {
	lines := strings.Split(string(data), "\n")
	gotHeader := false
	ret := []ParserRow{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if p.skipHeader && !gotHeader {
			gotHeader = true
			slog.Debug(fmt.Sprintf("Header of CSV is '%s'", line))
			continue
		}
		cells := strings.Split(line, p.delimiter)
		if p.desiredColumn >= len(cells) {
			return nil, fmt.Errorf("invalid data fetched from '%s', desired column #%d cannot be accessed (only %d available)", url, p.desiredColumn, len(cells))
		}
		ret = append(ret, *NewParserRow(cells[p.desiredColumn], nil))
	}
	return ret, nil
}
