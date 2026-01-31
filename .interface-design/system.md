# Dumper Interface Design System

## Intent

**Who:** Knowledge workers organizing captured content — someone reviewing saved links throughout the day, searching for articles, exploring connections between ideas.

**Task:** Quickly find saved items, review summaries, follow trails through tags and relationships. Capture on mobile, review on desktop.

**Feel:** Warm like a curated notebook, structured like a printed catalog, tactile like a well-made object. Not cold/futuristic — grounded like organizing a personal library.

---

## Foundation

### Palette

**Light mode:**
- Background: `hsl(40 25% 97%)` — cream paper, not white
- Surface: `hsl(35 20% 94%)` — subtle sepia tone
- Surface elevated: `hsl(38 22% 96%)` — slightly lighter for layering
- Borders: `hsl(35 15% 88%)` — subtle separators
- Border subtle: `hsl(35 15% 88%)` — even softer

**Dark mode:**
- Background: `hsl(30 12% 12%)` — warm charcoal
- Surface: `hsl(30 10% 16%)` — warmer than neutral gray
- Surface elevated: `hsl(30 12% 18%)` — layered depth
- Borders: `hsl(30 8% 24%)` — visible but not harsh

**Why:** Warm neutrals feel like aged paper and leather bindings. Avoids clinical white/blue-gray. Creates notebook aesthetic.

### Accent

**Colors:**
- Primary: `hsl(200 22% 42%)` — deeper teal (aged notebook covers)
- Hover: `hsl(25 65% 48%)` — richer rust/bookmark orange for interactive states
- Muted: `hsl(200 22% 42% / 0.08)` — subtle teal backgrounds

**Dark mode:**
- Primary: `hsl(200 20% 55%)` — slightly brighter teal
- Hover: `hsl(25 65% 58%)` — warmer rust
- Muted: `hsl(200 20% 55% / 0.12)` — more opacity for visibility

**Why:** Teal feels like faded notebook covers or library cards. Rust/orange for hover creates warmth. No indigo/violet gradients — those feel futuristic.

### Text

**Light:**
- Primary: `hsl(30 8% 20%)` — warm charcoal, not blue-black
- Secondary: `hsl(30 6% 42%)` — muted for metadata
- Muted: `hsl(30 5% 58%)` — subtle details

**Dark:**
- Primary: `hsl(40 20% 92%)` — warm off-white
- Secondary: `hsl(30 8% 62%)` — readable secondary text
- Muted: `hsl(30 6% 48%)` — subtle without disappearing

**Why:** Warm grays with slight brown undertones. Avoids cool blue-grays common in tech UIs.

---

## Depth

**Method:** Neumorphic-lite — stronger warm shadows with tactile depth. Cards clearly raised, inputs clearly recessed.

**Shadow tokens (v3 — increased intensity):**
```css
/* Light mode */
--shadow-sm: 0 1px 3px hsl(30 15% 20% / 0.08), 0 1px 2px hsl(30 15% 20% / 0.04);
--shadow-md: 0 4px 12px hsl(30 15% 20% / 0.1), 0 2px 4px hsl(30 15% 20% / 0.06);
--shadow-lg: 0 8px 24px hsl(30 15% 20% / 0.12), 0 4px 8px hsl(30 15% 20% / 0.06);

/* Dark mode */
--shadow-sm: 0 1px 3px hsl(0 0% 0% / 0.25), 0 1px 2px hsl(0 0% 0% / 0.15);
--shadow-md: 0 4px 12px hsl(0 0% 0% / 0.3), 0 2px 4px hsl(0 0% 0% / 0.2);
--shadow-lg: 0 8px 24px hsl(0 0% 0% / 0.35), 0 4px 8px hsl(0 0% 0% / 0.2);
```

