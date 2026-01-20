package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
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
	explicitTitle := extractTitleFromNote(raw.Text)
	explicitTags := extractHashTags(raw.Text)

	processed, err := p.llmClient.ProcessContent(ctx, "note", raw.Text, raw.Language, existingTags)
	if err != nil {
		slog.Warn("LLM processing failed", "error", err)
		// Fallback: save as-is
		title := explicitTitle
		if title == "" {
			title = raw.Text
			if len(title) > 50 {
				title = title[:50] + "..."
			}
		}
		return &store.Item{
			Type:    store.ItemTypeNote,
			Title:   title,
			Content: raw.Text,
			Tags:    mergeTags([]string{"uncategorized"}, explicitTags),
		}, nil
	}

	title := processed.Title
	if explicitTitle != "" {
		title = explicitTitle
	}

	return &store.Item{
		Type:    store.ItemTypeNote,
		Title:   title,
		Summary: processed.Summary,
		Content: raw.Text,
		Tags:    mergeTags(processed.Tags, explicitTags),
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
		explicitTags := extractHashTags(raw.Caption)

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
				Tags:      mergeTags([]string{"image", "uncategorized"}, explicitTags),
			}, nil
		}

		// Ensure "image" tag is always present
		tags := mergeTags(processed.Tags, append(explicitTags, "image"))

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
	// Fetch recent items (limit 1000 for graph generation)
	allItems, err := vault.ListItems(1000, 0)
	if err != nil {
		slog.Warn("failed to fetch items for relationships", "error", err)
		return
	}

	// Need at least one other item to create relationships
	if len(allItems) <= 1 {
		return
	}

	created := 0

	// Build title index for wikilinks
	titleIndex := buildTitleIndex(allItems)
	newTitleKey := normalizeTitle(item.Title)

	linkedIDs := make(map[string]struct{})

	// Create explicit wikilink relationships from the new item to existing items
	for _, targetKey := range extractWikiLinkTargets(item.Content) {
		target, ok := titleIndex[targetKey]
		if !ok || target.ID == item.ID {
			continue
		}

		if err := vault.CreateRelationship(&store.Relationship{
			SourceID:     item.ID,
			TargetID:     target.ID,
			RelationType: "link",
			Strength:     1.0,
		}); err != nil {
			slog.Warn("failed to create link relationship",
				"source", item.ID,
				"target", target.ID,
				"error", err)
			continue
		}
		linkedIDs[target.ID] = struct{}{}
		created++
	}

	// Create wikilink relationships from existing items to the new item
	if newTitleKey != "" {
		for _, other := range allItems {
			if other.ID == item.ID {
				continue
			}
			if !itemLinksToTitle(other, newTitleKey) {
				continue
			}

			if err := vault.CreateRelationship(&store.Relationship{
				SourceID:     other.ID,
				TargetID:     item.ID,
				RelationType: "link",
				Strength:     1.0,
			}); err != nil {
				slog.Warn("failed to create back link relationship",
					"source", other.ID,
					"target", item.ID,
					"error", err)
				continue
			}
			linkedIDs[other.ID] = struct{}{}
			created++
		}
	}

	// Create shared tag relationships (skip pairs already connected by explicit links)
	newTags := filterGraphTags(normalizeTags(item.Tags))
	if len(newTags) > 0 {
		newTagSet := make(map[string]struct{}, len(newTags))
		for _, tag := range newTags {
			newTagSet[tag] = struct{}{}
		}

		for _, other := range allItems {
			if other.ID == item.ID {
				continue
			}
			if _, linked := linkedIDs[other.ID]; linked {
				continue
			}

			overlap := countSharedTags(newTagSet, filterGraphTags(normalizeTags(other.Tags)))
			if overlap == 0 {
				continue
			}

			sourceID, targetID := orderedPair(item.ID, other.ID)
			if err := vault.CreateRelationship(&store.Relationship{
				SourceID:     sourceID,
				TargetID:     targetID,
				RelationType: "tag",
				Strength:     tagOverlapStrength(overlap),
			}); err != nil {
				slog.Warn("failed to create tag relationship",
					"source", sourceID,
					"target", targetID,
					"error", err)
				continue
			}
			created++
		}
	}

	if created > 0 {
		slog.Info("created relationships", "item_id", item.ID, "count", created)
	}
}

