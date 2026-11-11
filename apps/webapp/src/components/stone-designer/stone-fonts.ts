import "@fontsource-variable/cinzel";
import "@fontsource-variable/cormorant-garamond";
import "@fontsource/great-vibes";
import type { StoneTextFont } from "@glassact/data";

interface StoneFontInfo {
  label: string;
  family: string;
  // Cap height as a fraction of the em, so a text element's `size` (cap height
  // in inches) maps to a font-size that actually measures that tall.
  capHeight: number;
}

export const STONE_FONTS: Record<StoneTextFont, StoneFontInfo> = {
  roman: { label: "Roman capitals", family: "'Cinzel Variable', serif", capHeight: 0.7 },
  serif: { label: "Serif", family: "'Cormorant Garamond Variable', serif", capHeight: 0.63 },
  script: { label: "Script", family: "'Great Vibes', cursive", capHeight: 0.62 },
};
