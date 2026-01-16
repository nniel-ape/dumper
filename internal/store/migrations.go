package store

import (
	"database/sql"
	"fmt"
)

// Migration SQL - embedded directly since go:embed requires the files to be in the same package or below
const migrationSQL = `
-- User vault schema (applied to each user's vault.db)

CREATE TABLE IF NOT EXISTS items (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK(type IN ('link', 'note', 'image', 'search')),
    url TEXT,
    title TEXT NOT NULL,
    content TEXT,
    summary TEXT,
    raw_content TEXT,
    image_path TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS item_tags (
    item_id TEXT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (item_id, tag_id)
);

CREATE TABLE IF NOT EXISTS relationships (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_id TEXT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    target_id TEXT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL,
    strength REAL DEFAULT 1.0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source_id, target_id, relation_type)
);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Full-text search
CREATE VIRTUAL TABLE IF NOT EXISTS items_fts USING fts5(
    title,
    content,
    summary,
    content='items',
    content_rowid='rowid'
);

-- Triggers to keep FTS in sync
CREATE TRIGGER IF NOT EXISTS items_ai AFTER INSERT ON items BEGIN
    INSERT INTO items_fts(rowid, title, content, summary)
    VALUES (NEW.rowid, NEW.title, NEW.content, NEW.summary);
END;

CREATE TRIGGER IF NOT EXISTS items_ad AFTER DELETE ON items BEGIN
    INSERT INTO items_fts(items_fts, rowid, title, content, summary)
    VALUES ('delete', OLD.rowid, OLD.title, OLD.content, OLD.summary);
END;

CREATE TRIGGER IF NOT EXISTS items_au AFTER UPDATE ON items BEGIN
    INSERT INTO items_fts(items_fts, rowid, title, content, summary)
    VALUES ('delete', OLD.rowid, OLD.title, OLD.content, OLD.summary);
    INSERT INTO items_fts(rowid, title, content, summary)
    VALUES (NEW.rowid, NEW.title, NEW.content, NEW.summary);
END;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_items_type ON items(type);
CREATE INDEX IF NOT EXISTS idx_items_created ON items(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_relationships_source ON relationships(source_id);
CREATE INDEX IF NOT EXISTS idx_relationships_target ON relationships(target_id);
`

// Migration for existing databases to add image_path column
const migrationAddImagePath = `
ALTER TABLE items ADD COLUMN image_path TEXT;
`

// Migration to update CHECK constraint for existing databases
// SQLite doesn't support ALTER TABLE to modify CHECK constraints, so we recreate the table
const migrationUpdateTypeConstraint = `
-- Disable foreign keys temporarily
PRAGMA foreign_keys=OFF;

-- Create new table with updated CHECK constraint
CREATE TABLE IF NOT EXISTS items_new (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK(type IN ('link', 'note', 'image', 'search')),
    url TEXT,
    title TEXT NOT NULL,
    content TEXT,
    summary TEXT,
    raw_content TEXT,
    image_path TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Copy data from old table (handle missing image_path column)
INSERT OR IGNORE INTO items_new (id, type, url, title, content, summary, raw_content, created_at, updated_at)
SELECT id, type, url, title, content, summary, raw_content, created_at, updated_at FROM items;

-- Drop old table
DROP TABLE items;

-- Rename new table
ALTER TABLE items_new RENAME TO items;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_items_type ON items(type);
CREATE INDEX IF NOT EXISTS idx_items_created ON items(created_at DESC);

-- Re-enable foreign keys
PRAGMA foreign_keys=ON;
`

func RunMigrations(db *sql.DB) error {
	if _, err := db.Exec(migrationSQL); err != nil {
		return fmt.Errorf("exec migration: %w", err)
	}

	// Add image_path column for existing databases (ignore error if column exists)
	_, _ = db.Exec(migrationAddImagePath)

	// Check if we need to update the CHECK constraint
	// by checking if 'image' type is allowed
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('items') WHERE name='type'`).Scan(&count)
	if err == nil && count > 0 {
		// Try inserting a test row to see if constraint allows 'image'
		_, err := db.Exec(`INSERT INTO items (id, type, title) VALUES ('__test__', 'image', 'test')`)
		if err != nil {
			// Constraint doesn't allow 'image', need to migrate
			if _, err := db.Exec(migrationUpdateTypeConstraint); err != nil {
				return fmt.Errorf("update type constraint: %w", err)
			}
		} else {
			// Clean up test row
			_, _ = db.Exec(`DELETE FROM items WHERE id = '__test__'`)
		}
	}

	return nil
}
