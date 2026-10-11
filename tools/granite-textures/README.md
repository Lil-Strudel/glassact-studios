# granite-textures

Turns the supplier's granite sample PDFs into the seamless, true-scale backdrop
tiles the webapp previews inlays against.

```
input/<Name> 2x1.pdf      # supplier samples (not committed)
tint-source/granite-slab.jpg  # CC0 grayscale slab behind the custom-color backdrop
```

Writes to `apps/webapp/src/assets/images/granite/`:

- `<key>.webp` — the full seamless tile at the photo's native resolution
- `<key>-swatch.webp` — 96×96, 3" of stone, for the picker
- `tint-base.webp` — the slab re-centered on 50% gray for `overlay` tinting
- `granites.json` — name, physical tile size and average color per granite, in
  picker order

## What it does

Each "2x1" PDF is a 24"×12" page painted with a tiling pattern whose single
image is the stone photo, so the PDF itself tells us the photo's physical size
(12" for most samples, 8" for some). The JPEG is lifted out byte-for-byte rather
than rasterizing the page. Most photos are printed larger than life, so
`GRAIN_SHRINK_PERCENT` scales each one down to match the real stone.

1. **Decode to linear light** (sRGB, or Gray Gamma 2.2 for the grayscale samples).
2. **Destripe** — remove brightness that is constant along whole rows or columns
   (scanner striping; Galaxy Black is the worst).
3. **Flatten lighting** — divide out vignetting and gradients coarser than a
   quarter tile, which would otherwise repeat as a checkerboard.
4. **Make seamless** — swap a band around each wrap edge for a half-tile-shifted
   copy, bounded by minimum-error cuts that follow the grain (image quilting).
5. **Encode** WebP tiles, swatches, and the manifest.

The table it prints includes a seam ratio — the pixel difference across the wrap
edge relative to the 8px JPEG block edges inside the tile. Around 1 is invisible.

## Run

```bash
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt

python process.py
python process.py --review /tmp/granite-review   # also writes 3×3 tiled contact sheets
```

To add a granite, drop its `<Name> 2x1.pdf` into `input/` and re-run.
