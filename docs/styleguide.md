# GlassAct Styleguide — Illuminated Umber

The visual language for every GlassAct surface: the landing site, the webapp, emails and printed matter. It is drawn from the illuminated "S" emblem (`apps/landing/src/assets/images/s-emblem.png`): umber brown, gilt gold, crimson and the sage of its acanthus leaves, with stained-glass details used sparingly.

The company should feel **elegant, classy and regal**: an old craft house that makes something permanent for families on the worst week of their lives. Never playful, never techy, never loud.

The tokens live in `libs/ui/globals.css` (CSS variables) and `libs/ui/tailwind.config.ts` (Tailwind names). Both apps consume them. This document explains how to use them.

## Principles

1. **Browns carry the page; crimson marks the action.** A screen has one crimson thing to do next. Everything else is umber, vellum and rule lines.
2. **Gilt is ornament.** Rules, focus rings, lozenges, staff names and "needs your attention". Never a large fill, never body text.
3. **The glass is the colour.** Inlay photos, proofs and the stone designer are the most colourful things on any screen. The chrome around them stays quiet so they glow.
4. **One system, two volumes.** The landing site and the login page may be ornate (large Marcellus, leaded image grids, glass sheen, and on the landing site the headstone and arched frames). Inside the app, dealers work daily: calmer, denser, same tokens, and no headstone shapes at all.
5. **Dark only, for now.** Every colour is a token so a light theme can be a second set of values later. Never hard-code a colour.

## Colour

All colours are HSL triplets in `globals.css`, exposed to Tailwind as `hsl(var(--x) / <alpha-value>)`, so opacity modifiers work (`bg-primary/20`).

### Grounds and text

| Tailwind | Hex | Use |
| --- | --- | --- |
| `came` | `#0a0705` | Lead-line black: frame borders around imagery, the deepest wells |
| `background` | `#18120d` | Umber page ground |
| `card` / `muted` | `#221a14` | Panels, cards, sidebar, table header |
| `popover` / `secondary` | `#2e241c` | Raised: menus, dialogs, secondary buttons, chat bubbles |
| `accent` | `#3a2e24` | Hover and selected rows (visible even inside a popover) |
| `border` | `#44362a` | Default rules and borders |
| `input` | `#5e4b3a` | Field borders, strong rules, scrollbar thumb |
| `foreground` | `#f4ead7` | Vellum: headings and body text |
| `muted-foreground` | `#c6b298` | Descriptions, secondary text |
| `faint` | `#a28f77` | Meta, timestamps, hints, placeholders |
| `plaque` | `#f3eadb` | Vellum plaque: the light well behind black line artwork (catalog SVGs, sandblast masks, proof line art) and the original logo |

### Brand

| Tailwind | Hex | Use |
| --- | --- | --- |
| `primary` / `crimson` | `#a8242f` | Primary buttons, brand fills, active nav marker. Text on it is `primary-foreground` |
| `crimson-deep` | `#74141f` | Washes, tints, gradient ends, icon tiles |
| `crimson-light` | `#ea6f75` | Crimson **text** on dark: links, active tab underline, inlay tags in chat |
| `gilt` | `#dcad45` | Ornament, focus ring (`ring`), "awaiting you" state |
| `sage` | `#96aa84` | Supporting accent, used rarely (e.g. a secondary illustration tone) |

Crimson `#a8242f` is too dark for text on umber; use `crimson-light` for crimson-coloured words.

### Feedback

Each pair is a tinted background plus a bright foreground: `bg-success text-success-foreground`, bordered with `border-success-foreground/30`.

| Token | Foreground | Use |
| --- | --- | --- |
| `success` | `#9cbb85` | Approved, ready, saved |
| `warning` | `#e3b453` | Caution, blocked-but-fixable ("2 inlays still need approval") |
| `error` | `#ef7d4a` | Errors, failed validation |
| `info` | `#8eaecb` | Neutral information, "GlassAct is reviewing" |
| `destructive` | `#d9653a` fill, dark text | Destructive buttons (prefer the outline style below) |

**Errors are ember, never brand crimson.** Red errors would read as branding and branding would read as a problem. Errors always pair colour with an icon and words.

### Project status

Statuses borrow the jewel tones of glass so they never collide with the brand reds. Render them as the status pill (below) with `status-*` colours.

| Status | Tailwind | Hex |
| --- | --- | --- |
| draft | `status-draft` | `#a89a8d` stone |
| ordered | `status-ordered` | `#d4ae62` gold |
| in-production | `status-production` | `#7fa2d8` cobalt |
| shipped | `status-shipped` | `#6fb5ab` teal |
| invoiced | `status-invoiced` | `#b993c6` amethyst |
| completed | `status-completed` | `#9cbb85` green |
| cancelled | `status-cancelled` | `#8f7c76` faded, with line-through |

### Never

