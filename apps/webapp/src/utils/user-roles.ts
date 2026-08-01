import type { DealershipUserRole, InternalUserRole } from "@glassact/data";
import type { RoleOption } from "../components/user-table";

export const DEALERSHIP_ROLE_OPTIONS: (RoleOption & {
  value: DealershipUserRole;
})[] = [
  { label: "Viewer", value: "viewer" },
  { label: "Submitter", value: "submitter" },
  { label: "Approver", value: "approver" },
  { label: "Admin", value: "admin" },
];

export const INTERNAL_ROLE_OPTIONS: (RoleOption & {
  value: InternalUserRole;
})[] = [
  { label: "Designer", value: "designer" },
  { label: "Production", value: "production" },
  { label: "Billing", value: "billing" },
  { label: "Admin", value: "admin" },
];

const AVATAR_BACKGROUNDS = [
  "FFB3BA",
  "FFDFBA",
  "FFFFBA",
  "BAFFC9",
  "BAE1FF",
  "E1BAFF",
  "F0BAFF",
  "BAFFEF",
  "FFD4BA",
];

export function buildAvatarUrl(name: string) {
  const background =
    AVATAR_BACKGROUNDS[Math.floor(Math.random() * AVATAR_BACKGROUNDS.length)];
  return `https://ui-avatars.com/api/?name=${encodeURIComponent(name)}&background=${background}`;
}
