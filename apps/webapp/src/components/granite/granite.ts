import type { JSX } from "solid-js";
import manifest from "../../assets/images/granite/granites.json";
import tintBase from "../../assets/images/granite/tint-base.webp";
import type { GranitePlacement } from "./create-granite-placement";

// The customizer and the catalog preview an inlay against the stone it will be
// set into. The tiles, their physical size and picker order are generated from
// the supplier's samples by tools/granite-textures; for a stone we have no
// sample of, the user picks a color and a neutral tile is tinted to it.

const textureUrls = import.meta.glob<string>(
  "../../assets/images/granite/*.webp",
  { eager: true, query: "?url", import: "default" },
);

function textureUrl(file: string): string {
  const url = textureUrls[`../../assets/images/granite/${file}`];
  if (!url) {
    throw new Error(`granites.json lists ${file}, which was not generated`);
  }
  return url;
}

export interface GraniteSample {
  key: string;
  name: string;
  // Physical size of one tile of the photo, so the stone is drawn to scale.
  tileInches: number;
  // Average color, shown under the tile while it loads.
  meanHex: string;
  image: string;
  swatch: string;
}

export const GRANITE_SAMPLES: GraniteSample[] = manifest.samples.map(
  (sample) => ({
    ...sample,
    image: textureUrl(`${sample.key}.webp`),
    swatch: textureUrl(`${sample.key}-swatch.webp`),
  }),
);

export type GraniteChoice =
  | { kind: "plain" }
  | { kind: "sample"; sample: GraniteSample }
  | { kind: "custom"; hex: string };

export const PLAIN_GRANITE_KEY = "plain";
export const CUSTOM_GRANITE_KEY = "custom";
export const DEFAULT_GRANITE_KEY = "georgia-gray";
export const DEFAULT_CUSTOM_HEX = "#7a6a5e";

const PLAIN_HEX = "#f3f4f6";
// Until the inlay has been measured there is no true scale to draw at; this
// approximates a 4" inlay filling a typical preview.
const UNPLACED_PX_PER_INCH = 60;

export function isGraniteKey(key: string): boolean {
  return (
    key === PLAIN_GRANITE_KEY ||
    key === CUSTOM_GRANITE_KEY ||
    GRANITE_SAMPLES.some((sample) => sample.key === key)
  );
}

export function graniteChoice(key: string, customHex: string): GraniteChoice {
  if (key === PLAIN_GRANITE_KEY) return { kind: "plain" };
  if (key === CUSTOM_GRANITE_KEY) return { kind: "custom", hex: customHex };
  const sample =
    GRANITE_SAMPLES.find((s) => s.key === key) ??
    GRANITE_SAMPLES.find((s) => s.key === DEFAULT_GRANITE_KEY)!;
  return { kind: "sample", sample };
}

export function graniteChoiceName(choice: GraniteChoice): string {
  switch (choice.kind) {
    case "plain":
      return "Plain";
    case "custom":
      return "Custom color";
    case "sample":
      return choice.sample.name;
  }
}

// The backdrop behind an inlay. With a placement the stone is drawn at true
// scale and anchored to the inlay's corner, so it pans and zooms with it.
export function graniteBackdropStyle(
  choice: GraniteChoice,
  placement: GranitePlacement | null,
): JSX.CSSProperties {
  if (choice.kind === "plain") {
    return { "background-color": PLAIN_HEX };
  }

  const tileInches =
    choice.kind === "sample"
      ? choice.sample.tileInches
      : manifest.tint.tileInches;
  const tilePx = tileInches * (placement?.pxPerInch ?? UNPLACED_PX_PER_INCH);

  const image = choice.kind === "sample" ? choice.sample.image : tintBase;
  return {
    ...textureLayers(choice, image),
    "background-size": `${tilePx}px ${tilePx}px`,
    "background-position": placement
      ? `${placement.originX}px ${placement.originY}px`
      : "center",
    "background-repeat": "repeat",
  };
}

export function graniteSwatchStyle(choice: GraniteChoice): JSX.CSSProperties {
  if (choice.kind === "plain") {
    return { "background-color": PLAIN_HEX };
  }
  const image = choice.kind === "sample" ? choice.sample.swatch : tintBase;
  return {
    ...textureLayers(choice, image),
    "background-size": "cover",
  };
}

// The tint tile is centered on 50% gray, and `overlay` leaves 50% gray showing
// exactly the color beneath, so a custom backdrop averages to the picked color.
function textureLayers(
  choice: Exclude<GraniteChoice, { kind: "plain" }>,
  image: string,
): JSX.CSSProperties {
  if (choice.kind === "sample") {
    return {
      "background-color": choice.sample.meanHex,
      "background-image": `url(${image})`,
    };
  }
  return {
    "background-color": choice.hex,
    "background-image": `url(${image})`,
    "background-blend-mode": "overlay",
  };
}
