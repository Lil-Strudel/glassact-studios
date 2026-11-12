import type { ContentBBox } from "./catalog-admin";
import type { StandardTable } from "./helpers";

export type RelatedCatalogItem = {
  id: number;
  uuid: string;
  catalog_code: string;
  name: string;
  svg_url: string;
  default_width: number;
  default_height: number;
  is_active: boolean;
};

// A sandblast mask for the stone face. It has no size of its own; the stone
// designer scales it to the stone.
export type SandblastBackground = StandardTable<{
  code: string;
  name: string;
  description: string | null;
  svg_url: string;
  tags: string[];
  is_active: boolean;
  related_items: RelatedCatalogItem[];
}>;

export type SandblastBackgroundWrite = {
  code: string;
  name: string;
  description: string | null;
  tags: string[];
  is_active: boolean;
  // A freshly uploaded SVG, cleaned into a mask on save, or the stored URL
  // unchanged.
  svg_url: string;
  // Required whenever svg_url is a new upload.
  content_bbox: ContentBBox | null;
  related_catalog_item_ids: number[];
};

export type SandblastBackgroundUpdate = SandblastBackgroundWrite & {
  version: number;
};