**Card pattern:**
```css
.card-base {
  background: hsl(var(--surface));
  border: 1px solid hsl(var(--border-subtle));
  border-radius: 1rem; /* 16px */
  box-shadow: var(--shadow-sm);
}

.card-elevated {
  background: hsl(var(--surface-elevated));
  border: 1px solid hsl(var(--border));
  border-radius: 1rem;
  box-shadow: var(--shadow-md);
}
```

**Hover states:** Shift shadow up (`shadow-sm` → `shadow-md`) + border color to accent. Cards feel physically lifted on hover.

**Why:** Stronger warm-tinted shadows (v3) create unmistakable tactile depth. 16px radius (up from 12px) references soft neumorphic shapes. Deeper inset shadows on inputs create clear recessed feel.

---

## Typography

**Fonts:**
- Sans: Plus Jakarta Sans — warm, readable, not overused
- Mono: Roboto Mono — for dates, metadata, and dramatic stat numbers

**Scale:**
- App title: `text-2xl font-bold tracking-tight` — dominant editorial header
- Stat numbers: `text-5xl font-bold font-mono tabular-nums` — dramatic display (refs #4, #8)
- Card titles: `text-base font-bold leading-snug` — scannable, bold
- Section headers: `text-sm font-bold text-foreground` — editorial, not whispered
- Card section titles: `text-lg font-bold` — settings/about sections
- Body/summaries: `text-sm leading-relaxed` (cards) / `text-base leading-relaxed` (detail)
- Metadata: `text-xs font-mono font-medium tracking-tight tabular-nums` — catalog numbers

**Why:** Bold editorial typography inspired by printed catalogs. Dramatic stat numbers (text-5xl) create visual anchor on settings page. Section headers are full-weight foreground-colored. App title bumped to text-2xl for editorial presence.

---

## Spacing

**Base unit:** 16px (1rem = spacing-4 in Tailwind)

**Common patterns:**
- Card padding: `p-6` (24px) — generous whitespace (ref #5)
- Card margins: `mx-4 my-3` (16px horizontal, 12px vertical) — more breathing room
- Between elements: `gap-3` (12px) or `gap-2` (8px)
- Icon containers: `p-2.5` (10px) with `rounded-full`
- Section gaps: `space-y-5` (20px)
- Detail view padding: `p-5` (20px)

**Why:** More generous padding creates breathing room. Cards feel like objects, not cramped rows. 24px card padding matches editorial catalog spacing.

---

## Components

### ItemCard

```tsx
<div className="mx-4 my-3 p-6 card-base transition-all hover:shadow-md hover:border-accent">
  <div className="flex items-start gap-3">
    <div className="rounded-full bg-foreground p-2.5">
      <Icon className="h-4 w-4 text-background" />
    </div>
    <div className="flex-1 min-w-0">
      <h3 className="font-bold text-base leading-snug">...</h3>
      <p className="text-sm text-muted-foreground leading-relaxed">...</p>
      <span className="text-xs text-text-muted font-mono font-medium tracking-tight tabular-nums">
        {date}
      </span>
    </div>
  </div>
</div>
```

**Type icon:** Solid black circle (`bg-foreground`) with white icon (`text-background`). Creates strong visual anchor on each card — inspired by file cabinet index labels (#1), bold black elements (#5, #8).

**Why:** Bold icon badge + soft shadow + shadow lift on hover = tactile card. Date uses tabular-nums for aligned columns.

### TagPill

```tsx
<span className="rounded-full px-2.5 py-0.5 text-xs font-medium bg-accent-muted text-accent">
  {tag}
</span>
```

**Why:** Fully rounded pill shape (capsule). No border — uses background tint only. Inspired by capsule tags from catalog layouts. Cleaner than bordered rectangles.

### Button

**Default:** Solid accent background with shadow depth + hover lift
**Outline:** Border-based with hover shifting border to accent

```tsx
variant: {
  default: "bg-primary text-primary-foreground shadow-sm hover:bg-accent-hover hover:shadow-md",
  outline: "border border-border-subtle hover:border-accent hover:bg-surface-elevated",
  ghost: "hover:bg-accent-muted hover:text-accent"
}
```

**Radius:** `rounded-xl` (16px match)

**Why:** Buttons have subtle shadow for depth. Hover lifts shadow. Rounded-xl matches global 16px radius.

### Input

```tsx
<input className="rounded-xl px-3 py-2 bg-surface-elevated border border-border-subtle shadow-[inset_0_2px_4px_hsl(30_15%_20%/0.06)] focus:ring-1 focus:ring-accent" />
```

**Why:** Deeper inner shadow (2px 4px) creates clearly recessed neumorphic-lite feel. Rounded-xl matches global radius.

### BottomNav

```tsx
<nav className="bg-surface-elevated shadow-md border-t border-border-subtle/50 safe-area-bottom">
  ...
</nav>
```

**Active tab:** `font-semibold` on label for extra weight.

**Why:** Shadow + subtle border-top for double definition. Creates physical tab bar feel. Active labels have bolder weight for clear selection state.

### Catalog Divider

```css
.catalog-divider {
  border-top: 1px solid hsl(var(--border));
  margin-top: 1rem;
  padding-top: 1rem;
}
```

**Used in:** ItemDetail sections (Summary, Content, Source) to create structured catalog-style layout with horizontal rules between data sections. Inspired by Joshua Kaplan's (#6) catalog layout with clear data pair separation.

---

## Transitions

**Speed:** 150ms for colors, 150ms for shadows
**Easing:** Default ease-out
**Properties:** `transition-all duration-150` for components with shadow changes

```css
transition-all duration-150
```

**Why:** Fast, subtle. Shadow transitions need `transition-all` not just `transition-colors`.

---

## Removed

**Aurora orbs:** Deleted — felt futuristic/generic
**Glassmorphism:** All `backdrop-blur` removed. All `glass-card`, `glass-border`, `bg-glass` classes deleted.
**Gradients:** No `bg-gradient-to-r` — editorial flat colors only (exception: AI answer card uses subtle accent gradient)
**Scale transforms:** No `active:scale-[0.99]` — distracting

## Added (v3)

**Stronger shadows:** Increased shadow intensity across all tokens for unmistakable depth
**16px radius:** `--radius: 1rem`, `rounded-2xl` on cards, `rounded-xl` on buttons/inputs
**Black type badges:** Solid `bg-foreground` circles with `text-background` icons for bold visual anchors
**Catalog dividers:** Horizontal rules between detail view sections for structured data layout
**Dramatic stat numbers:** `text-5xl font-bold font-mono tabular-nums` for settings page counters
**Editorial title scale:** `text-2xl` app header, `text-lg` section titles
**Generous spacing:** `p-6` card padding, `my-3` card margins, `p-5` detail view
**Deeper inset inputs:** `inset 0 2px 4px` for clearly recessed form fields

---

## Graph-Specific

Knowledge graph uses ReactFlow. Nodes are ItemCards with same styling. Controls styled to match:

```css
.react-flow-controls {
  background: hsl(var(--surface-elevated));
  border: 1px solid hsl(var(--border));
  border-radius: 1rem;
}

.react-flow__edge:hover {
  stroke: hsl(var(--accent-hover)); /* Rust for emphasis */
}
```

**Why:** Graph is exploratory tool — rust hover on edges signals "follow this trail." Controls match card elevation.

---

## Mobile Considerations

**Safe areas:** Uses Telegram CSS vars for iOS notches
**Touch targets:** Minimum 44px (buttons, nav items)
**iOS Safari quirks:** ReactFlow uses `position: fixed` with calc-based dimensions (not flex)

**Pattern:**
```tsx
className="safe-area-bottom" // Applies padding-bottom with CSS var
```

**Why:** Telegram Mini App runs in iOS Safari WebView. Explicit safe area handling prevents UI behind notches.
