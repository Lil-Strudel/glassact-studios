import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

import type {
  GET,
  POST,
  PATCH,
  DealershipUser,
  InternalUser,
} from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getUserSelf(): Promise<
  GET<DealershipUser> | GET<InternalUser>
> {
  const res = await api.get("/user/self");
  return res.data;
}

export function getUserSelfOpts() {
  return queryOptions({
    queryKey: ["user", "self"],
    queryFn: getUserSelf,
  });
}

export async function getDealershipUsers(
  dealershipId?: number,
): Promise<GET<DealershipUser>[]> {
  const res = await api.get("/dealership-user", {
    params: dealershipId ? { dealership_id: dealershipId } : undefined,
  });
  return res.data;
}

export function getDealershipUsersOpts(dealershipId?: number) {
  return queryOptions({
    queryKey:
      dealershipId !== undefined
        ? ["dealership-user", { dealership_id: dealershipId }]
        : ["dealership-user"],
    queryFn: () => getDealershipUsers(dealershipId),
  });
}

export async function getDealershipUser(
  uuid: string,
): Promise<GET<DealershipUser>> {
  const res = await api.get(`/dealership-user/${uuid}`);
  return res.data;
}

export function getDealershipUserOpts(uuid: string) {
  return queryOptions({
    queryKey: ["dealership-user", uuid],
    queryFn: () => getDealershipUser(uuid),
  });
}

export async function postDealershipUser(
  body: POST<DealershipUser>,
): Promise<GET<DealershipUser>> {
  const res = await api.post("/dealership-user", body);
  return res.data;
}

export function postDealershipUserOpts() {
  return mutationOptions({
    mutationFn: postDealershipUser,
  });
}

export async function patchDealershipUser(args: {
  uuid: string;
  body: PATCH<DealershipUser>;
}): Promise<GET<DealershipUser>> {
  const res = await api.patch(`/dealership-user/${args.uuid}`, args.body);
  return res.data;
}

export function patchDealershipUserOpts() {
  return mutationOptions({
    mutationFn: patchDealershipUser,
  });
}

export async function deleteDealershipUser(
  uuid: string,
): Promise<GET<DealershipUser>> {
  const res = await api.delete(`/dealership-user/${uuid}`);
  return res.data;
}

export function deleteDealershipUserOpts() {
  return mutationOptions({
    mutationFn: deleteDealershipUser,
  });
}

export async function getInternalUsers(): Promise<GET<InternalUser>[]> {
  const res = await api.get("/internal-user");
  return res.data;
}

export function getInternalUsersOpts() {
  return queryOptions({
    queryKey: ["internal-user"],
    queryFn: getInternalUsers,
  });
}

export async function getInternalUser(
  uuid: string,
): Promise<GET<InternalUser>> {
  const res = await api.get(`/internal-user/${uuid}`);
  return res.data;
}

export function getInternalUserOpts(uuid: string) {
  return queryOptions({
    queryKey: ["internal-user", uuid],
    queryFn: () => getInternalUser(uuid),
  });
}

export async function postInternalUser(
  body: POST<InternalUser>,
): Promise<GET<InternalUser>> {
  const res = await api.post("/internal-user", body);
  return res.data;
}

export function postInternalUserOpts() {
  return mutationOptions({
    mutationFn: postInternalUser,
  });
}

export async function patchInternalUser(args: {
  uuid: string;
  body: PATCH<InternalUser>;
}): Promise<GET<InternalUser>> {
  const res = await api.patch(`/internal-user/${args.uuid}`, args.body);
  return res.data;
}

export function patchInternalUserOpts() {
  return mutationOptions({
    mutationFn: patchInternalUser,
  });
}

export async function deleteInternalUser(
  uuid: string,
): Promise<GET<InternalUser>> {
  const res = await api.delete(`/internal-user/${uuid}`);
  return res.data;
}

export function deleteInternalUserOpts() {
  return mutationOptions({
    mutationFn: deleteInternalUser,
  });
}
