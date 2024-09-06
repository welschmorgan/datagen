package seed

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/welschmorgan/datagen/pkg/cache"
)

type Fetcher interface {
	Fetch(url string) ([]byte, error)
}

type RemoteFetcher struct {
	Fetcher
}

func NewRemoteFetcher() *RemoteFetcher {
	return &RemoteFetcher{}
}

func (f *RemoteFetcher) Fetch(url string) ([]byte, error) {
	slog.Debug(fmt.Sprintf("Fetching %s", url))
	file, err := cache.GetCache().OpenOrRetrieve(url, func() (*bytes.Buffer, error) {
		var buffer *bytes.Buffer = bytes.NewBuffer([]byte{})
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch remote url '%s', %s", url, err)
		}
		defer resp.Body.Close()
		// cacheFile := os.OpenFile()
		if _, err := io.Copy(buffer, resp.Body); err != nil {
			return nil, fmt.Errorf("failed to write data to buffer, %s", err)
		}
		return buffer, nil
	})
	if err != nil {
		return nil, err
	}
	return io.ReadAll(file)
}

type StaticFetcher struct {
	Fetcher
	content []byte
}

func NewStaticFetcher(data []byte) *StaticFetcher {
	return &StaticFetcher{content: data}
}

func (f *StaticFetcher) Fetch(url string) ([]byte, error) {
	return f.content, nil
}
