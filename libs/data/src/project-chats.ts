import { StandardTable } from "./helpers";

export type ChatMessageType =
  | "text"
  | "image"
  | "proof_sent"
  | "proof_approved"
  | "proof_declined"
  | "system";

// A project has a single chat thread. `inlay_id` optionally tags a message with
// the inlay it is about so the UI can label it and link straight there — it is
// not a scope: every message on the project belongs to the same conversation.
export type ProjectChat = StandardTable<{
  project_id: number;
  inlay_id: number | null;
  dealership_user_id: number | null;
  internal_user_id: number | null;
  message_type: ChatMessageType;
  message: string;
  attachment_url: string | null;
}>;
