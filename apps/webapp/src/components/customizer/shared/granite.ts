import type { JSX } from "solid-js";
import graniteSlab from "../../../assets/images/granite/granite-slab.jpg";

// The customizer previews an inlay against the "stone" it will be set into.
// Rather than ship one photo per granite color, we tint a single neutral
// grayscale granite tile (see granite-slab.jpg / CREDITS.md) to each of the
// common monument granite colors via CSS `background-blend-mode`. This keeps
// the asset footprint tiny while still reading as real speckled granite in
// context. To use a real per-color photograph instead, give that preset an
// `image` and set `blend: "normal"`.
export interface GranitePreset {
  key: string;
  name: string;
  // Base color the granite tile is tinted toward.
  tint: string;
  // How the tint composites over the tile. "normal" with no image = flat color.
  blend: JSX.CSSProperties["background-blend-mode"];
  // Optional override texture (defaults to the shared grayscale slab).
  image?: string;
}

export const GRANITE_PRESETS: GranitePreset[] = [
  { key: "none", name: "Plain", tint: "#f3f4f6", blend: "normal" },
  { key: "black", name: "Absolute Black", tint: "#282828", blend: "multiply" },
  { key: "gray", name: "Barre Gray", tint: "#8f8f8f", blend: "multiply" },
  { key: "mahogany", name: "Dakota Mahogany", tint: "#7a4a37", blend: "multiply" },
  { key: "pink", name: "Salisbury Pink", tint: "#c39a95", blend: "multiply" },
  { key: "blue", name: "Blue Pearl", tint: "#46536b", blend: "multiply" },
  { key: "green", name: "Forest Green", tint: "#3a4a40", blend: "multiply" },
];

export const DEFAULT_GRANITE_KEY = "gray";

export function graniteByKey(key: string): GranitePreset {
  return (
    GRANITE_PRESETS.find((p) => p.key === key) ??
    GRANITE_PRESETS.find((p) => p.key === DEFAULT_GRANITE_KEY)!
  );
}

// The inline background style for the canvas backdrop element.
export function graniteBackgroundStyle(preset: GranitePreset): JSX.CSSProperties {
  if (preset.key === "none") {
    return { "background-color": preset.tint };
  }
  return {
    "background-color": preset.tint,
    "background-image": `url(${preset.image ?? graniteSlab})`,
    "background-blend-mode": preset.blend,
    "background-size": "260px 260px",
    "background-repeat": "repeat",
  };
}

// A small solid-color swatch representing a preset in the picker.
export function graniteSwatchStyle(preset: GranitePreset): JSX.CSSProperties {
  if (preset.key === "none") {
    return { "background-color": preset.tint };
  }
  return {
    "background-color": preset.tint,
    "background-image": `url(${preset.image ?? graniteSlab})`,
    "background-blend-mode": preset.blend,
    "background-size": "cover",
  };
}
