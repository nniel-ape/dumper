package export

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/nerdneilsfield/dumper/internal/store"
)

type ObsidianExporter struct{}

func NewObsidianExporter() *ObsidianExporter {
	return &ObsidianExporter{}
}

func (e *ObsidianExporter) Export(vault *store.VaultStore) (io.Reader, error) {
	items, relationships, err := vault.GetGraph()
	if err != nil {
		return nil, fmt.Errorf("get graph: %w", err)
	}

	// Build relationship map for wikilinks
	relMap := make(map[string][]string)
	for _, r := range relationships {
		if r.RelationType != "link" {
			continue
		}
		relMap[r.SourceID] = append(relMap[r.SourceID], r.TargetID)
	}

	// Item ID to title map
	titleMap := make(map[string]string)
	for _, item := range items {
		titleMap[item.ID] = item.Title
	}

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	for _, item := range items {
		content := e.itemToMarkdown(item, relMap[item.ID], titleMap)
		filename := fmt.Sprintf("notes/%s.md", sanitizeFilename(item.Title))

		f, err := zw.Create(filename)
		if err != nil {
			return nil, err
		}
		f.Write([]byte(content))
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf, nil
}

func (e *ObsidianExporter) itemToMarkdown(item store.Item, relatedIDs []string, titleMap map[string]string) string {
	var sb strings.Builder

	// YAML frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("id: %s\n", item.ID))
	sb.WriteString(fmt.Sprintf("type: %s\n", item.Type))
	if item.URL != "" {
		sb.WriteString(fmt.Sprintf("url: \"%s\"\n", item.URL))
	}
	sb.WriteString(fmt.Sprintf("created: %s\n", item.CreatedAt.Format(time.RFC3339)))
	if len(item.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(item.Tags, ", ")))
	}
	sb.WriteString("---\n\n")

	// Title
	sb.WriteString(fmt.Sprintf("# %s\n\n", item.Title))

	// Summary
	if item.Summary != "" {
		sb.WriteString(fmt.Sprintf("> %s\n\n", item.Summary))
	}

	// Source link
	if item.URL != "" {
		sb.WriteString(fmt.Sprintf("**Source:** [%s](%s)\n\n", item.URL, item.URL))
	}

	// Content
	if item.Content != "" {
		sb.WriteString("## Content\n\n")
		sb.WriteString(item.Content)
		sb.WriteString("\n\n")
	}

	// Related items as wikilinks
	if len(relatedIDs) > 0 {
		sb.WriteString("## Related\n\n")
		for _, id := range relatedIDs {
			if title, ok := titleMap[id]; ok {
				sb.WriteString(fmt.Sprintf("- [[%s]]\n", title))
			}
		}
	}

	return sb.String()
}

func sanitizeFilename(s string) string {
	// Replace invalid filename characters
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "-",
		"\n", " ",
		"\r", "",
	)
	s = replacer.Replace(s)

	// Trim whitespace
	s = strings.TrimSpace(s)

	// Limit length
	if len(s) > 100 {
		s = s[:100]
	}

	// If empty, use default
	if s == "" {
		s = "untitled"
	}

	return s
}
