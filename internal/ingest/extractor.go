package ingest

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-shiori/go-readability"
)

type Extractor struct {
	client *http.Client
}

func NewExtractor() *Extractor {
	return &Extractor{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type ExtractedContent struct {
	URL      string
	Title    string
	Content  string
	Excerpt  string
	SiteName string
	Favicon  string
}

func (e *Extractor) Extract(ctx context.Context, rawURL string) (*ExtractedContent, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Dumper/1.0; +https://github.com/dumper)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	article, err := readability.FromReader(resp.Body, parsed)
	if err != nil {
		return nil, fmt.Errorf("parse content: %w", err)
	}

	return &ExtractedContent{
		URL:      rawURL,
		Title:    article.Title,
		Content:  article.TextContent,
		Excerpt:  article.Excerpt,
		SiteName: article.SiteName,
		Favicon:  fmt.Sprintf("%s://%s/favicon.ico", parsed.Scheme, parsed.Host),
	}, nil
}

func IsURL(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
