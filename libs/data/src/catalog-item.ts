import { StandardTable } from "./helpers";
import { Manifest } from "./customizer";

export type CatalogItemTag = StandardTable<{
  catalog_item_id: number;
  tag: string;
}>;

export type CatalogItem = StandardTable<{
  catalog_code: string;
  name: string;
  description: string | null;
  category: string;
  default_width: number;
  default_height: number;
  min_width: number;
  min_height: number;
  default_price_group_id: number;
  svg_url: string;
  // Server-managed: baked by the catalog write step from the finalized manifest
  // (not part of create/update request bodies — those carry the manifest instead).
  manifest?: Manifest;
  is_active: boolean;
}>;
