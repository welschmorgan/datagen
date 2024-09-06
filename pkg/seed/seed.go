package seed

import "golang.org/x/text/encoding/charmap"

type Seed interface {
	Name() string
	Type() string
	Url() string
	Encoding() *charmap.Charmap

	Fetcher() Fetcher
	Parser() Parser
	Uploader() Uploader
}
