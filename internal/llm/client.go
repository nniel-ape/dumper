package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://openrouter.ai/api/v1",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	req := ChatRequest{
		Model:    c.model,
		Messages: messages,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("HTTP-Referer", "https://github.com/dumper")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("api error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (c *Client) ProcessContent(ctx context.Context, contentType, content, lang string, existingTags []string) (*ProcessedContent, error) {
	// Truncate content if too long (preserve first ~8000 chars)
	if len(content) > 8000 {
		content = content[:8000] + "..."
	}

	tagsContext := formatExistingTags(existingTags)
	prompt := fmt.Sprintf(ProcessContentPrompt, tagsContext, contentType, content)

	// Add language instruction for non-English
	if lang == "ru" {
		prompt += "\n\nIMPORTANT: Generate the title and summary in Russian (русский язык)."
	}

	response, err := c.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("chat: %w", err)
	}

	// Clean response - sometimes LLM adds markdown code blocks
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var result ProcessedContent
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("parse response: %w (raw: %s)", err, response)
	}
	return &result, nil
}

func (c *Client) AnswerQuestion(ctx context.Context, question string, items []string) (string, error) {
	itemsStr := strings.Join(items, "\n\n---\n\n")
	prompt := fmt.Sprintf(AnswerQuestionPrompt, question, itemsStr)

	response, err := c.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return "", fmt.Errorf("chat: %w", err)
	}

	return strings.TrimSpace(response), nil
}

func (c *Client) FindRelationships(ctx context.Context, title, summary string, tags []string, existingItems string) ([]RelationshipSuggestion, error) {
	prompt := fmt.Sprintf(FindRelationshipsPrompt, title, summary, tags, existingItems)

	response, err := c.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("chat: %w", err)
	}

	// Clean response
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var suggestions []RelationshipSuggestion
	if err := json.Unmarshal([]byte(response), &suggestions); err != nil {
		return nil, fmt.Errorf("parse response: %w (raw: %s)", err, response)
	}
	return suggestions, nil
}

// SummarizeSearchResults creates a knowledge entry from search results about a topic.
func (c *Client) SummarizeSearchResults(ctx context.Context, topic, searchResults, lang string, existingTags []string) (*ProcessedContent, error) {
	tagsContext := formatExistingTags(existingTags)
	prompt := fmt.Sprintf(SummarizeSearchPrompt, tagsContext, topic, searchResults)

	// Add language instruction for non-English
	if lang == "ru" {
		prompt += "\n\nIMPORTANT: Generate the title and summary in Russian (русский язык)."
	}

	response, err := c.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("chat: %w", err)
	}

	// Clean response
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var result ProcessedContent
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("parse response: %w (raw: %s)", err, response)
	}
	return &result, nil
}

// formatExistingTags formats existing tags for inclusion in prompts.
// Returns empty string if no tags, otherwise a formatted context block.
func formatExistingTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return fmt.Sprintf("\nExisting tags in user's knowledge base (prefer reusing when appropriate):\n%s\n", strings.Join(tags, ", "))
}
