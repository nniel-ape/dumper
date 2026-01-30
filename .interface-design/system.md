# Dumper Interface Design System

## Intent

**Who:** Knowledge workers organizing captured content — someone reviewing saved links throughout the day, searching for articles, exploring connections between ideas.

**Task:** Quickly find saved items, review summaries, follow trails through tags and relationships. Capture on mobile, review on desktop.

**Feel:** Warm like a notebook. Personal, comfortable, inviting to browse. Not cold/futuristic — grounded like organizing a personal library.

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
- Primary: `hsl(200 18% 46%)` — muted teal (like aged notebook covers)
- Hover: `hsl(25 60% 50%)` — rust/bookmark orange for interactive states
- Muted: `hsl(200 18% 46% / 0.08)` — subtle teal backgrounds

**Dark mode:**
- Primary: `hsl(200 20% 55%)` — slightly brighter teal
- Hover: `hsl(25 65% 58%)` — warmer rust
- Muted: `hsl(200 20% 55% / 0.12)` — more opacity for visibility

**Why:** Teal feels like faded notebook covers or library cards. Rust/orange for hover creates warmth (like bookmarks). No indigo/violet gradients — those feel futuristic.

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

**Method:** Borders only, no shadows or glassmorphism.

**Pattern:**
```css
.card-base {
  background: hsl(var(--surface));
  border: 1px solid hsl(var(--border-subtle));
  border-radius: 0.5rem; /* 8px */
}

.card-elevated {
  background: hsl(var(--surface-elevated));
  border: 1px solid hsl(var(--border));
  border-radius: 0.5rem;
}
```

**Hover states:** Change border color to accent, shift background slightly.

**Why:** Cards sit on the page, not float above it. Borders feel like pen-drawn boxes in a notebook. Removed all backdrop-blur, shadows, and glass effects.

---

## Typography

**Fonts:**
- Sans: Plus Jakarta Sans — warm, readable, not overused
- Mono: Roboto Mono — for dates and metadata labels

**Scale:**
- Headings: `leading-snug` (1.375) — tighter for density
- Body: `leading-relaxed` (1.625) — comfortable reading
- Metadata: `text-xs font-mono tracking-tight` — compact labels

**Why:** Plus Jakarta Sans has personality without being trendy. Tighter leading for headings helps knowledge workers scan quickly. Monospace dates feel like catalog labels.

---

## Spacing

**Base unit:** 16px (1rem = spacing-4 in Tailwind)

**Common patterns:**
- Card padding: `p-4` (16px)
- Card margins: `mx-4 my-2` (16px horizontal, 8px vertical)
- Between elements: `gap-3` (12px) or `gap-2` (8px)
- Icon containers: `p-2` (8px)

**Why:** 16px base creates comfortable density without cramping. Knowledge workers want information-rich views.

---

## Components

### ItemCard

```tsx
<div className="mx-4 my-2 p-4 card-base hover:border-accent">
  <div className="flex items-start gap-3">
    <div className="rounded bg-accent-muted p-2">
      <Icon className="h-4 w-4 text-accent" />
    </div>
    <div className="flex-1 min-w-0">
      <h3 className="font-semibold text-sm leading-snug">...</h3>
      <p className="text-xs text-muted-foreground leading-relaxed">...</p>
      <span className="text-xs text-text-muted font-mono tracking-tight">
        {date}
      </span>
    </div>
  </div>
</div>
```

**Why:** Icon in muted accent background (notebook tab). Metadata in monospace (catalog label). Border hover instead of shadow lift.

### TagPill

```tsx
<span className="rounded px-2 py-0.5 text-xs bg-accent-muted text-accent border border-border-subtle">
  {tag}
</span>
```

**Why:** Small radius (not fully rounded), subtle border, no backdrop-blur. Feels like handwritten labels.

### Button

**Default:** Solid accent background with hover to rust color
**Outline:** Border-based with hover shifting border to accent

```tsx
variant: {
  default: "bg-primary text-primary-foreground hover:bg-accent-hover",
  outline: "border border-border-subtle hover:border-accent hover:bg-surface-elevated",
  ghost: "hover:bg-accent-muted hover:text-accent"
}
```

**Why:** No gradients, no glows. Simple color transitions. Rust hover adds warmth.

### Input

```tsx
<input className="rounded px-3 py-2 bg-surface-elevated border border-border-subtle focus:ring-1 focus:ring-accent focus:border-accent" />
```

**Why:** Simple focus ring (1px, not 2px). Warm backgrounds. No blur effects.

### BottomNav

```tsx
<nav className="bg-surface-elevated border-t border-border">
  <button className={cn(
    'transition-colors duration-150',
    active ? 'text-accent' : 'text-muted-foreground hover:text-accent-hover'
  )}>
    <Icon />
    <span>{label}</span>
  </button>
</nav>
```

**Why:** Solid background (no blur), simple color transitions. Rust hover for warmth.

---

## Transitions

**Speed:** 150ms for colors, 200ms for movement
**Easing:** Default ease-out
**Properties:** Colors only (no scale, no transform except layout shifts)

```css
transition-colors duration-150
```

**Why:** Fast, subtle. Knowledge tools should feel immediate. No flashy animations.

---

## Removed

**Aurora orbs:** Deleted — felt futuristic/generic
**Glassmorphism:** All `backdrop-blur` removed
**Gradients:** No `bg-gradient-to-r` — flat colors only
**Shadows:** No `shadow-` utilities except in graph controls (functional)
**Scale transforms:** No `active:scale-[0.99]` — distracting

**Why:** These patterns create trendy/cold aesthetics. Notebook feel requires grounded, warm simplicity.

---

## Graph-Specific

Knowledge graph uses ReactFlow. Nodes are ItemCards with same styling. Controls styled to match:

```css
.react-flow-controls {
  background: hsl(var(--surface-elevated));
  border: 1px solid hsl(var(--border));
  border-radius: 0.5rem;
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
