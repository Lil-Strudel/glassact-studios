import type { GET, StandardTable } from "./helpers";
import type { PriceAdjustmentType, ProofStatus } from "./inlay-proofs";

export type InlayType = "catalog" | "custom";

export type ManufacturingStep =
  | "ordered"
  | "materials-prep"
  | "cutting"
  | "fire-polish"
  | "packaging"
  | "ready-to-ship";

export type InlayCatalogInfo = StandardTable<{
  inlay_id: number;
  catalog_item_id: number;
  customization_notes: string;
}>;

export type InlayCustomReferenceImage = {
  id: number;
  uuid: string;
  inlay_custom_info_id: number;
  image_url: string;
  sort_order: number;
};

export type InlayCustomInfo = StandardTable<{
  inlay_id: number;
  description: string;
  requested_width: number;
  requested_height: number;
  reference_images: InlayCustomReferenceImage[];
}>;

export type Inlay = StandardTable<{
  project_id: number;
  name: string;
  type: InlayType;
  is_customized: boolean;
  installation_kit: boolean;
  preview_url: string;
  sandblast_file_url: string | null;
  approved_proof_id: number | null;
  manufacturing_step: ManufacturingStep | null;
}>;

export type InlayWithInfo = GET<Inlay> & {
  catalog_info?: GET<InlayCatalogInfo> | null;
  custom_info?: GET<InlayCustomInfo> | null;
  has_pending_proof?: boolean;
  latest_proof_status?: ProofStatus | null;
  is_ready: boolean;
  price_group_id: number | null;
  price_group_name: string | null;
  price_cents: number | null;
  price_adjustment_type: PriceAdjustmentType;
  price_adjustment_value: number;
};
