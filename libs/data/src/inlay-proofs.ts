import { StandardTable } from "./helpers";

export type ProofStatus = "pending" | "approved" | "declined" | "superseded";

export type ProofApprovalAuthority = "dealership" | "internal";

// How a proof's price is derived from its price group's base price. For
// "percent", price_adjustment_value is percentage points (20 = +20%); for
// "fixed", it is cents (1221 = +$12.21). Both may be negative (discounts).
export type PriceAdjustmentType = "none" | "percent" | "fixed";

export const PRICE_ADJUSTMENT_TYPES: PriceAdjustmentType[] = [
  "none",
  "percent",
  "fixed",
];

export type InlayProof = StandardTable<{
  inlay_id: number;
  version_number: number;
  design_asset_url: string;
  width: number;
  height: number;
  price_group_id: number | null;
  price_adjustment_type: PriceAdjustmentType;
  price_adjustment_value: number;
  scale_factor: number;
  color_overrides: Record<string, unknown>;
  approval_authority: ProofApprovalAuthority;
  status: ProofStatus;
  approved_at: string | null;
  approved_by_dealership_user_id: number | null;
  approved_by_internal_user_id: number | null;
  declined_at: string | null;
  declined_by_dealership_user_id: number | null;
  declined_by_internal_user_id: number | null;
  decline_reason: string | null;
  sent_in_chat_id: number | null;
}>;
