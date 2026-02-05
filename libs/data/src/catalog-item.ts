import { StandardTable } from "./helpers";

export type CatalogItemTag = StandardTable<{
  catalog_item_id: number;
  tag: string;
}>;

export type CatalogItemImage = StandardTable<{
  catalog_item_id: number;
  image_url: string;
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
  is_active: boolean;
}>;
