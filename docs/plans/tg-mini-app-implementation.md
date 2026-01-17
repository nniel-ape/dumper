# Telegram Mini App Implementation Plan

## Overview

Build a React-based Telegram Mini App for Dumper that provides: (1) an infinite-scroll feed of saved items, (2) unified search with AI-powered Q&A, (3) an interactive knowledge graph using React Flow, and (4) a settings page with stats and export.

**Design doc:** [`docs/brainstorms/tg-mini-app.md`](../brainstorms/tg-mini-app.md)

---

## Prerequisites

### Required Tools
- **Node.js** 20+ (check: `node --version`)
- **npm** 10+ (check: `npm --version`)
- **Go** 1.25+ (for running the backend)

### Environment Setup
```bash
# Clone and enter the project
cd /path/to/dumper

# Backend env vars (create .env or export)
export TELEGRAM_BOT_TOKEN="your-bot-token"
export OPENROUTER_API_KEY="your-api-key"
export DATA_DIR="./data"
export HTTP_PORT="8080"
```

### Telegram Bot Setup
1. Talk to [@BotFather](https://t.me/BotFather) on Telegram
2. Create a bot or use existing one
3. Run `/newapp` to create a Mini App
4. Set the Web App URL to your deployment URL (or use ngrok for local dev)
5. Note: For local development, you'll need HTTPS. Use ngrok: `ngrok http 8080`

---

## Codebase Orientation

### Key Backend Files
| File | Purpose |
|------|---------|
| `internal/api/server.go` | HTTP routes, serves `web/dist/` at `/` |
| `internal/api/handlers.go` | All API endpoints: items, search, tags, graph, ask, export, stats |
| `internal/api/middleware.go` | Telegram init data validation via HMAC-SHA256 |
| `internal/store/models.go` | Data types: `Item`, `Relationship`, `SearchResult` |

### API Endpoints (all require `X-Telegram-Init-Data` header)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/items?limit=20&offset=0` | List items (paginated) |
| GET | `/api/items?tag=tagname` | Filter by tag |
| GET | `/api/items/{id}` | Get single item |
| DELETE | `/api/items/{id}` | Delete item |
| GET | `/api/search?q=query` | Full-text search |
| GET | `/api/tags` | All user's tags |
| GET | `/api/graph` | Nodes + edges for visualization |
| POST | `/api/ask` | Q&A with body `{"question": "..."}` |
| GET | `/api/stats` | Item count, tag count |
| GET | `/api/export` | Download Obsidian ZIP |

### Data Types (from `internal/store/models.go`)
```typescript
// Mirror these in TypeScript
interface Item {
  id: string;
  type: 'link' | 'note' | 'image' | 'search';
  url?: string;
  title: string;
  content?: string;
  summary?: string;
  image_path?: string;
  tags: string[];
  created_at: string; // ISO 8601
  updated_at: string;
}

interface Relationship {
  id: number;
  source_id: string;
  target_id: string;
  relation_type: string;
  strength: number;
}

interface SearchResult {
  item: Item;
  snippet?: string;
  score: number;
}
```

### Development Auth Shortcut
The middleware allows `?user_id=123` query param for dev (see `middleware.go:26-33`). Use this for local testing without Telegram.

---

## Implementation Tasks

### Phase 1: Project Scaffolding

---

### Task 1: Initialize Vite + React + TypeScript Project

**Goal:** Create the `web/` directory with a working Vite React TypeScript setup.

**Files to create:**
- `web/package.json`
- `web/tsconfig.json`
- `web/tsconfig.node.json`
- `web/vite.config.ts`
- `web/index.html`
- `web/src/main.tsx`
- `web/src/App.tsx`
- `web/src/vite-env.d.ts`

**Implementation steps:**
1. Create the web directory:
   ```bash
   mkdir -p web
   cd web
   ```
2. Initialize with Vite:
   ```bash
   npm create vite@latest . -- --template react-ts
   ```
3. Install dependencies:
   ```bash
   npm install
   ```
4. Update `vite.config.ts` to output to `dist/`:
   ```typescript
   import { defineConfig } from 'vite'
   import react from '@vitejs/plugin-react'
   import path from 'path'

   export default defineConfig({
     plugins: [react()],
     resolve: {
       alias: {
         '@': path.resolve(__dirname, './src'),
       },
     },
     build: {
       outDir: 'dist',
       emptyDirOnBuild: true,
     },
     server: {
       proxy: {
         '/api': 'http://localhost:8080',
       },
     },
   })
   ```
5. Update `tsconfig.json` paths:
   ```json
   {
     "compilerOptions": {
       "baseUrl": ".",
       "paths": {
         "@/*": ["./src/*"]
       }
     }
   }
   ```

**Verification:**
```bash
cd web
npm run dev
# Should open http://localhost:5173 with Vite welcome page
npm run build
# Should create web/dist/ with index.html and assets
```

**Commit:** `feat(web): initialize vite react typescript project`

---

### Task 2: Add Tailwind CSS

**Goal:** Configure Tailwind CSS with Telegram theme variable support.

**Files to touch:**
- `web/package.json` - add tailwind deps
- `web/tailwind.config.ts` - create config
- `web/postcss.config.js` - create postcss config
- `web/src/index.css` - tailwind directives + telegram vars

**Implementation steps:**
1. Install Tailwind and dependencies:
   ```bash
   cd web
   npm install -D tailwindcss postcss autoprefixer
   npx tailwindcss init -p --ts
   ```
2. Create `web/tailwind.config.ts`:
   ```typescript
   import type { Config } from 'tailwindcss'

   export default {
     darkMode: 'class',
     content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
     theme: {
       extend: {
         colors: {
           // Telegram theme colors
           tg: {
             bg: 'var(--tg-theme-bg-color, #ffffff)',
             'secondary-bg': 'var(--tg-theme-secondary-bg-color, #f0f0f0)',
             text: 'var(--tg-theme-text-color, #000000)',
             hint: 'var(--tg-theme-hint-color, #999999)',
             link: 'var(--tg-theme-link-color, #2481cc)',
             button: 'var(--tg-theme-button-color, #2481cc)',
             'button-text': 'var(--tg-theme-button-text-color, #ffffff)',
             'header-bg': 'var(--tg-theme-header-bg-color, #ffffff)',
             accent: 'var(--tg-theme-accent-text-color, #2481cc)',
             destructive: 'var(--tg-theme-destructive-text-color, #ff3b30)',
           },
         },
       },
     },
     plugins: [],
   } satisfies Config
   ```
3. Create `web/src/index.css`:
   ```css
   @tailwind base;
   @tailwind components;
   @tailwind utilities;

   :root {
     /* Fallback values when not in Telegram */
     --tg-theme-bg-color: #ffffff;
     --tg-theme-secondary-bg-color: #f7f7f7;
     --tg-theme-text-color: #000000;
     --tg-theme-hint-color: #999999;
     --tg-theme-link-color: #2481cc;
     --tg-theme-button-color: #2481cc;
     --tg-theme-button-text-color: #ffffff;
   }

   body {
     background-color: var(--tg-theme-bg-color);
     color: var(--tg-theme-text-color);
     font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
     margin: 0;
     padding: 0;
     min-height: 100vh;
     overflow-x: hidden;
   }
   ```
4. Import in `web/src/main.tsx`:
   ```typescript
   import './index.css'
   ```

**Verification:**
```bash
cd web
npm run dev
# Add some tailwind classes to App.tsx like:
# <div className="bg-tg-bg text-tg-text p-4">Hello</div>
# Should render with proper colors
```

**Commit:** `feat(web): add tailwind css with telegram theme colors`

---

### Task 3: Add shadcn/ui

**Goal:** Install and configure shadcn/ui for accessible, styled components.

**Files to touch:**
- `web/package.json` - deps
- `web/components.json` - shadcn config
- `web/src/lib/utils.ts` - cn helper
- `web/tailwind.config.ts` - extend for shadcn
- `web/src/index.css` - shadcn CSS variables

**Implementation steps:**
1. Install shadcn/ui dependencies:
   ```bash
   cd web
   npm install class-variance-authority clsx tailwind-merge lucide-react
   npm install -D @types/node
   ```
2. Create `web/src/lib/utils.ts`:
   ```typescript
   import { type ClassValue, clsx } from 'clsx'
   import { twMerge } from 'tailwind-merge'

   export function cn(...inputs: ClassValue[]) {
     return twMerge(clsx(inputs))
   }
   ```
3. Create `web/components.json`:
   ```json
   {
     "$schema": "https://ui.shadcn.com/schema.json",
     "style": "default",
     "rsc": false,
     "tsx": true,
     "tailwind": {
       "config": "tailwind.config.ts",
       "css": "src/index.css",
       "baseColor": "neutral",
       "cssVariables": true,
       "prefix": ""
     },
     "aliases": {
       "components": "@/components",
       "utils": "@/lib/utils",
       "ui": "@/components/ui",
       "lib": "@/lib",
       "hooks": "@/hooks"
     },
     "iconLibrary": "lucide"
   }
   ```
4. Update `web/src/index.css` to add shadcn CSS variables that map to Telegram:
   ```css
   @tailwind base;
   @tailwind components;
   @tailwind utilities;

   @layer base {
     :root {
       /* Telegram fallbacks */
       --tg-theme-bg-color: #ffffff;
       --tg-theme-secondary-bg-color: #f7f7f7;
       --tg-theme-text-color: #000000;
       --tg-theme-hint-color: #999999;
       --tg-theme-link-color: #2481cc;
       --tg-theme-button-color: #2481cc;
       --tg-theme-button-text-color: #ffffff;

       /* Map shadcn vars to Telegram */
       --background: var(--tg-theme-bg-color);
       --foreground: var(--tg-theme-text-color);
       --card: var(--tg-theme-secondary-bg-color);
       --card-foreground: var(--tg-theme-text-color);
       --popover: var(--tg-theme-bg-color);
       --popover-foreground: var(--tg-theme-text-color);
       --primary: var(--tg-theme-button-color);
       --primary-foreground: var(--tg-theme-button-text-color);
       --secondary: var(--tg-theme-secondary-bg-color);
       --secondary-foreground: var(--tg-theme-text-color);
       --muted: var(--tg-theme-secondary-bg-color);
       --muted-foreground: var(--tg-theme-hint-color);
       --accent: var(--tg-theme-accent-text-color, var(--tg-theme-link-color));
       --accent-foreground: var(--tg-theme-button-text-color);
       --destructive: var(--tg-theme-destructive-text-color, #ff3b30);
       --destructive-foreground: #ffffff;
       --border: var(--tg-theme-hint-color);
       --input: var(--tg-theme-secondary-bg-color);
       --ring: var(--tg-theme-button-color);
       --radius: 0.5rem;
     }
   }

   body {
     background-color: hsl(var(--background));
     color: hsl(var(--foreground));
     font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
   }
   ```
5. Add shadcn components we'll need:
   ```bash
   cd web
   npx shadcn@latest add button input sheet card skeleton tabs
   ```

**Verification:**
```bash
cd web
npm run dev
# Import and use a Button in App.tsx:
# import { Button } from '@/components/ui/button'
# <Button>Click me</Button>
# Should render styled button
```

**Commit:** `feat(web): add shadcn/ui with telegram theme integration`

---

### Task 4: Add Telegram SDK and TanStack Query

**Goal:** Set up Telegram WebApp SDK and React Query for data fetching.

**Files to touch:**
- `web/package.json` - deps
- `web/src/lib/telegram.ts` - telegram helpers
- `web/src/main.tsx` - providers
- `web/src/App.tsx` - SDK initialization

**Implementation steps:**
1. Install dependencies:
   ```bash
   cd web
   npm install @telegram-apps/sdk-react @tanstack/react-query
   ```
2. Create `web/src/lib/telegram.ts`:
   ```typescript
   import {
     init,
     miniApp,
     themeParams,
     viewport,
     backButton,
     mainButton,
   } from '@telegram-apps/sdk-react'

   export function initTelegramApp() {
     try {
       init()

       // Expand viewport
       if (viewport.mount.isAvailable()) {
         viewport.mount()
         viewport.expand()
       }

       // Mount theme params
       if (themeParams.mount.isAvailable()) {
         themeParams.mount()
       }

       // Mount mini app
       if (miniApp.mount.isAvailable()) {
         miniApp.mount()
         miniApp.ready()
       }

       // Mount back button
       if (backButton.mount.isAvailable()) {
         backButton.mount()
       }

       return true
     } catch (e) {
       console.warn('Not running in Telegram:', e)
       return false
     }
   }

   export function getInitData(): string {
     try {
       // @ts-expect-error - Telegram WebApp global
       return window.Telegram?.WebApp?.initData || ''
     } catch {
       return ''
     }
   }

   export function hapticFeedback(type: 'light' | 'medium' | 'heavy' | 'success' | 'error') {
     try {
       // @ts-expect-error - Telegram WebApp global
       const haptic = window.Telegram?.WebApp?.HapticFeedback
       if (haptic) {
         if (type === 'success' || type === 'error') {
           haptic.notificationOccurred(type)
         } else {
           haptic.impactOccurred(type)
         }
       }
     } catch {
       // Not in Telegram
     }
   }

   export function openLink(url: string) {
     try {
       // @ts-expect-error - Telegram WebApp global
       window.Telegram?.WebApp?.openLink(url)
     } catch {
       window.open(url, '_blank')
     }
   }
   ```
3. Update `web/src/main.tsx`:
   ```typescript
   import React from 'react'
   import ReactDOM from 'react-dom/client'
   import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
   import App from './App'
   import './index.css'
   import { initTelegramApp } from './lib/telegram'

   // Initialize Telegram SDK
   initTelegramApp()

   const queryClient = new QueryClient({
     defaultOptions: {
       queries: {
         staleTime: 1000 * 60, // 1 minute
         retry: 1,
       },
     },
   })

   ReactDOM.createRoot(document.getElementById('root')!).render(
     <React.StrictMode>
       <QueryClientProvider client={queryClient}>
         <App />
       </QueryClientProvider>
     </React.StrictMode>,
   )
   ```
4. Update `web/src/App.tsx` with basic structure:
   ```typescript
   export default function App() {
     return (
       <div className="min-h-screen bg-tg-bg text-tg-text">
         <main className="pb-16">
           <h1 className="p-4 text-xl font-semibold">Dumper</h1>
         </main>
       </div>
     )
   }
   ```

**Verification:**
```bash
cd web
npm run dev
# Should render without errors
# Check console - may see "Not running in Telegram" warning (expected in browser)
```

**Commit:** `feat(web): add telegram sdk and tanstack query`

---

### Phase 2: API Layer

---

### Task 5: Create API Client with Types

**Goal:** Build a typed API client that handles auth and all backend endpoints.

**Files to create:**
- `web/src/api/types.ts` - TypeScript interfaces
- `web/src/api/client.ts` - fetch wrapper with auth
- `web/src/api/items.ts` - items endpoints
- `web/src/api/search.ts` - search endpoints
- `web/src/api/graph.ts` - graph endpoint
- `web/src/api/index.ts` - barrel export

**Implementation steps:**

1. Create `web/src/api/types.ts`:
   ```typescript
   export type ItemType = 'link' | 'note' | 'image' | 'search'

   export interface Item {
     id: string
     type: ItemType
     url?: string
     title: string
     content?: string
     summary?: string
     image_path?: string
     tags: string[]
     created_at: string
     updated_at: string
   }

   export interface Relationship {
     id: number
     source_id: string
     target_id: string
     relation_type: string
     strength: number
   }

   export interface SearchResult {
     item: Item
     snippet?: string
     score: number
   }

   export interface GraphData {
     nodes: Item[]
     edges: Relationship[]
   }

   export interface Stats {
     items: number
     tags: number
   }

   export interface AskResponse {
     answer: string
     sources: SearchResult[]
   }

   export interface ApiError {
     error: string
   }
   ```

2. Create `web/src/api/client.ts`:
   ```typescript
   import { getInitData } from '@/lib/telegram'
   import type { ApiError } from './types'

   const BASE_URL = '/api'

   // For dev mode without Telegram
   const DEV_USER_ID = import.meta.env.DEV ? '123456789' : null

   class ApiClient {
     private getHeaders(): HeadersInit {
       const headers: HeadersInit = {
         'Content-Type': 'application/json',
       }

       const initData = getInitData()
       if (initData) {
         headers['X-Telegram-Init-Data'] = initData
       }

       return headers
     }

     private getUrl(path: string): string {
       const url = new URL(BASE_URL + path, window.location.origin)

       // Add dev user_id if no init data
       if (DEV_USER_ID && !getInitData()) {
         url.searchParams.set('user_id', DEV_USER_ID)
       }

       return url.toString()
     }

     async get<T>(path: string): Promise<T> {
       const response = await fetch(this.getUrl(path), {
         method: 'GET',
         headers: this.getHeaders(),
       })

       if (!response.ok) {
         const error: ApiError = await response.json().catch(() => ({ error: 'Unknown error' }))
         throw new Error(error.error)
       }

       return response.json()
     }

     async post<T, B = unknown>(path: string, body: B): Promise<T> {
       const response = await fetch(this.getUrl(path), {
         method: 'POST',
         headers: this.getHeaders(),
         body: JSON.stringify(body),
       })

       if (!response.ok) {
         const error: ApiError = await response.json().catch(() => ({ error: 'Unknown error' }))
         throw new Error(error.error)
       }

       return response.json()
     }

     async delete(path: string): Promise<void> {
       const response = await fetch(this.getUrl(path), {
         method: 'DELETE',
         headers: this.getHeaders(),
       })

       if (!response.ok) {
         const error: ApiError = await response.json().catch(() => ({ error: 'Unknown error' }))
         throw new Error(error.error)
       }
     }

     getExportUrl(): string {
       return this.getUrl('/export')
     }
   }

   export const apiClient = new ApiClient()
   ```

3. Create `web/src/api/items.ts`:
   ```typescript
   import { apiClient } from './client'
   import type { Item, Stats } from './types'

   export interface ListItemsParams {
     limit?: number
     offset?: number
     tag?: string
   }

   export async function listItems(params: ListItemsParams = {}): Promise<Item[]> {
     const searchParams = new URLSearchParams()
     if (params.limit) searchParams.set('limit', String(params.limit))
     if (params.offset) searchParams.set('offset', String(params.offset))
     if (params.tag) searchParams.set('tag', params.tag)

     const query = searchParams.toString()
     return apiClient.get<Item[]>(`/items${query ? `?${query}` : ''}`)
   }

   export async function getItem(id: string): Promise<Item> {
     return apiClient.get<Item>(`/items/${id}`)
   }

   export async function deleteItem(id: string): Promise<void> {
     return apiClient.delete(`/items/${id}`)
   }

   export async function getTags(): Promise<string[]> {
     return apiClient.get<string[]>('/tags')
   }

   export async function getStats(): Promise<Stats> {
     return apiClient.get<Stats>('/stats')
   }

   export function getExportUrl(): string {
     return apiClient.getExportUrl()
   }
   ```

4. Create `web/src/api/search.ts`:
   ```typescript
   import { apiClient } from './client'
   import type { SearchResult, AskResponse } from './types'

   export async function search(query: string): Promise<SearchResult[]> {
     return apiClient.get<SearchResult[]>(`/search?q=${encodeURIComponent(query)}`)
   }

   export async function ask(question: string): Promise<AskResponse> {
     return apiClient.post<AskResponse>('/ask', { question })
   }
   ```

5. Create `web/src/api/graph.ts`:
   ```typescript
   import { apiClient } from './client'
   import type { GraphData } from './types'

   export async function getGraph(): Promise<GraphData> {
     return apiClient.get<GraphData>('/graph')
   }
   ```

6. Create `web/src/api/index.ts`:
   ```typescript
   export * from './types'
   export * from './items'
   export * from './search'
   export * from './graph'
   ```

**Verification:**
```bash
# Start the Go backend
go run ./cmd/dumper

# In another terminal
cd web
npm run dev

# In browser console at http://localhost:5173:
# (after importing in App.tsx temporarily)
# import { listItems } from './api'
# listItems().then(console.log)
# Should see items array (or empty array if new user)
```

**Commit:** `feat(web): add typed api client with all endpoints`

---

### Task 6: Create React Query Hooks

**Goal:** Build custom hooks wrapping API calls with TanStack Query.

**Files to create:**
- `web/src/hooks/useItems.ts`
- `web/src/hooks/useSearch.ts`
- `web/src/hooks/useGraph.ts`
- `web/src/hooks/index.ts`

**Implementation steps:**

1. Create `web/src/hooks/useItems.ts`:
   ```typescript
   import {
     useQuery,
     useInfiniteQuery,
     useMutation,
     useQueryClient,
   } from '@tanstack/react-query'
   import {
     listItems,
     getItem,
     deleteItem,
     getTags,
     getStats,
     type ListItemsParams,
   } from '@/api'

   const ITEMS_KEY = 'items'
   const TAGS_KEY = 'tags'
   const STATS_KEY = 'stats'
   const PAGE_SIZE = 20

   export function useItems(tag?: string) {
     return useInfiniteQuery({
       queryKey: [ITEMS_KEY, { tag }],
       queryFn: ({ pageParam = 0 }) =>
         listItems({ limit: PAGE_SIZE, offset: pageParam, tag }),
       getNextPageParam: (lastPage, allPages) => {
         if (lastPage.length < PAGE_SIZE) return undefined
         return allPages.length * PAGE_SIZE
       },
       initialPageParam: 0,
     })
   }

   export function useItem(id: string) {
     return useQuery({
       queryKey: [ITEMS_KEY, id],
       queryFn: () => getItem(id),
       enabled: !!id,
     })
   }

   export function useDeleteItem() {
     const queryClient = useQueryClient()

     return useMutation({
       mutationFn: deleteItem,
       onSuccess: () => {
         queryClient.invalidateQueries({ queryKey: [ITEMS_KEY] })
         queryClient.invalidateQueries({ queryKey: [STATS_KEY] })
       },
     })
   }

   export function useTags() {
     return useQuery({
       queryKey: [TAGS_KEY],
       queryFn: getTags,
     })
   }

   export function useStats() {
     return useQuery({
       queryKey: [STATS_KEY],
       queryFn: getStats,
     })
   }
   ```

2. Create `web/src/hooks/useSearch.ts`:
   ```typescript
   import { useQuery, useMutation } from '@tanstack/react-query'
   import { search, ask } from '@/api'

   const SEARCH_KEY = 'search'

   export function useSearch(query: string) {
     return useQuery({
       queryKey: [SEARCH_KEY, query],
       queryFn: () => search(query),
       enabled: query.length > 0,
       staleTime: 1000 * 60 * 5, // 5 minutes
     })
   }

   export function useAsk() {
     return useMutation({
       mutationFn: ask,
     })
   }

   // Helper to detect question-like queries
   export function isQuestion(query: string): boolean {
     const q = query.toLowerCase().trim()
     return (
       q.includes('?') ||
       q.startsWith('what ') ||
       q.startsWith('how ') ||
       q.startsWith('why ') ||
       q.startsWith('when ') ||
       q.startsWith('where ') ||
       q.startsWith('who ') ||
       q.startsWith('which ') ||
       q.startsWith('can ') ||
       q.startsWith('do ') ||
       q.startsWith('does ') ||
       q.startsWith('is ') ||
       q.startsWith('are ')
     )
   }
   ```

3. Create `web/src/hooks/useGraph.ts`:
   ```typescript
   import { useQuery } from '@tanstack/react-query'
   import { getGraph } from '@/api'

   const GRAPH_KEY = 'graph'

   export function useGraph() {
     return useQuery({
       queryKey: [GRAPH_KEY],
       queryFn: getGraph,
       staleTime: 1000 * 60 * 5, // 5 minutes
     })
   }
   ```

4. Create `web/src/hooks/index.ts`:
   ```typescript
   export * from './useItems'
   export * from './useSearch'
   export * from './useGraph'
   ```

**Testing:**
- Unit test file: `web/src/hooks/__tests__/useItems.test.ts`
- Test mocking API calls with MSW or vi.mock
- Key test cases:
  - `useItems` returns paginated data
  - `useDeleteItem` invalidates queries on success
  - `isQuestion` correctly identifies question patterns

**Verification:**
```bash
cd web
npm run dev
# Add temporary test in App.tsx:
# const { data, fetchNextPage, hasNextPage } = useItems()
# console.log(data?.pages)
```

**Commit:** `feat(web): add react query hooks for data fetching`

---

### Phase 3: Core Components

---

### Task 7: Create Shared Components

**Goal:** Build reusable UI components used across the app.

**Files to create:**
- `web/src/components/TagPill.tsx`
- `web/src/components/LoadingSkeleton.tsx`
- `web/src/components/EmptyState.tsx`
- `web/src/components/ErrorState.tsx`

**Implementation steps:**

1. Create `web/src/components/TagPill.tsx`:
   ```typescript
   import { cn } from '@/lib/utils'

   interface TagPillProps {
     tag: string
     onClick?: () => void
     className?: string
   }

   // Generate consistent color from tag name
   function tagToColor(tag: string): string {
     let hash = 0
     for (let i = 0; i < tag.length; i++) {
       hash = tag.charCodeAt(i) + ((hash << 5) - hash)
     }
     const hue = Math.abs(hash % 360)
     return `hsl(${hue}, 70%, 85%)`
   }

   function tagToTextColor(tag: string): string {
     let hash = 0
     for (let i = 0; i < tag.length; i++) {
       hash = tag.charCodeAt(i) + ((hash << 5) - hash)
     }
     const hue = Math.abs(hash % 360)
     return `hsl(${hue}, 70%, 25%)`
   }

   export function TagPill({ tag, onClick, className }: TagPillProps) {
     return (
       <span
         role={onClick ? 'button' : undefined}
         tabIndex={onClick ? 0 : undefined}
         onClick={onClick}
         onKeyDown={(e) => {
           if (onClick && (e.key === 'Enter' || e.key === ' ')) {
             onClick()
           }
         }}
         className={cn(
           'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
           onClick && 'cursor-pointer hover:opacity-80 active:opacity-60',
           className
         )}
         style={{
           backgroundColor: tagToColor(tag),
           color: tagToTextColor(tag),
         }}
       >
         {tag}
       </span>
     )
   }
   ```

2. Create `web/src/components/LoadingSkeleton.tsx`:
   ```typescript
   import { Skeleton } from '@/components/ui/skeleton'

   export function ItemCardSkeleton() {
     return (
       <div className="p-4 border-b border-tg-hint/20">
         <Skeleton className="h-5 w-3/4 mb-2" />
         <Skeleton className="h-4 w-full mb-2" />
         <Skeleton className="h-4 w-2/3 mb-3" />
         <div className="flex gap-2">
           <Skeleton className="h-5 w-16 rounded-full" />
           <Skeleton className="h-5 w-20 rounded-full" />
         </div>
       </div>
     )
   }

   export function ItemsFeedSkeleton({ count = 5 }: { count?: number }) {
     return (
       <div>
         {Array.from({ length: count }).map((_, i) => (
           <ItemCardSkeleton key={i} />
         ))}
       </div>
     )
   }

   export function GraphSkeleton() {
     return (
       <div className="flex items-center justify-center h-full">
         <div className="text-center">
           <Skeleton className="h-8 w-8 rounded-full mx-auto mb-2" />
           <Skeleton className="h-4 w-24" />
         </div>
       </div>
     )
   }
   ```

3. Create `web/src/components/EmptyState.tsx`:
   ```typescript
   import { Inbox, Search, Share2 } from 'lucide-react'

   type EmptyStateType = 'items' | 'search' | 'graph'

   interface EmptyStateProps {
     type: EmptyStateType
     query?: string
   }

   const configs: Record<EmptyStateType, { icon: React.ElementType; title: string; description: string }> = {
     items: {
       icon: Inbox,
       title: 'No items yet',
       description: 'Forward links or send text to the bot to start building your knowledge vault.',
     },
     search: {
       icon: Search,
       title: 'No results found',
       description: 'Try a different search term or ask a question.',
     },
     graph: {
       icon: Share2,
       title: 'No connections yet',
       description: 'Save more items to discover relationships between them.',
     },
   }

   export function EmptyState({ type, query }: EmptyStateProps) {
     const config = configs[type]
     const Icon = config.icon

     return (
       <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
         <div className="rounded-full bg-tg-secondary-bg p-4 mb-4">
           <Icon className="h-8 w-8 text-tg-hint" />
         </div>
         <h3 className="text-lg font-medium mb-1">{config.title}</h3>
         <p className="text-sm text-tg-hint max-w-xs">
           {type === 'search' && query
             ? `No results for "${query}"`
             : config.description}
         </p>
       </div>
     )
   }
   ```

4. Create `web/src/components/ErrorState.tsx`:
   ```typescript
   import { AlertCircle, RefreshCw } from 'lucide-react'
   import { Button } from '@/components/ui/button'

   interface ErrorStateProps {
     message?: string
     onRetry?: () => void
   }

   export function ErrorState({ message = 'Something went wrong', onRetry }: ErrorStateProps) {
     return (
       <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
         <div className="rounded-full bg-red-100 p-4 mb-4">
           <AlertCircle className="h-8 w-8 text-red-600" />
         </div>
         <h3 className="text-lg font-medium mb-1">Error</h3>
         <p className="text-sm text-tg-hint max-w-xs mb-4">{message}</p>
         {onRetry && (
           <Button variant="outline" size="sm" onClick={onRetry}>
             <RefreshCw className="h-4 w-4 mr-2" />
             Try again
           </Button>
         )}
       </div>
     )
   }
   ```

**Verification:**
```bash
cd web
npm run dev
# Import and render each component to verify they display correctly
```

**Commit:** `feat(web): add shared components (TagPill, LoadingSkeleton, EmptyState, ErrorState)`

---

### Task 8: Create ItemCard and ItemSheet Components

**Goal:** Build the item display card and detail sheet.

**Files to create:**
- `web/src/components/ItemCard.tsx`
- `web/src/components/ItemSheet.tsx`

**Implementation steps:**

1. Create `web/src/components/ItemCard.tsx`:
   ```typescript
   import { Link2, FileText, Image, Search } from 'lucide-react'
   import { TagPill } from './TagPill'
   import { cn } from '@/lib/utils'
   import type { Item } from '@/api'

   interface ItemCardProps {
     item: Item
     onClick?: () => void
     onTagClick?: (tag: string) => void
   }

   const typeIcons: Record<string, React.ElementType> = {
     link: Link2,
     note: FileText,
     image: Image,
     search: Search,
   }

   function formatDate(dateString: string): string {
     const date = new Date(dateString)
     const now = new Date()
     const diffMs = now.getTime() - date.getTime()
     const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

     if (diffDays === 0) return 'Today'
     if (diffDays === 1) return 'Yesterday'
     if (diffDays < 7) return `${diffDays} days ago`

     return date.toLocaleDateString('en-US', {
       month: 'short',
       day: 'numeric',
       year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined,
     })
   }

   export function ItemCard({ item, onClick, onTagClick }: ItemCardProps) {
     const Icon = typeIcons[item.type] || FileText

     return (
       <div
         role="button"
         tabIndex={0}
         onClick={onClick}
         onKeyDown={(e) => {
           if (onClick && (e.key === 'Enter' || e.key === ' ')) {
             onClick()
           }
         }}
         className={cn(
           'p-4 border-b border-tg-hint/20 active:bg-tg-secondary-bg transition-colors',
           onClick && 'cursor-pointer'
         )}
       >
         <div className="flex items-start gap-3">
           <div className="rounded-lg bg-tg-secondary-bg p-2 shrink-0">
             <Icon className="h-4 w-4 text-tg-hint" />
           </div>
           <div className="flex-1 min-w-0">
             <h3 className="font-medium text-sm leading-tight mb-1 line-clamp-2">
               {item.title || 'Untitled'}
             </h3>
             {item.summary && (
               <p className="text-xs text-tg-hint line-clamp-2 mb-2">
                 {item.summary}
               </p>
             )}
             <div className="flex items-center gap-2 flex-wrap">
               {item.tags.slice(0, 3).map((tag) => (
                 <TagPill
                   key={tag}
                   tag={tag}
                   onClick={
                     onTagClick
                       ? (e) => {
                           e?.stopPropagation()
                           onTagClick(tag)
                         }
                       : undefined
                   }
                 />
               ))}
               {item.tags.length > 3 && (
                 <span className="text-xs text-tg-hint">
                   +{item.tags.length - 3}
                 </span>
               )}
               <span className="text-xs text-tg-hint ml-auto">
                 {formatDate(item.created_at)}
               </span>
             </div>
           </div>
         </div>
       </div>
     )
   }
   ```

2. Create `web/src/components/ItemSheet.tsx`:
   ```typescript
   import {
     Sheet,
     SheetContent,
     SheetHeader,
     SheetTitle,
   } from '@/components/ui/sheet'
   import { Button } from '@/components/ui/button'
   import { ExternalLink, Trash2, Link2, FileText, Image, Search } from 'lucide-react'
   import { TagPill } from './TagPill'
   import { openLink, hapticFeedback } from '@/lib/telegram'
   import { useDeleteItem } from '@/hooks'
   import type { Item } from '@/api'

   interface ItemSheetProps {
     item: Item | null
     open: boolean
     onOpenChange: (open: boolean) => void
     onTagClick?: (tag: string) => void
   }

   const typeIcons: Record<string, React.ElementType> = {
     link: Link2,
     note: FileText,
     image: Image,
     search: Search,
   }

   function formatDateTime(dateString: string): string {
     return new Date(dateString).toLocaleDateString('en-US', {
       weekday: 'short',
       month: 'short',
       day: 'numeric',
       year: 'numeric',
       hour: 'numeric',
       minute: '2-digit',
     })
   }

   export function ItemSheet({ item, open, onOpenChange, onTagClick }: ItemSheetProps) {
     const deleteItem = useDeleteItem()

     if (!item) return null

     const Icon = typeIcons[item.type] || FileText

     const handleDelete = async () => {
       if (!confirm('Delete this item?')) return

       hapticFeedback('medium')
       await deleteItem.mutateAsync(item.id)
       onOpenChange(false)
       hapticFeedback('success')
     }

     const handleOpenLink = () => {
       if (item.url) {
         hapticFeedback('light')
         openLink(item.url)
       }
     }

     return (
       <Sheet open={open} onOpenChange={onOpenChange}>
         <SheetContent side="bottom" className="h-[85vh] rounded-t-xl">
           <SheetHeader className="text-left pb-4 border-b border-tg-hint/20">
             <div className="flex items-start gap-3">
               <div className="rounded-lg bg-tg-secondary-bg p-2 shrink-0">
                 <Icon className="h-5 w-5 text-tg-hint" />
               </div>
               <div className="flex-1 min-w-0">
                 <SheetTitle className="text-base leading-tight pr-8">
                   {item.title || 'Untitled'}
                 </SheetTitle>
                 <p className="text-xs text-tg-hint mt-1">
                   {formatDateTime(item.created_at)}
                 </p>
               </div>
             </div>
           </SheetHeader>

           <div className="overflow-y-auto py-4 space-y-4">
             {/* Tags */}
             {item.tags.length > 0 && (
               <div className="flex flex-wrap gap-2">
                 {item.tags.map((tag) => (
                   <TagPill
                     key={tag}
                     tag={tag}
                     onClick={onTagClick ? () => {
                       onOpenChange(false)
                       onTagClick(tag)
                     } : undefined}
                   />
                 ))}
               </div>
             )}

             {/* Summary */}
             {item.summary && (
               <div>
                 <h4 className="text-xs font-medium text-tg-hint uppercase tracking-wider mb-2">
                   Summary
                 </h4>
                 <p className="text-sm leading-relaxed">{item.summary}</p>
               </div>
             )}

             {/* Content */}
             {item.content && (
               <div>
                 <h4 className="text-xs font-medium text-tg-hint uppercase tracking-wider mb-2">
                   Content
                 </h4>
                 <p className="text-sm leading-relaxed whitespace-pre-wrap">
                   {item.content}
                 </p>
               </div>
             )}

             {/* URL */}
             {item.url && (
               <div>
                 <h4 className="text-xs font-medium text-tg-hint uppercase tracking-wider mb-2">
                   Source
                 </h4>
                 <p className="text-sm text-tg-link break-all">{item.url}</p>
               </div>
             )}
           </div>

           {/* Actions */}
           <div className="absolute bottom-0 left-0 right-0 p-4 bg-tg-bg border-t border-tg-hint/20 flex gap-2">
             {item.url && (
               <Button
                 variant="default"
                 className="flex-1"
                 onClick={handleOpenLink}
               >
                 <ExternalLink className="h-4 w-4 mr-2" />
                 Open Original
               </Button>
             )}
             <Button
               variant="destructive"
               size="icon"
               onClick={handleDelete}
               disabled={deleteItem.isPending}
             >
               <Trash2 className="h-4 w-4" />
             </Button>
           </div>
         </SheetContent>
       </Sheet>
     )
   }
   ```

**Verification:**
```bash
cd web
npm run dev
# Test by rendering ItemCard with mock data
# Test ItemSheet opens/closes correctly
```

**Commit:** `feat(web): add ItemCard and ItemSheet components`

---

### Task 9: Create Bottom Navigation

**Goal:** Build the fixed bottom tab bar for navigation.

**Files to create:**
- `web/src/components/BottomNav.tsx`

**Implementation steps:**

1. Create `web/src/components/BottomNav.tsx`:
   ```typescript
   import { Home, Search, Share2, Settings } from 'lucide-react'
   import { cn } from '@/lib/utils'
   import { hapticFeedback } from '@/lib/telegram'

   export type TabId = 'items' | 'search' | 'graph' | 'settings'

   interface BottomNavProps {
     activeTab: TabId
     onTabChange: (tab: TabId) => void
   }

   const tabs: Array<{ id: TabId; label: string; icon: React.ElementType }> = [
     { id: 'items', label: 'Items', icon: Home },
     { id: 'search', label: 'Search', icon: Search },
     { id: 'graph', label: 'Graph', icon: Share2 },
     { id: 'settings', label: 'Settings', icon: Settings },
   ]

   export function BottomNav({ activeTab, onTabChange }: BottomNavProps) {
     const handleTabClick = (tabId: TabId) => {
       if (tabId !== activeTab) {
         hapticFeedback('light')
         onTabChange(tabId)
       }
     }

     return (
       <nav className="fixed bottom-0 left-0 right-0 bg-tg-bg border-t border-tg-hint/20 safe-area-bottom">
         <div className="flex justify-around items-center h-14">
           {tabs.map(({ id, label, icon: Icon }) => (
             <button
               key={id}
               onClick={() => handleTabClick(id)}
               className={cn(
                 'flex flex-col items-center justify-center flex-1 h-full transition-colors',
                 activeTab === id
                   ? 'text-tg-button'
                   : 'text-tg-hint active:text-tg-text'
               )}
             >
               <Icon className="h-5 w-5 mb-0.5" />
               <span className="text-[10px] font-medium">{label}</span>
             </button>
           ))}
         </div>
       </nav>
     )
   }
   ```

2. Add safe area CSS to `web/src/index.css`:
   ```css
   /* Add at the end */
   .safe-area-bottom {
     padding-bottom: env(safe-area-inset-bottom, 0);
   }
   ```

**Verification:**
```bash
cd web
npm run dev
# Add BottomNav to App.tsx with state
# Verify tabs switch and highlight correctly
```

**Commit:** `feat(web): add bottom navigation component`

---

### Phase 4: Pages

---

### Task 10: Create Items Page

**Goal:** Build the items feed with infinite scroll and pull-to-refresh.

**Files to create:**
- `web/src/pages/ItemsPage.tsx`

**Implementation steps:**

1. Create `web/src/pages/ItemsPage.tsx`:
   ```typescript
   import { useState, useRef, useCallback, useEffect } from 'react'
   import { useItems } from '@/hooks'
   import { ItemCard } from '@/components/ItemCard'
   import { ItemSheet } from '@/components/ItemSheet'
   import { ItemsFeedSkeleton } from '@/components/LoadingSkeleton'
   import { EmptyState } from '@/components/EmptyState'
   import { ErrorState } from '@/components/ErrorState'
   import { hapticFeedback } from '@/lib/telegram'
   import type { Item } from '@/api'

   interface ItemsPageProps {
     filterTag?: string
     onTagClick?: (tag: string) => void
   }

   export function ItemsPage({ filterTag, onTagClick }: ItemsPageProps) {
     const [selectedItem, setSelectedItem] = useState<Item | null>(null)
     const [isRefreshing, setIsRefreshing] = useState(false)
     const containerRef = useRef<HTMLDivElement>(null)
     const touchStartY = useRef(0)

     const {
       data,
       isLoading,
       isError,
       error,
       fetchNextPage,
       hasNextPage,
       isFetchingNextPage,
       refetch,
     } = useItems(filterTag)

     const items = data?.pages.flat() ?? []

     // Infinite scroll
     const handleScroll = useCallback(() => {
       if (!containerRef.current || isFetchingNextPage || !hasNextPage) return

       const { scrollTop, scrollHeight, clientHeight } = containerRef.current
       if (scrollHeight - scrollTop - clientHeight < 200) {
         fetchNextPage()
       }
     }, [fetchNextPage, hasNextPage, isFetchingNextPage])

     // Pull to refresh
     const handleTouchStart = (e: React.TouchEvent) => {
       touchStartY.current = e.touches[0].clientY
     }

     const handleTouchEnd = async (e: React.TouchEvent) => {
       const touchEndY = e.changedTouches[0].clientY
       const pullDistance = touchEndY - touchStartY.current

       if (
         containerRef.current?.scrollTop === 0 &&
         pullDistance > 100 &&
         !isRefreshing
       ) {
         setIsRefreshing(true)
         hapticFeedback('medium')
         await refetch()
         setIsRefreshing(false)
         hapticFeedback('success')
       }
     }

     const handleItemClick = (item: Item) => {
       hapticFeedback('light')
       setSelectedItem(item)
     }

     if (isLoading) {
       return <ItemsFeedSkeleton />
     }

     if (isError) {
       return (
         <ErrorState
           message={error?.message || 'Failed to load items'}
           onRetry={() => refetch()}
         />
       )
     }

     if (items.length === 0) {
       return <EmptyState type="items" />
     }

     return (
       <>
         <div
           ref={containerRef}
           className="h-full overflow-y-auto"
           onScroll={handleScroll}
           onTouchStart={handleTouchStart}
           onTouchEnd={handleTouchEnd}
         >
           {/* Pull to refresh indicator */}
           {isRefreshing && (
             <div className="flex justify-center py-4">
               <div className="animate-spin rounded-full h-6 w-6 border-2 border-tg-button border-t-transparent" />
             </div>
           )}

           {/* Filter indicator */}
           {filterTag && (
             <div className="px-4 py-2 bg-tg-secondary-bg border-b border-tg-hint/20 flex items-center justify-between">
               <span className="text-sm">
                 Filtered by: <strong>{filterTag}</strong>
               </span>
               <button
                 onClick={() => onTagClick?.('')}
                 className="text-sm text-tg-link"
               >
                 Clear
               </button>
             </div>
           )}

           {/* Items list */}
           {items.map((item) => (
             <ItemCard
               key={item.id}
               item={item}
               onClick={() => handleItemClick(item)}
               onTagClick={onTagClick}
             />
           ))}

           {/* Loading more indicator */}
           {isFetchingNextPage && (
             <div className="py-4">
               <ItemsFeedSkeleton count={2} />
             </div>
           )}

           {/* End of list */}
           {!hasNextPage && items.length > 0 && (
             <p className="text-center text-sm text-tg-hint py-8">
               No more items
             </p>
           )}
         </div>

         <ItemSheet
           item={selectedItem}
           open={!!selectedItem}
           onOpenChange={(open) => !open && setSelectedItem(null)}
           onTagClick={onTagClick}
         />
       </>
     )
   }
   ```

**Verification:**
```bash
cd web
npm run dev
# Render ItemsPage in App.tsx
# Test scrolling loads more items
# Test pull-to-refresh triggers reload
# Test clicking item opens sheet
```

**Commit:** `feat(web): add items page with infinite scroll and pull-to-refresh`

---

### Task 11: Create Search Page

**Goal:** Build unified search with Q&A detection.

**Files to create:**
- `web/src/pages/SearchPage.tsx`
- `web/src/components/AIAnswerCard.tsx`

**Implementation steps:**

1. Create `web/src/components/AIAnswerCard.tsx`:
   ```typescript
   import { useState } from 'react'
   import { Sparkles, ChevronDown, ChevronUp } from 'lucide-react'
   import { Card, CardContent } from '@/components/ui/card'
   import { ItemCard } from './ItemCard'
   import type { AskResponse, Item } from '@/api'

   interface AIAnswerCardProps {
     response: AskResponse
     isLoading?: boolean
     onSourceClick?: (item: Item) => void
   }

   export function AIAnswerCard({ response, isLoading, onSourceClick }: AIAnswerCardProps) {
     const [showSources, setShowSources] = useState(false)

     if (isLoading) {
       return (
         <Card className="mx-4 mb-4 bg-gradient-to-br from-tg-button/10 to-tg-button/5">
           <CardContent className="p-4">
             <div className="flex items-center gap-2 mb-3">
               <Sparkles className="h-4 w-4 text-tg-button animate-pulse" />
               <span className="text-sm font-medium text-tg-button">Thinking...</span>
             </div>
             <div className="space-y-2">
               <div className="h-4 bg-tg-hint/20 rounded animate-pulse" />
               <div className="h-4 bg-tg-hint/20 rounded animate-pulse w-3/4" />
             </div>
           </CardContent>
         </Card>
       )
     }

     return (
       <Card className="mx-4 mb-4 bg-gradient-to-br from-tg-button/10 to-tg-button/5">
         <CardContent className="p-4">
           <div className="flex items-center gap-2 mb-3">
             <Sparkles className="h-4 w-4 text-tg-button" />
             <span className="text-sm font-medium text-tg-button">AI Answer</span>
           </div>

           <p className="text-sm leading-relaxed mb-3">{response.answer}</p>

           {response.sources.length > 0 && (
             <div>
               <button
                 onClick={() => setShowSources(!showSources)}
                 className="flex items-center gap-1 text-xs text-tg-hint hover:text-tg-text"
               >
                 {showSources ? (
                   <ChevronUp className="h-3 w-3" />
                 ) : (
                   <ChevronDown className="h-3 w-3" />
                 )}
                 {response.sources.length} source{response.sources.length !== 1 ? 's' : ''}
               </button>

               {showSources && (
                 <div className="mt-2 -mx-4 border-t border-tg-hint/20">
                   {response.sources.map((result) => (
                     <ItemCard
                       key={result.item.id}
                       item={result.item}
                       onClick={() => onSourceClick?.(result.item)}
                     />
                   ))}
                 </div>
               )}
             </div>
           )}
         </CardContent>
       </Card>
     )
   }
   ```

2. Create `web/src/pages/SearchPage.tsx`:
   ```typescript
   import { useState, useEffect, useRef } from 'react'
   import { Search, X } from 'lucide-react'
   import { Input } from '@/components/ui/input'
   import { useSearch, useAsk, isQuestion } from '@/hooks'
   import { ItemCard } from '@/components/ItemCard'
   import { ItemSheet } from '@/components/ItemSheet'
   import { AIAnswerCard } from '@/components/AIAnswerCard'
   import { ItemsFeedSkeleton } from '@/components/LoadingSkeleton'
   import { EmptyState } from '@/components/EmptyState'
   import type { Item, AskResponse } from '@/api'

   interface SearchPageProps {
     onTagClick?: (tag: string) => void
   }

   export function SearchPage({ onTagClick }: SearchPageProps) {
     const [query, setQuery] = useState('')
     const [debouncedQuery, setDebouncedQuery] = useState('')
     const [selectedItem, setSelectedItem] = useState<Item | null>(null)
     const [aiResponse, setAiResponse] = useState<AskResponse | null>(null)
     const inputRef = useRef<HTMLInputElement>(null)

     const { data: searchResults, isLoading: isSearching } = useSearch(debouncedQuery)
     const askMutation = useAsk()

     // Debounce search input
     useEffect(() => {
       const timer = setTimeout(() => {
         setDebouncedQuery(query.trim())
       }, 300)
       return () => clearTimeout(timer)
     }, [query])

     // Trigger AI when query looks like a question
     useEffect(() => {
       if (debouncedQuery && isQuestion(debouncedQuery) && !askMutation.isPending) {
         setAiResponse(null)
         askMutation.mutate(debouncedQuery, {
           onSuccess: setAiResponse,
         })
       } else if (!isQuestion(debouncedQuery)) {
         setAiResponse(null)
       }
     }, [debouncedQuery])

     const handleClear = () => {
       setQuery('')
       setAiResponse(null)
       inputRef.current?.focus()
     }

     const showAI = isQuestion(debouncedQuery)
     const hasResults = searchResults && searchResults.length > 0
     const showEmpty = debouncedQuery && !isSearching && !hasResults && !askMutation.isPending

     return (
       <>
         <div className="flex flex-col h-full">
           {/* Search input */}
           <div className="p-4 border-b border-tg-hint/20">
             <div className="relative">
               <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-tg-hint" />
               <Input
                 ref={inputRef}
                 type="text"
                 placeholder="Search or ask a question..."
                 value={query}
                 onChange={(e) => setQuery(e.target.value)}
                 className="pl-10 pr-10"
               />
               {query && (
                 <button
                   onClick={handleClear}
                   className="absolute right-3 top-1/2 -translate-y-1/2 text-tg-hint hover:text-tg-text"
                 >
                   <X className="h-4 w-4" />
                 </button>
               )}
             </div>
           </div>

           {/* Results area */}
           <div className="flex-1 overflow-y-auto">
             {/* AI Answer */}
             {showAI && (askMutation.isPending || aiResponse) && (
               <div className="pt-4">
                 <AIAnswerCard
                   response={aiResponse || { answer: '', sources: [] }}
                   isLoading={askMutation.isPending}
                   onSourceClick={setSelectedItem}
                 />
               </div>
             )}

             {/* Search results */}
             {isSearching && <ItemsFeedSkeleton count={3} />}

             {hasResults && (
               <div>
                 {!showAI && (
                   <p className="px-4 py-2 text-xs text-tg-hint">
                     {searchResults.length} result{searchResults.length !== 1 ? 's' : ''}
                   </p>
                 )}
                 {searchResults.map((result) => (
                   <ItemCard
                     key={result.item.id}
                     item={result.item}
                     onClick={() => setSelectedItem(result.item)}
                     onTagClick={onTagClick}
                   />
                 ))}
               </div>
             )}

             {/* Empty state */}
             {showEmpty && <EmptyState type="search" query={debouncedQuery} />}

             {/* Initial state */}
             {!debouncedQuery && (
               <div className="flex flex-col items-center justify-center py-12 text-center">
                 <Search className="h-12 w-12 text-tg-hint/50 mb-4" />
                 <p className="text-sm text-tg-hint">
                   Search your vault or ask a question
                 </p>
               </div>
             )}
           </div>
         </div>

         <ItemSheet
           item={selectedItem}
           open={!!selectedItem}
           onOpenChange={(open) => !open && setSelectedItem(null)}
           onTagClick={onTagClick}
         />
       </>
     )
   }
   ```

**Verification:**
```bash
cd web
npm run dev
# Test search returns results
# Test question-like queries trigger AI answer
# Test clicking results opens sheet
```

**Commit:** `feat(web): add search page with unified search and AI Q&A`

---

### Task 12: Create Graph Page

**Goal:** Build interactive knowledge graph with React Flow.

**Files to create:**
- `web/src/pages/GraphPage.tsx`
- `web/src/components/ItemNode.tsx`

**Implementation steps:**

1. Install React Flow:
   ```bash
   cd web
   npm install @xyflow/react
   ```

2. Create `web/src/components/ItemNode.tsx`:
   ```typescript
   import { memo } from 'react'
   import { Handle, Position, type NodeProps } from '@xyflow/react'
   import type { Item } from '@/api'

   interface ItemNodeData {
     item: Item
     connectionCount: number
   }

   function tagToColor(tag: string): string {
     let hash = 0
     for (let i = 0; i < tag.length; i++) {
       hash = tag.charCodeAt(i) + ((hash << 5) - hash)
     }
     const hue = Math.abs(hash % 360)
     return `hsl(${hue}, 60%, 70%)`
   }

   export const ItemNode = memo(function ItemNode({
     data,
   }: NodeProps<{ data: ItemNodeData }>) {
     const { item, connectionCount } = data.data
     const primaryTag = item.tags[0]
     const bgColor = primaryTag ? tagToColor(primaryTag) : 'var(--tg-theme-secondary-bg-color)'

     // Size based on connection count
     const size = Math.min(80, 40 + connectionCount * 5)

     return (
       <>
         <Handle type="target" position={Position.Top} className="opacity-0" />
         <div
           className="rounded-lg p-2 shadow-md border border-white/20 cursor-pointer transition-transform hover:scale-105"
           style={{
             backgroundColor: bgColor,
             width: size,
             minHeight: size,
           }}
         >
           <p
             className="text-[10px] font-medium leading-tight text-center line-clamp-3"
             style={{ color: 'rgba(0,0,0,0.8)' }}
           >
             {item.title || 'Untitled'}
           </p>
         </div>
         <Handle type="source" position={Position.Bottom} className="opacity-0" />
       </>
     )
   })
   ```

3. Create `web/src/pages/GraphPage.tsx`:
   ```typescript
   import { useMemo, useState, useCallback } from 'react'
   import {
     ReactFlow,
     Background,
     Controls,
     MiniMap,
     useNodesState,
     useEdgesState,
     type Node,
     type Edge,
   } from '@xyflow/react'
   import '@xyflow/react/dist/style.css'

   import { useGraph } from '@/hooks'
   import { ItemNode } from '@/components/ItemNode'
   import { ItemSheet } from '@/components/ItemSheet'
   import { GraphSkeleton } from '@/components/LoadingSkeleton'
   import { EmptyState } from '@/components/EmptyState'
   import { ErrorState } from '@/components/ErrorState'
   import type { Item } from '@/api'

   const nodeTypes = {
     item: ItemNode,
   }

   interface GraphPageProps {
     onTagClick?: (tag: string) => void
   }

   export function GraphPage({ onTagClick }: GraphPageProps) {
     const [selectedItem, setSelectedItem] = useState<Item | null>(null)
     const { data, isLoading, isError, error, refetch } = useGraph()

     // Convert graph data to React Flow format
     const { initialNodes, initialEdges } = useMemo(() => {
       if (!data) return { initialNodes: [], initialEdges: [] }

       // Count connections per node
       const connectionCounts = new Map<string, number>()
       data.edges.forEach((edge) => {
         connectionCounts.set(
           edge.source_id,
           (connectionCounts.get(edge.source_id) || 0) + 1
         )
         connectionCounts.set(
           edge.target_id,
           (connectionCounts.get(edge.target_id) || 0) + 1
         )
       })

       // Position nodes in a grid (simple layout)
       const cols = Math.ceil(Math.sqrt(data.nodes.length))
       const spacing = 150

       const nodes: Node[] = data.nodes.map((item, index) => ({
         id: item.id,
         type: 'item',
         position: {
           x: (index % cols) * spacing + Math.random() * 30,
           y: Math.floor(index / cols) * spacing + Math.random() * 30,
         },
         data: {
           item,
           connectionCount: connectionCounts.get(item.id) || 0,
         },
       }))

       const edges: Edge[] = data.edges.map((rel) => ({
         id: `${rel.source_id}-${rel.target_id}`,
         source: rel.source_id,
         target: rel.target_id,
         label: rel.relation_type,
         animated: rel.strength > 0.7,
         style: {
           strokeWidth: Math.max(1, rel.strength * 3),
           opacity: 0.6,
         },
       }))

       return { initialNodes: nodes, initialEdges: edges }
     }, [data])

     const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes)
     const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges)

     // Update when data changes
     useMemo(() => {
       setNodes(initialNodes)
       setEdges(initialEdges)
     }, [initialNodes, initialEdges, setNodes, setEdges])

     const handleNodeClick = useCallback(
       (_: React.MouseEvent, node: Node) => {
         const item = data?.nodes.find((n) => n.id === node.id)
         if (item) setSelectedItem(item)
       },
       [data]
     )

     if (isLoading) {
       return <GraphSkeleton />
     }

     if (isError) {
       return (
         <ErrorState
           message={error?.message || 'Failed to load graph'}
           onRetry={() => refetch()}
         />
       )
     }

     if (!data || data.nodes.length === 0) {
       return <EmptyState type="graph" />
     }

     return (
       <>
         <div className="h-full w-full">
           <ReactFlow
             nodes={nodes}
             edges={edges}
             onNodesChange={onNodesChange}
             onEdgesChange={onEdgesChange}
             onNodeClick={handleNodeClick}
             nodeTypes={nodeTypes}
             fitView
             minZoom={0.1}
             maxZoom={2}
             defaultViewport={{ x: 0, y: 0, zoom: 0.8 }}
           >
             <Background color="var(--tg-theme-hint-color)" gap={20} />
             <Controls
               showInteractive={false}
               className="bg-tg-bg border border-tg-hint/20 rounded-lg"
             />
             <MiniMap
               nodeColor={(n) => {
                 const item = n.data?.item as Item | undefined
                 if (item?.tags[0]) {
                   let hash = 0
                   for (let i = 0; i < item.tags[0].length; i++) {
                     hash = item.tags[0].charCodeAt(i) + ((hash << 5) - hash)
                   }
                   const hue = Math.abs(hash % 360)
                   return `hsl(${hue}, 60%, 70%)`
                 }
                 return 'var(--tg-theme-secondary-bg-color)'
               }}
               className="bg-tg-bg border border-tg-hint/20 rounded-lg"
             />
           </ReactFlow>
         </div>

         <ItemSheet
           item={selectedItem}
           open={!!selectedItem}
           onOpenChange={(open) => !open && setSelectedItem(null)}
           onTagClick={onTagClick}
         />
       </>
     )
   }
   ```

**Verification:**
```bash
cd web
npm run dev
# Render GraphPage
# Verify nodes appear and are draggable
# Verify clicking node opens item sheet
# Test pinch-to-zoom on mobile
```

**Commit:** `feat(web): add graph page with react flow visualization`

---

### Task 13: Create Settings Page

**Goal:** Build settings page with stats and export.

**Files to create:**
- `web/src/pages/SettingsPage.tsx`

**Implementation steps:**

1. Create `web/src/pages/SettingsPage.tsx`:
   ```typescript
   import { Download, FileText, Tag, ExternalLink, Info } from 'lucide-react'
   import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
   import { Button } from '@/components/ui/button'
   import { Skeleton } from '@/components/ui/skeleton'
   import { useStats, useTags } from '@/hooks'
   import { getExportUrl } from '@/api'
   import { hapticFeedback, openLink } from '@/lib/telegram'

   export function SettingsPage() {
     const { data: stats, isLoading: statsLoading } = useStats()
     const { data: tags, isLoading: tagsLoading } = useTags()

     const handleExport = () => {
       hapticFeedback('medium')
       const url = getExportUrl()
       openLink(url)
     }

     return (
       <div className="p-4 space-y-4">
         {/* Stats */}
         <Card>
           <CardHeader className="pb-2">
             <CardTitle className="text-sm font-medium">Your Vault</CardTitle>
           </CardHeader>
           <CardContent>
             <div className="grid grid-cols-2 gap-4">
               <div className="flex items-center gap-3">
                 <div className="rounded-lg bg-tg-secondary-bg p-2">
                   <FileText className="h-5 w-5 text-tg-hint" />
                 </div>
                 <div>
                   {statsLoading ? (
                     <Skeleton className="h-6 w-12" />
                   ) : (
                     <p className="text-2xl font-semibold">{stats?.items ?? 0}</p>
                   )}
                   <p className="text-xs text-tg-hint">Items</p>
                 </div>
               </div>
               <div className="flex items-center gap-3">
                 <div className="rounded-lg bg-tg-secondary-bg p-2">
                   <Tag className="h-5 w-5 text-tg-hint" />
                 </div>
                 <div>
                   {tagsLoading ? (
                     <Skeleton className="h-6 w-12" />
                   ) : (
                     <p className="text-2xl font-semibold">{tags?.length ?? 0}</p>
                   )}
                   <p className="text-xs text-tg-hint">Tags</p>
                 </div>
               </div>
             </div>
           </CardContent>
         </Card>

         {/* Export */}
         <Card>
           <CardHeader className="pb-2">
             <CardTitle className="text-sm font-medium">Export</CardTitle>
           </CardHeader>
           <CardContent>
             <p className="text-sm text-tg-hint mb-3">
               Download your vault as Obsidian-compatible markdown files.
             </p>
             <Button onClick={handleExport} className="w-full">
               <Download className="h-4 w-4 mr-2" />
               Export to Obsidian
             </Button>
           </CardContent>
         </Card>

         {/* About */}
         <Card>
           <CardHeader className="pb-2">
             <CardTitle className="text-sm font-medium">About</CardTitle>
           </CardHeader>
           <CardContent className="space-y-2">
             <div className="flex items-center justify-between">
               <span className="text-sm text-tg-hint">Version</span>
               <span className="text-sm">1.0.0</span>
             </div>
             <div className="flex items-center justify-between">
               <span className="text-sm text-tg-hint">Source</span>
               <button
                 onClick={() => openLink('https://github.com/nerdneilsfield/dumper')}
                 className="flex items-center gap-1 text-sm text-tg-link"
               >
                 GitHub
                 <ExternalLink className="h-3 w-3" />
               </button>
             </div>
           </CardContent>
         </Card>

         {/* Tips */}
         <Card>
           <CardContent className="py-4">
             <div className="flex gap-3">
               <Info className="h-5 w-5 text-tg-hint shrink-0 mt-0.5" />
               <div>
                 <p className="text-sm font-medium mb-1">Pro tip</p>
                 <p className="text-xs text-tg-hint">
                   Forward links or send text messages to the bot to save them to your vault.
                   The AI will automatically generate summaries and tags.
                 </p>
               </div>
             </div>
           </CardContent>
         </Card>
       </div>
     )
   }
   ```

**Verification:**
```bash
cd web
npm run dev
# Render SettingsPage
# Verify stats display
# Test export button triggers download
```

**Commit:** `feat(web): add settings page with stats and export`

---

### Task 14: Wire Up App with Navigation

**Goal:** Connect all pages with tab navigation.

**Files to touch:**
- `web/src/App.tsx` - main app component

**Implementation steps:**

1. Update `web/src/App.tsx`:
   ```typescript
   import { useState } from 'react'
   import { BottomNav, type TabId } from '@/components/BottomNav'
   import { ItemsPage } from '@/pages/ItemsPage'
   import { SearchPage } from '@/pages/SearchPage'
   import { GraphPage } from '@/pages/GraphPage'
   import { SettingsPage } from '@/pages/SettingsPage'

   export default function App() {
     const [activeTab, setActiveTab] = useState<TabId>('items')
     const [filterTag, setFilterTag] = useState<string | undefined>()

     const handleTagClick = (tag: string) => {
       if (tag) {
         setFilterTag(tag)
         setActiveTab('items')
       } else {
         setFilterTag(undefined)
       }
     }

     return (
       <div className="h-screen flex flex-col bg-tg-bg text-tg-text">
         {/* Header */}
         <header className="shrink-0 px-4 py-3 border-b border-tg-hint/20">
           <h1 className="text-lg font-semibold">Dumper</h1>
         </header>

         {/* Content */}
         <main className="flex-1 overflow-hidden pb-14">
           {activeTab === 'items' && (
             <ItemsPage filterTag={filterTag} onTagClick={handleTagClick} />
           )}
           {activeTab === 'search' && (
             <SearchPage onTagClick={handleTagClick} />
           )}
           {activeTab === 'graph' && (
             <GraphPage onTagClick={handleTagClick} />
           )}
           {activeTab === 'settings' && <SettingsPage />}
         </main>

         {/* Navigation */}
         <BottomNav activeTab={activeTab} onTabChange={setActiveTab} />
       </div>
     )
   }
   ```

**Verification:**
```bash
cd web
npm run dev
# Test tab navigation
# Test tag filtering from any page
# Verify all pages render correctly
```

**Commit:** `feat(web): wire up app with tab navigation`

---

### Phase 5: Polish & Integration

---

### Task 15: Add Error Boundaries and Loading States

**Goal:** Add global error handling and polish loading experience.

**Files to create:**
- `web/src/components/ErrorBoundary.tsx`

**Files to touch:**
- `web/src/App.tsx` - wrap with error boundary
- `web/src/main.tsx` - add suspense

**Implementation steps:**

1. Create `web/src/components/ErrorBoundary.tsx`:
   ```typescript
   import { Component, type ReactNode } from 'react'
   import { ErrorState } from './ErrorState'

   interface Props {
     children: ReactNode
   }

   interface State {
     hasError: boolean
     error?: Error
   }

   export class ErrorBoundary extends Component<Props, State> {
     constructor(props: Props) {
       super(props)
       this.state = { hasError: false }
     }

     static getDerivedStateFromError(error: Error): State {
       return { hasError: true, error }
     }

     componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
       console.error('Error caught by boundary:', error, errorInfo)
     }

     render() {
       if (this.state.hasError) {
         return (
           <div className="h-screen flex items-center justify-center bg-tg-bg">
             <ErrorState
               message={this.state.error?.message || 'Something went wrong'}
               onRetry={() => {
                 this.setState({ hasError: false })
                 window.location.reload()
               }}
             />
           </div>
         )
       }

       return this.props.children
     }
   }
   ```

2. Update `web/src/main.tsx`:
   ```typescript
   import React, { Suspense } from 'react'
   import ReactDOM from 'react-dom/client'
   import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
   import { ErrorBoundary } from '@/components/ErrorBoundary'
   import App from './App'
   import './index.css'
   import { initTelegramApp } from './lib/telegram'

   initTelegramApp()

   const queryClient = new QueryClient({
     defaultOptions: {
       queries: {
         staleTime: 1000 * 60,
         retry: 1,
       },
     },
   })

   function LoadingFallback() {
     return (
       <div className="h-screen flex items-center justify-center bg-tg-bg">
         <div className="animate-spin rounded-full h-8 w-8 border-2 border-tg-button border-t-transparent" />
       </div>
     )
   }

   ReactDOM.createRoot(document.getElementById('root')!).render(
     <React.StrictMode>
       <ErrorBoundary>
         <QueryClientProvider client={queryClient}>
           <Suspense fallback={<LoadingFallback />}>
             <App />
           </Suspense>
         </QueryClientProvider>
       </ErrorBoundary>
     </React.StrictMode>,
   )
   ```

**Verification:**
```bash
cd web
npm run dev
# Test error boundary by throwing error in a component
# Verify loading state appears on slow network (throttle in devtools)
```

**Commit:** `feat(web): add error boundary and loading states`

---

### Task 16: Build and Integration Test

**Goal:** Build production bundle and test with Go backend.

**Files to touch:**
- `web/package.json` - verify scripts
- `.gitignore` - ensure dist is committed or handled

**Implementation steps:**

1. Verify `web/package.json` scripts:
   ```json
   {
     "scripts": {
       "dev": "vite",
       "build": "tsc && vite build",
       "preview": "vite preview",
       "lint": "eslint . --ext ts,tsx"
     }
   }
   ```

2. Build production bundle:
   ```bash
   cd web
   npm run build
   ```

3. Verify output structure:
   ```bash
   ls -la web/dist/
   # Should have index.html and assets/
   ```

4. Test with Go backend:
   ```bash
   # Terminal 1: Start backend
   go run ./cmd/dumper

   # Terminal 2: Open browser
   open http://localhost:8080
   # Should serve the Mini App
   ```

5. Test in Telegram (optional):
   - Use ngrok to expose local server: `ngrok http 8080`
   - Update Mini App URL in BotFather to ngrok URL
   - Open Mini App in Telegram

**Verification:**
- [ ] `npm run build` succeeds without errors
- [ ] `web/dist/index.html` exists
- [ ] Go backend serves the app at `/`
- [ ] All API calls work with auth header
- [ ] Pages load and display data

**Commit:** `feat(web): production build configuration`

---

## Testing Strategy

### Unit Tests
- **Location:** `web/src/**/__tests__/*.test.ts(x)`
- **Run:** `npm test` (after adding Vitest)
- **Coverage:** Hooks, utility functions, component logic

### Component Tests
- Use React Testing Library
- Test user interactions (clicks, typing)
- Test loading and error states

### Integration Tests
- Test API client with MSW (Mock Service Worker)
- Test full page flows

### Manual Testing Checklist
- [ ] Items page loads and scrolls
- [ ] Pull-to-refresh works
- [ ] Search returns results
- [ ] Question queries trigger AI
- [ ] Graph displays nodes and edges
- [ ] Node click opens item
- [ ] Settings displays stats
- [ ] Export downloads file
- [ ] Tab navigation works
- [ ] Tag filtering works across pages
- [ ] Works in Telegram (light + dark mode)

---

## Documentation Updates

### README.md
Add section:
```markdown
## Mini App Development

```bash
cd web
npm install
npm run dev    # Start dev server with hot reload
npm run build  # Build for production
```

The built files in `web/dist/` are served by the Go backend at `/`.
```

### CLAUDE.md
Add:
```markdown
## Frontend (web/)

React + TypeScript Mini App with:
- Vite build tool
- Tailwind CSS + shadcn/ui
- TanStack Query for data fetching
- React Flow for graph visualization
- @telegram-apps/sdk-react for Telegram integration

Key commands:
- `cd web && npm run dev` - Start frontend dev server
- `cd web && npm run build` - Build production bundle
```

---

## Definition of Done

- [ ] All 16 tasks implemented
- [ ] `npm run build` succeeds
- [ ] Go backend serves Mini App at `/`
- [ ] Items page with infinite scroll works
- [ ] Search with AI Q&A works
- [ ] Graph visualization works
- [ ] Settings with export works
- [ ] Tab navigation works
- [ ] Tag filtering works
- [ ] Telegram theme integration works
- [ ] No TypeScript errors
- [ ] No console errors in production
- [ ] Manual testing checklist passed

---

*Generated via /brainstorm-plan on 2026-01-16*
*Design doc: [docs/brainstorms/tg-mini-app.md](../brainstorms/tg-mini-app.md)*
