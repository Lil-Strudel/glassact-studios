# Granite texture credits

`granite-slab.jpg` is derived from **"Granite001A"** by [ambientCG](https://ambientcg.com/view?id=Granite001A).

- License: **CC0 1.0 (Public Domain)** — free for commercial use, no attribution required.
- Modifications: the 1K color map was cropped/scaled to 512×512, desaturated to a neutral
  grayscale, and lightly contrast-adjusted so it can be tinted to any monument granite color at
  runtime via CSS `background-blend-mode`.

The customizer uses this single seamless tile as the base for every granite background preset
(see `apps/webapp/src/components/customizer/shared/granite.ts`). To offer a real per-color
photograph instead of a tint, drop a new image here and point that preset's config at it.
