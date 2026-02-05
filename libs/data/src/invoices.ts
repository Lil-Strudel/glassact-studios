import { StandardTable } from "./helpers";

export type InvoiceStatus = "draft" | "sent" | "paid" | "void";

export type InvoiceLineItem = StandardTable<{
  invoice_id: number;
  inlay_id: number | null;
  description: string;
  quantity: number;
  unit_price_cents: number;
  total_cents: number;
  sort_order: number;
}>;

export type Invoice = StandardTable<{
  project_id: number;
  invoice_number: string;
  subtotal_cents: number;
  tax_cents: number;
  total_cents: number;
  status: InvoiceStatus;
  sent_at: string | null;
  sent_to_email: string | null;
  paid_at: string | null;
  notes: string | null;
}>;
