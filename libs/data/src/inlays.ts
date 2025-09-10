import { GET, StandardTable } from "./helpers";

export type InlayCatalogInfo = StandardTable<{
  inlay_id: number;
  catalog_item_id: number;
}>;

export type InlayCustomInfo = StandardTable<{
  inlay_id: number;
  description: string;
  width: number;
  height: number;
}>;

export type InlayCustomImage = StandardTable<{
  url: string;
}>;

export type Inlay = StandardTable<
  {
    project_id: number;
    name: string;
    preview_url: string;
    price_group: number;
  } & (
    | {
        type: "catalog";
        catalog_info: InlayCatalogInfo;
      }
    | {
        type: "custom";
        custom_info: InlayCustomInfo;
      }
  )
>;
