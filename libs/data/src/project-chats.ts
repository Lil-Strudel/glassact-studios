import { StandardTable } from "./helpers";

export type ProjectChatMessageType = "text" | "image" | "system";

export type ProjectChat = StandardTable<{
  project_id: number;
  dealership_user_id: number | null;
  internal_user_id: number | null;
  message_type: ProjectChatMessageType;
  message: string;
  attachment_url: string | null;
}>;
