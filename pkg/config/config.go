package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/kirsle/configdir"
	"golang.org/x/text/encoding/charmap"
)

type SeedType int64

const (
	SeedTypeUnknown SeedType = iota
	SeedTypeSchema
	SeedTypeRemote
	SeedTypeMax
)

func (st SeedType) String() string {
	switch st {
	case SeedTypeSchema:
		return "schema"
	case SeedTypeRemote:
		return "remote"
	default:
		return "unknown"
	}
}

func (st SeedType) MarshalJSON() ([]byte, error) {
	return json.Marshal(st.String())
}

func (st *SeedType) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" || string(data) == `""` {
		*st = SeedTypeUnknown
		return nil
	}
	data = bytes.ReplaceAll(data, []byte("\""), []byte(""))
	// Fractional seconds are handled implicitly by Parse.
	for i := range SeedTypeMax {
		t := SeedType(i)
		if strings.EqualFold(t.String(), string(data)) {
			*st = t
			return nil
		}
	}
	return fmt.Errorf("invalid seed type %s", string(data))
}

type SeedConfig struct {
	Type      SeedType
	Name      string
	PropTable string
	PropType  string
	Url       string
	Encoding  string
	Locale    string
	Parser    string
}

type Config struct {
	Seeds []SeedConfig
}

var defaultConfig *Config = &Config{
	Seeds: []SeedConfig{
		{
			Type:      SeedTypeRemote,
			Name:      "[fr] person.firstName",
			PropTable: "person",
			PropType:  "firstName",
			Url:       "https://www.data.gouv.fr/fr/datasets/r/55cd803a-998d-4a5c-9741-4cd0ee0a7699",
			Locale:    "fr-FR",
			Encoding:  charmap.Windows1252.String(),
			Parser:    "csv(skip_header,delim=;,column=0)",
		},
	},
}

func New(seeds []SeedConfig) *Config {
	return &Config{
		Seeds: seeds,
	}
}

func Default() *Config {
	return defaultConfig
}

func DefaultPath() string {
	return configdir.LocalConfig("datagen/config.json")
}

func (c *Config) Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to load config from '%s', %s", path, err)
	}
	content, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to load config from '%s', %s", path, err)
	}
	if err := json.Unmarshal(content, c); err != nil {
		return fmt.Errorf("failed to load config from '%s', %s", path, err)
	}
	return nil
}

func (c *Config) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to save config to '%s', %s", path, err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to save config to '%s', %s", path, err)
	}
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to save config to '%s', %s", path, err)
	}
	return nil
}

func (c *Config) Init(path string) error {
	cfg := Default()
	dir := filepath.Dir(path)
	var err error
	if _, err = os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	if _, err = os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		slog.Info("Creating user configuration", "path", path)
		err = cfg.Save(path)
	} else {
		slog.Info("Loading user configuration", "path", path)
		err = cfg.Load(path)
	}
	return err
}