- Raw Tailwind palette classes: `bg-white`, `text-gray-*`, `border-gray-*`, `bg-red-50`, `text-blue-600`, `bg-black`… Map them to tokens (`text-gray-900` → `text-foreground`, `text-gray-500/600` → `text-muted-foreground`, `text-gray-400` → `text-faint`, `bg-white` → `bg-card` or `bg-background`, `bg-gray-50/100` → `bg-muted` or `bg-accent`, `border-gray-200/300` → `border-border` or `border-input`).
- Literal hex in components, except real-world colours (granite samples, glass colours, grout, the stone designer canvas) which depict physical materials and must not be themed.

## Typography

| Role | Face | Tailwind |
| --- | --- | --- |
| Display | **Marcellus** (400 only), a flared Roman face close to the wordmark | `font-display` |
| Body and UI | **Spectral** (400, 500, 600, 700, italic 400), a serif built for screens | `font-body` (the default on `html`/`body`) |

`h1`–`h4` get `font-display font-normal tracking-display` and balanced wrapping from the base layer. Marcellus has one weight: never bold it (`font-semibold` on a heading fakes a bold and looks wrong). Hierarchy comes from size and colour.

### Scale

| Role | Size | Notes |
| --- | --- | --- |
| Hero (landing only) | `text-5xl`–`text-7xl` | Marcellus, `leading-[1.05]` |
| Page title | `text-3xl` (`sm:text-4xl`) | Marcellus |
| Section title | `text-2xl` | Marcellus |
| Card / item title | `text-lg`–`text-xl` | Marcellus |
| Body | `text-base` | Spectral, `leading-relaxed` for prose |
| Small / meta | `text-sm` | Spectral, `text-muted-foreground` or `text-faint` |
| Eyebrow / label | `.eyebrow` | Spectral 600, `text-xs`, uppercase, `tracking-label` (0.16em) |
| Figures | `tabular-nums` | Prices, dimensions, catalog codes, dates in columns |

Totals (an order total, an invoice amount) may be set in Marcellus at `text-2xl`+ with `tabular-nums`.

Keep running text near 65 characters wide (`max-w-prose`).

## Shape, space and depth

- **Radius:** `rounded-md` for controls, `rounded-lg` for cards, panels and images, `rounded-full` for pills and avatars.
- **Images in the webapp** (catalog artwork, inlay previews, proofs, sandblast backgrounds): `came-frame rounded-lg`, with even padding inside a plaque well. Never a headstone or arch shape: the app is a working tool, and a screen of headstone-shaped tiles reads as a cemetery.
- **Headstone frame (landing only):** `HeadstoneFrame` (`apps/landing/src/components/headstone-frame.astro`) masks its content to the GlassAct headstone outline, a 3:2 tablet with a gently swelling serpentine top, set in a came border with a gilt hairline. It is reserved for the landing hero. Don't reuse the outline elsewhere without a deliberate reason.
- **Arched frame (landing only):** `.arched-frame` (`apps/landing/src/styles/landing.css`) gives the top-rounded church-window shape for a few single featured images on the landing site. Grids of many photos never arch, because rows of arches read as headstones: the gallery and the home page strip use square panes in a leaded panel, and the catalog category cards use `came-frame rounded-lg`.
- **Borders over shadows.** Panels are `bg-card border border-border`. Shadows are reserved for floating layers (menus, dialogs, toasts): deep and soft, e.g. `shadow-2xl shadow-black/50`.
- **Spacing:** Tailwind's scale. Lay out siblings with `flex`/`grid` + `gap`, not per-element margins.

## Motifs

Use one or two per screen, never all at once.

- **Lozenge** (`.lozenge`): a small rotated square in `currentColor`, from the knot logo. Used as the dot in status pills, list bullets, the marker on manufacturing steps.
- **Ornament rule** (`.ornament-rule` with a `<span class="lozenge" />` child): a hairline that fades in from both sides around a gilt lozenge. Section dividers on the landing site, the login card, empty states.
- **Came lines:** stained glass is held by dark lead. `.came-frame` around images; on the landing site, grids of imagery can sit on `bg-came` with small gaps so they read like a leaded window.
- **Glass sheen** (`.glass-sheen`): a faint diagonal highlight layered over a coloured fill. For decorative panes, never on text.
- **Crimson wash:** `bg-gradient-to-br from-crimson-deep/40 to-transparent` behind a hero, the order bar or the login card to warm the umber.

## Components

### Page header and empty state

`PageHeader` (eyebrow, Marcellus title, description, actions) opens every page. `EmptyState` (ornament rule, one Marcellus line, one sentence, optional action) fills every empty list. Both come from `@glassact/ui`.

### Buttons

