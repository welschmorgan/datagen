package seed

import (
	"bytes"
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"strings"

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
