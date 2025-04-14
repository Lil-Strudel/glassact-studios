import { StandardTable } from "./helpers";

export interface InlayCatalogInfo extends StandardTable {
  inlay_id: number;
  catalog_item_id: number;
}

export interface InlayCustomInfo extends StandardTable {
  inlay_id: number;
  description: string;
  width: number;
  height: number;
  images: InlayCustomImage[];
}

export interface InlayCustomImage extends StandardTable {
  url: string;
}

export type Inlay = StandardTable & {
  project_id: number;

  preview_url: string;
  name: string;
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
  );
