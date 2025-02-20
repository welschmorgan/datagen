package seed

import (
	"archive/zip"
	"bytes"
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/welschmorgan/datagen/pkg/cache"
	"github.com/welschmorgan/datagen/pkg/config"
	"github.com/welschmorgan/datagen/pkg/models"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var DEFAULT_SEED_SCHEMA *string = nil

type Seeder struct {
	seeds []Seed
	db    *sql.DB
}

func NewSeeder(db *sql.DB, seeds []Seed) *Seeder {
	return &Seeder{
		seeds: seeds,
		db:    db,
	}
}

func NewSeederFromConfig(db *sql.DB, c *config.Config) (*Seeder, error) {
	seeds := []Seed{}
	if DEFAULT_SEED_SCHEMA != nil {
		seeds = append(seeds, NewStdSeed(config.SeedTypeSchema, "schema", "resource", "assets/seed.sql", nil, nil, nil, nil, nil, NewQueryUploader(db, *DEFAULT_SEED_SCHEMA)))
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
		case config.SeedTypeSchema:
			return nil, fmt.Errorf("unsupported seed type '%s'", seed.Type)
		}
		if parser, _, err = NewParserFromDecl(seed.Parser); err != nil {
			return nil, err
		}
		uploader = NewBasicUploader(db, fmt.Sprintf("%s_prop", seed.PropTable), seed.PropType)
		var loc *models.Locale = &models.Locale{
			Id:   0,
			Name: seed.Locale,
		}
		seeds = append(seeds, NewStdSeed(seed.Type, seed.Name, seed.PropType, seed.Url, loc, cm, seed.ExtractFile, fetcher, parser, uploader))
	}
	return NewSeeder(db, seeds), nil
}

func (seeder *Seeder) Seed() error {
	errs := []error{}
	locales := []*models.Locale{}
	locale := func(name string) *models.Locale {
		for _, loc := range locales {
			if strings.EqualFold(name, loc.Name) {
				return loc
			}
		}
		return nil
	}
	archiveExts := []string{
		".zip",
		".7z",
		".tar",
		".tar.gz",
		".tar.bz",
		".tar.bz2",
		".tar.xz",
	}
	isArchive := func(path string) bool {
		ret := false
		for _, forbiddenExt := range archiveExts {
			if strings.EqualFold(forbiddenExt, filepath.Ext(filepath.Base(path))) {
				ret = true
				break
			}
		}
		return ret
	}
	for _, s := range seeder.seeds {
		slog.Warn(fmt.Sprintf("Seeding '%s'", s.Name()))
		var content []byte
		var err error
		var rows []*ParserRow
		var loc *models.Locale
		var path string
		if s.Locale() != nil {
			loc = locale(s.Locale().Name)
		}
		if isArchive(s.Url()) && s.ExtractFile() == nil {
			err = fmt.Errorf("cannot extract archive if ExtractFile property is not defined")
			slog.Error(fmt.Sprint(err), "url", s.Url())
			errs = append(errs, err)
			continue
		}
		if s.Fetcher() != nil {
			content, err = s.Fetcher().Fetch(s.Url(), &path)
			if err != nil {
				slog.Error("Failed to fetch seed", "err", err, "url", s.Url())
				errs = append(errs, err)
				continue
			}
			if isArchive(s.Url()) {
				if content, err = seeder.Extract(content, *s.ExtractFile()); err != nil {
					slog.Error("Failed to extract seed's archive", "err", err, "url", s.Url())
					errs = append(errs, err)
					continue
				}
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

			slog.Debug("Parsing seed", "name", s.Name())
			rows, err = s.Parser().Parse(loc, s.Url(), utf8Data)
			if err != nil {
				slog.Error("Failed to parse seed", "err", err, "url", s.Url())
				errs = append(errs, err)
				continue
			}
			// for id, row := range rows {
			// 	slog.Debug("Parsed row", "id", id, "value", row)
			// }
		}
		slog.Debug("Uploading seed", "name", s.Name())
		if s.Uploader() != nil {
			if err = s.Uploader().Upload(rows); err != nil {
				slog.Error("Failed to upload seed to DB", "err", err, "url", s.Url(), "numRows", len(rows))
				errs = append(errs, err)
				continue
			}
		}
		if s.Type() == config.SeedTypeSchema {
			// only load locales after seeding schema
			if locales, err = models.LoadLocales(seeder.db); err != nil {
				return err
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

func (seeder *Seeder) Extract(archiveContent []byte, fileNameToExtract string) (extractedContent []byte, err error) {
	slog.Debug("Extracting seed file", "name", fileNameToExtract)
	cacheFile, err := cache.GetCache().Create(fileNameToExtract, 0)
	if err != nil {
		return nil, err
	}
	defer cacheFile.Close()
	var r *zip.Reader
	if r, err = zip.NewReader(bytes.NewReader(archiveContent), int64(len(archiveContent))); err != nil {
		return nil, err
	}
	var found *zip.File
	for _, f := range r.File {
		if strings.EqualFold(f.Name, fileNameToExtract) {
			found = f
			break
		}
	}
	if found == nil {
		return nil, fmt.Errorf("couldn't find file '%s'", fileNameToExtract)
	}
	f, err := found.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open archive entry '%s', %s", fileNameToExtract, err)
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if _, err := cacheFile.Write(content); err != nil {
		return nil, err
	}
	return content, nil
}
