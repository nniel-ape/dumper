# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Dumper is a personal knowledge capture system. Users forward links and text to a Telegram bot, which extracts content, generates LLM-powered summaries/tags via OpenRouter, and stores everything in per-user SQLite databases. Features include full-text search, knowledge graph relationships, and Obsidian markdown export.

## Build & Run Commands

```bash
# Build (uses CGO-free modernc.org/sqlite)
go build ./cmd/dumper

# Run (requires environment variables)
./dumper

# Test (no tests exist yet - create with table-driven tests)
go test ./...

# Docker
docker build -t dumper .
docker run -e TELEGRAM_BOT_TOKEN=... -e OPENROUTER_API_KEY=... dumper
```

## Environment Variables

Required:
- `TELEGRAM_BOT_TOKEN` - From @BotFather
- `OPENROUTER_API_KEY` - From openrouter.ai/keys

Optional:
- `DATA_DIR` (default: `./data`) - Where per-user SQLite databases are stored
- `HTTP_PORT` (default: `8080`)
- `LOG_LEVEL` (default: `info`) - debug/info/warn/error
- `OPENROUTER_MODEL` (default: `anthropic/claude-3-haiku`)
- `WEBAPP_URL` - TG Mini App URL for bot buttons

## Architecture

```
cmd/dumper/main.go     # Entry point - wires dependencies, runs bot + HTTP server concurrently
internal/
├── config/            # Environment-based config via caarlos0/env
├── bot/               # Telegram bot (go-telegram-bot-api/v5)
│   ├── bot.go         # Bot setup, update loop, routing
│   └── handlers.go    # Command and message handlers
├── ingest/            # Content processing pipeline
│   ├── source.go      # RawContent type, ContentType enum
│   ├── extractor.go   # URL content extraction (go-readability)
│   └── pipeline.go    # Extract → LLM process → Store flow
├── llm/               # OpenRouter API client
│   ├── client.go      # HTTP client with Chat() and ProcessContent()
│   ├── types.go       # Request/response structs
│   └── prompts.go     # Prompt templates for summarization/tagging
├── store/             # SQLite storage layer
│   ├── store.go       # Manager (multi-tenant), VaultStore
│   ├── models.go      # Item, Relationship, SearchResult types
│   ├── items.go       # CRUD + FTS5 search
│   ├── relationships.go # Graph edges
│   └── migrations.go  # Embedded SQL migrations
├── api/               # HTTP API for TG Mini App
│   ├── server.go      # Routes, CORS, mux setup
│   ├── middleware.go  # Telegram init data validation
│   └── handlers.go    # REST endpoints
└── export/            # Obsidian markdown export
    └── obsidian.go    # ZIP generation with YAML frontmatter
migrations/
└── 001_init.sql       # Schema: items, tags, relationships, FTS5
```

## Key Patterns

**Multi-tenancy**: Each Telegram user gets isolated SQLite file at `data/users/{user_id}/vault.db`. Manager uses double-checked locking for lazy vault initialization.

**Processing Flow**: Message → `ingest.Pipeline.Process()` → URL extraction (readability) → LLM summarization → Store with auto-generated UUID.

**Graceful Degradation**: LLM failures fall back to "uncategorized" tag. Extraction failures save URL with minimal metadata.

**API Auth**: Mini App validates Telegram init data via HMAC-SHA256 (X-Telegram-Init-Data header).

## Database Schema

SQLite with FTS5 for search. Key tables:
- `items` (id, type, url, title, content, summary, raw_content, timestamps)
- `tags` (id, name)
- `item_tags` (junction table)
- `relationships` (source_id, target_id, relation_type, strength)
- `items_fts` (FTS5 virtual table with triggers for sync)

## Dependencies

Uses CGO-free `modernc.org/sqlite` (not mattn/go-sqlite3). Key deps:
- `github.com/go-telegram-bot-api/telegram-bot-api/v5`
- `github.com/go-shiori/go-readability`
- `github.com/caarlos0/env/v11`
- `golang.org/x/sync/errgroup`

## API Endpoints

All `/api/*` routes require `X-Telegram-Init-Data` header:
- `GET /api/items` - List items (limit, offset params)
- `GET /api/items/{id}` - Get single item
- `DELETE /api/items/{id}` - Delete item
- `GET /api/search?q=` - FTS5 search
- `GET /api/tags` - All user tags
- `GET /api/graph` - Items + relationships for visualization
- `GET /api/stats` - User statistics
- `POST /api/ask` - Q&A (not implemented)
- `GET /api/export` - Obsidian export (not implemented in handler)

## Current Status

MVP backend is functional. Missing:
- Unit tests
- TG Mini App frontend (web/dist)
- Q&A endpoint implementation
- Export endpoint wiring
- Relationship inference in pipeline
