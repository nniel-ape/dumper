# Dumper - Personal Knowledge Capture System

## Overview

**Dumper** is a personal knowledge capture and organization system. Users forward links, text, and eventually other content to a Telegram bot, which automatically extracts, summarizes, and categorizes the content using LLMs (via OpenRouter). Data is stored in per-user SQLite databases with a multi-tenant architecture ready for SaaS distribution.

The system provides three interfaces:
1. **Telegram Bot** - Primary input method, with conversational Q&A
2. **TG Mini App** - Browse, search, and visualize the knowledge graph
3. **Obsidian Export** - On-demand markdown vault generation

The architecture uses a clean `InputSource` interface to enable future input methods (browser extension, email, REST API) without core changes.

## Goals

**MVP Goals:**
- Accept links and text messages via Telegram bot
- Extract full content from URLs (title, body text, metadata)
- Generate LLM-powered summaries and auto-assign tags/categories
- Store in per-user SQLite database with full-text search
- AI-inferred relationships between items for knowledge graph
- TG Mini App with three core features:
  - Browse/search saved items with tag filtering
  - Interactive knowledge graph visualization
  - Q&A chat interface to query saved knowledge
- On-demand export to Obsidian-compatible markdown vault
- Docker Compose deployment for easy self-hosting

**SaaS-Ready Architecture Goals:**
- Multi-tenant data isolation from day one (SQLite per user)
- Clean `InputSource` interface for adding future input methods
- User authentication and session management
- Usage tracking for future billing integration

## Non-Goals

