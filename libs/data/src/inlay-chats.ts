import { StandardTable } from "./helpers";

export type InlayChat = StandardTable<{
  inlay_id: number;
  user_id: number;
  sender_type: "glassact" | "customer";
  message: string;
}>;
