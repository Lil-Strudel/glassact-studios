// Request/response contracts for the admin catalog manifest-editor flow:
// upload -> analyze (best-guess manifest) -> edit -> create/update (bake + store).

import { Manifest } from "./customizer";

// Known catalog categories, mirrored from tools/svg-to-catalog/input/* directory
// names. Used to seed the freesolo category autocomplete (the field still accepts
// arbitrary custom values).
export const CATALOG_CATEGORIES = [
  "A-ANIMALS",
  "B-OUTDOORS",
  "C-CHILDREN",
  "D-FLOWERS",
  "E-ORNAMENTS",
  "F-FAITH",
  "G-PROFESSIONS",
  "SB-BACKGROUNDS",
] as const;

// Phase A: analyze an uploaded SVG into a working structure SVG + best-guess
// manifest. No catalog row is created.
export interface AnalyzeRequest {
  svg_url: string;
}

export interface AnalyzeResponse {
  // Structure SVG (stable ids p0.., group classes, original viewBox) as text.
  // Held client-side in the editor and re-uploaded at save time.
  structure_svg: string;
  manifest: Manifest;
  // Human-readable notes: groups left unassigned, parse concerns, etc.
  warnings: string[];
}

// Browser-measured content bounding box of the structure SVG, used server-side to
// recompute the viewBox (300 units/inch) and fit+center the artwork at bake.
export interface ContentBBox {
  x: number;
  y: number;
  width: number;
  height: number;
}

// Phase C: create (POST /api/catalog) or full update (PUT /api/catalog/{uuid}).
// The finalized manifest must have every glass/grout id assigned (no nulls); the
// server bakes the structure SVG, swaps svg_url to the baked asset, and stores it.
export interface CatalogWriteRequest {
  catalog_code: string;
  name: string;
  description: string | null;
  category: string;
  default_width: number;
  default_height: number;
  min_width: number;
  min_height: number;
  default_price_group_id: number;
  svg_url: string; // working structure SVG; bake swaps it to the baked asset URL
  manifest: Manifest;
  content_bbox: ContentBBox;
  is_active: boolean;
  tags: string[];
}
