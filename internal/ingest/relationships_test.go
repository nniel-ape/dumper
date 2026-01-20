package ingest

import (
	"context"
	"reflect"
	"testing"

	"github.com/nerdneilsfield/dumper/internal/store"
)

func TestExtractWikiLinkTargets(t *testing.T) {
	input := "See [[Note]] and [[Second Note|alias]] plus [[Third#Section]] and [[note]]."
	got := extractWikiLinkTargets(input)
	want := []string{"note", "second note", "third"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractWikiLinkTargets mismatch: got %v want %v", got, want)
	}
}

func TestExtractHashTags(t *testing.T) {
	input := "Some #Tag and #tag/sub plus https://example.com/#anchor"
	got := extractHashTags(input)
	want := []string{"tag", "tag/sub"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractHashTags mismatch: got %v want %v", got, want)
	}
}

func TestExtractTitleFromNote(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "h1", input: "# My Title\nBody", want: "My Title"},
		{name: "not_h1", input: "Intro\n# Title", want: ""},
		{name: "h2", input: "## Heading\nBody", want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractTitleFromNote(tc.input)
			if got != tc.want {
				t.Fatalf("extractTitleFromNote mismatch: got %q want %q", got, tc.want)
			}
		})
	}
}

func TestFindAndCreateRelationshipsObsidian(t *testing.T) {
	manager, err := store.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	vault, err := manager.GetVault(1)
	if err != nil {
		t.Fatalf("failed to get vault: %v", err)
	}
	t.Cleanup(func() {
		_ = manager.Close()
	})

	ctx := context.Background()
	pipeline := &Pipeline{}

	itemA := &store.Item{
		Type:    store.ItemTypeNote,
		Title:   "Alpha",
		Content: "See [[Beta]]",
		Tags:    []string{"go", "dev"},
	}
	if err := vault.CreateItem(itemA); err != nil {
		t.Fatalf("create item A: %v", err)
	}
	pipeline.findAndCreateRelationships(ctx, vault, itemA)

	itemB := &store.Item{
		Type:    store.ItemTypeNote,
		Title:   "Beta",
		Content: "No links here",
		Tags:    []string{"go", "notes"},
	}
	if err := vault.CreateItem(itemB); err != nil {
		t.Fatalf("create item B: %v", err)
	}
	pipeline.findAndCreateRelationships(ctx, vault, itemB)

	itemC := &store.Item{
		Type:    store.ItemTypeNote,
		Title:   "Gamma",
		Content: "No links here either",
		Tags:    []string{"go"},
	}
	if err := vault.CreateItem(itemC); err != nil {
		t.Fatalf("create item C: %v", err)
	}
	pipeline.findAndCreateRelationships(ctx, vault, itemC)

	_, relationships, err := vault.GetGraph()
	if err != nil {
		t.Fatalf("get graph: %v", err)
	}

	linkCount := 0
	tagCount := 0
	for _, rel := range relationships {
		switch rel.RelationType {
		case "link":
			linkCount++
		case "tag":
			tagCount++
		}
	}

	if linkCount != 1 {
		t.Fatalf("expected 1 link relationship, got %d", linkCount)
	}
	if tagCount != 2 {
		t.Fatalf("expected 2 tag relationships, got %d", tagCount)
	}

	if !hasRelationship(relationships, itemA.ID, itemB.ID, "link") {
		t.Fatalf("expected link from Alpha to Beta")
	}

	if hasTagBetween(vault, itemA.ID, itemB.ID) {
		t.Fatalf("unexpected tag relationship between Alpha and Beta")
	}
}

func hasRelationship(relationships []store.Relationship, sourceID, targetID, relType string) bool {
	for _, rel := range relationships {
		if rel.RelationType == relType && rel.SourceID == sourceID && rel.TargetID == targetID {
			return true
		}
	}
	return false
}

func hasTagBetween(vault *store.VaultStore, sourceID, targetID string) bool {
	var count int
	err := vault.DB().QueryRow(
		`SELECT COUNT(*) FROM relationships WHERE relation_type = 'tag' AND ((source_id = ? AND target_id = ?) OR (source_id = ? AND target_id = ?))`,
		sourceID, targetID, targetID, sourceID,
	).Scan(&count)
	if err != nil {
		return true
	}
	return count > 0
}