**Explicitly out of scope for MVP:**
- Image, PDF, audio, or video content processing (text and links only)
- Vector embeddings and semantic search (keyword/FTS only for now)
- Real-time or scheduled Obsidian sync (on-demand export only)
- Browser extension, email forwarding, or REST API inputs
- Collaborative features (sharing, team vaults)
- Mobile apps (TG Mini App is the mobile interface)
- Offline mode or local-first sync
- Custom LLM model selection (OpenRouter with sensible defaults)
- User-defined taxonomy or manual tagging workflows
- Bi-directional Obsidian sync (export only, not import)
- Payment/billing system (architect for it, don't implement)

**Deferred to v2:**
- Vector embeddings for semantic similarity search
- Additional input sources (browser extension, email, API)
- Image/PDF processing with OCR
- Advanced graph analytics (clusters, recommendations)
- Obsidian plugin for live sync

## User Experience

**Telegram Bot Interaction:**
```
User forwards a link → Bot replies "Got it! Processing..."
→ Bot extracts content, summarizes, categorizes
→ Bot replies with: title, summary, auto-assigned tags
→ Item is saved, no user action required
```

```
User sends text message → Bot saves as note
→ Auto-generates title from content
→ Assigns relevant tags
→ Confirms with brief summary
```

```
User types "/search golang concurrency"
→ Bot returns top matches with snippets
→ User can tap to view full item in Mini App
```

**TG Mini App Experience:**
- **Home/Browse**: Card grid of recent items, filter by tags, sort by date/relevance
- **Search**: Search bar with instant results, highlights matching text
- **Graph**: Force-directed graph visualization, tap nodes to preview, pinch to zoom
- **Q&A**: Chat interface - "What did I save about X?" returns synthesized answer with source links
- **Settings**: Obsidian export button, account preferences

**Obsidian Export:**
- User taps "Export to Obsidian" in settings
- Downloads a `.zip` containing:
  - `/notes/` - One `.md` file per item with YAML frontmatter (tags, date, source)
  - `/attachments/` - Future: images, PDFs
  - Graph links as `[[wikilinks]]` in note bodies
- User unzips into their Obsidian vault

## Technical Approach

**Language & Framework:**
- Go 1.22+ for all backend services
- Single binary containing: TG bot, HTTP API, Mini App server
- Standard library + minimal dependencies philosophy

**Data Storage:**
- SQLite per user (e.g., `data/users/{user_id}/vault.db`)
- Schema: `items`, `tags`, `item_tags`, `relationships`, `settings`
- FTS5 for full-text search
- User metadata in a central `users.db` (auth, preferences, usage)

**AI/LLM Integration:**
- OpenRouter API for all LLM calls (model flexibility)
- Processing pipeline: Extract → Summarize → Categorize → Relate
- Single prompt chain or separate calls (configurable)
- Graceful degradation if API unavailable

**Content Extraction:**
- `go-readability` or `colly` for HTML parsing
- Fallback to basic metadata (title, description, favicon)
- Rate limiting and timeout handling for external fetches

**TG Mini App:**
- Served as static files from Go backend
- Frontend: lightweight (vanilla JS + minimal framework, or Svelte)
- Graph visualization: `d3-force` or `cytoscape.js`
- Communicates with backend via JSON API

**Multi-tenancy:**
- Telegram user ID as primary identifier
- Each user gets isolated SQLite file
- Connection pool per active user, lazy initialization

## Key Components

```
┌─────────────────────────────────────────────────────────────┐
│                        Dumper                               │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐ │
│  │ InputSource │    │ InputSource │    │   InputSource   │ │
│  │  (Telegram) │    │   (Future)  │    │    (Future)     │ │
│  └──────┬──────┘    └──────┬──────┘    └────────┬────────┘ │
│         │                  │                    │          │
│         └──────────────────┼────────────────────┘          │
│                            ▼                               │
│                   ┌────────────────┐                       │
│                   │  Ingestion Hub │                       │
│                   └────────┬───────┘                       │
│                            ▼                               │
│         ┌──────────────────────────────────────┐           │
│         │        Processing Pipeline           │           │
│         │  ┌─────────┐ ┌─────────┐ ┌─────────┐│           │
│         │  │Extract  │→│Summarize│→│   Tag   ││           │
│         │  └─────────┘ └─────────┘ └─────────┘│           │
│         └──────────────────┬───────────────────┘           │
│                            ▼                               │
│    ┌───────────────────────────────────────────────┐       │
│    │              Storage Layer                    │       │
│    │  ┌─────────┐  ┌────────────┐  ┌────────────┐ │       │
│    │  │users.db │  │user_X/vault│  │user_Y/vault│ │       │
│    │  └─────────┘  └────────────┘  └────────────┘ │       │
│    └───────────────────────┬───────────────────────┘       │
│                            ▼                               │
│         ┌──────────────────────────────────────┐           │
│         │           HTTP API                   │           │
│         └──────────────────┬───────────────────┘           │
│                            ▼                               │
│    ┌────────────┐  ┌────────────┐  ┌────────────────┐      │
│    │ TG MiniApp │  │  Exporter  │  │ Q&A Service    │      │
│    │  (Static)  │  │ (Markdown) │  │ (LLM + Search) │      │
│    └────────────┘  └────────────┘  └────────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

**Key Interfaces:**
- `InputSource` - Abstraction for content ingestion (TG implements this)
- `Processor` - Content extraction, summarization, tagging steps
- `VaultStore` - Per-user SQLite operations (CRUD, search, relationships)
- `Exporter` - Generates Obsidian-compatible markdown

**Core Packages:**
- `cmd/dumper` - Main entrypoint, wires everything together
- `internal/bot` - Telegram bot implementation
- `internal/ingest` - InputSource interface + processing pipeline
- `internal/llm` - OpenRouter client, prompt templates
- `internal/store` - SQLite operations, user vault management
- `internal/api` - HTTP handlers for Mini App
- `internal/export` - Markdown/Obsidian export logic
- `web/` - Mini App static assets (HTML/JS/CSS)

## Open Questions

**Product/UX:**
- Should the bot support inline mode for quick saves from any chat?
- How to handle duplicate links? Auto-merge, warn, or allow duplicates?
- Should users be able to delete items via bot commands or only in Mini App?
- What happens when LLM categorization fails? Save with "uncategorized" tag?

**Technical:**
- Which OpenRouter model to default to? (Claude Haiku for speed vs Sonnet for quality)
- How to handle rate limits from OpenRouter? Queue with retries or fail fast?
- SQLite file location: local disk vs S3-compatible storage for SaaS scale?
- Mini App framework choice: Vanilla JS (smaller), Svelte (DX), or Vue (familiarity)?
- Graph relationship storage: separate table or JSON field in items table?

**SaaS/Business:**
- Free tier limits? (items per month, storage, API calls)
- How to handle user data export/deletion for GDPR compliance?
- Should self-hosted version be open source or source-available?

**Security:**
- How to validate Telegram Mini App init data?
- Rate limiting strategy per user to prevent abuse?
- Content sanitization for user-submitted text/links?

---
*Generated via /brainstorm on 2026-01-16*
