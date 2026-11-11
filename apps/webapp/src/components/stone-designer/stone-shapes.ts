import type { StoneShape } from "@glassact/data";

export interface StoneSize {
  width: number;
  height: number;
}

interface StoneShapeInfo {
  label: string;
  defaultSize: StoneSize;
  presets: StoneSize[];
}

const UPRIGHT_PRESETS: StoneSize[] = [
  { width: 24, height: 20 },
  { width: 30, height: 22 },
  { width: 36, height: 24 },
  { width: 42, height: 26 },
  { width: 48, height: 28 },
];

export const STONE_SHAPE_INFO: Record<StoneShape, StoneShapeInfo> = {
  "flat-top": {
    label: "Flat top",
    defaultSize: { width: 36, height: 24 },
    presets: UPRIGHT_PRESETS,
  },
  serpentine: {
    label: "Serpentine",
    defaultSize: { width: 36, height: 24 },
    presets: UPRIGHT_PRESETS,
  },
  "half-serpentine": {
    label: "Half serpentine",
    defaultSize: { width: 36, height: 24 },
    presets: UPRIGHT_PRESETS,
  },
  "oval-top": {
    label: "Oval top",
    defaultSize: { width: 36, height: 24 },
    presets: UPRIGHT_PRESETS,
  },
  cathedral: {
    label: "Cathedral",
    defaultSize: { width: 36, height: 24 },
    presets: UPRIGHT_PRESETS,
  },
  heart: {
    label: "Heart",
    defaultSize: { width: 30, height: 26 },
    presets: [
      { width: 24, height: 22 },
      { width: 30, height: 26 },
      { width: 36, height: 30 },
    ],
  },
  "flat-marker": {
    label: "Flat marker",
    defaultSize: { width: 24, height: 12 },
    presets: [
      { width: 16, height: 8 },
      { width: 20, height: 10 },
      { width: 24, height: 12 },
      { width: 30, height: 14 },
    ],
  },
  slant: {
    label: "Slant",
    defaultSize: { width: 24, height: 16 },
    presets: [
      { width: 20, height: 12 },
      { width: 24, height: 16 },
      { width: 30, height: 18 },
    ],
  },
};

const f = (n: number) => Number(n.toFixed(3));

// The stone's face outline in inches, origin at its top-left, sized w by h.
export function stoneOutlinePath(
  shape: StoneShape,
  w: number,
  h: number,
): string {
  switch (shape) {
    case "flat-top": {
      const r = Math.min(w, h) * 0.01;
      return roundedRect(w, h, r);
    }
    case "flat-marker": {
      const r = Math.min(w, h) * 0.04;
      return roundedRect(w, h, r);
    }
    case "serpentine": {
      const a = h * 0.2;
      return [
        `M0 ${f(h)} L0 ${f(a)}`,
        `C0 ${f(a * 0.45)} ${f(w * 0.03)} ${f(a * 0.25)} ${f(w * 0.07)} ${f(a * 0.25)}`,
        `C${f(w * 0.14)} ${f(a * 0.25)} ${f(w * 0.18)} ${f(a * 0.6)} ${f(w * 0.25)} ${f(a * 0.6)}`,
        `C${f(w * 0.33)} ${f(a * 0.6)} ${f(w * 0.38)} 0 ${f(w * 0.5)} 0`,
        `C${f(w * 0.62)} 0 ${f(w * 0.67)} ${f(a * 0.6)} ${f(w * 0.75)} ${f(a * 0.6)}`,
        `C${f(w * 0.82)} ${f(a * 0.6)} ${f(w * 0.86)} ${f(a * 0.25)} ${f(w * 0.93)} ${f(a * 0.25)}`,
        `C${f(w * 0.97)} ${f(a * 0.25)} ${f(w)} ${f(a * 0.45)} ${f(w)} ${f(a)}`,
        `L${f(w)} ${f(h)} Z`,
      ].join(" ");
    }
    case "half-serpentine": {
      const a = h * 0.14;
      return [
        `M0 ${f(h)} L0 ${f(a)}`,
        `C0 ${f(a * 0.4)} ${f(w * 0.04)} ${f(a * 0.45)} ${f(w * 0.08)} ${f(a * 0.3)}`,
        `C${f(w * 0.11)} ${f(a * 0.15)} ${f(w * 0.12)} 0 ${f(w * 0.17)} 0`,
        `L${f(w * 0.83)} 0`,
        `C${f(w * 0.88)} 0 ${f(w * 0.89)} ${f(a * 0.15)} ${f(w * 0.92)} ${f(a * 0.3)}`,
        `C${f(w * 0.96)} ${f(a * 0.45)} ${f(w)} ${f(a * 0.4)} ${f(w)} ${f(a)}`,
        `L${f(w)} ${f(h)} Z`,
      ].join(" ");
    }
    case "oval-top": {
      const a = h * 0.3;
      return `M0 ${f(h)} L0 ${f(a)} A${f(w / 2)} ${f(a)} 0 0 1 ${f(w)} ${f(a)} L${f(w)} ${f(h)} Z`;
    }
    case "cathedral": {
      const a = h * 0.38;
      return [
        `M0 ${f(h)} L0 ${f(a)}`,
        `C0 ${f(a * 0.45)} ${f(w * 0.28)} ${f(a * 0.12)} ${f(w * 0.5)} 0`,
        `C${f(w * 0.72)} ${f(a * 0.12)} ${f(w)} ${f(a * 0.45)} ${f(w)} ${f(a)}`,
        `L${f(w)} ${f(h)} Z`,
      ].join(" ");
    }
    case "heart":
      return [
        `M${f(w / 2)} ${f(h)}`,
        `C${f(w * 0.2)} ${f(h * 0.77)} 0 ${f(h * 0.56)} 0 ${f(h * 0.3)}`,
        `C0 ${f(h * 0.12)} ${f(w * 0.13)} 0 ${f(w * 0.28)} 0`,
        `C${f(w * 0.39)} 0 ${f(w * 0.47)} ${f(h * 0.07)} ${f(w / 2)} ${f(h * 0.16)}`,
        `C${f(w * 0.53)} ${f(h * 0.07)} ${f(w * 0.61)} 0 ${f(w * 0.72)} 0`,
        `C${f(w * 0.87)} 0 ${f(w)} ${f(h * 0.12)} ${f(w)} ${f(h * 0.3)}`,
        `C${f(w)} ${f(h * 0.56)} ${f(w * 0.8)} ${f(h * 0.77)} ${f(w / 2)} ${f(h)} Z`,
      ].join(" ");
    case "slant":
      // The sloped face narrows toward the top, which is what tells a slant
      // apart from an upright at a glance.
      return `M0 ${f(h)} L${f(w * 0.04)} 0 L${f(w * 0.96)} 0 L${f(w)} ${f(h)} Z`;
  }
}

function roundedRect(w: number, h: number, r: number): string {
  return [
    `M${f(r)} 0 L${f(w - r)} 0 Q${f(w)} 0 ${f(w)} ${f(r)}`,
    `L${f(w)} ${f(h - r)} Q${f(w)} ${f(h)} ${f(w - r)} ${f(h)}`,
    `L${f(r)} ${f(h)} Q0 ${f(h)} 0 ${f(h - r)}`,
    `L0 ${f(r)} Q0 0 ${f(r)} 0 Z`,
  ].join(" ");
}
