package seed

import (
	"github.com/welschmorgan/datagen/pkg/config"
	"github.com/welschmorgan/datagen/pkg/models"
	"golang.org/x/text/encoding/charmap"
)

type StdSeed struct {
	Seed

	name          string
	typ           config.SeedType
	propType      string
	url           string
	locale        *models.Locale
	encoding      *charmap.Charmap
	extractedName *string

	fetcher  Fetcher
	parser   Parser
	uploader Uploader
}

func NewStdSeed(typ config.SeedType, name, propTyp, url string, locale *models.Locale, encoding *charmap.Charmap, extractedName *string, fetcher Fetcher, parser Parser, uploader Uploader) *StdSeed {
	return &StdSeed{
		name:          name,
		typ:           typ,
		propType:      propTyp,
		url:           url,
		locale:        locale,
		encoding:      encoding,
		extractedName: extractedName,
		fetcher:       fetcher,
		parser:        parser,
		uploader:      uploader,
	}
}

func (s *StdSeed) Name() string {
	return s.name
}
func (s *StdSeed) Type() config.SeedType {
	return s.typ
}
func (s *StdSeed) PropType() string {
	return s.propType
}
func (s *StdSeed) Url() string {
	return s.url
}
func (s *StdSeed) Locale() *models.Locale {
	return s.locale
}
func (s *StdSeed) Encoding() *charmap.Charmap {
	return s.encoding
}

func (s *StdSeed) Fetcher() Fetcher {
	return s.fetcher
}
func (s *StdSeed) Parser() Parser {
	return s.parser
}
func (s *StdSeed) Uploader() Uploader {
	return s.uploader
}
func (s *StdSeed) ExtractedName() *string {
	return s.extractedName
}
