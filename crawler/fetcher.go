package crawler

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Fetcher handles downloading web pages over HTTP
type Fetcher struct {
	Client *http.Client
}

// NewFetcher creates a Fetcher with sane timeout defaults
func NewFetcher() *Fetcher {
	return &Fetcher{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Fetch downloads the content at the given URL and returns it as a string.
// It returns an error if the request fails or the server responds with a bad status code.
func (f *Fetcher) Fetch(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Identify our crawler politely — many sites block requests with no User-Agent
	req.Header.Set("User-Agent", "GoSearchBot/1.0 (+https://github.com/Gout-ham24/gosearch)")

	resp, err := f.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status for %s: %d %s", url, resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body of %s: %w", url, err)
	}

	return string(body), nil
}
