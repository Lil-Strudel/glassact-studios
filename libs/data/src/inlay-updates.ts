import { StandardTable } from "./helpers";
import { ManufacturingStep } from "./inlays";

export type InlayUpdateType = "info" | "issue";

export const INLAY_UPDATE_TYPES: InlayUpdateType[] = ["info", "issue"];

export type InlayUpdate = StandardTable<{
  inlay_id: number;
  update_type: InlayUpdateType;
  message: string;
  step: ManufacturingStep | null;
  created_by: number | null;
}>;
