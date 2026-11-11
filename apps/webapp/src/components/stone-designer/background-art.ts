import DOMPurify from "dompurify";

export interface BackgroundArt {
  viewBox: string;
  // height / width of the artwork.
  aspect: number;
  markup: string;
}

// Turns a stored background SVG into markup that can be nested inside the
// stone's own <svg>: sanitized, with the cut-list metadata and Illustrator
// class styles dropped since the designer paints every shape as frost.
export function parseBackgroundArt(svgText: string): BackgroundArt | null {
  const clean = DOMPurify.sanitize(svgText, {
    USE_PROFILES: { svg: true, svgFilters: true },
    FORBID_TAGS: ["style", "metadata", "image", "foreignObject"],
  });
  const doc = new DOMParser().parseFromString(clean, "image/svg+xml");
  const root = doc.documentElement;
  if (root.nodeName.toLowerCase() !== "svg") return null;

  const viewBox = root.getAttribute("viewBox");
  const parts = viewBox?.trim().split(/[\s,]+/).map(Number);
  if (!viewBox || !parts || parts.length !== 4 || parts.some(Number.isNaN)) {
    return null;
  }
  const [, , vbWidth, vbHeight] = parts;
  if (vbWidth <= 0 || vbHeight <= 0) return null;

  root.querySelectorAll("[id]").forEach((el) => el.removeAttribute("id"));

  return {
    viewBox,
    aspect: vbHeight / vbWidth,
    markup: root.innerHTML,
  };
}