var (
	wikiLinkPattern = regexp.MustCompile(`\[\[([^\[\]]+)\]\]`)
	hashTagPattern  = regexp.MustCompile(`(?:^|[\s])#([\p{L}\p{N}][\p{L}\p{N}_-]*(?:/[\p{L}\p{N}_-]+)*)`)
)

func extractTitleFromNote(text string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if !strings.HasPrefix(trimmed, "#") {
			return ""
		}

		hashCount := 0
		for hashCount < len(trimmed) && trimmed[hashCount] == '#' {
			hashCount++
		}
		if hashCount == 1 {
			title := strings.TrimSpace(trimmed[1:])
			if title != "" {
				return title
			}
		}
		return ""
	}
	return ""
}

func extractWikiLinkTargets(content string) []string {
	matches := wikiLinkPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(matches))
	var targets []string
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		target := normalizeWikiTarget(match[1])
		if target == "" {
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		targets = append(targets, target)
	}
	return targets
}

func normalizeWikiTarget(raw string) string {
	parts := strings.Split(raw, "|")
	base := parts[0]
	base = strings.Split(base, "#")[0]
	base = strings.TrimSpace(base)
	return normalizeTitle(base)
}

func extractHashTags(text string) []string {
	matches := hashTagPattern.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(matches))
	var tags []string
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		tag := normalizeTag(match[1])
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}
	return tags
}

func normalizeTitle(title string) string {
	if title == "" {
		return ""
	}
	title = strings.ToLower(strings.TrimSpace(title))
	fields := strings.Fields(title)
	return strings.Join(fields, " ")
}

func normalizeTag(tag string) string {
	return strings.ToLower(strings.TrimSpace(tag))
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = normalizeTag(tag)
		if tag == "" {
			continue
		}
		normalized = append(normalized, tag)
	}
	return normalized
}

func mergeTags(primary []string, secondary []string) []string {
	if len(primary) == 0 && len(secondary) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(primary)+len(secondary))
	var merged []string
	for _, tag := range append(primary, secondary...) {
		tag = normalizeTag(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		merged = append(merged, tag)
	}
	return merged
}

func filterGraphTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}

	ignore := map[string]struct{}{
		"uncategorized": {},
		"image":         {},
		"search":        {},
		"link":          {},
		"note":          {},
	}

	filtered := make([]string, 0, len(tags))
	for _, tag := range tags {
		if _, ok := ignore[tag]; ok {
			continue
		}
		filtered = append(filtered, tag)
	}
	return filtered
}

func countSharedTags(left map[string]struct{}, right []string) int {
	count := 0
	for _, tag := range right {
		if _, ok := left[tag]; ok {
			count++
		}
	}
	return count
}

func tagOverlapStrength(overlap int) float64 {
	if overlap <= 0 {
		return 0
	}
	strength := 0.4 + (0.15 * float64(overlap))
	if strength > 1.0 {
		return 1.0
	}
	return strength
}

func orderedPair(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}

func buildTitleIndex(items []store.Item) map[string]store.Item {
	index := make(map[string]store.Item, len(items))
	for _, item := range items {
		key := normalizeTitle(item.Title)
		if key == "" {
			continue
		}
		if _, exists := index[key]; exists {
			continue
		}
		index[key] = item
	}
	return index
}

func itemLinksToTitle(item store.Item, titleKey string) bool {
	if titleKey == "" || item.Content == "" {
		return false
	}

	for _, target := range extractWikiLinkTargets(item.Content) {
		if target == titleKey {
			return true
		}
	}
	return false
}
