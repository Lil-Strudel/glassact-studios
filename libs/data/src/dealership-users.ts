import { StandardTable } from "./helpers";

export type DealershipUserRole = 
  | "viewer" 
  | "submitter" 
  | "approver" 
  | "admin";

export type DealershipUser = StandardTable<{
  dealership_id: number;
  name: string;
  email: string;
  avatar: string;
  role: DealershipUserRole;
  is_active: boolean;
}>;
