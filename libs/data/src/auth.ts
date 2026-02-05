import { DealershipUser } from "./dealership-users";
import { InternalUser } from "./internal-users";

export type User = DealershipUser | InternalUser;

export function isDealershipUser(user: User): user is DealershipUser {
  return "dealership_id" in user;
}

export function isInternalUser(user: User): user is InternalUser {
  return !("dealership_id" in user);
}

export const PERMISSION_ACTIONS = {
  CREATE_PROJECT: "create_project",
  APPROVE_PROOF: "approve_proof",
  PLACE_ORDER: "place_order",
  PAY_INVOICE: "pay_invoice",
  MANAGE_DEALERSHIP_USERS: "manage_dealership_users",
  VIEW_PROJECTS: "view_projects",
  VIEW_INVOICES: "view_invoices",
  CREATE_PROOF: "create_proof",
  MANAGE_KANBAN: "manage_kanban",
  CREATE_BLOCKER: "create_blocker",
  CREATE_INVOICE: "create_invoice",
  MANAGE_INTERNAL_USERS: "manage_internal_users",
  VIEW_ALL: "view_all",
} as const;
