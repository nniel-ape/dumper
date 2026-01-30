package store

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (v *VaultStore) CreateItem(item *Item) error {
	if item.ID == "" {
		item.ID = uuid.NewString()
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = item.CreatedAt

	tx, err := v.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Rollback is always safe to call; ignore error as commit handles success
	}()

	_, err = tx.Exec(`
		INSERT INTO items (id, type, url, title, content, summary, raw_content, image_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.Type, item.URL, item.Title, item.Content, item.Summary, item.RawContent, item.ImagePath,
		item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert item: %w", err)
	}

	if err := v.setItemTags(tx, item.ID, item.Tags); err != nil {
		return fmt.Errorf("set tags: %w", err)
	}

	return tx.Commit()
}

func (v *VaultStore) GetItem(id string) (*Item, error) {
	item := &Item{}
	var url, content, summary, imagePath sql.NullString
	err := v.db.QueryRow(`
		SELECT id, type, url, title, content, summary, image_path, created_at, updated_at
		FROM items WHERE id = ?`, id,
	).Scan(&item.ID, &item.Type, &url, &item.Title, &content, &summary, &imagePath,
		&item.CreatedAt, &item.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query item: %w", err)
	}

	item.URL = url.String
	item.Content = content.String
	item.Summary = summary.String
	item.ImagePath = imagePath.String

	tags, err := v.getItemTags(item.ID)
	if err != nil {
		return nil, fmt.Errorf("get tags: %w", err)
	}
	item.Tags = tags
	return item, nil
}

func (v *VaultStore) ListItems(limit, offset int) ([]Item, error) {
	rows, err := v.db.Query(`
		SELECT id, type, url, title, content, summary, image_path, created_at, updated_at
		FROM items ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer func() {
		_ = rows.Close() // Close is always safe to call
	}()

	var items []Item
	for rows.Next() {
		var item Item
		var url, content, summary, imagePath sql.NullString
		if err := rows.Scan(&item.ID, &item.Type, &url, &item.Title, &content,
			&summary, &imagePath, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		item.URL = url.String
		item.Content = content.String
		item.Summary = summary.String
		item.ImagePath = imagePath.String
		tags, _ := v.getItemTags(item.ID)
		item.Tags = tags
		items = append(items, item)
	}
	return items, nil
}

func (v *VaultStore) ListItemsByTag(tag string, limit, offset int) ([]Item, error) {
	rows, err := v.db.Query(`
		SELECT DISTINCT i.id, i.type, i.url, i.title, i.content, i.summary, i.image_path, i.created_at, i.updated_at
		FROM items i
		JOIN item_tags it ON i.id = it.item_id
		JOIN tags t ON it.tag_id = t.id
		WHERE t.name = ?
		ORDER BY i.created_at DESC LIMIT ? OFFSET ?`, tag, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query items by tag: %w", err)
	}
	defer func() {
		_ = rows.Close() // Close is always safe to call
	}()

	var items []Item
	for rows.Next() {
		var item Item
		var url, content, summary, imagePath sql.NullString
		if err := rows.Scan(&item.ID, &item.Type, &url, &item.Title, &content,
			&summary, &imagePath, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		item.URL = url.String
		item.Content = content.String
		item.Summary = summary.String
		item.ImagePath = imagePath.String
		tags, _ := v.getItemTags(item.ID)
		item.Tags = tags
		items = append(items, item)
	}
	return items, nil
}

func (v *VaultStore) Search(query string, limit int) ([]SearchResult, error) {
	// Sanitize FTS5 query to prevent syntax errors with special chars
	sanitizedQuery := sanitizeFTS5Query(query)

	rows, err := v.db.Query(`
		SELECT i.id, i.type, i.url, i.title, i.content, i.summary, i.image_path, i.created_at, i.updated_at,
		       snippet(items_fts, 1, '<mark>', '</mark>', '...', 32) as snippet,
		       bm25(items_fts) as score
		FROM items_fts
		JOIN items i ON items_fts.rowid = i.rowid
		WHERE items_fts MATCH ?
		ORDER BY score
		LIMIT ?`, sanitizedQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer func() {
		_ = rows.Close() // Close is always safe to call
	}()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var url, content, summary, imagePath sql.NullString
		if err := rows.Scan(&r.Item.ID, &r.Item.Type, &url, &r.Item.Title,
			&content, &summary, &imagePath, &r.Item.CreatedAt, &r.Item.UpdatedAt,
			&r.Snippet, &r.Score); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		r.Item.URL = url.String
		r.Item.Content = content.String
		r.Item.Summary = summary.String
		r.Item.ImagePath = imagePath.String
		tags, _ := v.getItemTags(r.Item.ID)
		r.Item.Tags = tags
		results = append(results, r)
	}
	return results, nil
}

func (v *VaultStore) DeleteItem(id string) error {
	// Get item to check for image path before deletion
	item, err := v.GetItem(id)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("get item: %w", err)
	}

	// Delete from database (CASCADE handles relationships and tags)
	_, err = v.db.Exec("DELETE FROM items WHERE id = ?", id)
	if err != nil {
		return err
	}

	// Clean up image file if it exists
	if item != nil && item.ImagePath != "" {
		fullPath := filepath.Join(v.userDir, item.ImagePath)
		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			// Log error but don't fail the deletion - database record is already gone
			slog.Warn("failed to delete image file", "path", fullPath, "error", err)
		}
	}

	return nil
}

// sanitizeFTS5Query escapes special FTS5 characters to prevent syntax errors
// while preserving boolean search capability. Splits query into words,
// wraps each word in quotes to escape special characters, and joins with OR.
func sanitizeFTS5Query(query string) string {
	words := strings.Fields(query)
	if len(words) == 0 {
		return `""`
	}

	var escaped []string
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}
		// Escape internal quotes and wrap each word in quotes to prevent
		// special characters like +, -, *, :, etc. from being interpreted as operators
		word = strings.ReplaceAll(word, `"`, `""`)
		escaped = append(escaped, `"`+word+`"`)
	}

	// Join with OR for boolean search (matches any word)
	return strings.Join(escaped, " OR ")
}

func (v *VaultStore) setItemTags(tx *sql.Tx, itemID string, tags []string) error {
	_, err := tx.Exec("DELETE FROM item_tags WHERE item_id = ?", itemID)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		_, err := tx.Exec(`INSERT OR IGNORE INTO tags (name) VALUES (?)`, tag)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`
			INSERT INTO item_tags (item_id, tag_id)
			SELECT ?, id FROM tags WHERE name = ?`, itemID, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *VaultStore) getItemTags(itemID string) ([]string, error) {
	rows, err := v.db.Query(`
		SELECT t.name FROM tags t
		JOIN item_tags it ON t.id = it.tag_id
		WHERE it.item_id = ?`, itemID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close() // Close is always safe to call
	}()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (v *VaultStore) GetAllTags() ([]string, error) {
	rows, err := v.db.Query(`SELECT name FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close() // Close is always safe to call
	}()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (v *VaultStore) ItemCount() (int, error) {
	var count int
	err := v.db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&count)
	return count, err
}
