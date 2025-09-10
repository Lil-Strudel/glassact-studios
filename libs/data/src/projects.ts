import { StandardTable } from "./helpers";

export type ProjectStatus =
  | "awaiting-proof"
  | "proof-in-revision"
  | "all-proofs-accepted"
  | "cancelled"
  | "ordered"
  | "in-production"
  | "awaiting-payment"
  | "completed";

export type Project = StandardTable<{
  name: string;
  status: ProjectStatus;
  approved: boolean;
  dealership_id: number;
}>;
