import type { StandardTable } from "./helpers";

export type CatalogCompoundMemberItem = {
  id: number;
  uuid: string;
  catalog_code: string;
  name: string;
  category: string;
  svg_url: string;
  default_width: number;
  default_height: number;
  is_active: boolean;
};

// center_x/center_y are inches from the compound's own center.
export type CatalogCompoundMember = {
  id: number;
  catalog_item_id: number;
  center_x: number;
  center_y: number;
  rotation_deg: number;
  sort_order: number;
  catalog_item: CatalogCompoundMemberItem;
};

export type CatalogCompound = StandardTable<{
  name: string;
  description: string | null;
  is_active: boolean;
  display_order: number | null;
  members: CatalogCompoundMember[];
}>;

export type CatalogCompoundMemberWrite = {
  catalog_item_id: number;
  center_x: number;
  center_y: number;
  rotation_deg: number;
};

export type CatalogCompoundWrite = {
  name: string;
  description: string | null;
  is_active: boolean;
  display_order: number | null;
  members: CatalogCompoundMemberWrite[];
};
