import { queryOptions, SolidMutationOptions } from "@tanstack/solid-query";
import api from "./api";

import type { GET, POST, User } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getUserSelf(): Promise<GET<User>> {
  const res = await api.get("/user/self");
  return res.data;
}

export function getUserSelfOpts() {
  return queryOptions({
    queryKey: ["user", "self"],
    queryFn: getUserSelf,
  });
}

export async function getUsers(): Promise<GET<User>[]> {
  const res = await api.get("/user");
  return res.data;
}

export function getUsersOpts() {
  return queryOptions({
    queryKey: ["user"],
    queryFn: getUsers,
  });
}

export async function getUser(uuid: string): Promise<GET<User>> {
  const res = await api.get(`/user/${uuid}`);
  return res.data;
}

export function getUserOpts(uuid: string) {
  return () =>
    queryOptions({
      queryKey: ["user", uuid],
      queryFn: () => getUser(uuid),
    });
}

export async function postUser(body: POST<User>): Promise<GET<User>> {
  const res = await api.post("/user", body);
  return res.data;
}

export function postUserOpts() {
  return mutationOptions({
    mutationFn: postUser,
  });
}
