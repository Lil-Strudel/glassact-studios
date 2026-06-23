import { StandardTable } from "./helpers";
import type { PriceAdjustmentType } from "./inlay-proofs";

export type OrderSnapshot = StandardTable<{
  project_id: number;
  inlay_id: number;
  proof_id: number | null;
  price_group_id: number;
  price_cents: number;
  price_adjustment_type: PriceAdjustmentType;
  price_adjustment_value: number;
  width: number;
  height: number;
}>;
