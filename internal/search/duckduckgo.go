package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a DuckDuckGo Instant Answer API client.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new DuckDuckGo search client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Result represents a search result from DuckDuckGo.
type Result struct {
	Abstract     string        // Main abstract/summary
	AbstractText string        // Plain text version
	AbstractURL  string        // URL to the source
	Source       string        // Source name (e.g., "Wikipedia")
	Heading      string        // Title/heading
	RelatedItems []RelatedItem // Related topics
}

// RelatedItem represents a related topic from DuckDuckGo.
type RelatedItem struct {
	Text string
	URL  string
}

// ddgResponse is the raw DuckDuckGo API response structure.
type ddgResponse struct {
	Abstract       string `json:"Abstract"`
	AbstractText   string `json:"AbstractText"`
	AbstractSource string `json:"AbstractSource"`
	AbstractURL    string `json:"AbstractURL"`
	Heading        string `json:"Heading"`
	RelatedTopics  []struct {
		Text     string `json:"Text,omitempty"`
		FirstURL string `json:"FirstURL,omitempty"`
		Topics   []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"Topics,omitempty"`
	} `json:"RelatedTopics"`
	Definition       string `json:"Definition"`
	DefinitionSource string `json:"DefinitionSource"`
	DefinitionURL    string `json:"DefinitionURL"`
}

// Search queries DuckDuckGo Instant Answer API for the given topic.
func (c *Client) Search(ctx context.Context, query string) (*Result, error) {
	// Build URL with query parameters
	apiURL := "https://api.duckduckgo.com/"
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("no_html", "1")
	params.Set("skip_disambig", "1")

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Dumper/1.0 (Knowledge capture bot)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var ddg ddgResponse
	if err := json.Unmarshal(body, &ddg); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	result := &Result{
		Abstract:     ddg.Abstract,
		AbstractText: ddg.AbstractText,
		AbstractURL:  ddg.AbstractURL,
		Source:       ddg.AbstractSource,
		Heading:      ddg.Heading,
	}

	// Use definition if no abstract available
	if result.Abstract == "" && ddg.Definition != "" {
		result.Abstract = ddg.Definition
		result.AbstractText = ddg.Definition
		result.AbstractURL = ddg.DefinitionURL
		result.Source = ddg.DefinitionSource
	}

	// Extract related topics (flatten nested structure)
	for _, rt := range ddg.RelatedTopics {
		if rt.Text != "" && rt.FirstURL != "" {
			result.RelatedItems = append(result.RelatedItems, RelatedItem{
				Text: rt.Text,
				URL:  rt.FirstURL,
			})
		}
		for _, topic := range rt.Topics {
			if topic.Text != "" && topic.FirstURL != "" {
				result.RelatedItems = append(result.RelatedItems, RelatedItem{
					Text: topic.Text,
					URL:  topic.FirstURL,
				})
			}
		}
	}

	// Limit related items
	if len(result.RelatedItems) > 5 {
		result.RelatedItems = result.RelatedItems[:5]
	}

	return result, nil
}

// HasContent returns true if the result has meaningful content.
func (r *Result) HasContent() bool {
	return r.Abstract != "" || r.AbstractText != "" || len(r.RelatedItems) > 0
}

// FormatForLLM formats the search result as text suitable for LLM processing.
func (r *Result) FormatForLLM() string {
	var sb strings.Builder

	if r.Heading != "" {
		fmt.Fprintf(&sb, "Topic: %s\n\n", r.Heading)
	}

	if r.Abstract != "" {
		fmt.Fprintf(&sb, "Summary: %s\n", r.Abstract)
		if r.Source != "" {
			fmt.Fprintf(&sb, "Source: %s\n", r.Source)
		}
		if r.AbstractURL != "" {
			fmt.Fprintf(&sb, "URL: %s\n", r.AbstractURL)
		}
		sb.WriteString("\n")
	}

	if len(r.RelatedItems) > 0 {
		sb.WriteString("Related Topics:\n")
		for _, item := range r.RelatedItems {
			fmt.Fprintf(&sb, "- %s\n", item.Text)
		}
	}

	return sb.String()
}
