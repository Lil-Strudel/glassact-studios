import type { ContentBBox, Manifest } from "@glassact/data";

// Everything the catalog form knows about an item's artwork: the working
// structure SVG, its manifest (colors), the browser-measured content bounds, and
// the default display size. Edited only through the upload / modify dialogs.
export interface CatalogArtwork {
  structureSvg: string;
  manifest: Manifest;
  warnings: string[];
  // Trimmed content bounds of the structure SVG. Null only when measuring a
  // stored item's SVG failed on load; the form re-measures at save time.
  contentBBox: ContentBBox | null;
  defaultWidth: number;
  defaultHeight: number;
}

// Height per unit width of the trimmed artwork — what the size fields lock to.
// Falls back to the stored dimensions when the content box is unavailable.
export function artworkAspect(artwork: CatalogArtwork): number {
  const bbox = artwork.contentBBox;
  if (bbox && bbox.width > 0) return bbox.height / bbox.width;
  return artwork.defaultHeight / artwork.defaultWidth;
}
