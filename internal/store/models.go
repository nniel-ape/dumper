package store

import "time"

type ItemType string

const (
	ItemTypeLink   ItemType = "link"
	ItemTypeNote   ItemType = "note"
	ItemTypeImage  ItemType = "image"
	ItemTypeSearch ItemType = "search"
)

type Item struct {
	ID         string    `json:"id"`
	Type       ItemType  `json:"type"`
	URL        string    `json:"url,omitempty"`
	Title      string    `json:"title"`
	Content    string    `json:"content,omitempty"`
	Summary    string    `json:"summary,omitempty"`
	RawContent string    `json:"-"`
	ImagePath  string    `json:"image_path,omitempty"` // relative path from user dir
	Tags       []string  `json:"tags"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Relationship struct {
	ID           int64   `json:"id"`
	SourceID     string  `json:"source_id"`
	TargetID     string  `json:"target_id"`
	RelationType string  `json:"relation_type"`
	Strength     float64 `json:"strength"`
}

type SearchResult struct {
	Item    Item    `json:"item"`
	Snippet string  `json:"snippet,omitempty"`
	Score   float64 `json:"score"`
}
