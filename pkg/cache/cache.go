package cache

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/kirsle/configdir"
)

const CACHE_EXPIRATION_DELAY = time.Hour * 1

func CacheDir() string {
	return configdir.LocalCache("datagen/downloads")
}

func CacheFile(name string) string {
	nonWordsPattern := regexp.MustCompile(`\W+`)
	return fmt.Sprintf("%s/%s", CacheDir(), nonWordsPattern.ReplaceAllString(name, "_"))
}

type Cache struct{}

const FAIL_IF_EXPIRED int = 1 << 0

type Stat struct {
	path        string
	lastModTime time.Time
	expiresAt   time.Time
	isExpired   bool
}

func (c *Cache) assertParentExists(name string) error {
	cachePath := CacheFile(name)
	cacheDir := filepath.Dir(cachePath)
	_, cacheDirExistsErr := os.Stat(cacheDir)
	if errors.Is(cacheDirExistsErr, fs.ErrNotExist) {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cache) Stat(name string) (*Stat, error) {
	cachePath := CacheFile(name)
	fi, err := os.Stat(cachePath)
	if err != nil {
		slog.Debug(fmt.Sprintf("Stat '%s'", name), "err", err)
		return nil, err
	}
	var lastModTime time.Time = fi.ModTime().Local()
	var expiresAt time.Time = lastModTime.Add(CACHE_EXPIRATION_DELAY)
	slog.Debug(fmt.Sprintf("Stat '%s'", name), "lastModTime", lastModTime, "expiresAt", expiresAt)
	return &Stat{
		path:        cachePath,
		lastModTime: lastModTime,
		expiresAt:   expiresAt,
		isExpired:   time.Now().Local().After(expiresAt),
	}, nil
}

func (c *Cache) Open(name string, flags int) (*os.File, error) {
	path := CacheFile(name)
	slog.Debug("Opening cache file for reading", "name", name, "dir", filepath.Dir(path))
	st, err := c.Stat(name)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to open cache file '%s', %s", path, err)
	} else if (flags&FAIL_IF_EXPIRED) != 0 && st != nil && st.isExpired {
		return nil, fmt.Errorf("failed to open cache file '%s', expired at %s", path, st.expiresAt)
	}
	return os.Open(path)
}

func (c *Cache) Create(name string, flags int) (*os.File, error) {
	c.assertParentExists(name)
	path := CacheFile(name)
	slog.Debug("Opening cache file for writing", "name", name, "dir", filepath.Dir(path))
	st, err := c.Stat(name)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to create cache file '%s', %s", path, err)
	} else if (flags&FAIL_IF_EXPIRED) != 0 && st != nil && !st.isExpired {
		return nil, fmt.Errorf("failed to create cache file '%s', not expired yet %s", path, st.expiresAt)
	}
	return os.Create(path)
}

func (c *Cache) OpenOrRetrieve(name string, retriever func() (*bytes.Buffer, error)) (*os.File, error) {
	st, err := c.Stat(name)
	bustNow := false
	if errors.Is(err, os.ErrNotExist) {
		bustNow = true
	} else if st != nil && st.isExpired {
		bustNow = true
	}
	if bustNow {
		content, err := retriever()
		if err != nil {
			return nil, err
		}
		f, err := c.Create(name, FAIL_IF_EXPIRED)
		if err != nil {
			return nil, err
		}
		r := bytes.NewReader(content.Bytes())
		if _, err := io.Copy(f, r); err != nil {
			return nil, err
		}
	}
	return c.Open(name, FAIL_IF_EXPIRED)
}

var cache Cache = Cache{}

func GetCache() *Cache {
	return &cache
}
