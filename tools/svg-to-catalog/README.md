# svg-to-catalog

Bulk-imports the stained-glass catalog from source SVGs into the GlassAct API.

```
input/<CATEGORY>/<catalog_code>.svg   # source artwork (Illustrator SVG export)
input/<CATEGORY>/<catalog_code>.json  # { name, description, tags, category }
```

Two steps, independent:

1. **`analyze.py`** — generates the `<catalog_code>.json` metadata for each SVG
   using a local Ollama vision model (cairosvg → image → model). Already run for
   the current `input/` set; re-run only for new/changed art.
2. **`seed.py`** — uploads each item through the catalog pipeline and creates the
   database rows.

## seed.py

Replicates, headlessly, the browser-driven admin flow
(`upload → analyze → manifest editor → create`):

1. `POST /api/upload` the source SVG.
2. `POST /api/catalog/analyze` → structure SVG + best-guess manifest (glass groups
   + grout region, each matched to a real `glass_color`/`grout` id where close).
3. Fill any unmatched color ids with a nearest-color match against the live
   palettes (`GET /api/glass-colors`, `GET /api/grouts`) — the create endpoint
   rejects null ids.
4. Measure the content bounding box by rasterizing the structure SVG (cairosvg +
   Pillow) and mapping the opaque-pixel box back to user units — the headless
   stand-in for the browser's `getBBox`.
5. `POST /api/catalog` with the finalized manifest, content bbox, aspect-derived
   physical dimensions (longer side 4"), price group PG-1, and tags. The server
   bakes the fitted, colored SVG and stores it.

### Prerequisites

- API running (default `http://localhost:4100` via `pnpm dev`) with **S3
  configured** in `apps/api/.env` — analyze and bake round-trip through S3.
- Reference data seeded: `pnpm db:seed` (glass colors, grouts, price groups).
- An internal **admin/designer** access token (the catalog routes require the
  `manage_catalog` permission), supplied via `AUTH_TOKEN`.

### Run

```bash
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt

export AUTH_TOKEN=<internal admin/designer access token>
export API_BASE=http://localhost:4100   # optional, this is the default

python seed.py                # seeds ./input
python seed.py ./input/A-ANIMALS   # or a single subfolder (good for a smoke test)
```

Progress is tracked in `seed_progress.json` (keyed by `catalog_code`), so re-runs
skip already-seeded items. Delete it to re-seed from scratch. Sources with
embedded raster or gradients still import best-effort (logged as warnings).
