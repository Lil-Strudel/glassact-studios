import { StandardTable } from "./helpers";

export type NotificationEventType =
  | "proof_ready"
  | "proof_approved"
  | "proof_declined"
  | "project_submitted"
  | "order_placed"
  | "inlay_step_changed"
  | "inlay_blocked"
  | "inlay_unblocked"
  | "project_shipped"
  | "project_delivered"
  | "invoice_sent"
  | "payment_received"
  | "chat_message";

export const NOTIFICATION_EVENT_TYPES: NotificationEventType[] = [
  "proof_ready",
  "proof_approved",
  "proof_declined",
  "project_submitted",
  "order_placed",
  "inlay_step_changed",
  "inlay_blocked",
  "inlay_unblocked",
  "project_shipped",
  "project_delivered",
  "invoice_sent",
  "payment_received",
  "chat_message",
];

export const DEALERSHIP_NOTIFICATION_EVENT_TYPES: NotificationEventType[] = [
  "proof_ready",
  "proof_approved",
  "proof_declined",
  "inlay_step_changed",
  "inlay_blocked",
  "inlay_unblocked",
  "project_shipped",
  "project_delivered",
  "invoice_sent",
  "payment_received",
  "chat_message",
];

export const INTERNAL_NOTIFICATION_EVENT_TYPES: NotificationEventType[] = [
  "project_submitted",
  "order_placed",
  "proof_ready",
  "proof_approved",
  "proof_declined",
  "project_delivered",
  "chat_message",
];

export const NOTIFICATION_EVENT_LABELS: Record<NotificationEventType, string> =
  {
    proof_ready: "Proof Ready for Review",
    proof_approved: "Proof Approved",
    proof_declined: "Proof Declined",
    project_submitted: "Project Submitted for Design",
    order_placed: "Order Placed",
    inlay_step_changed: "Inlay Step Changed",
    inlay_blocked: "Inlay Blocked",
    inlay_unblocked: "Inlay Unblocked",
    project_shipped: "Project Shipped",
    project_delivered: "Project Delivered",
    invoice_sent: "Invoice Sent",
    payment_received: "Payment Received",
    chat_message: "New Chat Message",
  };

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
