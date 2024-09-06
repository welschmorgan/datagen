package seed

import (
	"bytes"
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/welschmorgan/datagen/pkg/config"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type Seeder []Seed

func NewSeeder(seeds []Seed) *Seeder {
	var s Seeder = seeds
	return &s
}

var DEFAULT_SEED_SCHEMA *string = nil

func NewDefaultSeeder(db *sql.DB) *Seeder {
	seeds := []Seed{}
	if DEFAULT_SEED_SCHEMA != nil {
		seeds = append(seeds, NewStdSeed("schema", "resource", "assets/seed.sql", nil, nil, nil, NewQueryUploader(db, *DEFAULT_SEED_SCHEMA)))
	}
	seeds = append(seeds, NewStdSeed("[fr] person.firstName", "firstName", "https://www.data.gouv.fr/fr/datasets/r/55cd803a-998d-4a5c-9741-4cd0ee0a7699", charmap.Windows1252, NewRemoteFetcher(), NewCSVParser(true, ";", 0), NewBasicUploader(db, "person_prop", "firstName")))
	return NewSeeder(seeds)
}

func NewSeederFromConfig(db *sql.DB, c *config.Config) (*Seeder, error) {
	seeds := []Seed{}
	if DEFAULT_SEED_SCHEMA != nil {
		seeds = append(seeds, NewStdSeed("schema", "resource", "assets/seed.sql", nil, nil, nil, NewQueryUploader(db, *DEFAULT_SEED_SCHEMA)))
	}
	charmap := func(name string) (*charmap.Charmap, error) {
		for _, enc := range charmap.All {
			cm, ok := enc.(*charmap.Charmap)
			if ok && strings.EqualFold(cm.String(), name) {
				return cm, nil
			}
		}
		return nil, fmt.Errorf("failed to find charmap '%s'", name)
	}
	for _, seed := range c.Seeds {
		cm, err := charmap(seed.Encoding)
		if err != nil {
			return nil, err
		}
		var fetcher Fetcher
		var parser Parser
		var uploader Uploader
		switch seed.Type {
		case config.SeedTypeRemote:
			fetcher = NewRemoteFetcher()
		case config.SeedTypeStatic:
			return nil, fmt.Errorf("unsupported seed type '%s'", seed.Type)
		}
		if parser, _, err = NewParserFromDecl(seed.Parser); err != nil {
			return nil, err
		}
		uploader = NewBasicUploader(db, fmt.Sprintf("%s_prop", seed.PropTable), seed.PropType)
		seeds = append(seeds, NewStdSeed(seed.Name, seed.PropType, seed.Url, cm, fetcher, parser, uploader))
	}
	return NewSeeder(seeds), nil
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

func (s *Seeder) Seed() error {
	errs := []error{}
	for _, s := range *s {
		slog.Warn(fmt.Sprintf("Seeding '%s' from %s", s.Name(), s.Url()))
		var content []byte
		var err error
		var rows []ParserRow
		if s.Fetcher() != nil {
			content, err = s.Fetcher().Fetch(s.Url())
			if err != nil {
				slog.Error("Failed to fetch seed", "err", err, "url", s.Url())
				errs = append(errs, err)
				continue
			}
		}
		if s.Parser() != nil {
			var utf8Data []byte
			if s.Encoding() != nil {
				utf8Reader := transform.NewReader(bytes.NewReader(content), s.Encoding().NewDecoder())
				if utf8Data, err = io.ReadAll(utf8Reader); err != nil {
					slog.Error("Failed to convert seed to utf-8", "err", err, "url", s.Url())
					errs = append(errs, err)
					continue
				}
			} else {
				utf8Data = content
			}

			rows, err = s.Parser().Parse(s.Url(), utf8Data)
			if err != nil {
				slog.Error("Failed to parse seed", "err", err, "url", s.Url())
				errs = append(errs, err)
				continue
			}
			// for id, row := range rows {
			// 	slog.Debug("Parsed row", "id", id, "value", row)
			// }
		}
		if s.Uploader() != nil {
			if err = s.Uploader().Upload(rows); err != nil {
				slog.Error("Failed to upload seed to DB", "err", err, "url", s.Url(), "numRows", len(rows))
				errs = append(errs, err)
				continue
			}
		}
	}
	if len(errs) > 0 {
		msg := []string{}
		for _, e := range errs {
			msg = append(msg, fmt.Sprint(e))
		}
		return fmt.Errorf("there were errors while seeding DB:\n%s", strings.Join(msg, "\n - "))
	}
	return nil
}
