package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

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

	// Fetch existing tags for LLM context (ignore error, empty list is fine)
	existingTags, _ := vault.GetAllTags()

	var item *store.Item

	switch raw.Type {
	case ContentTypeLink:
		item, err = p.processLink(ctx, raw, existingTags)
	case ContentTypeNote:
		item, err = p.processNote(ctx, raw, existingTags)
	case ContentTypeImage:
		item, err = p.processImage(ctx, raw, existingTags)
	case ContentTypeSearch:
		item, err = p.processSearch(ctx, raw, existingTags)
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

	// Find and create relationships with existing items (best-effort)
	p.findAndCreateRelationships(ctx, vault, item)

	return item, nil
}

func (p *Pipeline) processLink(ctx context.Context, raw RawContent, existingTags []string) (*store.Item, error) {
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
	processed, err := p.llmClient.ProcessContent(ctx, "web article", extracted.Content, raw.Language, existingTags)
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

func (p *Pipeline) processNote(ctx context.Context, raw RawContent, existingTags []string) (*store.Item, error) {
	processed, err := p.llmClient.ProcessContent(ctx, "note", raw.Text, raw.Language, existingTags)
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

func (p *Pipeline) processImage(ctx context.Context, raw RawContent, existingTags []string) (*store.Item, error) {
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

	slog.Info("saved image", "id", itemID, "path", imagePath, "size", len(raw.ImageData))

	// If caption exists, process through LLM
	if raw.Caption != "" {
		processed, err := p.llmClient.ProcessContent(ctx, "note with image", raw.Caption, raw.Language, existingTags)
		if err != nil {
			slog.Warn("LLM processing failed for image caption", "error", err)
			// Fallback: use caption as-is
			title := raw.Caption
			if len(title) > 100 {
				title = title[:100] + "..."
			}
			return &store.Item{
				ID:        itemID,
				Type:      store.ItemTypeImage,
				Title:     title,
				Content:   raw.Caption,
				ImagePath: imagePath,
				Tags:      []string{"image", "uncategorized"},
			}, nil
		}

		// Ensure "image" tag is always present
		tags := processed.Tags
		if !slices.Contains(tags, "image") {
			tags = append(tags, "image")
		}

		return &store.Item{
			ID:        itemID,
			Type:      store.ItemTypeImage,
			Title:     processed.Title,
			Summary:   processed.Summary,
			Content:   raw.Caption,
			ImagePath: imagePath,
			Tags:      tags,
		}, nil
	}

	// No caption: save image with minimal metadata
	return &store.Item{
		ID:        itemID,
		Type:      store.ItemTypeImage,
		Title:     "Image",
		ImagePath: imagePath,
		Tags:      []string{"image"},
	}, nil
}

func (p *Pipeline) processSearch(ctx context.Context, raw RawContent, existingTags []string) (*store.Item, error) {
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
	processed, err := p.llmClient.SummarizeSearchResults(ctx, topic, searchText, raw.Language, existingTags)
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

// findAndCreateRelationships finds related items and creates graph edges.
// This is best-effort; failures are logged but never fail ingestion.
func (p *Pipeline) findAndCreateRelationships(ctx context.Context, vault *store.VaultStore, item *store.Item) {
	// Fetch existing items (limit 200 for reasonable LLM context)
	allItems, err := vault.ListItems(200, 0)
	if err != nil {
		slog.Warn("failed to fetch items for relationships", "error", err)
		return
	}

	// Need at least one other item to create relationships
	if len(allItems) <= 1 {
		return
	}

	// Select up to 50 most relevant items (by tag overlap + recency)
	relevantItems := selectRelevantItems(allItems, item.Tags, item.ID)
	if len(relevantItems) == 0 {
		return
	}

	// Format items for LLM
	itemsContext := formatItemsForLLM(relevantItems)

	// Call LLM to find relationships
	suggestions, err := p.llmClient.FindRelationships(ctx, item.Title, item.Summary, item.Tags, itemsContext)
	if err != nil {
		slog.Warn("LLM relationship finding failed", "error", err)
		return
	}

	// Create relationships for strength >= 0.5
	created := createRelationshipsFromSuggestions(vault, item.ID, suggestions)
	if created > 0 {
		slog.Info("created relationships", "item_id", item.ID, "count", created)
	}
}

// formatItemsForLLM formats items for the relationship finding prompt.
func formatItemsForLLM(items []store.Item) string {
	var sb strings.Builder
	for _, item := range items {
		fmt.Fprintf(&sb, "ID: %s\nTitle: %s\nSummary: %s\nTags: %s\n\n",
			item.ID, item.Title, item.Summary, strings.Join(item.Tags, ", "))
	}
	return sb.String()
}

// selectRelevantItems filters and sorts items by relevance to the new item.
// Returns up to 50 items prioritized by tag overlap and recency.
func selectRelevantItems(allItems []store.Item, newItemTags []string, excludeID string) []store.Item {
	type scoredItem struct {
		item     store.Item
		tagScore int
	}

	// Build tag set for O(1) lookup
	tagSet := make(map[string]struct{}, len(newItemTags))
	for _, tag := range newItemTags {
		tagSet[strings.ToLower(tag)] = struct{}{}
	}

	// Score items by tag overlap
	var scored []scoredItem
	for _, item := range allItems {
		if item.ID == excludeID {
			continue
		}

		tagScore := 0
		for _, tag := range item.Tags {
			if _, ok := tagSet[strings.ToLower(tag)]; ok {
				tagScore++
			}
		}

		// Only include items with some tag overlap to avoid nonsensical relationships
		if tagScore > 0 {
			scored = append(scored, scoredItem{item: item, tagScore: tagScore})
		}
	}

	// Sort by tag score (desc), then by recency (already sorted desc from ListItems)
	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].tagScore > scored[j].tagScore
	})

	// Take top 50
	limit := min(50, len(scored))

	result := make([]store.Item, limit)
	for i := range limit {
		result[i] = scored[i].item
	}
	return result
}

// createRelationshipsFromSuggestions creates graph edges for suggestions with strength >= 0.7.
func createRelationshipsFromSuggestions(vault *store.VaultStore, sourceID string, suggestions []llm.RelationshipSuggestion) int {
	created := 0
	for _, s := range suggestions {
		if s.Strength < 0.7 {
			continue
		}

		rel := &store.Relationship{
			SourceID:     sourceID,
			TargetID:     s.TargetID,
			RelationType: s.RelationType,
			Strength:     s.Strength,
		}

		if err := vault.CreateRelationship(rel); err != nil {
			slog.Warn("failed to create relationship",
				"source", sourceID,
				"target", s.TargetID,
				"type", s.RelationType,
				"error", err)
			continue
		}
		created++
	}
	return created
}
