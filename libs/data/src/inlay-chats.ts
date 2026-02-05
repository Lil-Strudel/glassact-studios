import { StandardTable } from "./helpers";

export type ChatMessageType =
  | "text"
  | "image"
  | "proof_sent"
  | "proof_approved"
  | "proof_declined"
  | "system";

export type InlayChat = StandardTable<{
  inlay_id: number;
  dealership_user_id: number | null;
  internal_user_id: number | null;
  message_type: ChatMessageType;
  message: string;
  attachment_url: string | null;
}>;