| Variant | Look | Use |
| --- | --- | --- |
| default (primary) | `bg-primary text-primary-foreground`, subtle top highlight, brighten on hover | The one next action: Place order, Approve, Save |
| secondary | `bg-secondary border border-input`, border turns gilt on hover | Supporting actions |
| outline | transparent, `border-input` | Toolbars, filters |
| ghost | `text-muted-foreground`, `hover:bg-accent hover:text-foreground` | Low-emphasis, icon buttons |
| destructive | `text-error-foreground border border-error-foreground/45`, tinted hover | Cancel project, Delete. Outline style so it never competes with crimson |
| destructive-solid | ember fill | Only where a filled destructive button is unavoidable |
| gilt | gilt outline | Rare emphasis that isn't the primary action |
| link | `text-crimson-light`, underline on hover | Inline links |

Disabled: `opacity-45`, no hover effect, and say why nearby ("2 inlays still need approval").

### Fields

`bg-background border-input rounded-md`, focus `border-gilt ring-2 ring-gilt/25`. Labels `text-sm font-medium`, hints `text-xs text-faint`, errors `text-xs text-error-foreground` with an icon. Invalid fields get `border-error-foreground`. Checkboxes fill crimson with a vellum tick.

### Status pill

```html
<span class="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium
             text-status-shipped bg-status-shipped/15 border-status-shipped/35">
  <span class="lozenge"></span>Shipped
</span>
```

The same shape serves readiness (`success` = Ready, `info` = GlassAct is reviewing, `gilt` = Proof awaiting you).

In code: `<Badge variant="shipped" lozenge>Shipped</Badge>`. `Badge` variants: `success | info | error | gilt | crimson` and the status tones `draft | ordered | production | shipped | invoiced | completed | cancelled`. `PROJECT_STATUS_TONES` in `apps/webapp/src/utils/project-status.ts` maps a project status to its tone.

### Cards and panels

`bg-card border border-border rounded-lg`. A card that needs the user's attention gets `border-gilt/45`. No coloured side rails.

### Tables

Header row `bg-muted`, `.eyebrow`-style column labels, rows divided by `border-border`, hover `bg-accent/60`. Numbers right-aligned, `tabular-nums`.

### Navigation

App chrome sits on `bg-card` with a `border-border` rule. Active item: `text-foreground` with a crimson-light underline (top nav) or inset left marker (side nav); inactive: `text-muted-foreground`.

### Dialogs, menus, toasts

`bg-popover border border-border`, deep shadow, a `bg-black/60` backdrop with a slight blur behind dialogs. Dialog titles in Marcellus.

### Chat

Others' messages `bg-secondary`, your own a crimson-deep tint (`bg-crimson-deep/50 border-primary/40`). Staff names in gilt. Inlay tags in `crimson-light` with a lozenge outline. Proof events as a dashed `border-gilt/45` card with a thumbnail.

### Manufacturing steps

Lozenge markers joined by a rule. Done: gilt. Current: crimson with a soft crimson-light halo. To come: outlined in `input`.

### Empty states

An ornament rule, one Marcellus line saying what will appear here, one sentence on how to add the first item, and the primary action.

## Imagery and logo

- **Inlay photos, proofs, granite and the stone designer are never tinted, filtered or recoloured.** They are the product.
- **Headstone shapes are decoration for the landing site only** (see Shape, space and depth). The stone designer draws real stone shapes because it depicts the customer's stone; that is content, not decoration.
- **Knot mark:** `logo-mark.svg` is the official vector knot for dark grounds, with `foreground` vellum outlines and bars in the original GlassAct red (`#84153a`). Use it wherever the knot stands alone. Don't recolour it.
- **Full logo on dark:** `logo-full.svg`, the official vector lockup beside the knot: "GlassAct" in `foreground` vellum (the same as its knot outlines) and "STUDIOS" in the knot's original GlassAct red (`#84153a`), deliberately not the `crimson` token, with a faint vellum glow beneath it so the dark red holds up on umber. The default wherever the name appears: app header, login, landing header and footer. Don't recolour it.
- **Original logo** (`logo.png`, black lettering) only sits on the vellum plaque (`bg-plaque`), never directly on umber.
- **Black line artwork** (catalog design SVGs, sandblast masks, proof line art) sits in a `bg-plaque` well; on umber its lead lines vanish. Photos of finished inlays in granite need no well.
- **The illuminated "S" emblem** is the source of the palette, not a display element. Don't place it on pages.

## Voice

Write plainly, warmly and with dignity. Name things the way a monument dealer would: "proof", "inlay", "stone", "installation kit". Buttons state exactly what happens ("Place order", "Request changes"). Errors say what went wrong and how to fix it. No exclamation marks, no jokes, no emoji. Remember who the end customer is: a grieving family.

## Accessibility

- Body text at least 4.5:1 on its ground (`foreground`, `muted-foreground` and `faint` all pass on `background`, `card` and `popover`).
- Focus is always visible: a 2px gilt outline.
- Never encode meaning in colour alone: pills have words, errors have icons.
- Respect `prefers-reduced-motion`; motion is short (150–200ms) and functional.
