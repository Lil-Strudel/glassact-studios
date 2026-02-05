import { StandardTable } from "./helpers";

export type InternalUserRole = 
  | "designer" 
  | "production" 
  | "billing" 
  | "admin";

export type InternalUser = StandardTable<{
  name: string;
  email: string;
  avatar: string;
  role: InternalUserRole;
  is_active: boolean;
}>;
