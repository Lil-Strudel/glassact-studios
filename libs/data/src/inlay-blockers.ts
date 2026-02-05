import { StandardTable } from "./helpers";

export type BlockerType = "soft" | "hard";

export type InlayBlocker = StandardTable<{
  inlay_id: number;
  blocker_type: BlockerType;
  reason: string;
  step_blocked: string;
  created_by: number;
  resolved_at: string | null;
  resolved_by: number | null;
  resolution_notes: string | null;
}>;
