package seed

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/welschmorgan/datagen/pkg/models"
)

type ParserRow struct {
	locale  *models.Locale
	value   string
	context map[string]string
}

func NewParserRow(locale *models.Locale, value string, context map[string]string) *ParserRow {
	return &ParserRow{
		locale:  locale,
		value:   value,
		context: context,
	}
}

type Parser interface {
	Parse(locale *models.Locale, url string, data []byte) ([]ParserRow, error)
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

func (p *CSVParser) Parse(locale *models.Locale, url string, data []byte) ([]ParserRow, error) {
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
		ret = append(ret, *NewParserRow(locale, cells[p.desiredColumn], nil))
	}
	return ret, nil
}

type ParserArg int64

const (
	ParserArgUnknown ParserArg = iota
	ParserSkipHeader
	ParserDelim
	ParserColumn
	ParserMax
)

func (a ParserArg) String() string {
	switch a {
	case ParserSkipHeader:
		return "skip_header"
	case ParserDelim:
		return "delim"
	case ParserColumn:
		return "column"
	}
	return "unknown"
}

func NewParserFromDecl(s string) (Parser, map[ParserArg]string, error) {
	ret := map[ParserArg]string{}
	pstr := strings.ToLower(s)
	pargs := []string{
		"",
		"",
	}
	if pos := strings.Index(pstr, "("); pos != -1 {
		pargs[0] = strings.TrimSpace(pstr[0:pos])
		rest := strings.TrimSpace(pstr[pos+1:])
		if pos = strings.Index(rest, ")"); pos != -1 {
			pargs[1] = strings.TrimSpace(rest[0:pos])
		}
	}
	var parser Parser
	switch strings.ToLower(pargs[0]) {
	case "csv":
		inQuote := false
		var quoteCh rune = 0
		realArgs := []string{}
		accu := ""
		for _, ch := range pargs[1] {
			switch ch {
			case '"', '\'':
				accu += fmt.Sprintf("%c", ch)
				if !inQuote {
					inQuote = !inQuote
					quoteCh = ch
				} else if quoteCh == ch {
					inQuote = !inQuote
					quoteCh = ch
				}
			case ',', ')':
				if !inQuote {
					realArgs = append(realArgs, strings.TrimSpace(accu))
					accu = ""
				} else {
					accu += fmt.Sprintf("%c", ch)
				}
			default:
				accu += fmt.Sprintf("%c", ch)
			}
		}
		if inQuote {
			return nil, nil, fmt.Errorf("unterminated quote: %s", pstr)
		}
		accu = strings.TrimSpace(accu)
		if len(accu) > 0 {
			realArgs = append(realArgs, accu)
		}
		for _, arg := range realArgs {
			parts := strings.SplitN(arg, "=", 2)
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			found := ParserArgUnknown
			for knownArg := range ParserMax {
				if strings.EqualFold(parts[0], knownArg.String()) {
					found = knownArg
					break
				}
			}
			if found == ParserArgUnknown {
				return nil, nil, fmt.Errorf("unknown parser argument '%s'", parts[0])
			}
			if len(parts) > 1 {
				ret[found] = parts[1]
			} else {
				ret[found] = ""
			}
		}
		_, header := ret[ParserSkipHeader]
		delim := ret[ParserDelim]
		colstr := ret[ParserColumn]
		col, err := strconv.ParseInt(colstr, 10, 32)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid desired column '%s', %s", colstr, err)
		}
		parser = NewCSVParser(header, delim, int(col))
	default:
		return nil, nil, fmt.Errorf("unsupported parser type '%s' (full = '%s')", pargs[0], pstr)
	}
	return parser, ret, nil
}
