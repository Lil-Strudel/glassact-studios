import { StandardTable } from "./helpers";

export type NotificationEventType =
  | "proof_ready"
  | "proof_approved"
  | "proof_declined"
  | "order_placed"
  | "inlay_step_changed"
  | "inlay_blocked"
  | "inlay_unblocked"
  | "project_shipped"
  | "project_delivered"
  | "invoice_sent"
  | "payment_received"
  | "chat_message";

export type Notification = StandardTable<{
  dealership_user_id: number | null;
  internal_user_id: number | null;
  event_type: NotificationEventType;
  title: string;
  body: string;
  project_id: number | null;
  inlay_id: number | null;
  read_at: string | null;
  email_sent_at: string | null;
}>;

export type NotificationPreference = {
  id: number;
  dealership_user_id?: number;
  internal_user_id?: number;
  event_type: NotificationEventType;
  email_enabled: boolean;
};
