package seed

import "golang.org/x/text/encoding/charmap"

type StdSeed struct {
	Seed

	name     string
	typ      string
	url      string
	encoding *charmap.Charmap

	fetcher  Fetcher
	parser   Parser
	uploader Uploader
}

func NewStdSeed(name, typ, url string, encoding *charmap.Charmap, fetcher Fetcher, parser Parser, uploader Uploader) *StdSeed {
	return &StdSeed{
		name:     name,
		typ:      typ,
		url:      url,
		encoding: encoding,

		fetcher:  fetcher,
		parser:   parser,
		uploader: uploader,
	}
}

func (s *StdSeed) Name() string {
	return s.name
}
func (s *StdSeed) Type() string {
	return s.typ
}
func (s *StdSeed) Url() string {
	return s.url
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
