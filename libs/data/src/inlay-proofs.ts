import { StandardTable } from "./helpers";

export type ProofStatus = "pending" | "approved" | "declined" | "superseded";

export type InlayProof = StandardTable<{
  inlay_id: number;
  version_number: number;
  design_asset_url: string;
  width: number;
  height: number;
  price_group_id: number | null;
  price_cents: number | null;
  scale_factor: number;
  color_overrides: Record<string, unknown>;
  status: ProofStatus;
  approved_at: string | null;
  approved_by: number | null;
  declined_at: string | null;
  declined_by: number | null;
  decline_reason: string | null;
  sent_in_chat_id: number;
}>;
