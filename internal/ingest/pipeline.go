package ingest

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nerdneilsfield/dumper/internal/llm"
	"github.com/nerdneilsfield/dumper/internal/store"
)

type Pipeline struct {
	extractor *Extractor
	llmClient *llm.Client
	stores    *store.Manager
}

func NewPipeline(llmClient *llm.Client, stores *store.Manager) *Pipeline {
	return &Pipeline{
		extractor: NewExtractor(),
		llmClient: llmClient,
		stores:    stores,
	}
}

func (p *Pipeline) Process(ctx context.Context, raw RawContent) (*store.Item, error) {
	vault, err := p.stores.GetVault(raw.UserID)
	if err != nil {
		return nil, fmt.Errorf("get vault: %w", err)
	}

	var item *store.Item

	switch raw.Type {
	case ContentTypeLink:
		item, err = p.processLink(ctx, raw)
	case ContentTypeNote:
		item, err = p.processNote(ctx, raw)
	default:
		return nil, fmt.Errorf("unknown content type: %s", raw.Type)
	}

	if err != nil {
		return nil, err
	}

	if err := vault.CreateItem(item); err != nil {
		return nil, fmt.Errorf("save item: %w", err)
	}

	slog.Info("processed item", "id", item.ID, "title", item.Title, "tags", item.Tags)

	return item, nil
}

func (p *Pipeline) processLink(ctx context.Context, raw RawContent) (*store.Item, error) {
	extracted, err := p.extractor.Extract(ctx, raw.URL)
	if err != nil {
		slog.Warn("extraction failed, using basic info", "url", raw.URL, "error", err)
		// Fallback: save with just URL
		return &store.Item{
			Type:    store.ItemTypeLink,
			URL:     raw.URL,
			Title:   raw.URL,
			Content: raw.Text,
			Tags:    []string{"uncategorized"},
		}, nil
	}

	// Process with LLM
	processed, err := p.llmClient.ProcessContent(ctx, "web article", extracted.Content)
	if err != nil {
		slog.Warn("LLM processing failed", "error", err)
		return &store.Item{
			Type:       store.ItemTypeLink,
			URL:        raw.URL,
			Title:      extracted.Title,
			Content:    extracted.Excerpt,
			RawContent: extracted.Content,
			Tags:       []string{"uncategorized"},
		}, nil
	}

	return &store.Item{
		Type:       store.ItemTypeLink,
		URL:        raw.URL,
		Title:      processed.Title,
		Summary:    processed.Summary,
		Content:    extracted.Excerpt,
		RawContent: extracted.Content,
		Tags:       processed.Tags,
	}, nil
}

func (p *Pipeline) processNote(ctx context.Context, raw RawContent) (*store.Item, error) {
	processed, err := p.llmClient.ProcessContent(ctx, "note", raw.Text)
	if err != nil {
		slog.Warn("LLM processing failed", "error", err)
		// Fallback: save as-is
		title := raw.Text
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		return &store.Item{
			Type:    store.ItemTypeNote,
			Title:   title,
			Content: raw.Text,
			Tags:    []string{"uncategorized"},
		}, nil
	}

	return &store.Item{
		Type:    store.ItemTypeNote,
		Title:   processed.Title,
		Summary: processed.Summary,
		Content: raw.Text,
		Tags:    processed.Tags,
	}, nil
}
