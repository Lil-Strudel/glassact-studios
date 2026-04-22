import { StandardTable } from "./helpers";

export type InvoiceStatus = "draft" | "sent" | "paid" | "void";

export type Invoice = StandardTable<{
  project_id: number;
  invoice_url: string | null;
  status: InvoiceStatus;
  paid_at: string | null;
}>;
