# Granite texture credits

The named granite tiles (`<key>.webp`, `<key>-swatch.webp`) are generated from the
supplier sample PDFs provided to GlassAct, by `tools/granite-textures/process.py`.

`tint-base.webp` is derived from **"Granite001A"** by [ambientCG](https://ambientcg.com/view?id=Granite001A).

- License: **CC0 1.0 (Public Domain)** — free for commercial use, no attribution required.
- Modifications: the 1K color map was cropped/scaled to 512×512 and desaturated to a neutral
  grayscale, then re-centered on 50% gray so it can be tinted to any color at runtime via CSS
  `background-blend-mode: overlay` (see `apps/webapp/src/components/granite/granite.ts`).
