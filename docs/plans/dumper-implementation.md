# Dumper - Implementation Plan

**Design Doc:** [docs/brainstorms/dumper-knowledge-capture-system.md](../brainstorms/dumper-knowledge-capture-system.md)

## Overview

Build a Telegram bot that captures links and text, processes them with LLMs (OpenRouter) for summarization and tagging, stores in per-user SQLite databases, and provides a TG Mini App for browsing, searching, graph visualization, and Q&A. Includes Obsidian-compatible markdown export.

---

## Prerequisites

### Required Tools
- Go 1.22+ (`brew install go`)
- Node.js 20+ (`brew install node`) - for Mini App frontend build
- Docker & Docker Compose (`brew install docker`)
- SQLite3 CLI (`brew install sqlite3`) - for debugging

### Accounts & API Keys
1. **Telegram Bot Token**: Create via [@BotFather](https://t.me/botfather)
   - Enable inline mode: `/setinline`
   - Enable Mini App: `/newapp` → select your bot → provide webapp URL
2. **OpenRouter API Key**: Get from [openrouter.ai/keys](https://openrouter.ai/keys)

### Environment Variables
Create `.env` file (never commit):
```bash
TELEGRAM_BOT_TOKEN=your_bot_token
OPENROUTER_API_KEY=your_api_key
DATA_DIR=./data
HTTP_PORT=8080
LOG_LEVEL=debug
```

---

## Codebase Orientation

### Directory Structure (to create)
```
dumper/
├── cmd/
│   └── dumper/
│       └── main.go           # Entrypoint, wires dependencies
├── internal/
│   ├── bot/                  # Telegram bot handlers
│   ├── ingest/               # InputSource interface, pipeline
│   ├── llm/                  # OpenRouter client
│   ├── store/                # SQLite storage layer
│   ├── api/                  # HTTP handlers for Mini App
│   └── export/               # Obsidian markdown export
├── web/                      # Mini App frontend (Svelte)
├── migrations/               # SQL schema files
├── docs/
│   ├── brainstorms/
│   └── plans/
├── docker-compose.yml
├── Dockerfile
├── .env.example
├── .gitignore
└── go.mod
```

### Patterns to Follow
- **Error handling**: Always wrap errors with context: `fmt.Errorf("operation: %w", err)`
- **Logging**: Use `log/slog` (stdlib structured logging)
- **Config**: Use `github.com/caarlos0/env/v11` for env parsing
- **Testing**: Table-driven tests, `testify/assert` for assertions
- **SQL**: Use `github.com/mattn/go-sqlite3` with raw SQL (no ORM)

---

## Implementation Tasks

### Phase 1: Project Foundation

---

### Task 1: Initialize Go Module and Project Structure

**Goal:** Set up the Go module, directory structure, and basic configuration.

**Files to create:**
- `go.mod` - Go module definition
- `cmd/dumper/main.go` - Entry point stub
- `internal/config/config.go` - Configuration struct
- `.env.example` - Example environment file
- `.gitignore` - Ignore patterns

**Implementation steps:**

1. Initialize Go module:
```bash
go mod init github.com/yourusername/dumper
```

2. Create directory structure:
```bash
mkdir -p cmd/dumper internal/{bot,ingest,llm,store,api,export,config} web migrations docs/{brainstorms,plans}
```

3. Create `internal/config/config.go`:
```go
package config

import (
    "fmt"
    "github.com/caarlos0/env/v11"
)

type Config struct {
    TelegramToken   string `env:"TELEGRAM_BOT_TOKEN,required"`
    OpenRouterKey   string `env:"OPENROUTER_API_KEY,required"`
    DataDir         string `env:"DATA_DIR" envDefault:"./data"`
    HTTPPort        int    `env:"HTTP_PORT" envDefault:"8080"`
    LogLevel        string `env:"LOG_LEVEL" envDefault:"info"`
    OpenRouterModel string `env:"OPENROUTER_MODEL" envDefault:"anthropic/claude-3-haiku"`
}

func Load() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }
    return cfg, nil
}
```

4. Create `cmd/dumper/main.go`:
```go
package main

import (
    "log/slog"
    "os"

    "github.com/yourusername/dumper/internal/config"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(logger)

    cfg, err := config.Load()
    if err != nil {
        slog.Error("failed to load config", "error", err)
        os.Exit(1)
    }

    slog.Info("starting dumper", "port", cfg.HTTPPort)
    // TODO: wire up components
}
```

5. Create `.env.example`:
```bash
TELEGRAM_BOT_TOKEN=
OPENROUTER_API_KEY=
DATA_DIR=./data
HTTP_PORT=8080
LOG_LEVEL=debug
OPENROUTER_MODEL=anthropic/claude-3-haiku
```

6. Create `.gitignore`:
```
.env
data/
*.db
web/dist/
node_modules/
```

7. Install dependencies:
```bash
go get github.com/caarlos0/env/v11
```

**Testing:**
```bash
# Should compile without errors
go build ./cmd/dumper

# Should fail gracefully without env vars
./dumper
# Expected: error about missing TELEGRAM_BOT_TOKEN
```

**Verification:**
- `go build ./cmd/dumper` succeeds
- Running without env vars shows clear error message

**Commit:** `feat: initialize project structure and config`

---

### Task 2: Create SQLite Storage Layer - Schema and Migrations

**Goal:** Define database schema and create migration system.

**Files to create:**
- `migrations/001_init.sql` - Initial schema
- `internal/store/migrations.go` - Migration runner
- `internal/store/store.go` - Store interface and factory

**Implementation steps:**

1. Create `migrations/001_init.sql`:
```sql
-- User vault schema (applied to each user's vault.db)

CREATE TABLE IF NOT EXISTS items (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK(type IN ('link', 'note')),
    url TEXT,
    title TEXT NOT NULL,
    content TEXT,
    summary TEXT,
    raw_content TEXT,
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
```

2. Create `internal/store/migrations.go`:
```go
package store

import (
    "database/sql"
    "embed"
    "fmt"
    "io/fs"
    "sort"
    "strings"
)

//go:embed ../../migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(db *sql.DB) error {
    entries, err := fs.ReadDir(migrationsFS, "migrations")
    if err != nil {
        return fmt.Errorf("read migrations dir: %w", err)
    }

    var files []string
    for _, e := range entries {
        if strings.HasSuffix(e.Name(), ".sql") {
            files = append(files, e.Name())
        }
    }
    sort.Strings(files)

    for _, f := range files {
        content, err := fs.ReadFile(migrationsFS, "migrations/"+f)
        if err != nil {
            return fmt.Errorf("read migration %s: %w", f, err)
        }
        if _, err := db.Exec(string(content)); err != nil {
            return fmt.Errorf("exec migration %s: %w", f, err)
        }
    }
    return nil
}
```

3. Create `internal/store/store.go`:
```go
package store

import (
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "sync"

    _ "github.com/mattn/go-sqlite3"
)

type VaultStore struct {
    db *sql.DB
}

type Manager struct {
    dataDir string
    vaults  map[int64]*VaultStore
    mu      sync.RWMutex
}

func NewManager(dataDir string) (*Manager, error) {
    if err := os.MkdirAll(dataDir, 0755); err != nil {
        return nil, fmt.Errorf("create data dir: %w", err)
    }
    return &Manager{
        dataDir: dataDir,
        vaults:  make(map[int64]*VaultStore),
    }, nil
}

func (m *Manager) GetVault(userID int64) (*VaultStore, error) {
    m.mu.RLock()
    if v, ok := m.vaults[userID]; ok {
        m.mu.RUnlock()
        return v, nil
    }
    m.mu.RUnlock()

    m.mu.Lock()
    defer m.mu.Unlock()

    // Double-check after acquiring write lock
    if v, ok := m.vaults[userID]; ok {
        return v, nil
    }

    vault, err := m.openVault(userID)
    if err != nil {
        return nil, err
    }
    m.vaults[userID] = vault
    return vault, nil
}

func (m *Manager) openVault(userID int64) (*VaultStore, error) {
    userDir := filepath.Join(m.dataDir, "users", fmt.Sprintf("%d", userID))
    if err := os.MkdirAll(userDir, 0755); err != nil {
        return nil, fmt.Errorf("create user dir: %w", err)
    }

    dbPath := filepath.Join(userDir, "vault.db")
    db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
    if err != nil {
        return nil, fmt.Errorf("open db: %w", err)
    }

    if err := RunMigrations(db); err != nil {
        db.Close()
        return nil, fmt.Errorf("run migrations: %w", err)
    }

    return &VaultStore{db: db}, nil
}

func (m *Manager) Close() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    for _, v := range m.vaults {
        v.db.Close()
    }
    return nil
}
```

**Testing:**
- Create `internal/store/store_test.go` with tests for vault creation and migration

**Verification:**
```bash
go test ./internal/store/...
```

**Commit:** `feat: add SQLite storage layer with migrations`

---

### Task 3: Implement Item CRUD Operations

**Goal:** Add methods to create, read, update, delete items in the vault.

**Files to modify:**
- `internal/store/store.go` - Add CRUD methods
- `internal/store/models.go` - Define domain models

**Files to create:**
- `internal/store/models.go` - Item, Tag, Relationship structs
- `internal/store/items.go` - Item-specific queries
- `internal/store/items_test.go` - Tests

**Implementation steps:**

1. Create `internal/store/models.go`:
```go
package store

import "time"

type ItemType string

const (
    ItemTypeLink ItemType = "link"
    ItemTypeNote ItemType = "note"
)

type Item struct {
    ID         string    `json:"id"`
    Type       ItemType  `json:"type"`
    URL        string    `json:"url,omitempty"`
    Title      string    `json:"title"`
    Content    string    `json:"content,omitempty"`
    Summary    string    `json:"summary,omitempty"`
    RawContent string    `json:"-"`
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
```

2. Create `internal/store/items.go`:
```go
package store

import (
    "database/sql"
    "fmt"
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
    defer tx.Rollback()

    _, err = tx.Exec(`
        INSERT INTO items (id, type, url, title, content, summary, raw_content, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        item.ID, item.Type, item.URL, item.Title, item.Content, item.Summary, item.RawContent,
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
    err := v.db.QueryRow(`
        SELECT id, type, url, title, content, summary, created_at, updated_at
        FROM items WHERE id = ?`, id,
    ).Scan(&item.ID, &item.Type, &item.URL, &item.Title, &item.Content, &item.Summary,
        &item.CreatedAt, &item.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("query item: %w", err)
    }

    tags, err := v.getItemTags(item.ID)
    if err != nil {
        return nil, fmt.Errorf("get tags: %w", err)
    }
    item.Tags = tags
    return item, nil
}

func (v *VaultStore) ListItems(limit, offset int) ([]Item, error) {
    rows, err := v.db.Query(`
        SELECT id, type, url, title, content, summary, created_at, updated_at
        FROM items ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
    if err != nil {
        return nil, fmt.Errorf("query items: %w", err)
    }
    defer rows.Close()

    var items []Item
    for rows.Next() {
        var item Item
        if err := rows.Scan(&item.ID, &item.Type, &item.URL, &item.Title, &item.Content,
            &item.Summary, &item.CreatedAt, &item.UpdatedAt); err != nil {
            return nil, fmt.Errorf("scan item: %w", err)
        }
        tags, _ := v.getItemTags(item.ID)
        item.Tags = tags
        items = append(items, item)
    }
    return items, nil
}

func (v *VaultStore) Search(query string, limit int) ([]SearchResult, error) {
    rows, err := v.db.Query(`
        SELECT i.id, i.type, i.url, i.title, i.content, i.summary, i.created_at, i.updated_at,
               snippet(items_fts, 1, '<mark>', '</mark>', '...', 32) as snippet,
               bm25(items_fts) as score
        FROM items_fts
        JOIN items i ON items_fts.rowid = i.rowid
        WHERE items_fts MATCH ?
        ORDER BY score
        LIMIT ?`, query, limit)
    if err != nil {
        return nil, fmt.Errorf("search: %w", err)
    }
    defer rows.Close()

    var results []SearchResult
    for rows.Next() {
        var r SearchResult
        if err := rows.Scan(&r.Item.ID, &r.Item.Type, &r.Item.URL, &r.Item.Title,
            &r.Item.Content, &r.Item.Summary, &r.Item.CreatedAt, &r.Item.UpdatedAt,
            &r.Snippet, &r.Score); err != nil {
            return nil, fmt.Errorf("scan result: %w", err)
        }
        tags, _ := v.getItemTags(r.Item.ID)
        r.Item.Tags = tags
        results = append(results, r)
    }
    return results, nil
}

func (v *VaultStore) DeleteItem(id string) error {
    _, err := v.db.Exec("DELETE FROM items WHERE id = ?", id)
    return err
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
    defer rows.Close()

    var tags []string
    for rows.Next() {
        var tag string
        rows.Scan(&tag)
        tags = append(tags, tag)
    }
    return tags, nil
}

func (v *VaultStore) GetAllTags() ([]string, error) {
    rows, err := v.db.Query(`SELECT name FROM tags ORDER BY name`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tags []string
    for rows.Next() {
        var tag string
        rows.Scan(&tag)
        tags = append(tags, tag)
    }
    return tags, nil
}
```

**Testing:**
Create `internal/store/items_test.go` with table-driven tests:
- Test CreateItem + GetItem roundtrip
- Test ListItems pagination
- Test Search with FTS
- Test tag operations

**Verification:**
```bash
go get github.com/google/uuid
go test -v ./internal/store/...
```

**Commit:** `feat: implement item CRUD and search operations`

---

### Task 4: Implement Relationship Storage

**Goal:** Store and query AI-inferred relationships between items.

**Files to create:**
- `internal/store/relationships.go` - Relationship queries

**Implementation steps:**

1. Create `internal/store/relationships.go`:
```go
package store

import "fmt"

func (v *VaultStore) CreateRelationship(rel *Relationship) error {
    _, err := v.db.Exec(`
        INSERT OR REPLACE INTO relationships (source_id, target_id, relation_type, strength)
        VALUES (?, ?, ?, ?)`,
        rel.SourceID, rel.TargetID, rel.RelationType, rel.Strength)
    return err
}

func (v *VaultStore) GetRelationships(itemID string) ([]Relationship, error) {
    rows, err := v.db.Query(`
        SELECT id, source_id, target_id, relation_type, strength
        FROM relationships
        WHERE source_id = ? OR target_id = ?`, itemID, itemID)
    if err != nil {
        return nil, fmt.Errorf("query relationships: %w", err)
    }
    defer rows.Close()

    var rels []Relationship
    for rows.Next() {
        var r Relationship
        if err := rows.Scan(&r.ID, &r.SourceID, &r.TargetID, &r.RelationType, &r.Strength); err != nil {
            return nil, err
        }
        rels = append(rels, r)
    }
    return rels, nil
}

// GetGraph returns all items and relationships for graph visualization
func (v *VaultStore) GetGraph() ([]Item, []Relationship, error) {
    items, err := v.ListItems(1000, 0) // TODO: pagination for large vaults
    if err != nil {
        return nil, nil, err
    }

    rows, err := v.db.Query(`SELECT id, source_id, target_id, relation_type, strength FROM relationships`)
    if err != nil {
        return nil, nil, err
    }
    defer rows.Close()

    var rels []Relationship
    for rows.Next() {
        var r Relationship
        rows.Scan(&r.ID, &r.SourceID, &r.TargetID, &r.RelationType, &r.Strength)
        rels = append(rels, r)
    }
    return items, rels, nil
}
```

**Commit:** `feat: add relationship storage for knowledge graph`

---

### Phase 2: LLM Integration

---

### Task 5: Create OpenRouter Client

**Goal:** Build a client to call OpenRouter API for summarization, tagging, and relationship inference.

**Files to create:**
- `internal/llm/client.go` - HTTP client for OpenRouter
- `internal/llm/prompts.go` - Prompt templates
- `internal/llm/types.go` - Request/response types

**Implementation steps:**

1. Create `internal/llm/types.go`:
```go
package llm

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
}

type ChatResponse struct {
    Choices []struct {
        Message Message `json:"message"`
    } `json:"choices"`
    Error *struct {
        Message string `json:"message"`
    } `json:"error,omitempty"`
}

type ProcessedContent struct {
    Title        string   `json:"title"`
    Summary      string   `json:"summary"`
    Tags         []string `json:"tags"`
    RelatedTopics []string `json:"related_topics"`
}
```

2. Create `internal/llm/prompts.go`:
```go
package llm

const ProcessContentPrompt = `Analyze the following content and extract structured information.

Content Type: %s
Content:
---
%s
---

Respond with ONLY valid JSON (no markdown, no explanation):
{
  "title": "concise descriptive title (max 10 words)",
  "summary": "2-3 sentence summary capturing key points",
  "tags": ["tag1", "tag2", "tag3"],
  "related_topics": ["topic that might connect to other saved items"]
}

Rules:
- Tags should be lowercase, single words or short phrases
- Generate 3-7 relevant tags
- Summary should be informative but concise
- Related topics help build knowledge graph connections`

const FindRelationshipsPrompt = `Given a new item and existing items, identify semantic relationships.

New item:
Title: %s
Summary: %s
Tags: %v

Existing items:
%s

Respond with ONLY valid JSON array of relationships:
[
  {"target_id": "id", "relation_type": "type", "strength": 0.8}
]

Relation types: "similar_topic", "references", "contradicts", "extends", "prerequisite"
Strength: 0.0-1.0 (how strong the connection is)
Only include relationships with strength >= 0.5`
```

3. Create `internal/llm/client.go`:
```go
package llm

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
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

func (c *Client) ProcessContent(ctx context.Context, contentType, content string) (*ProcessedContent, error) {
    prompt := fmt.Sprintf(ProcessContentPrompt, contentType, content)

    response, err := c.Chat(ctx, []Message{
        {Role: "user", Content: prompt},
    })
    if err != nil {
        return nil, fmt.Errorf("chat: %w", err)
    }

    var result ProcessedContent
    if err := json.Unmarshal([]byte(response), &result); err != nil {
        return nil, fmt.Errorf("parse response: %w (raw: %s)", err, response)
    }
    return &result, nil
}
```

**Testing:**
Create mock tests that don't call the real API.

**Verification:**
```bash
go test ./internal/llm/...
```

**Commit:** `feat: add OpenRouter LLM client`

---

### Task 6: Create Content Extraction Pipeline

**Goal:** Extract content from URLs using readability.

**Files to create:**
- `internal/ingest/extractor.go` - URL content extraction
- `internal/ingest/pipeline.go` - Processing pipeline
- `internal/ingest/source.go` - InputSource interface

**Implementation steps:**

1. Create `internal/ingest/source.go`:
```go
package ingest

import "context"

type ContentType string

const (
    ContentTypeLink ContentType = "link"
    ContentTypeNote ContentType = "note"
)

type RawContent struct {
    Type    ContentType
    URL     string // for links
    Text    string // raw text or note content
    UserID  int64
}

type InputSource interface {
    Name() string
    // Sources push to the channel, pipeline consumes
}

type ProcessedItem struct {
    UserID   int64
    Item     interface{} // *store.Item
    Error    error
}
```

2. Create `internal/ingest/extractor.go`:
```go
package ingest

import (
    "context"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/go-shiori/go-readability"
)

type Extractor struct {
    client *http.Client
}

func NewExtractor() *Extractor {
    return &Extractor{
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

type ExtractedContent struct {
    URL         string
    Title       string
    Content     string
    Excerpt     string
    SiteName    string
    Favicon     string
}

func (e *Extractor) Extract(ctx context.Context, rawURL string) (*ExtractedContent, error) {
    parsed, err := url.Parse(rawURL)
    if err != nil {
        return nil, fmt.Errorf("parse url: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Dumper/1.0)")

    resp, err := e.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("fetch url: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
    }

    article, err := readability.FromReader(resp.Body, parsed)
    if err != nil {
        return nil, fmt.Errorf("parse content: %w", err)
    }

    return &ExtractedContent{
        URL:      rawURL,
        Title:    article.Title,
        Content:  article.TextContent,
        Excerpt:  article.Excerpt,
        SiteName: article.SiteName,
        Favicon:  fmt.Sprintf("%s://%s/favicon.ico", parsed.Scheme, parsed.Host),
    }, nil
}

func IsURL(s string) bool {
    s = strings.TrimSpace(s)
    return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
```

3. Create `internal/ingest/pipeline.go`:
```go
package ingest

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/yourusername/dumper/internal/llm"
    "github.com/yourusername/dumper/internal/store"
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

    // TODO: Find relationships with existing items
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
```

**Dependencies:**
```bash
go get github.com/go-shiori/go-readability
```

**Commit:** `feat: add content extraction and processing pipeline`

---

### Phase 3: Telegram Bot

---

### Task 7: Implement Telegram Bot Core

**Goal:** Create the Telegram bot that handles messages and commands.

**Files to create:**
- `internal/bot/bot.go` - Bot setup and handler routing
- `internal/bot/handlers.go` - Message handlers

**Implementation steps:**

1. Create `internal/bot/bot.go`:
```go
package bot

import (
    "context"
    "fmt"
    "log/slog"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/yourusername/dumper/internal/ingest"
    "github.com/yourusername/dumper/internal/store"
)

type Bot struct {
    api      *tgbotapi.BotAPI
    pipeline *ingest.Pipeline
    stores   *store.Manager
    webAppURL string
}

func New(token string, pipeline *ingest.Pipeline, stores *store.Manager, webAppURL string) (*Bot, error) {
    api, err := tgbotapi.NewBotAPI(token)
    if err != nil {
        return nil, fmt.Errorf("create bot api: %w", err)
    }

    slog.Info("authorized telegram bot", "username", api.Self.UserName)

    return &Bot{
        api:       api,
        pipeline:  pipeline,
        stores:    stores,
        webAppURL: webAppURL,
    }, nil
}

func (b *Bot) Run(ctx context.Context) error {
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates := b.api.GetUpdatesChan(u)

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case update := <-updates:
            go b.handleUpdate(ctx, update)
        }
    }
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
    if update.Message == nil {
        return
    }

    msg := update.Message
    userID := msg.From.ID

    slog.Debug("received message",
        "user_id", userID,
        "text", msg.Text,
        "has_entities", len(msg.Entities) > 0,
    )

    // Handle commands
    if msg.IsCommand() {
        b.handleCommand(ctx, msg)
        return
    }

    // Handle regular messages
    b.handleMessage(ctx, msg)
}

func (b *Bot) send(chatID int64, text string) {
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ParseMode = "HTML"
    if _, err := b.api.Send(msg); err != nil {
        slog.Error("failed to send message", "error", err)
    }
}

func (b *Bot) sendWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ParseMode = "HTML"
    msg.ReplyMarkup = keyboard
    if _, err := b.api.Send(msg); err != nil {
        slog.Error("failed to send message", "error", err)
    }
}
```

2. Create `internal/bot/handlers.go`:
```go
package bot

import (
    "context"
    "fmt"
    "strings"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/yourusername/dumper/internal/ingest"
)

func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
    switch msg.Command() {
    case "start":
        b.handleStart(msg)
    case "help":
        b.handleHelp(msg)
    case "search":
        b.handleSearch(ctx, msg)
    case "recent":
        b.handleRecent(ctx, msg)
    case "tags":
        b.handleTags(ctx, msg)
    case "export":
        b.handleExport(ctx, msg)
    case "app":
        b.handleApp(msg)
    default:
        b.send(msg.Chat.ID, "Unknown command. Use /help to see available commands.")
    }
}

func (b *Bot) handleStart(msg *tgbotapi.Message) {
    text := `👋 <b>Welcome to Dumper!</b>

I help you capture and organize knowledge from the web.

<b>How to use:</b>
• Send me any link - I'll extract, summarize, and tag it
• Send me text notes - I'll categorize them too
• Use /search to find saved items
• Use /app to open the Mini App

All your data is stored privately and can be exported to Obsidian.`

    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonWebApp("📱 Open App", tgbotapi.WebAppInfo{URL: b.webAppURL}),
        ),
    )
    b.sendWithKeyboard(msg.Chat.ID, text, keyboard)
}

func (b *Bot) handleHelp(msg *tgbotapi.Message) {
    text := `<b>Commands:</b>
/search [query] - Search your saved items
/recent - Show recent items
/tags - List all your tags
/export - Export to Obsidian format
/app - Open Mini App

<b>Saving content:</b>
Just send me any URL or text message!`
    b.send(msg.Chat.ID, text)
}

func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
    text := strings.TrimSpace(msg.Text)
    if text == "" {
        return
    }

    // Send processing message
    b.send(msg.Chat.ID, "⏳ Processing...")

    var raw ingest.RawContent
    raw.UserID = msg.From.ID

    if ingest.IsURL(text) {
        raw.Type = ingest.ContentTypeLink
        raw.URL = text
    } else {
        raw.Type = ingest.ContentTypeNote
        raw.Text = text
    }

    item, err := b.pipeline.Process(ctx, raw)
    if err != nil {
        b.send(msg.Chat.ID, fmt.Sprintf("❌ Failed to process: %v", err))
        return
    }

    // Format response
    var tagsStr string
    if len(item.Tags) > 0 {
        tagsStr = "#" + strings.Join(item.Tags, " #")
    }

    response := fmt.Sprintf(`✅ <b>Saved!</b>

<b>%s</b>

%s

%s`, item.Title, item.Summary, tagsStr)

    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonWebApp("View in App", tgbotapi.WebAppInfo{URL: b.webAppURL + "?item=" + item.ID}),
        ),
    )
    b.sendWithKeyboard(msg.Chat.ID, response, keyboard)
}

func (b *Bot) handleSearch(ctx context.Context, msg *tgbotapi.Message) {
    query := msg.CommandArguments()
    if query == "" {
        b.send(msg.Chat.ID, "Usage: /search [query]\nExample: /search golang concurrency")
        return
    }

    vault, err := b.stores.GetVault(msg.From.ID)
    if err != nil {
        b.send(msg.Chat.ID, "❌ Failed to access your vault")
        return
    }

    results, err := vault.Search(query, 5)
    if err != nil {
        b.send(msg.Chat.ID, fmt.Sprintf("❌ Search failed: %v", err))
        return
    }

    if len(results) == 0 {
        b.send(msg.Chat.ID, "No results found.")
        return
    }

    var text strings.Builder
    text.WriteString(fmt.Sprintf("🔍 <b>Results for \"%s\":</b>\n\n", query))

    for i, r := range results {
        text.WriteString(fmt.Sprintf("%d. <b>%s</b>\n", i+1, r.Item.Title))
        if r.Snippet != "" {
            text.WriteString(fmt.Sprintf("   %s\n", r.Snippet))
        }
        text.WriteString("\n")
    }

    b.send(msg.Chat.ID, text.String())
}

func (b *Bot) handleRecent(ctx context.Context, msg *tgbotapi.Message) {
    vault, err := b.stores.GetVault(msg.From.ID)
    if err != nil {
        b.send(msg.Chat.ID, "❌ Failed to access your vault")
        return
    }

    items, err := vault.ListItems(5, 0)
    if err != nil {
        b.send(msg.Chat.ID, fmt.Sprintf("❌ Failed to list items: %v", err))
        return
    }

    if len(items) == 0 {
        b.send(msg.Chat.ID, "No items saved yet. Send me a link or note to get started!")
        return
    }

    var text strings.Builder
    text.WriteString("📚 <b>Recent items:</b>\n\n")

    for i, item := range items {
        text.WriteString(fmt.Sprintf("%d. <b>%s</b>\n", i+1, item.Title))
        if len(item.Tags) > 0 {
            text.WriteString(fmt.Sprintf("   #%s\n", strings.Join(item.Tags, " #")))
        }
        text.WriteString("\n")
    }

    b.send(msg.Chat.ID, text.String())
}

func (b *Bot) handleTags(ctx context.Context, msg *tgbotapi.Message) {
    vault, err := b.stores.GetVault(msg.From.ID)
    if err != nil {
        b.send(msg.Chat.ID, "❌ Failed to access your vault")
        return
    }

    tags, err := vault.GetAllTags()
    if err != nil {
        b.send(msg.Chat.ID, fmt.Sprintf("❌ Failed to get tags: %v", err))
        return
    }

    if len(tags) == 0 {
        b.send(msg.Chat.ID, "No tags yet.")
        return
    }

    b.send(msg.Chat.ID, fmt.Sprintf("🏷 <b>Your tags:</b>\n\n#%s", strings.Join(tags, " #")))
}

func (b *Bot) handleExport(ctx context.Context, msg *tgbotapi.Message) {
    // TODO: Implement export
    b.send(msg.Chat.ID, "Export feature coming soon! Use the Mini App for now.")
}

func (b *Bot) handleApp(msg *tgbotapi.Message) {
    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonWebApp("📱 Open Mini App", tgbotapi.WebAppInfo{URL: b.webAppURL}),
        ),
    )
    b.sendWithKeyboard(msg.Chat.ID, "Open the Mini App to browse, search, and visualize your knowledge:", keyboard)
}
```

**Dependencies:**
```bash
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
```

**Commit:** `feat: implement Telegram bot with message handling`

---

### Phase 4: HTTP API

---

### Task 8: Create HTTP API for Mini App

**Goal:** Build REST API endpoints for the TG Mini App.

**Files to create:**
- `internal/api/server.go` - HTTP server setup
- `internal/api/handlers.go` - API handlers
- `internal/api/middleware.go` - Auth middleware

**Implementation steps:**

1. Create `internal/api/middleware.go`:
```go
package api

import (
    "context"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "net/http"
    "net/url"
    "sort"
    "strings"
)

type contextKey string
const userIDKey contextKey = "userID"

func (s *Server) authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        initData := r.Header.Get("X-Telegram-Init-Data")
        if initData == "" {
            http.Error(w, "missing init data", http.StatusUnauthorized)
            return
        }

        userID, err := s.validateInitData(initData)
        if err != nil {
            http.Error(w, "invalid init data", http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), userIDKey, userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func (s *Server) validateInitData(initData string) (int64, error) {
    // Parse init data
    values, err := url.ParseQuery(initData)
    if err != nil {
        return 0, err
    }

    hash := values.Get("hash")
    if hash == "" {
        return 0, fmt.Errorf("missing hash")
    }

    // Build data check string
    var keys []string
    for k := range values {
        if k != "hash" {
            keys = append(keys, k)
        }
    }
    sort.Strings(keys)

    var dataCheckString strings.Builder
    for i, k := range keys {
        if i > 0 {
            dataCheckString.WriteString("\n")
        }
        dataCheckString.WriteString(k + "=" + values.Get(k))
    }

    // Validate HMAC
    secretKey := hmac.New(sha256.New, []byte("WebAppData"))
    secretKey.Write([]byte(s.botToken))

    h := hmac.New(sha256.New, secretKey.Sum(nil))
    h.Write([]byte(dataCheckString.String()))

    if hex.EncodeToString(h.Sum(nil)) != hash {
        return 0, fmt.Errorf("invalid hash")
    }

    // Extract user ID from user JSON
    // Simplified: parse user JSON to get ID
    // In production, properly parse the JSON
    userJSON := values.Get("user")
    // Parse userJSON to extract ID...

    return 0, nil // Placeholder - implement proper parsing
}

func getUserID(ctx context.Context) int64 {
    if id, ok := ctx.Value(userIDKey).(int64); ok {
        return id
    }
    return 0
}
```

2. Create `internal/api/server.go`:
```go
package api

import (
    "encoding/json"
    "net/http"

    "github.com/yourusername/dumper/internal/store"
)

type Server struct {
    stores   *store.Manager
    botToken string
    mux      *http.ServeMux
}

func NewServer(stores *store.Manager, botToken string) *Server {
    s := &Server{
        stores:   stores,
        botToken: botToken,
        mux:      http.NewServeMux(),
    }
    s.routes()
    return s
}

func (s *Server) routes() {
    // API routes (protected)
    api := http.NewServeMux()
    api.HandleFunc("GET /items", s.handleListItems)
    api.HandleFunc("GET /items/{id}", s.handleGetItem)
    api.HandleFunc("DELETE /items/{id}", s.handleDeleteItem)
    api.HandleFunc("GET /search", s.handleSearch)
    api.HandleFunc("GET /tags", s.handleGetTags)
    api.HandleFunc("GET /graph", s.handleGetGraph)
    api.HandleFunc("POST /ask", s.handleAsk)
    api.HandleFunc("GET /export", s.handleExport)

    s.mux.Handle("/api/", http.StripPrefix("/api", s.authMiddleware(api)))

    // Static files for Mini App
    s.mux.Handle("/", http.FileServer(http.Dir("web/dist")))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    s.mux.ServeHTTP(w, r)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
```

3. Create `internal/api/handlers.go`:
```go
package api

import (
    "net/http"
    "strconv"
)

func (s *Server) handleListItems(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context())

    vault, err := s.stores.GetVault(userID)
    if err != nil {
        jsonError(w, "failed to access vault", http.StatusInternalServerError)
        return
    }

    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit <= 0 || limit > 100 {
        limit = 20
    }
    offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

    items, err := vault.ListItems(limit, offset)
    if err != nil {
        jsonError(w, "failed to list items", http.StatusInternalServerError)
        return
    }

    jsonResponse(w, items)
}

func (s *Server) handleGetItem(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context())
    itemID := r.PathValue("id")

    vault, err := s.stores.GetVault(userID)
    if err != nil {
        jsonError(w, "failed to access vault", http.StatusInternalServerError)
        return
    }

    item, err := vault.GetItem(itemID)
    if err != nil {
        jsonError(w, "failed to get item", http.StatusInternalServerError)
        return
    }
    if item == nil {
        jsonError(w, "item not found", http.StatusNotFound)
        return
    }

    jsonResponse(w, item)
}

func (s *Server) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context())
    itemID := r.PathValue("id")

    vault, err := s.stores.GetVault(userID)
    if err != nil {
        jsonError(w, "failed to access vault", http.StatusInternalServerError)
        return
    }

    if err := vault.DeleteItem(itemID); err != nil {
        jsonError(w, "failed to delete item", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context())
    query := r.URL.Query().Get("q")
    if query == "" {
        jsonError(w, "query required", http.StatusBadRequest)
        return
    }

    vault, err := s.stores.GetVault(userID)
    if err != nil {
        jsonError(w, "failed to access vault", http.StatusInternalServerError)
        return
    }

    results, err := vault.Search(query, 20)
    if err != nil {
        jsonError(w, "search failed", http.StatusInternalServerError)
        return
    }

    jsonResponse(w, results)
}

func (s *Server) handleGetTags(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context())

    vault, err := s.stores.GetVault(userID)
    if err != nil {
        jsonError(w, "failed to access vault", http.StatusInternalServerError)
        return
    }

    tags, err := vault.GetAllTags()
    if err != nil {
        jsonError(w, "failed to get tags", http.StatusInternalServerError)
        return
    }

    jsonResponse(w, tags)
}

func (s *Server) handleGetGraph(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context())

    vault, err := s.stores.GetVault(userID)
    if err != nil {
        jsonError(w, "failed to access vault", http.StatusInternalServerError)
        return
    }

    items, relationships, err := vault.GetGraph()
    if err != nil {
        jsonError(w, "failed to get graph", http.StatusInternalServerError)
        return
    }

    jsonResponse(w, map[string]interface{}{
        "nodes": items,
        "edges": relationships,
    })
}

func (s *Server) handleAsk(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement Q&A with LLM
    jsonError(w, "not implemented", http.StatusNotImplemented)
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement Obsidian export
    jsonError(w, "not implemented", http.StatusNotImplemented)
}
```

**Commit:** `feat: add HTTP API for Mini App`

---

### Task 9: Wire Everything Together in Main

**Goal:** Connect all components in main.go.

**Files to modify:**
- `cmd/dumper/main.go` - Wire dependencies and start services

**Implementation:**

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "golang.org/x/sync/errgroup"

    "github.com/yourusername/dumper/internal/api"
    "github.com/yourusername/dumper/internal/bot"
    "github.com/yourusername/dumper/internal/config"
    "github.com/yourusername/dumper/internal/ingest"
    "github.com/yourusername/dumper/internal/llm"
    "github.com/yourusername/dumper/internal/store"
)

func main() {
    // Setup logging
    level := slog.LevelInfo
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
    slog.SetDefault(logger)

    if err := run(); err != nil {
        slog.Error("fatal error", "error", err)
        os.Exit(1)
    }
}

func run() error {
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }

    // Initialize store manager
    stores, err := store.NewManager(cfg.DataDir)
    if err != nil {
        return fmt.Errorf("create store manager: %w", err)
    }
    defer stores.Close()

    // Initialize LLM client
    llmClient := llm.NewClient(cfg.OpenRouterKey, cfg.OpenRouterModel)

    // Initialize processing pipeline
    pipeline := ingest.NewPipeline(llmClient, stores)

    // Initialize bot
    webAppURL := fmt.Sprintf("https://your-domain.com") // TODO: configure
    tgBot, err := bot.New(cfg.TelegramToken, pipeline, stores, webAppURL)
    if err != nil {
        return fmt.Errorf("create bot: %w", err)
    }

    // Initialize API server
    apiServer := api.NewServer(stores, cfg.TelegramToken)

    // Setup graceful shutdown
    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    g, ctx := errgroup.WithContext(ctx)

    // Run bot
    g.Go(func() error {
        slog.Info("starting telegram bot")
        return tgBot.Run(ctx)
    })

    // Run HTTP server
    g.Go(func() error {
        addr := fmt.Sprintf(":%d", cfg.HTTPPort)
        slog.Info("starting http server", "addr", addr)

        server := &http.Server{Addr: addr, Handler: apiServer}

        go func() {
            <-ctx.Done()
            server.Shutdown(context.Background())
        }()

        return server.ListenAndServe()
    })

    return g.Wait()
}
```

**Dependencies:**
```bash
go get golang.org/x/sync/errgroup
```

**Commit:** `feat: wire all components in main`

---

### Phase 5: Mini App Frontend (Deferred)

The Mini App frontend is a significant effort. For MVP, create a minimal Svelte app with:
- Item list view
- Search functionality
- Basic graph visualization with d3-force

This is detailed in separate tasks (10-15) but can be implemented after backend is stable.

---

### Phase 6: Export & Docker

---

### Task 16: Implement Obsidian Export

**Goal:** Export user's vault to Obsidian-compatible markdown.

**Files to create:**
- `internal/export/obsidian.go` - Markdown export logic

**Implementation:**

```go
package export

import (
    "archive/zip"
    "bytes"
    "fmt"
    "io"
    "strings"
    "time"

    "github.com/yourusername/dumper/internal/store"
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
        sb.WriteString(fmt.Sprintf("url: %s\n", item.URL))
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
    s = strings.ReplaceAll(s, "/", "-")
    s = strings.ReplaceAll(s, "\\", "-")
    s = strings.ReplaceAll(s, ":", "-")
    s = strings.ReplaceAll(s, "*", "")
    s = strings.ReplaceAll(s, "?", "")
    s = strings.ReplaceAll(s, "\"", "")
    s = strings.ReplaceAll(s, "<", "")
    s = strings.ReplaceAll(s, ">", "")
    s = strings.ReplaceAll(s, "|", "-")
    if len(s) > 100 {
        s = s[:100]
    }
    return s
}
```

**Commit:** `feat: add Obsidian markdown export`

---

### Task 17: Create Docker Configuration

**Goal:** Containerize the application.

**Files to create:**
- `Dockerfile` - Multi-stage build
- `docker-compose.yml` - Development setup

**Implementation:**

1. Create `Dockerfile`:
```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o dumper ./cmd/dumper

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates sqlite

WORKDIR /app
COPY --from=builder /app/dumper .
COPY --from=builder /app/web/dist ./web/dist
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./dumper"]
```

2. Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  dumper:
    build: .
    ports:
      - "8080:8080"
    environment:
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
      - OPENROUTER_API_KEY=${OPENROUTER_API_KEY}
      - DATA_DIR=/data
      - HTTP_PORT=8080
      - LOG_LEVEL=info
    volumes:
      - dumper-data:/data
    restart: unless-stopped

volumes:
  dumper-data:
```

**Commit:** `feat: add Docker configuration`

---

## Testing Strategy

### Unit Tests
- `internal/store/*_test.go` - Test CRUD, search, migrations
- `internal/llm/*_test.go` - Test prompt formatting (mock API)
- `internal/ingest/*_test.go` - Test extraction, pipeline
- `internal/export/*_test.go` - Test markdown generation

### Integration Tests
- Test full flow: message → extraction → LLM → storage → retrieval
- Use test fixtures and mock LLM responses

### Running Tests
```bash
go test ./... -v
go test ./... -race  # Check for race conditions
go test ./... -cover # Check coverage
```

---

## Documentation Updates

After implementation:
- [ ] Update README with setup instructions
- [ ] Add API documentation (OpenAPI/Swagger)
- [ ] Document environment variables
- [ ] Add deployment guide

---

## Definition of Done

- [ ] All Phase 1-4 tasks implemented
- [ ] Bot responds to messages and saves items
- [ ] Search returns relevant results
- [ ] API endpoints functional
- [ ] Tests passing with >70% coverage
- [ ] Docker build succeeds
- [ ] Can deploy and run locally
- [ ] Basic error handling in place

---

*Generated via /brainstorm-plan on 2026-01-16*
