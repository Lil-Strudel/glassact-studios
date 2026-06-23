import { StandardTable } from "./helpers";

export type ProofStatus = "pending" | "approved" | "declined" | "superseded";

export type ProofApprovalAuthority = "dealership" | "internal";

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
