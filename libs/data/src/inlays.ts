import type { GET, StandardTable } from "./helpers";
import type { Manifest } from "./customizer";
import type {
  InlayProof,
  PriceAdjustmentType,
  ProofStatus,
} from "./inlay-proofs";
import type { OrderSnapshot } from "./order-snapshots";

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
  preview_url: string;
  sandblast_file_url: string | null;
  approved_proof_id: number | null;
  manufacturing_step: ManufacturingStep | null;
}>;

// Kinds of dependent record that make an inlay undeletable, mirroring the
// ON DELETE RESTRICT foreign keys pointing at inlays. Chat messages cascade and
// never block.
export type InlayDeleteBlocker = "proof" | "milestone" | "update" | "order";

export const INLAY_DELETE_BLOCKERS: InlayDeleteBlocker[] = [
  "proof",
  "milestone",
  "update",
  "order",
];

export type InlayWithInfo = GET<Inlay> & {
  catalog_info?: GET<InlayCatalogInfo> | null;
  custom_info?: GET<InlayCustomInfo> | null;
  has_pending_proof?: boolean;
  latest_proof_status?: ProofStatus | null;
  is_ready: boolean;
  can_delete: boolean;
  delete_blockers: InlayDeleteBlocker[];
  price_group_id: number | null;
  price_group_name: string | null;
  price_cents: number | null;
  price_adjustment_type: PriceAdjustmentType;
  price_adjustment_value: number;
};

// The slice of a catalog item the inlay page needs: enough to identify the
// design to a human, link to it, and resolve color overrides against its
// manifest.
export type InlayCatalogItemRef = {
  uuid: string;
  catalog_code: string;
  name: string;
  category: string;
  svg_url: string;
  manifest?: Manifest;
  default_width: number;
  default_height: number;
};

// Response shape of GET /inlay/:uuid — the list shape plus everything the inlay
// detail page renders without further round-trips.
export type InlayDetail = InlayWithInfo & {
  catalog_item: InlayCatalogItemRef | null;
  approved_proof: GET<InlayProof> | null;
  latest_proof: GET<InlayProof> | null;
  order_snapshot: GET<OrderSnapshot> | null;
};
