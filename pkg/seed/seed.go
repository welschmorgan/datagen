package seed

import (
	"github.com/welschmorgan/datagen/pkg/config"
	"github.com/welschmorgan/datagen/pkg/models"
	"golang.org/x/text/encoding/charmap"
)

type Seed interface {
	Name() string
	Type() config.SeedType
	PropType() string
	ExtractFile() *string
	Url() string
	Locale() *models.Locale
	Encoding() *charmap.Charmap

	Fetcher() Fetcher
	Parser() Parser
	Uploader() Uploader
}
