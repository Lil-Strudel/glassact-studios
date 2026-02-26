import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

import type { GET, POST, DealershipUser, InternalUser } from "@glassact/data";
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

export async function getDealershipUsers(): Promise<GET<DealershipUser>[]> {
  const res = await api.get("/dealership-user");
  return res.data;
}

export function getDealershipUsersOpts() {
  return queryOptions({
    queryKey: ["dealership-user"],
    queryFn: getDealershipUsers,
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
