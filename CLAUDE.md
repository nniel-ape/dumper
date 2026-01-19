# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Dumper is a personal knowledge capture system. Users forward links and text to a Telegram bot, which extracts content, generates LLM-powered summaries/tags via OpenRouter, and stores everything in per-user SQLite databases. Features include full-text search, knowledge graph relationships, and Obsidian markdown export.

## Build & Run Commands

```bash
# Verify build compiles (use go test, never go build - don't create binary artifacts)
go test ./...

# Run (requires environment variables)
go run ./cmd/dumper

# Docker
docker build -t dumper .
docker run -e TELEGRAM_BOT_TOKEN=... -e OPENROUTER_API_KEY=... dumper
```

**IMPORTANT**: Never run `go build` directly - use `go test ./...` to verify compilation. Do not commit binary artifacts.

## Configuration

Configure via CLI flags or environment variables (flags take precedence):

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `--telegram-token` | `TELEGRAM_BOT_TOKEN` | (required) | Telegram bot token |
| `--openrouter-key` | `OPENROUTER_API_KEY` | (required) | OpenRouter API key |
| `--data-dir` | `DATA_DIR` | `./data` | SQLite database directory |
| `--http-port` | `HTTP_PORT` | `8080` | HTTP server port |
| `--log-level` | `LOG_LEVEL` | `info` | debug/info/warn/error |
| `--openrouter-model` | `OPENROUTER_MODEL` | `anthropic/claude-3-haiku` | LLM model |
| `--webapp-url` | `WEBAPP_URL` | (empty) | Mini App URL |

Run with `--help` to see all options.

## Architecture

```
cmd/dumper/main.go     # Entry point - wires dependencies, runs bot + HTTP server concurrently
internal/
├── config/            # CLI/env config via jessevdk/go-flags
├── bot/               # Telegram bot (go-telegram-bot-api/v5)
│   ├── bot.go         # Bot setup, update loop, routing
│   └── handlers.go    # Command and message handlers
├── ingest/            # Content processing pipeline
│   ├── source.go      # RawContent type, ContentType enum
│   ├── extractor.go   # URL content extraction (go-readability)
│   ├── detector.go    # Short topic detection for web search
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
├── i18n/              # Internationalization (EN, RU)
│   └── i18n.go        # Localizer with fallback to English
├── search/            # External search integration
│   └── duckduckgo.go  # Instant Answer API for topic enrichment
└── export/            # Obsidian markdown export
    └── obsidian.go    # ZIP generation with YAML frontmatter
migrations/
└── 001_init.sql       # Schema: items, tags, relationships, FTS5
```

## Key Patterns

**Multi-tenancy**: Each Telegram user gets isolated SQLite file at `data/users/{user_id}/vault.db`. Manager uses double-checked locking for lazy vault initialization.

**Processing Flow**: Message → `ingest.Pipeline.Process()` → URL extraction (readability) OR topic detection (short messages trigger DuckDuckGo lookup) → LLM summarization → Store with auto-generated UUID.

**Graceful Degradation**: LLM failures fall back to "uncategorized" tag. Extraction failures save URL with minimal metadata.

**API Auth**: Mini App validates Telegram init data via HMAC-SHA256 (X-Telegram-Init-Data header).

**i18n**: Bot messages are localized. `i18n.New(langCode)` creates a Localizer; Russian/Ukrainian/Belarusian users get Russian UI, others get English. Language preferences are cached per-user in memory.

## Database Schema

SQLite with FTS5 for search. Key tables:
- `items` (id, type, url, title, content, summary, raw_content, timestamps)
- `tags` (id, name)
- `item_tags` (junction table)
- `relationships` (source_id, target_id, relation_type, strength)
- `items_fts` (FTS5 virtual table with triggers for sync)

## Dependencies

Requires **Go 1.25+** (uses `sync.WaitGroup.Go()` added in 1.25).

Uses CGO-free `modernc.org/sqlite` (not mattn/go-sqlite3). Key deps:
- `github.com/go-telegram-bot-api/telegram-bot-api/v5`
- `github.com/go-shiori/go-readability`
- `github.com/jessevdk/go-flags`
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

## Mini App Development

The `mini-app/` directory contains the Telegram Mini App frontend (Vite + React + TypeScript + Tailwind).

```bash
cd mini-app

# Install dependencies
bun install

# Development server
bun run dev

# Build for production
bun run build

# Lint
bun run lint

# Preview production build
bun run preview
```

**IMPORTANT**: Always use `bun` instead of `npm` or `yarn` for package management.

### iOS Safari Compatibility

- ReactFlow requires explicit pixel dimensions; iOS Safari fails with `flex-1 + min-h-0 + absolute`
- Use `position: fixed` with calc-based top/bottom to bypass flex layout bugs
- Safe area CSS vars: `--tg-total-safe-area-top`, `--tg-total-safe-area-bottom`
- Call `swipeBehavior.disableVertical()` to prevent swipe-to-close during graph/canvas interactions
- Always test graph features on real iPhone in Telegram (simulators miss some WebView quirks)

## Current Status

MVP backend is functional. Missing:
- Unit tests
- TG Mini App frontend (mini-app/dist)
- Q&A endpoint implementation
- Export endpoint wiring
- Relationship inference in pipeline
