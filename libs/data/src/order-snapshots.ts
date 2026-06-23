import { StandardTable } from "./helpers";

export type OrderSnapshot = StandardTable<{
  project_id: number;
  inlay_id: number;
  proof_id: number | null;
  price_group_id: number;
  price_cents: number;
  width: number;
  height: number;
}>;
