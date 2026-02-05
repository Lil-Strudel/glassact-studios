import { StandardTable } from "./helpers";

export type InlayType = "catalog" | "custom";

export type ManufacturingStep =
  | "ordered"
  | "materials-prep"
  | "cutting"
  | "fire-polish"
  | "packaging"
  | "shipped"
  | "delivered";

export type InlayCatalogInfo = StandardTable<{
  inlay_id: number;
  catalog_item_id: number;
  customization_notes: string;
}>;

export type InlayCustomInfo = StandardTable<{
  inlay_id: number;
  description: string;
  requested_width: number;
  requested_height: number;
}>;

export type Inlay = StandardTable<{
  project_id: number;
  name: string;
  type: InlayType;
  preview_url: string;
  approved_proof_id: number | null;
  manufacturing_step: ManufacturingStep | null;
}>;
