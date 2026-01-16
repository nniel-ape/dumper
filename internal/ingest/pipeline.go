package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/nerdneilsfield/dumper/internal/llm"
	"github.com/nerdneilsfield/dumper/internal/search"
	"github.com/nerdneilsfield/dumper/internal/store"
)

type Pipeline struct {
	extractor    *Extractor
	llmClient    *llm.Client
	searchClient *search.Client
	stores       *store.Manager
}

func NewPipeline(llmClient *llm.Client, searchClient *search.Client, stores *store.Manager) *Pipeline {
	return &Pipeline{
		extractor:    NewExtractor(),
		llmClient:    llmClient,
		searchClient: searchClient,
		stores:       stores,
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
	case ContentTypeImage:
		item, err = p.processImage(ctx, raw)
	case ContentTypeSearch:
		item, err = p.processSearch(ctx, raw)
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
	processed, err := p.llmClient.ProcessContent(ctx, "web article", extracted.Content, raw.Language)
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
	processed, err := p.llmClient.ProcessContent(ctx, "note", raw.Text, raw.Language)
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

func (p *Pipeline) processImage(ctx context.Context, raw RawContent) (*store.Item, error) {
	itemID := uuid.NewString()

	// Create images directory under user folder
	userDir := p.stores.UserDir(raw.UserID)
	imagesDir := filepath.Join(userDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return nil, fmt.Errorf("create images dir: %w", err)
	}

	// Write image file
	imagePath := fmt.Sprintf("images/%s.%s", itemID, raw.ImageExt)
	fullPath := filepath.Join(userDir, imagePath)
	if err := os.WriteFile(fullPath, raw.ImageData, 0644); err != nil {
		return nil, fmt.Errorf("write image: %w", err)
	}

	// Use caption as title if available
	title := "Image"
	if raw.Caption != "" {
		title = raw.Caption
		if len(title) > 100 {
			title = title[:100] + "..."
		}
	}

	slog.Info("saved image", "id", itemID, "path", imagePath, "size", len(raw.ImageData))

	return &store.Item{
		ID:        itemID,
		Type:      store.ItemTypeImage,
		Title:     title,
		Content:   raw.Caption,
		ImagePath: imagePath,
		Tags:      []string{"image"},
	}, nil
}

func (p *Pipeline) processSearch(ctx context.Context, raw RawContent) (*store.Item, error) {
	topic := raw.Text

	// Search DuckDuckGo
	searchResult, err := p.searchClient.Search(ctx, topic)
	if err != nil {
		slog.Warn("search failed, using LLM knowledge", "topic", topic, "error", err)
		searchResult = &search.Result{} // Empty result, LLM will use general knowledge
	}

	// Format search results for LLM
	searchText := searchResult.FormatForLLM()
	if searchText == "" {
		searchText = "(No search results found)"
	}

	// Summarize with LLM
	processed, err := p.llmClient.SummarizeSearchResults(ctx, topic, searchText, raw.Language)
	if err != nil {
		slog.Warn("LLM summarization failed", "error", err)
		// Fallback: save raw search result
		title := topic
		summary := searchResult.Abstract
		if summary == "" {
			summary = fmt.Sprintf("Search result for: %s", topic)
		}
		return &store.Item{
			Type:    store.ItemTypeSearch,
			URL:     searchResult.AbstractURL,
			Title:   title,
			Summary: summary,
			Content: searchText,
			Tags:    []string{"search", "uncategorized"},
		}, nil
	}

	return &store.Item{
		Type:    store.ItemTypeSearch,
		URL:     searchResult.AbstractURL,
		Title:   processed.Title,
		Summary: processed.Summary,
		Content: searchText,
		Tags:    processed.Tags,
	}, nil
}
