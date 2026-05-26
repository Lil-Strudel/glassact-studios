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
  // Server-managed by the SVG ingest step (not part of create/update requests).
  manifest?: Manifest;
  is_quarantined?: boolean;
  quarantine_reason?: string | null;
  is_active: boolean;
}>;
