package calendar

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Fetcher handles retrieving calendar data from various sources
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a new calendar fetcher
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Fetch retrieves calendar data from a feed source
func (f *Fetcher) Fetch(feed Feed) ([]byte, error) {
	if feed.IsURL {
		return f.fetchURL(feed.Source)
	}
	return f.fetchFile(feed.Source)
}

// fetchURL retrieves calendar data from a URL
func (f *Fetcher) fetchURL(url string) ([]byte, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// fetchFile retrieves calendar data from a local file
func (f *Fetcher) fetchFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}
