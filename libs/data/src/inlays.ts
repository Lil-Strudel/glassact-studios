import type { GET, StandardTable } from "./helpers";
import type { ProofStatus } from "./inlay-proofs";

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
  excluded_from_order: boolean;
}>;

export type InlayWithInfo = GET<Inlay> & {
  catalog_info?: GET<InlayCatalogInfo> | null;
  custom_info?: GET<InlayCustomInfo> | null;
  has_pending_proof?: boolean;
  latest_proof_status?: ProofStatus | null;
  approved_proof_price_group_id?: number | null;
  approved_proof_price_group_name?: string | null;
  approved_proof_price_cents?: number | null;
  has_active_blocker?: boolean;
};
